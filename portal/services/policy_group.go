// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func CreatePolicyGroup(tx *db.Session, group *models.PolicyGroup) (*models.PolicyGroup, e.Error) {
	if err := models.Create(tx, group); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyGroupAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return group, nil
}

func GetPolicyGroupById(tx *db.Session, id models.Id) (*models.PolicyGroup, e.Error) {
	group := models.PolicyGroup{}
	if err := tx.Model(models.PolicyGroup{}).Where("id = ?", id).First(&group); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyGroupNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &group, nil
}

func SearchPolicyGroup(dbSess *db.Session, orgId models.Id, q string) *db.Session {
	pgTable := models.PolicyGroup{}.TableName()
	query := dbSess.Table(pgTable).
		Joins(fmt.Sprintf("left join (%s) as p on p.group_id = %s.id",
			fmt.Sprintf("select count(group_id) as policy_count,group_id from %s group by group_id",
				models.Policy{}.TableName()), pgTable))
		//Where(fmt.Sprintf("%s.org_id = ?", pgTable), orgId)
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where(fmt.Sprintf("%s.name like ?", pgTable), qs)
	}
	return query.LazySelectAppend(fmt.Sprintf("%s.*,p.policy_count", pgTable))
}

func UpdatePolicyGroup(query *db.Session, group *models.PolicyGroup, attr models.Attrs) e.Error {
	if _, err := models.UpdateAttr(query, group, attr); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.PolicyGroupAlreadyExist, err)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func DeletePolicyGroup(tx *db.Session, groupId models.Id) e.Error {
	if _, err := tx.Where("id = ?", groupId).
		Delete(&models.PolicyGroup{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func DetailPolicyGroup(dbSess *db.Session, groupId models.Id) (*models.PolicyGroup, e.Error) {
	pg := &models.PolicyGroup{}
	if err := dbSess.
		Where("id = ?", groupId).
		First(pg); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyGroupAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return pg, nil
}

type NewPolicyGroup struct {
	models.PolicyGroup
	OrgId     models.Id `json:"orgId"`
	ProjectId models.Id `json:"projectId" `
	TplId     models.Id `json:"tplId"`
	EnvId     models.Id `json:"envId"`
	Scope     string    `json:"scope"`
}

func GetPolicyGroupByTplIds(tx *db.Session, ids []models.Id) ([]NewPolicyGroup, e.Error) {
	group := make([]NewPolicyGroup, 0)
	if len(ids) == 0 {
		return group, nil
	}
	rel := models.PolicyRel{}.TableName()
	if err := tx.Model(models.PolicyRel{}).
		Joins(fmt.Sprintf("left join %s as pg on pg.id = %s.group_id",
			models.PolicyGroup{}.TableName(), rel)).
		Where(fmt.Sprintf("%s.tpl_id in (?)", rel), ids).
		Where(fmt.Sprintf("%s.scope = ?", rel), models.PolicyRelScopeTpl).
		LazySelectAppend(fmt.Sprintf("%s.org_id,%s.project_id,%s.tpl_id,%s.env_id,%s.scope",
			rel, rel, rel, rel, rel), "pg.*").
		Find(&group); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return group, nil
}

func GetPolicyGroupByEnvIds(tx *db.Session, ids []models.Id) ([]NewPolicyGroup, e.Error) {
	group := make([]NewPolicyGroup, 0)
	if len(ids) == 0 {
		return group, nil
	}
	rel := models.PolicyRel{}.TableName()
	if err := tx.Model(models.PolicyRel{}).
		Joins(fmt.Sprintf("left join %s as pg on pg.id = %s.group_id",
			models.PolicyGroup{}.TableName(), rel)).
		Where(fmt.Sprintf("%s.env_id in (?)", rel), ids).
		Where(fmt.Sprintf("%s.scope = ?", rel), models.PolicyRelScopeEnv).
		LazySelectAppend("pg.*").
		LazySelectAppend(fmt.Sprintf("%s.scope, %s.org_id, %s.project_id, %s.tpl_id, %s.env_id",
			rel, rel, rel, rel, rel)).
		Find(&group); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return group, nil
}

func GetPolicyGroupByTplId(tx *db.Session, id models.Id) ([]models.PolicyGroup, e.Error) {
	groups := make([]models.PolicyGroup, 0)
	if err := tx.Model(models.PolicyGroup{}).
		Joins("join iac_policy_rel on iac_policy_group.id = iac_policy_rel.group_id").
		Where("iac_policy_rel.tpl_id = ?", id).
		Find(&groups); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyGroupNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return groups, nil
}

type PolicyGroupImportError struct {
	File  string `json:"file"`
	Error string `json:"error"`
}

type PolicyGroupImportSummary struct {
	PolicyCount int                      `json:"policyCount"`
	ErrorCount  int                      `json:"errorCount"`
	Errors      []PolicyGroupImportError `json:"errors"`
}

type DownloadPolicyGroupResult struct {
	Error   e.Error                   `json:"-"`
	Group   *models.PolicyGroup       `json:"group"`
	Summary *PolicyGroupImportSummary `json:"summary"`
}

func DownloadPolicyGroup(tx *db.Session, tmpDir string, result *DownloadPolicyGroupResult, wg *sync.WaitGroup) {
	logger := logs.Get().WithField("func", "DownloadPolicyGroup")
	group := result.Group

	defer wg.Done()

	// 1. git download
	logger.Debugf("downloading git")
	branch := group.GitTags
	if group.Branch != "" {
		branch = group.Branch
	}
	repoAddr, commitId, err := GetPolicyGroupCommitId(tx, group.VcsId, group.RepoId, branch)
	if err != nil {
		result.Error = e.New(e.InternalError, errors.Wrapf(err, "get commit id"), http.StatusInternalServerError)
		return
	}
	logger.Debugf("downloading git %s@%s to %s", repoAddr, commitId, filepath.Join(tmpDir, "code"))
	er := GitCheckout(filepath.Join(tmpDir, "code"), repoAddr, commitId)
	if err != nil {
		result.Error = e.New(e.InternalError, errors.Wrapf(er, "checkout repo"), http.StatusInternalServerError)
		return
	}
	logger.Debugf("download git complete")
}

func GetPolicyGroupCommitId(tx *db.Session, vcsId models.Id, repoId string, branch string) (repoAddr, commitId string, err e.Error) {
	vcs, err := QueryVcsByVcsId(vcsId, tx)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return "", "", e.New(e.VcsNotExists, err)
		}
		return "", "", e.New(e.DBError, err)
	}

	repo, er := vcsrv.GetRepo(vcs, repoId)
	if er != nil {
		return "", "", e.New(e.VcsError, er)
	}

	commitId, er = repo.BranchCommitId(branch)
	if er != nil {
		return "", "", e.New(e.VcsError, er)
	}

	repoAddr, er = vcsrv.GetRepoAddress(repo)
	if er != nil {
		return "", "", e.New(e.VcsError, er)
	}

	token, er := vcs.DecryptToken()
	if er != nil {
		return "", "", e.New(e.VcsError, er)
	}

	if repoAddr == "" {
		return "", "", e.New(e.BadParam, fmt.Errorf("repo address is blank"))
	}

	u, er := url.Parse(repoAddr)
	if er != nil {
		return "", "", e.New(e.InternalError, errors.Wrapf(er, "parse url: %v", repoAddr))
	} else if token != "" {
		u.User = url.UserPassword("token", token)
	}
	repoAddr = u.String()

	return repoAddr, commitId, nil
}

// GitCheckout 从 repoUrl 的 git 仓库 checkout 到本地目录 localDir，可以设置对应的 commitId
func GitCheckout(localDir string, repoUrl string, commitId string) error {
	opt := git.CloneOptions{
		URL:      repoUrl,
		Progress: logs.Writer(),
	}
	repo, err := git.PlainClone(localDir, false, &opt)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitId),
	})
	return err
}

