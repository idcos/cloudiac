// Copyright 2021 CloudJ Company Limited. All rights reserved.

package vcsrv

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
)

/*
version control service 接口
*/

type UserInfo struct {
	Login string `json:"login" form:"login" `
	Id    int    `json:"id" form:"id" `
	Name  string `json:"name" form:"name" `
}

const (
	WebhookUrlGitlab = "/webhooks/gitlab"
	WebhookUrlGitea  = "/webhooks/gitea"
	WebhookUrlGitee  = "/webhooks/gitee"
	WebhookUrlGithub = "/webhooks/github"
)

type VcsIfaceOptions struct {
	Ref       string
	Path      string
	Search    string
	Recursive bool
	Limit     int
	Offset    int
}

type VcsIface interface {
	// GetRepo 列出仓库
	// param idOrPath: 仓库id或者路径
	GetRepo(idOrPath string) (RepoIface, error)

	// ListRepos 列出仓库列表
	// param namespace: namespace 可用于表示用户、组织等
	// param search: 搜索字符串
	// param limit: 限制返回的文件数，传 0 表示无限制
	// return in64(分页total数量)
	ListRepos(namespace, search string, limit, offset int) ([]RepoIface, int64, error)

	// UserInfo 获取用户信息
	UserInfo() (UserInfo, error)
}

type RepoIface interface {
	// ListBranches 获取分支列表
	ListBranches() ([]string, error)
	ListTags() ([]string, error)

	// BranchCommitId
	//param branch: 分支
	BranchCommitId(branch string) (string, error)

	// ListFiles 列出指定路径下的文件
	// param ref: 分支或者 commit id
	// param filename: 文件名称部分的点
	// param path: 路径
	// param search: 搜索字符串
	// param recursive: 是否递归遍历子目录
	// param limit: 限制返回的文件数，传 0 表示无限制
	// return: 返回文件路径列表，路径为完整路径(即包含传入的 path 部分)
	ListFiles(option VcsIfaceOptions) ([]string, error)

	// ReadFileContent
	// param path: 路径
	// param branch: 分支
	ReadFileContent(branch, path string) (content []byte, err error)

	// FormatRepoSearch 格式化输出前端需要的内容
	FormatRepoSearch() (project *Projects, err e.Error)

	// DefaultBranch 获取默认分支
	DefaultBranch() string

	//ListWebhook 查询Webhook列表
	ListWebhook() ([]ProjectsHook, error)

	//DeleteWebhook 查询Webhook列表
	DeleteWebhook(id int) error

	//AddWebhook 查询Webhook列表
	AddWebhook(url string) error

	//CreatePrComment 添加PR评论
	CreatePrComment(prId int, comment string) error
}

func GetVcsInstance(vcs *models.Vcs) (VcsIface, error) {
	switch vcs.VcsType {
	case consts.GitTypeLocal:
		return newLocalVcs(vcs.Address), nil
	case consts.GitTypeGitLab:
		return newGitlabInstance(vcs)
	case consts.GitTypeGitEA:
		return newGiteaInstance(vcs)
	case consts.GitTypeGithub:
		return newGithubInstance(vcs)
	case consts.GitTypeGitee:
		return newGiteeInstance(vcs)
	case consts.GitTypeRegistry:
		return newRegistryVcs(vcs)
	default:
		return nil, errors.New("vcs type doesn't exist")
	}
}

func GetRepo(vcs *models.Vcs, repoId string) (RepoIface, error) {
	v, err := GetVcsInstance(vcs)
	if err != nil {
		return nil, err
	}
	return v.GetRepo(repoId)
}

func matchGlob(search, name string) bool {
	if search == "" {
		return true
	}

	pattern := fmt.Sprintf("*%s*", search)
	matched, err := path.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

//校验ref是否为空 空则返回默认分支
func getBranch(repo RepoIface, ref string) string {
	if ref == "" {
		return repo.DefaultBranch()
	}
	return ref
}

func GetRepoAddress(repo RepoIface) (string, error) {
	p, err := repo.FormatRepoSearch()
	if err != nil {
		return "", err
	}
	return p.HTTPURLToRepo, nil
}

func SetWebhook(vcs *models.Vcs, repoId, apiToken string, triggers []string) error {
	webhookUrl := GetWebhookUrl(vcs, apiToken)
	repo, err := GetRepo(vcs, repoId)
	if err != nil {
		return err
	}
	webhooks, err := repo.ListWebhook()
	if err != nil {
		return err
	}
	var webhookId int
	isExist := false
	for _, webhook := range webhooks {
		// 如果url相同，证明仓库中存在webhook；
		if webhook.URL == webhookUrl {
			isExist = true
			webhookId = webhook.ID
			break
		}
	}
	//空值时删除
	if len(triggers) == 0 {
		// 判断同vcs、仓库的环境是否存在
		envExist, err := db.Get().Model(&models.Env{}).
			Joins("left join iac_template as tpl on iac_env.tpl_id = tpl.id").
			Where("tpl.vcs_id = ?", vcs.Id).
			Where("iac_env.triggers IS NOT NULL or iac_env.triggers != '{}'").Exists()
		if err != nil {
			return err
		}
		// 判断同vcs、仓库的环境是否存在
		tplExist, err := db.Get().Model(&models.Template{}).
			Where("tpl.vcs_id = ?", vcs.Id).
			Where("iac_template.triggers IS NOT NULL or iac_template.triggers != '{}'").Exists()
		if err != nil {
			return err
		}
		//如果同vcs、仓库的环境和云模板不存在，则删除代码仓库中的webhook
		if !envExist && !tplExist {
			if err := repo.DeleteWebhook(webhookId); err != nil {
				return err
			}
		}
		return nil
	} else {
		// 存在则忽略，不存在则添加
		if !isExist {
			if err := repo.AddWebhook(webhookUrl); err != nil {
				return err
			}
		}
		return nil
	}
}

func GetVcsToken(token string) (string, error) {
	return utils.DecryptSecretVar(token)
}

func GetWebhookUrl(vcs *models.Vcs, apiToken string) string {
	webhookUrl := configs.Get().Portal.Address + "/api/v1"
	switch vcs.VcsType {
	case models.VcsGitlab:
		webhookUrl += WebhookUrlGitlab
	case models.VcsGitea:
		webhookUrl += WebhookUrlGitea
	case models.VcsGitee:
		webhookUrl += WebhookUrlGitee
	case models.VcsGithub:
		webhookUrl += WebhookUrlGithub
	}
	webhookUrl += fmt.Sprintf("/%s?token=%s", vcs.Id.String(), apiToken)
	return webhookUrl
}

func IsNotFoundErr(err error) bool {
	if err == nil {
		return false
	}

	if er, ok := err.(e.Error); ok && er.Code() == e.ObjectNotExists {
		return true
	}
	if strings.Contains(err.Error(), "not found") {
		return true
	}
	return false
}

func GetUser(vcs *models.Vcs) (UserInfo, error) {
	v, err := GetVcsInstance(vcs)
	if err != nil {
		return UserInfo{}, err
	}
	return v.UserInfo()
}
