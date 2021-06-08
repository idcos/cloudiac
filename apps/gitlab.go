package apps

//import (
//	"cloudiac/consts"
//	"cloudiac/consts/e"
//	"cloudiac/libs/page"
//	"cloudiac/models"
//	"cloudiac/models/forms"
//	"cloudiac/services/vcsrv"
//	"encoding/json"
//	"time"
//)
//
//type Projects struct {
//	ID             int        `json:"id"`
//	Description    string     `json:"description"`
//	DefaultBranch  string     `json:"default_branch"`
//	SSHURLToRepo   string     `json:"ssh_url_to_repo"`
//	HTTPURLToRepo  string     `json:"http_url_to_repo"`
//	Name           string     `json:"name"`
//	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
//}
//
//func ListOrganizationRepos(vcs *models.Vcs, form *forms.GetGitProjectsForm) (interface{}, e.Error) {
//	projects, total, err := vcsrv.ListOrganizationReposById(vcs, form)
//	if err != nil {
//		return nil, err
//	}
//
//	jsonProjects, er := json.Marshal(projects)
//	if er != nil {
//		return nil, e.New(e.JSONParseError, er)
//	}
//	repos := make([]*Projects, 0)
//	er = json.Unmarshal(jsonProjects, &repos)
//	if er != nil {
//		return nil, e.New(e.JSONParseError, er)
//	}
//
//	return page.PageResp{
//		Total:    int64(total),
//		PageSize: form.CurrentPage(),
//		List:     repos,
//	}, nil
//}
//
//type Branches struct {
//	Name string `json:"name"`
//}
//
//func ListRepositoryBranches(vcs *models.Vcs, form *forms.GetGitBranchesForm) (brans []*Branches, err e.Error) {
//	branches, err := vcsrv.ListRepositoryBranches(vcs, form)
//	if err != nil {
//		return nil, err
//	}
//
//	jsonBranches, er := json.Marshal(branches)
//	if er != nil {
//		return nil, e.New(e.JSONParseError, er)
//	}
//
//	er = json.Unmarshal(jsonBranches, &brans)
//	if er != nil {
//		return nil, e.New(e.JSONParseError, er)
//	}
//	return brans, nil
//}
//
//func GetReadmeContent(vcs *models.Vcs, form *forms.GetReadmeForm) (interface{}, e.Error) {
//	content, err := vcsrv.GetReadmeContent(vcs, form)
//	if err != nil {
//		return nil, nil
//	}
//	return content, nil
//}
//
//func TemplateTfvarsSearch(vcs *models.Vcs, form *forms.TemplateTfvarsSearchForm) (interface{}, e.Error) {
//	return vcsrv.TemplateTfvarsSearch(vcs, form.RepoId, form.RepoBranch, []string{consts.TfVarFileExt})
//}
//
//func TemplatePlaybookSearch(vcs *models.Vcs, form *forms.TemplatePlaybookSearchForm) (interface{}, e.Error) {
//	return vcsrv.TemplateTfvarsSearch(vcs, form.RepoId, form.RepoBranch, []string{consts.PlaybookPrefixYml, consts.PlaybookPrefixYaml})
//}
//
////func TemplateVariableSearch(vcs *models.Vcs, form *forms.TemplateVariableSearchForm) (interface{}, e.Error) {
////	return services.TemplateVariableSearch()
////}

