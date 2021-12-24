// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
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

func DownloadPolicyGroup(tx *db.Session, result *DownloadPolicyGroupResult, wg *sync.WaitGroup) {
	group := result.Group

	defer wg.Done()

	// 生成临时工作目录
	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		result.Error = e.New(e.InternalError, errors.Wrapf(err, "create tmp dir"), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	// 1. git download
	branch := group.GitTags
	if group.Branch != "" {
		branch = group.Branch
	}
	repoAddr, commitId, err := GetPolicyGroupCommitId(tx, group.VcsId, group.RepoId, branch)
	if err != nil {
		result.Error = e.New(e.InternalError, errors.Wrapf(err, "create tmp dir"), http.StatusInternalServerError)
		return
	}
	// generate checkout command line
	cmdline := genGitCheckoutScript(tmpDir, repoAddr, commitId, group.Dir)

	// git checkout
	cmd := exec.Command("sh", "-c", cmdline)
	cmd.Dir = tmpDir
	// setpgid 保证可以杀死进程及子进程
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	done := make(chan error)
	go func() {
		if err := cmd.Start(); err != nil {
			logrus.Errorf("error start cmd %s, err: %v", cmd.Path, err)
			done <- err
		}
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(consts.PolicyGroupDownloadTimeoutSecond):
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			result.Error = e.New(e.InternalError, errors.Wrapf(err, fmt.Sprintf("kill timeout process %s error %v", cmd.Path, err)), http.StatusInternalServerError)
		} else {
			result.Error = e.New(e.InternalError, errors.Wrapf(err, fmt.Sprintf("kill timeout process %s with timeout %s seconds", cmd.Path, consts.PolicyGroupDownloadTimeoutSecond/time.Second)), http.StatusInternalServerError)
		}
		return
	case err = <-done:
		if err != nil {
			result.Error = e.New(e.InternalError, errors.Wrapf(err, fmt.Sprintf("command complete with error %v", err)), http.StatusInternalServerError)
			return
		}
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("git checkout error %v, output %s", err, output)
		result.Error = e.New(e.InternalError, errors.Wrapf(err, "git checkout"), http.StatusInternalServerError)
		return
	}

	// TODO: 2. parse policy
	_, er := ParsePolicyGroup(filepath.Join(tmpDir, "code"))
	if er != nil {
		result.Error = e.New(e.InternalError, errors.Wrapf(er, "parse rego"), http.StatusInternalServerError)
		return
	}

	// TODO: 3. insert policy to db
}

func GetPolicyGroupCommitId(tx *db.Session, vcsId models.Id, repoId models.Id, branch string) (repoAddr, commitId string, err e.Error) {
	vcs, err := QueryVcsByVcsId(vcsId, tx)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return "", "", e.New(e.VcsNotExists, err)
		}
		return "", "", e.New(e.DBError, err)
	}

	repo, er := vcsrv.GetRepo(vcs, repoId.String())
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

var gitCheckoutScriptCommandTpl = `#!/bin/sh
cd '{{.WorkingDir}}'
git clone '{{.RepoAddress}}' code && \
cd code && \
echo "checkout $(git rev-parse --short HEAD)." && \
git checkout -q '{{.Revision}}' && \
cd '{{.SubDir}}'
`

func genGitCheckoutScript(wd string, repoAddr string, revision string, dir string) (command string) {
	if revision == "" || revision == "/" {
		revision = "./"
	}
	cmdline := utils.SprintTemplate(gitCheckoutScriptCommandTpl, map[string]interface{}{
		"WorkingDir":  wd,
		"RepoAddress": repoAddr,
		"SubDir":      dir,
		"Revision":    revision,
	})

	return cmdline
}

type RegoFile struct {
	MetaFile string
	RegoFile string
}

func ParsePolicyGroup(dirname string) ([]*models.Policy, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	var regoFiles []fs.FileInfo
	copy(regoFiles, files)
	var regos []RegoFile
	// 遍历当前目录
	for _, f := range files {
		// 优先处理 json 的 meta 及对应的 rego 文件
		if filepath.Ext(f.Name()) == "json" {
			if utils.FileExist(filepath.Base(f.Name()) + ".rego") {
				regos = append(regos, RegoFile{
					MetaFile: filepath.Base(f.Name()) + ".json",
					RegoFile: filepath.Base(f.Name()) + ".rego",
				})
				// 将已经处理的 rego 排除
				for i, rf := range regoFiles {
					if rf.Name() == filepath.Base(f.Name())+".rego" {
						regoFiles[i] = regoFiles[len(regoFiles)-1]
						regoFiles = regoFiles[:len(regoFiles)-1]
						break
					}
				}
			}
		}
	}
	// 遍历其他没有 json meta 的 rego 文件
	for _, f := range regoFiles {
		if filepath.Ext(f.Name()) == "rego" {
			regos = append(regos, RegoFile{
				RegoFile: filepath.Base(f.Name()) + ".rego",
			})
		}
	}

	// TODO: parse rego header

	return nil, nil
}
