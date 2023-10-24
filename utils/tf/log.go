package tf

import (
	"cloudiac/common"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var planWordZHMap = map[string]string{
	"Terraform will perform the following actions": "Terraform将执行以下操作",
	"will be created":             "将会被创建",
	"must be replaced":            "必须被修改",
	"will be updated in-place":    "将会被修改",
	"unchanged attributes hidden": "未改变属性被隐藏",
}

var envScanZHMap = map[string]string{
	"Violated": "不通过",
	"Passed":   "通过",
}

const (
	simplePlanPreText = `
使用指定的Provider生成以下资源执行计划。资源操作以下符号表示：
～ 资源属性更改
+ 资源新增
- 资源销毁
`
)

// SimpleLog 将log简单化
func SimpleLog(log, logType string) string {
	// 去掉颜色字符，方便处理
	//log = removeColorWord(log)
	if logType == common.TaskStepTfPlan {
		// \[[0-9;]*[a-zA-Z]
		regex := regexp.MustCompile(`Terraform will perform the following actions:(?s:.*?)\[[0-9;]*[a-zA-Z]Plan:\[[0-9;]*[a-zA-Z] \d+ to add, \d+ to change, \d+ to destroy\.`)
		//regex := regexp.MustCompile(`Terraform will perform the following actions:(?s:.*?)\[\d+mPlan: \d+ to add, \d+ to change, \d+ to destroy\.`)

		matches := regex.FindAllString(log, -1)
		if len(matches) > 0 {
			log = matches[0]
		} else {
			log = ""
		}
		log = simplePlanPreText + log
	}
	return log
}

// TranslateLogToZH 翻译tf日志中关键字为中文
func TranslateLogToZH(log, logType string) (string, error) {
	//log = removeColorWord(log)
	if logType == common.TaskStepTfPlan {
		newLog, err := replacePlanText(log)
		if err != nil {
			return log, err
		}
		log = newLog
	} else if logType == common.TaskStepEnvScan {
		log = replaceEnvScan(log)
	}
	return log, nil
}

func removeColorWord(text string) string {
	re := regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
	return re.ReplaceAllString(text, "")
}

func replaceEnvScan(log string) string {
	for key, val := range envScanZHMap {
		log = strings.ReplaceAll(log, key, val)
	}
	return log
}
func replacePlanText(log string) (string, error) {
	// 中英替换
	for key, val := range planWordZHMap {
		log = strings.ReplaceAll(log, key, val)
	}
	// 提取plan结果字符串
	regex := regexp.MustCompile(`\[[0-9;]*[a-zA-Z]Plan:\[[0-9;]*[a-zA-Z] \d+ to add, \d+ to change, \d+ to destroy.`)
	lineMatches := regex.FindAllString(log, -1)
	replaceLineMap := map[string]string{}
	for _, lineMatch := range lineMatches {
		numberText, err := replacePlanTextNumber(lineMatch)
		if err != nil {
			return "", err
		}
		replaceLineMap[lineMatch] = numberText
	}
	for key, value := range replaceLineMap {
		log = strings.ReplaceAll(log, key, value)
	}
	return log, nil
}

func replacePlanTextNumber(line string) (string, error) {
	//regex := regexp.MustCompile(`\d+`)
	regex := regexp.MustCompile(`\[[0-9;]*[a-zA-Z]Plan:\[[0-9;]*[a-zA-Z] (\d+) to add, (\d+) to change, (\d+) to destroy.`)
	matches := regex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return line, nil
	}
	numbers := make([]int, 3)
	for index := 0; index < 3; index++ {
		match := matches[index+1]
		i, err := strconv.Atoi(match)
		if err != nil {
			return "", err
		}
		numbers[index] = i
	}
	return fmt.Sprintf("预览：新增 %d ，更改 %d ，销毁 %d 。", numbers[0], numbers[1], numbers[2]), nil
}