type RegoFile struct {
	MetaFile string
	RegoFile string
}

func ParsePolicyGroup(dirname string) ([]*Policy, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	otherFiles := append(files)
	var regoFiles []RegoFile
	// 遍历当前目录
	for _, f := range files {
		// 优先处理 json 的 meta 及对应的 rego 文件
		if filepath.Ext(f.Name()) == ".json" {
			regoFileName := utils.FileNameWithoutExt(f.Name()) + ".rego"
			regoFilePath := filepath.Join(dirname, regoFileName)
			if utils.FileExist(regoFilePath) {
				regoFiles = append(regoFiles, RegoFile{
					MetaFile: filepath.Join(dirname, f.Name()),
					RegoFile: filepath.Join(dirname, regoFileName),
				})
				// 将已经处理的 rego 排除
				for i, rf := range otherFiles {
					if rf.Name() == regoFileName {
						otherFiles[i] = otherFiles[len(otherFiles)-1]
						otherFiles = otherFiles[:len(otherFiles)-1]
						break
					}
				}
			}
		}
	}

	// 遍历其他没有 json meta 的 rego 文件
	for _, f := range otherFiles {
		if filepath.Ext(f.Name()) == ".rego" {
			regoFiles = append(regoFiles, RegoFile{
				RegoFile: filepath.Join(dirname, f.Name()),
			})
		}
	}

	// 解析 rego 元信息
	var policies []*Policy
	for _, r := range regoFiles {
		p, err := ParseMeta(r.RegoFile, r.MetaFile)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}

	return policies, nil
}

type Policy struct {
	Id   string `json:"Id"`
	Meta Meta   `json:"meta"`
	Rego string `json:"rego"`
}

