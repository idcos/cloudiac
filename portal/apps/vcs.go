// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateVcs(c *ctx.ServiceContext, form *forms.CreateVcsForm) (interface{}, e.Error) {
	token, err := utils.EncryptSecretVar(form.VcsToken)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}
	v := models.Vcs{
		OrgId:    c.OrgId,
		Name:     form.Name,
		VcsType:  form.VcsType,
		Address:  form.Address,
		VcsToken: token,
	}
	if err := vcsrv.VerifyVcsToken(&v); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	vcs, err := services.CreateVcs(c.DB(), models.Vcs{
		OrgId:    c.OrgId,
		Name:     form.Name,
		VcsType:  form.VcsType,
		Address:  form.Address,
		VcsToken: token,
	})

	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return vcs, nil
}

// 判断前端传递组织id是否具有该vcs仓库读写权限
func checkOrgVcsAuth(c *ctx.ServiceContext, id models.Id) (vcs *models.Vcs, err e.Error) {
	vcs, err = services.QueryVcsByVcsId(id, c.DB())
	if err != nil {
		return nil, err
	}
	if vcs.OrgId != c.OrgId && vcs.OrgId != "" {
		return nil, e.New(e.VcsNotExists, http.StatusForbidden, fmt.Errorf("The organization does not have the Vcs permission"))
	}
	return vcs, nil

}

func UpdateVcs(c *ctx.ServiceContext, form *forms.UpdateVcsForm) (vcs *models.Vcs, err e.Error) {
	vcs, err = checkOrgVcsAuth(c, form.Id) //nolint
	if err != nil {
		return nil, err
	}
	attrs := models.Attrs{}

	setAttrIfExist := func(k, v string) {
		if form.HasKey(k) {
			attrs[k] = v
		}
	}
	setAttrIfExist("status", form.Status)
	setAttrIfExist("name", form.Name)
	setAttrIfExist("vcsType", form.VcsType)
	setAttrIfExist("address", form.Address)
	if form.HasKey("vcsToken") && form.VcsToken != "" {
		token, err := utils.EncryptSecretVar(form.VcsToken)
		if err != nil {
			return nil, e.New(e.VcsError, err)
		}
		if err := services.VscTokenCheckByID(c.DB(), form.Id, token); err != nil {
			return nil, e.AutoNew(err, e.VcsInvalidToken)
		}
		attrs["vcs_Token"] = token
	}
	return services.UpdateVcs(c.DB(), form.Id, attrs)
}

func SearchVcs(c *ctx.ServiceContext, form *forms.SearchVcsForm) (interface{}, e.Error) {
	rs, err := getPage(services.QueryVcs(c.OrgId, form.Status, form.Q, form.IsShowDefaultVcs, false, c.DB()), form, models.Vcs{})
	if err != nil {
		return nil, err
	}
	return rs, nil

}

func DeleteVcs(c *ctx.ServiceContext, form *forms.DeleteVcsForm) (result interface{}, re e.Error) {
	_, err := checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	// 根据vcsId查询是否相关云模版已经被全部清除
	exist, err := services.QueryTplByVcsId(c.DB(), form.Id)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, e.New(e.VcsDeleteError, fmt.Errorf("vcs cannot be deleted. Please delete the dependent cloud template first"))
	}

	if err := services.DeleteVcs(c.DB(), form.Id); err != nil {
		return nil, err
	}
	return
}

func ListEnableVcs(c *ctx.ServiceContext) (interface{}, e.Error) {
	return services.QueryEnableVcs(c.OrgId, c.DB())

}

func GetReadme(c *ctx.ServiceContext, form *forms.GetReadmeForm) (interface{}, e.Error) {
	vcs, err := checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	repo, er := vcsService.GetRepo(form.RepoId)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	// 如果路径以 "/" 开头，部分 vcs 会报错
	dir := strings.TrimLeft(form.Dir, "/")
	b, er := repo.ReadFileContent(form.RepoRevision, path.Join(dir, "README.md"))
	if er != nil && vcsrv.IsNotFoundErr(er) {
		// README.md 文件不存在时尝试读 README 文件
		b, er = repo.ReadFileContent(form.RepoRevision, path.Join(dir, "README"))
	}
	if er != nil {
		if vcsrv.IsNotFoundErr(er) {
			b = make([]byte, 0)
		} else {
			return nil, e.AutoNew(er, e.VcsError)
		}
	}

	res := gin.H{"content": string(b)}
	return res, nil
}

