package services

import (
	"cloudiac/consts"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/services/vcsrv"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	fPath "path"
)

type MetaTemplate struct {
	Version   string    `yaml:"version"`
	Templates Templates `yaml:"Templates"`
}

type Templates struct {
	Name      string                 `yaml:"name"`
	Terraform Terraform              `yaml:"terraform"`
	Ansible   Ansible                `yaml:"ansible"`
	Env       map[string]interface{} `yaml:"env"`
}

type Terraform struct {
	Workdir string                 `yaml:"workdir"`
	Var     map[string]interface{} `yaml:"var"`
	VarFile string                 `yaml:"var_file"`
}

type Ansible struct {
	Workdir   string `yaml:"workdir"`
	Playbook  string `yaml:"playbook"`
	Inventory string `yaml:"inventory"`
}

func MetaAnalysis(content []byte) (MetaTemplate, error) {
	var mt MetaTemplate
	content = []byte(os.ExpandEnv(string(content)))

	if err := yaml.Unmarshal(content, &mt); err != nil {
		return MetaTemplate{}, fmt.Errorf("yaml.Unmarshal: %v", err)
	}

	return mt, nil
}

func InitMetaTemplate() {
	dbSess := db.Get()
	logger := logs.Get()
	vcs, err := GetDefaultVcs(dbSess)
	if err != nil {
		logger.Errorf("vcs query err: %v", err)
		return
	}
	vcsService, err := vcsrv.GetVcsInstance(vcs)
	if err != nil {
		logger.Errorf("vcs service new err: %v", err)
		return
	}

	repos, err := vcsService.ListRepos("", "", 0, 0)
	if err != nil {
		logger.Errorf("vcs service new err: %v", err)
		return
	}

	for _, repo := range repos {
		files, err := repo.ListFiles(vcsrv.VcsIfaceOptions{
			Search: consts.TfVarFileMatch,
		})
		fileNameMatch2Analysis(files, repo, vcs)
		if err != nil {
			logger.Debug("vcs get files err: %v", err)
			continue
		}
	}
}
func fileNameMatch2Analysis(files []string, repo vcsrv.RepoIface, vcs *models.Vcs)  {
	for _, file := range files {
		matched, err := fPath.Match(consts.MetaYmlMatch, file)
		if err != nil {
			logs.Get().Debug("file name match err: %v", err)
			continue
		}
		if matched {
			content, err := repo.ReadFileContent("", file)
			if err != nil {
				logs.Get().Debug("ReadFileContent err: %v", err)
				continue
			}
			mt, err := MetaAnalysis(content)
			if err != nil {
				logs.Get().Debug("MetaAnalysis err: %v", err)
				continue
			}
			varb, _ := json.Marshal(mt.Templates.Terraform.Var)
			//envb,_:=json.Marshal(mt.Templates.Env)
			if _, err := CreateTemplate(db.Get(), models.Template{
				Name:     mt.Templates.Name,
				Guid:     utils.GenGuid("ct"),
				OrgId:    0,
				Vars:     models.JSON(varb),
				VcsId:    vcs.Id,
				Playbook: mt.Templates.Ansible.Playbook,
			}); err != nil {
				logs.Get().Debug("CreateTemplate err: %v", err)
			}
		}
	}
}
