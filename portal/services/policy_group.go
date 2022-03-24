// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils/logs"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
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
	query := dbSess.Model(models.PolicyGroup{}).
		Joins(fmt.Sprintf("left join (%s) as p on p.group_id = %s.id",
			fmt.Sprintf("select count(group_id) as policy_count,group_id from %s group by group_id",
				models.Policy{}.TableName()), pgTable)).
		Joins(fmt.Sprintf("left join (%s) as rel on rel.group_id = %s.id",
			fmt.Sprintf("select count(group_id) as rel_count, group_id from %s group by group_id",
				models.PolicyRel{}.TableName()), pgTable)).
		Where(fmt.Sprintf("%s.org_id = ?", pgTable), orgId)
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where(fmt.Sprintf("%s.name like ?", pgTable), qs)
	}
	return query.LazySelectAppend(fmt.Sprintf("%s.*,p.policy_count,rel.rel_count", pgTable))
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

func GetPolicyGroupByTplIds(tx *db.Session, ids []models.Id) ([]resps.NewPolicyGroup, e.Error) {
	group := make([]resps.NewPolicyGroup, 0)
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

func GetPolicyGroupByEnvIds(tx *db.Session, ids []models.Id) ([]resps.NewPolicyGroup, e.Error) {
	group := make([]resps.NewPolicyGroup, 0)
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

func DownloadPolicyGroup(sess *db.Session, tmpDir string, result *DownloadPolicyGroupResult, wg *sync.WaitGroup) {
	logger := logs.Get().WithField("func", "DownloadPolicyGroup")
	group := result.Group

	defer wg.Done()

	// 1. git download
	logger.Debugf("downloading git")
	branch := group.GitTags
	if group.Branch != "" {
		branch = group.Branch
	}
	repoAddr, commitId, err := GetPolicyGroupCommitId(sess, group.VcsId, group.RepoId, branch)
	if err != nil {
		result.Error = e.New(e.InternalError, errors.Wrapf(err, "get commit id"), http.StatusInternalServerError)
		return
	}
	logger.Debugf("downloading git %s@%s to %s", repoAddr, commitId, filepath.Join(tmpDir, "code"))
	er := GitCheckout(filepath.Join(tmpDir, "code"), repoAddr, commitId)
	if er != nil {
		result.Error = e.New(e.BadRequest, errors.Wrapf(er, "checkout repo"), http.StatusBadRequest)
		return
	}
	logger.Debugf("download git complete")
}

func GetPolicyGroupCommitId(sess *db.Session, vcsId models.Id, repoId string, branch string) (repoAddr, commitId string, err e.Error) {
	vcs, err := QueryVcsByVcsId(vcsId, sess)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return "", "", e.New(e.VcsNotExists, err)
		}
		return "", "", e.New(e.DBError, err)
	}

	var repoUser = models.RepoUser
	vcsInstance, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return "", "", e.New(e.VcsError, er)
	}

	if vcs.VcsType == models.VcsGitee {
		user, er := vcsInstance.UserInfo()
		if er != nil {
			return "", "", e.New(e.VcsError, er)
		}
		repoUser = user.Login
	}

	repo, er := vcsInstance.GetRepo(repoId)
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
		u.User = url.UserPassword(repoUser, token)
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