func ListRepos(c *ctx.ServiceContext, form *forms.GetGitProjectsForm) (interface{}, e.Error) {
	vcs, err := checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.AutoNew(er, e.VcsError)
	}
	limit := form.PageSize()
	offset := utils.PageSize2Offset(form.CurrentPage(), limit)
	repo, total, er := vcsService.ListRepos("", form.Q, limit, offset)
	if er != nil {
		return nil, e.AutoNew(er, e.VcsError)
	}
	project := make([]*vcsrv.Projects, 0)
	for _, repo := range repo {
		proj, er := repo.FormatRepoSearch()
		if er != nil {
			return nil, er
		}
		project = append(project, proj)
	}

	return page.PageResp{
		Total:    total,
		PageSize: form.PageSize(),
		List:     project,
	}, nil
}

type Revision struct {
	Name string `json:"name"`
}

func listRepoRevision(c *ctx.ServiceContext, form *forms.GetGitRevisionForm, revisionType string) (revision []*Revision, err e.Error) {
	vcs, err := checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	repo, er := vcsService.GetRepo(form.RepoId)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	var revisionList []string
	if revisionType == "tags" {
		revisionList, er = repo.ListTags()
	} else if revisionType == "branches" {
		revisionList, er = repo.ListBranches()
	}
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	revision = make([]*Revision, 0)
	for _, v := range revisionList {
		revision = append(revision, &Revision{v})
	}
	return revision, nil
}

func ListRepoBranches(c *ctx.ServiceContext, form *forms.GetGitRevisionForm) (brans []*Revision, err e.Error) {
	brans, err = listRepoRevision(c, form, "branches")
	return brans, err
}

func ListRepoTags(c *ctx.ServiceContext, form *forms.GetGitRevisionForm) (tags []*Revision, err e.Error) {
	tags, err = listRepoRevision(c, form, "tags")
	return tags, err

}

func VcsRepoFileSearch(c *ctx.ServiceContext, form *forms.RepoFileSearchForm, searchDir string, pattern string) ([]string, e.Error) {
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.DB())

	if err != nil {
		return nil, err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	repo, er := vcsService.GetRepo(form.RepoId)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	listFiles, er := repo.ListFiles(vcsrv.VcsIfaceOptions{
		Ref:    form.RepoRevision,
		Search: pattern,
		Path:   filepath.Join(form.Workdir, searchDir),
	})

	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	return utils.StrSliceTrimPrefix(listFiles, form.Workdir+"/"), nil
}

func VcsVariableSearch(c *ctx.ServiceContext, form *forms.TemplateVariableSearchForm) (interface{}, e.Error) {
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.DB())
	if err != nil {
		return nil, err
	}

	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	repo, er := vcsService.GetRepo(form.RepoId)
	if er != nil {
		return nil, e.New(e.VcsError, err)
	}
	listFiles, er := repo.ListFiles(vcsrv.VcsIfaceOptions{
		Ref:    form.RepoRevision,
		Search: consts.VariablePrefix,
		Path:   form.Workdir,
	})
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	tvl := make([]services.TemplateVariable, 0)
	for _, file := range listFiles {
		content, er := repo.ReadFileContent(form.RepoRevision, file)
		if er != nil {
			return nil, e.New(e.VcsError, er)
		}
		tvs, er := services.ParseTfVariables(file, content)
		if er != nil {
			return nil, e.AutoNew(er, e.VcsError)
		}
		tvl = append(tvl, tvs...)
	}

	return tvl, nil
}

func GetVcsRepoFile(c *ctx.ServiceContext, form *forms.GetVcsRepoFileForm) (interface{}, e.Error) {
	vcs, err := checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	repo, er := vcsService.GetRepo(form.RepoId)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	b, er := repo.ReadFileContent(form.Branch, filepath.Join(form.Workdir, form.FileName))
	if er != nil {
		if vcsrv.IsNotFoundErr(er) {
			b = make([]byte, 0)
		} else {
			return nil, e.AutoNew(er, e.VcsError)
		}
	}

	res := gin.H{"content": string(b)}
	return res, nil
}
