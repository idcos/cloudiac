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
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CreateVcs(c *ctx.ServiceCtx, form *forms.CreateVcsForm) (interface{}, e.Error) {
	vcs, err := services.CreateVcs(c.DB(), models.Vcs{
		OrgId:    c.OrgId,
		Name:     form.Name,
		VcsType:  form.VcsType,
		Address:  form.Address,
		VcsToken: form.VcsToken,
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return vcs, nil
}

// 判断前端传递组织id是否具有该vcs仓库读写权限
func checkOrgVcsAuth(c *ctx.ServiceCtx, id models.Id) (vcs *models.Vcs, err e.Error) {
	vcs, err = services.QueryVcsByVcsId(id, c.DB())
	if err != nil {
		return nil, err
	}
	if vcs.OrgId != c.OrgId {
		return nil, e.New(e.VcsNotExists, http.StatusForbidden, fmt.Errorf("The organization does not have the Vcs permission"))
	}
	return vcs, nil

}

func UpdateVcs(c *ctx.ServiceCtx, form *forms.UpdateVcsForm) (vcs *models.Vcs, err e.Error) {
	vcs, err = checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	attrs := models.Attrs{}
	if form.HasKey("status") {
		attrs["status"] = form.Status
	}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}
	if form.HasKey("vcsType") {
		attrs["vcsType"] = form.VcsType
	}
	if form.HasKey("address") {
		attrs["address"] = form.Address
	}
	if form.HasKey("vcsToken") {
		attrs["vcsToken"] = form.VcsToken
	}
	vcs, err = services.UpdateVcs(c.DB(), form.Id, attrs)
	return
}

func SearchVcs(c *ctx.ServiceCtx, form *forms.SearchVcsForm) (interface{}, e.Error) {
	rs, err := getPage(services.QueryVcs(c.OrgId, form.Status, form.Q, c.DB()), form, models.Vcs{})
	if err != nil {
		return nil, err
	}
	return rs, nil

}

func DeleteVcs(c *ctx.ServiceCtx, form *forms.DeleteVcsForm) (result interface{}, re e.Error) {
	_, err := checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	if err := services.DeleteVcs(c.DB(), form.Id); err != nil {
		return nil, err
	}
	return
}

func ListEnableVcs(c *ctx.ServiceCtx) (interface{}, e.Error) {
	return services.QueryEnableVcs(c.OrgId, c.DB())

}

func GetReadme(c *ctx.ServiceCtx, form *forms.GetReadmeForm) (interface{}, e.Error) {
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
	b, er := repo.ReadFileContent(form.Branch, "README.md")
	if er != nil {
		if strings.Contains(er.Error(), "not found") {
			b = make([]byte, 0)
		} else {
			return nil, e.New(e.VcsError, er)
		}
	}

	res := gin.H{"content": string(b)}
	return res, nil
}

func ListRepos(c *ctx.ServiceCtx, form *forms.GetGitProjectsForm) (interface{}, e.Error) {
	vcs, err := checkOrgVcsAuth(c, form.Id)
	if err != nil {
		return nil, err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	limit := form.PageSize()
	offset := utils.PageSize2Offset(form.CurrentPage(), limit)
	repo, total, er := vcsService.ListRepos("", form.Q, limit, offset)
	if er != nil {
		return nil, e.New(e.VcsError, er)
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

func listRepoRevision(c *ctx.ServiceCtx, form *forms.GetGitRevisionForm, revisionType string) (revision []*Revision, err e.Error) {
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
	for _, v := range revisionList {
		revision = append(revision, &Revision{
			v,
		})
	}
	return revision, nil

}

func ListRepoBranches(c *ctx.ServiceCtx, form *forms.GetGitRevisionForm) (brans []*Revision, err e.Error) {
	brans, err = listRepoRevision(c, form, "branches")
	return brans, err
}

func ListRepoTags(c *ctx.ServiceCtx, form *forms.GetGitRevisionForm) (tags []*Revision, err e.Error) {
	tags, err = listRepoRevision(c, form, "tags")
	return tags, err

}

func VcsTfVarsSearch(c *ctx.ServiceCtx, form *forms.TemplateTfvarsSearchForm) (interface{}, e.Error) {
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
		Search: consts.TfVarFileMatch,
	})
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	return listFiles, nil
}

func VcsPlaybookSearch(c *ctx.ServiceCtx, form *forms.TemplatePlaybookSearchForm) (interface{}, e.Error) {
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
		Ref:       form.RepoRevision,
		Search:    consts.PlaybookMatch,
		Recursive: true,
		Path:      consts.Ansible,
	})
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	return listFiles, nil
}

func VcsVariableSearch(c *ctx.ServiceCtx, form *forms.TemplateVariableSearchForm) (interface{}, e.Error) {
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
