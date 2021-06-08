package apps

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/services/vcsrv"
	"fmt"
)

func CreateVcs(c *ctx.ServiceCtx, form *forms.CreateVcsForm) (interface{}, e.Error) {
	vcs, err := services.CreateVcs(c.DB(), models.Vcs{
		OrgId:    c.OrgId,
		Name:     form.Name,
		VcsType:  form.VcsType,
		Status:   form.Status,
		Address:  form.Address,
		VcsToken: form.VcsToken,
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return vcs, nil
}

func UpdateVcs(c *ctx.ServiceCtx, form *forms.UpdateVcsForm) (vcs *models.Vcs, err e.Error) {
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
	if err := services.DeleteVcs(c.DB(), form.Id); err != nil {
		return nil, err
	}
	return
}

func ListEnableVcs(c *ctx.ServiceCtx) (interface{}, e.Error) {
	return services.QueryEnableVcs(c.OrgId, c.DB())

}

func GetReadme(c *ctx.ServiceCtx, form *forms.GetReadmeForm) (interface{}, e.Error) {
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.DB())
	if err != nil {
		return nil, err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	repo, er := vcsService.GetRepo(repoIdOrPath(form.RepoId, form.RepoPath))
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	b, er := repo.ReadFileContent(form.Branch, "README.md")
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	res := models.FileContent{
		Content: string(b[:]),
	}
	return res, nil
}

func ListRepos(c *ctx.ServiceCtx, form *forms.GetGitProjectsForm) (interface{}, e.Error) {
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.DB())
	if err != nil {
		return nil, err
	}

	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.GitLabError, err)
	}
	repo, er := vcsService.ListRepos("", form.Q, uint(form.PageSize_), uint(form.CurrentPage_))
	if er != nil {
		return nil, e.New(e.GitLabError, err)
	}
	project := make([]*vcsrv.Projects, 0)
	var total int64
	for _, repo := range repo {
		total++
		proj, er := repo.FormatRepoSearch()
		if er != nil {
			return nil, err
		}
		project = append(project, proj)
	}

	return page.PageResp{
		Total:    total,
		PageSize: form.CurrentPage(),
		List:     project,
	}, nil
}

type Branches struct {
	Name string `json:"name"`
}

func repoIdOrPath(id int, path string) string {
	if id != 0 {
		return fmt.Sprintf("%d", id)
	}
	return path
}

func ListRepoBranches(c *ctx.ServiceCtx, form *forms.GetGitBranchesForm) (brans []*Branches, err e.Error) {
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.DB())
	if err != nil {
		return nil, err
	}

	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}

	repo, er := vcsService.GetRepo(repoIdOrPath(form.RepoId, form.RepoPath))
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	branchList, er := repo.ListBranches("", 0, 0)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	for _, v := range branchList {
		brans = append(brans, &Branches{
			v,
		})
	}
	return brans, nil
}

func VcsTfVarsSearch(c *ctx.ServiceCtx, form *forms.TemplateTfvarsSearchForm) (interface{}, e.Error) {
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.DB())
	if err != nil {
		return nil, err
	}

	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.GitLabError, err)
	}
	repo, er := vcsService.GetRepo(repoIdOrPath(int(form.RepoId), form.RepoPath))
	if er != nil {
		return nil, e.New(e.GitLabError, err)
	}
	listFiles, er := repo.ListFiles(vcsrv.VcsIfaceOptions{
		Ref:    form.RepoBranch,
		Search: consts.TfVarFileMatch,
	})
	if er != nil {
		return nil, e.New(e.GitLabError, err)
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
		return nil, e.New(e.GitLabError, err)
	}
	repo, er := vcsService.GetRepo(repoIdOrPath(int(form.RepoId), form.RepoPath))
	if er != nil {
		return nil, e.New(e.GitLabError, err)
	}
	listFiles, er := repo.ListFiles(vcsrv.VcsIfaceOptions{
		Ref:    form.RepoBranch,
		Search: consts.PlaybookMatch,
	})
	if er != nil {
		return nil, e.New(e.GitLabError, err)
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
		return nil, e.New(e.GitLabError, err)
	}
	repo, er := vcsService.GetRepo(repoIdOrPath(int(form.RepoId), form.RepoPath))
	if er != nil {
		return nil, e.New(e.GitLabError, err)
	}
	listFiles, er := repo.ListFiles(vcsrv.VcsIfaceOptions{
		Ref:    form.RepoBranch,
		Search: consts.VariablePrefix,
	})
	if er != nil {
		return nil, e.New(e.GitLabError, err)
	}
	tvl := make([]services.TemplateVariable, 0)
	for _, file := range listFiles {
		content, er := repo.ReadFileContent(form.RepoBranch, file)
		if er != nil {
			return nil, e.New(e.GitLabError, err)
		}
		tvs, er := services.TemplateVariableSearch(content)
		if er != nil {
			return nil, e.New(e.GitLabError, err)
		}
		tvl = append(tvl, tvs...)
	}

	return tvl, nil
}