type Meta struct {
	Category      string `json:"category"`
	File          string `json:"file"`
	Id            string `json:"id"`
	Name          string `json:"name"`
	PolicyType    string `json:"policy_type"`
	ReferenceId   string `json:"reference_id"`
	ResourceType  string `json:"resource_type"`
	Severity      string `json:"severity"`
	Version       int    `json:"version"`
	Description   string `json:"description"`
	FixSuggestion string `json:"fixSuggestion"`
}

//ParseMeta 解析 rego metadata，如果存在 file.json 则从 json 文件读取 metadat，否则通过头部注释读取 metadata
func ParseMeta(regoFilePath string, metaFilePath string) (policy *Policy, err e.Error) {
	var meta Meta
	buf, er := os.ReadFile(regoFilePath)
	if er != nil {
		return nil, e.New(e.PolicyRegoInvalid, fmt.Errorf("read rego file: %v", err))
	}
	regoContent := string(buf)

	// 1. 如果存在 json metadata，则解析 json 文件
	policy = &Policy{}
	if metaFilePath != "" {
		content, er := os.ReadFile(metaFilePath)
		if er != nil {
			return nil, e.New(e.PolicyMetaInvalid, fmt.Errorf("read meta file: %v", err))
		}
		er = json.Unmarshal(content, &meta)
		if er != nil {
			return nil, e.New(e.PolicyMetaInvalid, fmt.Errorf("unmarshal meta file: %v", err))
		}
		policy.Meta = meta
		policy.Id = meta.Id
		policy.Rego = string(regoContent)

		return policy, nil
	}

	// 2. 无 json metadata，通过头部注释解析信息
	//	## id 为策略在策略组中的唯一标识，由大小写英文字符、数字、"."、"_"、"-" 组成
	//	## 建议按`组织_云商_资源名称/分类_编号`的格式进行命名
	//	# @id: cloudiac_alicloud_security_p001
	//
	//	# @name: 策略名称A
	//	# @description: 这是策略的描述
	//
	//	## 策略类型，如 aws, k8s, github, alicloud, ...
	//	# @policy_type: alicloud
	//
	//	## 资源类型，如 aws_ami, k8s_pod, alicloud_ecs, ...
	//	# @resource_type: aliyun_ami
	//
	//	## 策略严重级别: 可选 HIGH/MEDIUM/LOW
	//	# @severity: HIGH
	//
	//	## 策略分类(或者叫标签)，多个分类使用逗号分隔
	//	# @category: cat1,cat2
	//
	//	## 策略修复建议（支持多行）
	//	# @fix_suggestion:
	//	Terraform 代码去掉`associate_public_ip_address`配置
	//	```
	//resource "aws_instance" "bar" {
	//  ...
	//- associate_public_ip_address = true
	//}
	//```
	//	# @fix_suggestion_end

	meta = Meta{
		Id:           ExtractStr("id", regoContent),
		File:         filepath.Base(regoFilePath),
		Name:         utils.FileNameWithoutExt(regoFilePath),
		Description:  ExtractStr("description", regoContent),
		PolicyType:   ExtractStr("policy_type", regoContent),
		ResourceType: ExtractStr("resource_type", regoContent),
		Category:     ExtractStr("label", regoContent),
		ReferenceId:  ExtractStr("reference_id", regoContent),
		Severity:     ExtractStr("severity", regoContent),
	}
	ver := ExtractStr("version", regoContent)
	meta.Version, _ = strconv.Atoi(ver)

	// 多行注释提取
	regex := regexp.MustCompile("(?s)@fix_suggestion:\\s*(.*)\\s*#+\\s*@fix_suggestion_end")
	match := regex.FindStringSubmatch(regoContent)
	if len(match) == 2 {
		meta.FixSuggestion = strings.TrimSpace(match[1])
	} else {
		// 单行注释提取
		meta.FixSuggestion = ExtractStr("fix_suggestion", regoContent)
	}

	if meta.ResourceType == "" {
		return nil, e.New(e.PolicyRegoMissingComment, fmt.Errorf("missing resource type info"))
	}
	if meta.PolicyType == "" {
		// alicloud_instance => alicloud
		meta.PolicyType = meta.ResourceType[:strings.Index(meta.ResourceType, "_")]
	}
	if meta.Severity == "" {
		meta.Severity = consts.PolicySeverityMedium
	}

	policy.Id = meta.Id
	policy.Meta = meta
	policy.Rego = regoContent
	return policy, nil
}

// ExtractStr 提取 # @keyword: xxx 格式字符串
func ExtractStr(keyword string, input string) string {
	regex := regexp.MustCompile(fmt.Sprintf("(?m)^\\s*#+\\s*@%s:\\s*(.*)$)", keyword))
	match := regex.FindStringSubmatch(input)
	if len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}
