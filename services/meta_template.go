package services

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/services/vcsrv"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	fPath "path"
)

func SearchMetaTemplate(query *db.Session) *db.Session {
	return query.Table(models.MetaTemplate{}.TableName()).Order("created_at DESC")
}

func GetMetaTemplateById(query *db.Session, id uint) (models.MetaTemplate, e.Error) {
	tplLib := models.MetaTemplate{}
	if err := query.Table(models.MetaTemplate{}.TableName()).Where("id = ?", id).First(&tplLib); err != nil {
		return models.MetaTemplate{}, e.New(e.DBError, err)
	}

	return tplLib, nil
}

func CreateMetaTemplate(tx *db.Session, metaTemplate models.MetaTemplate) (*models.MetaTemplate, e.Error) {
	if err := models.Create(tx, &metaTemplate); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TemplateAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &metaTemplate, nil
}

type MetaTemplateParse struct {
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
	Workdir   string                 `yaml:"workdir"`
	Var       map[string]interface{} `yaml:"var"`
	VarFile   string                 `yaml:"var_file"`
	SaveState bool                   `json:"saveState"`
}

type Ansible struct {
	Workdir   string `yaml:"workdir"`
	Playbook  string `yaml:"playbook"`
	Inventory string `yaml:"inventory"`
}

func MetaAnalysis(content []byte) (MetaTemplateParse, error) {
	var mt MetaTemplateParse
	content = []byte(os.ExpandEnv(string(content)))

	if err := yaml.Unmarshal(content, &mt); err != nil {
		return MetaTemplateParse{}, fmt.Errorf("yaml.Unmarshal: %v", err)
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

	repos, _, err := vcsService.ListRepos("", "", 0, 0)
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
func fileNameMatch2Analysis(files []string, repo vcsrv.RepoIface, vcs *models.Vcs) {
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
			if _, err := CreateMetaTemplate(db.Get(), models.MetaTemplate{
				Name:      mt.Templates.Name,
				Vars:      models.JSON(varb),
				VcsId:     vcs.Id,
				Playbook:  mt.Templates.Ansible.Playbook,
				SaveState: mt.Templates.Terraform.SaveState,
			}); err != nil {
				logs.Get().Debug("CreateTemplate err: %v", err)
			}
		}
	}
}
