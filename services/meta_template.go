package services

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/services/vcsrv"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
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

func DeleteMetaTemplate(tx *db.Session) error {
	_, err := tx.Delete(&models.MetaTemplate{})
	return err
}

type MetaFile struct {
	Version   string             `yaml:"version"`
	Templates []MetaFileTemplate `yaml:"templates"`
}

type MetaFileTemplate struct {
	Name        string            `yaml:"name"`
	Terraform   MetaFileTerraform `yaml:"terraform"`
	Ansible     MetaFileAnsible   `yaml:"ansible"`
	Env         map[string]string `yaml:"env"`
	Branch      string            `yaml:"branch"`
	Description string            `yaml:"description"`
	Timeout     int64             `yaml:"timeout"`
}

type MetaFileTerraform struct {
	Workdir   string            `yaml:"workdir"`
	Var       map[string]string `yaml:"var"`
	VarFile   string            `yaml:"var_file"`
	SaveState bool              `yaml:"save_state"`
}

type MetaFileAnsible struct {
	Workdir   string `yaml:"workdir"`
	Playbook  string `yaml:"playbook"`
	Inventory string `yaml:"inventory"`
}

func MetaAnalysis(content []byte) (MetaFile, error) {
	var mt MetaFile
	content = []byte(os.ExpandEnv(string(content)))

	if err := yaml.Unmarshal(content, &mt); err != nil {
		return MetaFile{}, fmt.Errorf("yaml.Unmarshal: %v", err)
	}

	return mt, nil
}

func InitMetaTemplate(tx *db.Session) error {
	logger := logs.Get()
	//清空meta template 数据库
	err := DeleteMetaTemplate(tx)
	if err != nil {
		logger.Errorf("meta template delete err: %v", err)
		return err
	}
	vcs, err := GetDefaultVcs(tx)
	if err != nil {
		logger.Errorf("vcs query err: %v", err)
		return err
	}
	vcsService, err := vcsrv.GetVcsInstance(vcs)
	if err != nil {
		logger.Errorf("vcs service new err: %v", err)
		return err
	}

	// 内置代码库统一在 cloud-iac 目录下，`make repos` 会将代码库 clone 到该目录
	repos, _, err := vcsService.ListRepos("cloud-iac", "", 0, 0)
	if err != nil {
		logger.Errorf("vcs service new err: %v", err)
		return err
	}

	for _, repo := range repos {
		project, _ := repo.FormatRepoSearch()
		files, err := repo.ListFiles(vcsrv.VcsIfaceOptions{
			Search: consts.MetaYmlMatch,
		})
		if err != nil {
			logger.Debugf("vcs get files err: %v", err)
			continue
		}

		fileNameMatch2Analysis(files, repo, vcs, project, tx)
	}
	return nil
}
func fileNameMatch2Analysis(files []string, repo vcsrv.RepoIface, vcs *models.Vcs, project *vcsrv.Projects, tx *db.Session) {
	for _, file := range files {
		content, err := repo.ReadFileContent("master", file)
		if err != nil {
			logs.Get().Debugf("ReadFileContent err: %v", err)
			continue
		}
		mt, err := MetaAnalysis(content)
		if err != nil {
			logs.Get().Debugf("MetaAnalysis err: %v", err)
			continue
		}
		for _, template := range mt.Templates {
			var branch string = template.Branch
			if branch == "" {
				branch = project.DefaultBranch
			}
			if _, err := CreateMetaTemplate(tx.Debug(), models.MetaTemplate{
				Name:        template.Name,
				Vars:        models.JSON(var2TerraformVar(template.Terraform.Var, template.Env)),
				Playbook:    template.Ansible.Playbook,
				SaveState:   template.Terraform.SaveState,
				VcsId:       vcs.Id,
				RepoBranch:  branch,
				RepoAddr:    project.HTTPURLToRepo,
				RepoId:      project.ID,
				Timeout:     template.Timeout,
				Description: template.Description,
			}); err != nil {
				logs.Get().Debugf("CreateTemplate err: %v", err)
			}
		}
	}
}

//将terraform变量和环境变量进行格式转换
func var2TerraformVar(vars, env map[string]string) []byte {
	//{"id": "7894bfc3-813d-453a-8f12-8d6be1428408", "key": "ALICLOUD_ACCESS_KEY", "value": "be7baaff819dc6edc3ee71022ed5310c03636f174e828a3069d52884243a33332e53938c431fa8dc", "isSecret": true}
	envNew := make([]map[string]interface{}, 0)
	for k, v := range vars {
		envNew = append(envNew, map[string]interface{}{
			"key":      k,
			"value":    v,
			"type":     consts.Terraform,
			"id":       utils.GenGuid(""),
			"isSecret": false,
		})
	}
	for k, v := range env {
		envNew = append(envNew, map[string]interface{}{
			"key":      k,
			"value":    v,
			"type":     consts.Env,
			"id":       utils.GenGuid(""),
			"isSecret": false,
		})
	}

	b, _ := json.Marshal(envNew)
	return b

}
