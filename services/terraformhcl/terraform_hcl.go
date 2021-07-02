package terraformhcl

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/services/vcsrv"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"strings"
)

func LoadModuleFromFile(content []byte) (*tfconfig.Module, e.Error) {
	parser := hclparse.NewParser()
	file, diag := parser.ParseHCL(content, "")
	if diag.HasErrors() {
		return nil, e.New(e.TerraformHclErr, diag.Error())
	}
	mod := tfconfig.NewModule("")
	diag = tfconfig.LoadModuleFromFile(file, mod)
	if diag.Errs() != nil {

	}
	return mod, nil
}

func GetProvider(dbSess *db.Session, form *forms.CreateTemplateForm, metaTpl models.MetaTemplate) (string, e.Error) {
	var (
		repoId string = form.RepoId
		branch string = form.RepoBranch
	)
	if repoId == "" {
		repoId = metaTpl.RepoId
	}
	if branch == "" {
		branch = metaTpl.RepoBranch
	}

	vcs, err := services.QueryVcsByVcsId(form.VcsId, dbSess)
	if err != nil {
		return "", err
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return "", e.New(e.GitLabError, er)
	}

	repo, er := vcsService.GetRepo(repoId)
	if er != nil {
		return "", e.New(e.GitLabError, er)
	}
	b, er := repo.ReadFileContent(branch, "versions.tf")
	if er != nil {
		if strings.Contains(er.Error(), "not found") {
			b = make([]byte, 0)
		} else {
			return "", e.New(e.GitLabError, er)
		}
	}
	providers := make([]string, 0)
	mod, err := LoadModuleFromFile(b[:])
	if err != nil {
		return "", err
	}

	for provider, _ := range mod.RequiredProviders {
		providers = append(providers, provider)
	}

	return strings.Join(providers, ","), nil

}
