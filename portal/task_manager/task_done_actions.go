// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package task_manager

import (
	"cloudiac/common"
	"cloudiac/policy"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/forecast/pricecalculator"
	"cloudiac/portal/services/forecast/providers/terraform"
	"cloudiac/portal/services/forecast/schema"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

//taskDoneProcessState 分析环境资源、outputs
func taskDoneProcessState(dbSess *db.Session, task *models.Task) error {
	if bs, err := readIfExist(task.StateJsonPath()); err != nil {
		return fmt.Errorf("read state json: %v", err)
	} else if len(bs) > 0 {
		tfState, err := services.UnmarshalStateJson(bs)

		if err != nil {
			return fmt.Errorf("unmarshal state json: %v", err)
		}
		ps, err := readIfExist(task.ProviderSchemaJsonPath())
		proMap := runner.ProviderSensitiveAttrMap{}
		if err != nil {
			return fmt.Errorf("read provider schema json: %v", err)
		}
		if len(ps) > 0 {
			if err = json.Unmarshal(ps, &proMap); err != nil {
				return err
			}
		}

		// parse tfplan to get sensitive keys
		planBytes, err := readIfExist(task.PlanJsonPath())
		if err != nil {
			return fmt.Errorf("read tfplan json: %v", err)
		}
		sensitiveKeys := services.GetSensitiveKeysFromTfPlan(planBytes)
		fmt.Println("-------------------------")
		fmt.Printf("sensitiveKeys: %v\n", sensitiveKeys)
		fmt.Println("-------------------------")

		if err = services.SaveTaskResources(dbSess, task, tfState.Values, proMap, sensitiveKeys); err != nil {
			return fmt.Errorf("save task resources: %v", err)
		}
		if err = services.SaveTaskOutputs(dbSess, task, tfState.Values.Outputs); err != nil {
			return fmt.Errorf("save task outputs: %v", err)
		}
	}

	return nil
}

func taskDoneProcessPlan(dbSess *db.Session, task *models.Task, isPlanResult bool) error {
	if bs, err := readIfExist(task.PlanJsonPath()); err != nil {
		return fmt.Errorf("read plan json: %v", err)
	} else if len(bs) > 0 {
		tfPlan, err := services.UnmarshalPlanJson(bs)
		if err != nil {
			return fmt.Errorf("unmarshal plan json: %v", err)
		}

		var costs []float32
		var forecastFailed []string
		if isPlanResult {
			costs, forecastFailed, err = getForecastCostWhenTaskPlan(dbSess, task, bs)
			if err != nil {
				logs.Get().Warnf("get prices after plan error: %v", err)
			}
		}

		if err = services.SaveTaskChanges(dbSess, task, tfPlan.ResourceChanges, isPlanResult, costs, forecastFailed); err != nil {
			return fmt.Errorf("save task changes: %v", err)
		}
	}
	return nil
}

func getForecastCostWhenTaskPlan(dbSess *db.Session, task *models.Task, bs []byte) ([]float32, []string, error) {
	var (
		addedCost        float32 // 新增资源的费用
		updateBeforeCost float32 // 变更前的资源费用
		updateAfterCost  float32 // 变更后的资源费用
		destroyedCost    float32 // 删除资源的费用
		err              error
		aliRegion        string //阿里云region
	)

	// 通过环境变量获取区域
	for _, i := range task.Variables {
		if i.Name == "ALICLOUD_REGION" && i.Type == consts.VarTypeEnv {
			if i.Sensitive {
				value, err := utils.DecryptSecretVarForce(i.Value)
				if err != nil {
					logs.Get().Errorf("DecryptSecretVarForce err: %s", err)
					break
				}
				aliRegion = value
			} else {
				aliRegion = i.Value
			}
			break
		}
	}

	// get resources
	createResources, deleteResources, updateBeforeResources, updateAfterResources := terraform.ParserPlanJson(bs, aliRegion)
	// compute cost
	addedCost, addedForecast, err := computeResourceCost(createResources)
	if err != nil {
		return nil, nil, err
	}

	updateBeforeCost, updateBeforeForecast, err := computeResourceCost(updateBeforeResources)
	if err != nil {
		return nil, nil, err
	}

	updateAfterCost, updateAfterForecast, err := computeResourceCost(updateAfterResources)
	if err != nil {
		return nil, nil, err
	}

	destroyedCost, destroyedForecast, err := computeResourceCost(deleteResources)
	if err != nil {
		return nil, nil, err
	}

	// 整理询价失败的resource
	failedForecast := make([]string, 0)
	failedForecast = append(failedForecast, destroyedForecast...)
	failedForecast = append(failedForecast, updateBeforeForecast...)
	failedForecast = append(failedForecast, updateAfterForecast...)
	failedForecast = append(failedForecast, addedForecast...)

	// 获取的费用是以小时计算的，乘以730算月费用
	return []float32{addedCost * 730, -1 * destroyedCost * 730, (updateAfterCost - updateBeforeCost) * 730}, utils.RemoveDuplicateElement(failedForecast), err
}

func computeResourceCost(resources []*schema.Resource) (float32, []string, error) {
	var cost float32
	// 询价失败产品的address
	forecastFailed := make([]string, 0)

	for _, res := range resources {
		priceResp, err := pricecalculator.GetResourcePrice(res)
		if err != nil {
			forecastFailed = append(forecastFailed, res.Name)
			logs.Get().WithField("cost_forecast", "GetResourcePrice").Error(err)
			continue
		}

		price, err := pricecalculator.GetPriceFromResponse(priceResp)
		if err != nil {
			forecastFailed = append(forecastFailed, res.Name)
			logs.Get().WithField("cost_forecast", "GetPriceFromResponse").Error(err)
			continue
		}

		cost = cost + price
	}

	return cost, forecastFailed, nil
}

func taskDoneProcessDriftTask(logger logs.Logger, dbSess *db.Session, task *models.Task) error {
	// 判断是否是偏移检测任务，如果是，解析log文件并写入表
	step, err := services.GetTaskPlanStep(db.Get(), task.Id)
	if err != nil {
		// 解析失败任务不停止不影响主流程
		logger.Errorf("read plan output log: %v", err)
	} else {
		if bs, err := readIfExist(task.TFPlanOutputLogPath(fmt.Sprintf("step%d", step.Index))); err != nil {
			logger.Errorf("read plan output log: %v", err)
		} else {
			// 解析并保存资源漂移信息
			env, err := services.GetEnv(dbSess, task.EnvId)
			if err != nil {
				logger.Errorf("get env '%s': %v", task.EnvId, err)
				return err
			}
			driftInfoMap := ParseResourceDriftInfo(bs)
			if len(driftInfoMap) == 0 {
				err = services.DeleteEnvResourceDrift(dbSess, env.LastResTaskId)
				if err != nil {
					logs.Get().Error("Failed to delete all resoruce drift information in the environment")
				}
			} else {
				addressList := []string{}
				for address := range driftInfoMap {
					addressList = append(addressList, address)
				}
				err = services.DeleteEnvResourceDriftByAddressList(dbSess, env.LastResTaskId, addressList)
				if err != nil {
					logs.Get().Error("Failed to delete already repair resoruce drift information in the environment")
				}
				for address, driftInfo := range driftInfoMap {
					res, err := services.GetResourceIdByAddressAndTaskId(dbSess, address, env.LastResTaskId)
					if err != nil {
						logs.Get().Error("Failed to query resource table while writing drift resource")
						continue
					} else {
						driftInfo.ResId = res.Id
						// TODO 后续使用batch 改进
						services.InsertOrUpdateCronTaskInfo(db.Get(), driftInfo)
					}
				}
			}

			if len(driftInfoMap) > 0 {
				// 发送邮件通知
				services.TaskStatusChangeSendMessage(task, consts.EvenvtCronDrift)
			}
		}
	}
	return nil
}

func taskDoneProcessAutoDestroy(dbSess *db.Session, task *models.Task) error {
	env, err := services.GetEnv(dbSess, task.EnvId)
	if err != nil {
		return errors.Wrapf(err, "get env '%s'", task.EnvId)
	}

	updateAttrs := models.Attrs{}

	if task.Type == models.TaskTypeDestroy && env.Status == models.EnvStatusDestroyed {
		// 环境销毁后清空自动销毁设置，以支持通过再次部署重建环境。
		// ttl 需要保留，做为重建环境的默认 ttl
		updateAttrs["AutoDestroyAt"] = nil
		updateAttrs["AutoDestroyTaskId"] = ""
	}

	// 如果设置了环境的 ttl，则在部署成功后自动根据 ttl 设置销毁时间。
	// 该逻辑只在环境从非活跃状态变为活跃时执行，活跃环境修改 ttl 会立即计算 AutoDestroyAt
	if task.Type == models.TaskTypeApply && env.Status == models.EnvStatusActive &&
		env.AutoDestroyAt == nil && env.TTL != "" && env.TTL != "0" {
		ttl, err := services.ParseTTL(env.TTL)
		if err != nil {
			return err
		}
		at := models.Time(time.Now().Add(ttl))
		updateAttrs["AutoDestroyAt"] = &at
	}

	_, err = services.UpdateEnv(dbSess, env.Id, updateAttrs)
	if err != nil {
		return errors.Wrapf(err, "update environment")
	}

	return nil
}

func StopTaskContainers(sess *db.Session, taskId, envId models.Id) error {
	return stopTaskContainers(sess, taskId, envId, false)
}

func StopScanTaskContainers(sess *db.Session, taskId, envId models.Id) error {
	return stopTaskContainers(sess, taskId, envId, true)
}

func stopTaskContainers(sess *db.Session, taskId, envId models.Id, isScanTask bool) error {
	logs.Get().Infof("stop task container, taskId=%s", taskId)

	var (
		runnerId    string
		containerId string
	)
	if isScanTask {
		task, er := services.GetScanTaskById(sess, taskId)
		if er != nil {
			return er
		}
		runnerId = task.RunnerId
		containerId = task.ContainerId
	} else {
		task, er := services.GetTaskById(sess, taskId)
		if er != nil {
			return er
		}
		runnerId = task.RunnerId
		containerId = task.ContainerId
	}

	runnerAddr, err := services.GetRunnerAddress(runnerId)
	if err != nil {
		return err
	}

	requestUrl := utils.JoinURL(runnerAddr, consts.RunnerStopTaskURL)
	req := runner.TaskStopReq{
		EnvId:        envId.String(),
		TaskId:       taskId.String(),
		ContainerIds: []string{},
	}
	req.ContainerIds = append(req.ContainerIds, containerId)

	header := &http.Header{}
	header.Set("Content-Type", "application/json")
	timeout := int(consts.RunnerConnectTimeout.Seconds())
	_, err = utils.HttpService(requestUrl, "POST", header, req, timeout, timeout)
	return err
}

func sacnTaskDoneProcessTfResult(dbSess *db.Session, task *models.ScanTask) error {
	var (
		tsResult policy.TsResult
		bs       []byte
		err      error
	)

	if task.PolicyStatus == common.PolicyStatusPassed || task.PolicyStatus == common.PolicyStatusViolated {
		if bs, err = readIfExist(task.TfResultJsonPath()); err == nil && len(bs) > 0 {
			if tfResultJson, err := policy.UnmarshalTfResultJson(bs); err == nil {
				tsResult = tfResultJson.Results
			}
		}

		if err := services.UpdateScanResult(dbSess, task, tsResult, task.PolicyStatus); err != nil {
			return fmt.Errorf("save scan result: %v", err)
		}
	} else if task.PolicyStatus == common.PolicyStatusFailed {
		if err := services.CleanScanResult(dbSess, task); err != nil {
			return fmt.Errorf("clean scan result err: %v", err)
		}
	}

	return err
}
