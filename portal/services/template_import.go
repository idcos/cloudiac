// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
	"time"
)

type TplImporter struct {
	OrgId           models.Id
	CreatorId       models.Id
	ProjectIds      []models.Id
	Data            TplExportedData
	WhenIdDuplicate string

	Logger logs.Logger

	result   TplImportResult
	newIdMap map[models.Id]models.Id
}

type TplImportResult struct {
	Created   tplImportCount `json:"created"`
	Updated   tplImportCount `json:"updated"`
	Skipped   tplImportCount `json:"skipped"`
	Copied    tplImportCount `json:"copied"`
	Renamed   tplImportCount `json:"renamed"`
	Duplicate tplImportCount `json:"duplicate"`
}

type tplImportCount struct {
	Templates []tplImportCountItem `json:"templates"`
	Vcs       []tplImportCountItem `json:"vcs"`
	VarGroups []tplImportCountItem `json:"varGroups"`
}

type tplImportCountItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (t *TplImporter) Import(tx *db.Session) (*TplImportResult, e.Error) {
	t.newIdMap = make(map[models.Id]models.Id)

	if t.Data.Version != TplExportVersion {
		t.Logger.Infof("data.version '%s' != '%s'", t.Data.Version, TplExportVersion)
		return nil, e.New(e.InvalidExportVersion)
	}

	getResult := func(er e.Error) *TplImportResult {
		if er == nil {
			return &t.result
		}
		// 当发生 error 时只需要返回 Duplicate 数据用于回显，其他操作已经回滚
		return &TplImportResult{
			Duplicate: t.result.Duplicate,
		}
	}

	if er := t.ImportVcs(tx); er != nil {
		return getResult(er), er
	}
	if er := t.ImportVarGroups(tx); er != nil {
		return getResult(er), er
	}
	if er := t.ImportTemplates(tx); er != nil {
		return getResult(er), er
	}
	return getResult(nil), nil
}

func (t *TplImporter) ImportTemplates(tx *db.Session) e.Error {
	// 用于导入云模板与项目的关联
	bs := utils.NewBatchSQL(1024, "REPLACE INTO", models.ProjectTemplate{}.TableName(),
		"project_id", "template_id")

	for i := range t.Data.Templates {
		tpl := t.Data.Templates[i]
		if skip, err := t.importExportedTpl(tx, i, tpl); err != nil {
			return err
		} else if skip {
			continue
		}

		// 处理云模板与项目的关联
		for _, pid := range t.ProjectIds {
			bs.MustAddRow(pid, t.getImportedId(tpl.Id))
		}
	}

	for bs.HasNext() {
		sql, args := bs.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return e.AutoNew(err, e.DBError)
		}
	}
	return nil
}

// importExportedTpl
func (t *TplImporter) importExportedTpl(tx *db.Session, i int, exportedTpl exportedTpl) (skip bool, er e.Error) {
	tpl, err := t.getTplFromExportData(exportedTpl)
	if err != nil {
		return false, e.AutoNew(err, e.InternalError)
	}

	if er := t.renameTplIf(tx, tpl); er != nil {
		return false, er
	}

	dbTpl := models.Template{}
	if err := QueryTemplate(tx.Unscoped().Where("id = ?", tpl.Id)).Find(&dbTpl); err != nil {
		return false, e.AutoNew(err, e.DBError)
	}

	tplIdDuplicate := false
	if dbTpl.Id != "" {
		tplIdDuplicate = true
		if t.WhenIdDuplicate == "update" && t.OrgId != dbTpl.OrgId {
			return false, e.New(e.ImportUpdateOrgId)
		}
	}

	if tplIdDuplicate {
		t.Logger.Debugf("template %s, id duplicate", tpl.Id)
	}
	if er := t.doImport(tx, tpl.Id, tpl, tplIdDuplicate); er != nil {
		return false, er
	}

	//// 云模板跳过处理了，其关联关系及变量也就不需要处理
	if tplIdDuplicate && t.WhenIdDuplicate == "skip" {
		return true, nil
	}

	// 处理云模板变量
	er1 := t.processTplVars(tx, tpl, i, tplIdDuplicate)
	if er1 != nil {
		return false, er1
	}

	// 处理云模板关联的变量组
	if err := t.processTplVarsGroup(i, tplIdDuplicate, tx, tpl); err != nil {
		return false, err
	}

	return false, nil
}

func (t *TplImporter) processTplVarsGroup(i int, tplIdDuplicate bool, tx *db.Session, tpl *models.Template) e.Error {
	vgIds := t.Data.Templates[i].VarGroupIds
	importedVgIds := make([]models.Id, 0, len(vgIds))
	for _, id := range vgIds {
		importedVgIds = append(importedVgIds, t.getImportedId(id.String()))
	}

	if !tplIdDuplicate || t.WhenIdDuplicate == "copy" {
		if er := BatchUpdateVarGroupObjectRel(tx, importedVgIds, nil, consts.ScopeTemplate, tpl.Id); er != nil {
			return er
		}
	} else { // update
		// 选择 update 策略时，会删除所有己关联的变量组，然后重新导入关联关系
		if er := DeleteVarGroupRel(tx, consts.ScopeTemplate, tpl.Id); er != nil {
			return er
		}
		if er := BatchUpdateVarGroupObjectRel(tx, importedVgIds, nil, consts.ScopeTemplate, tpl.Id); er != nil {
			return er
		}
	}
	return nil
}

// 处理云模版变量
func (t *TplImporter) processTplVars(tx *db.Session, tpl *models.Template, i int, tplIdDuplicate bool) e.Error {
	tplVars, er := SearchVariableByTemplateId(tx, tpl.Id)
	if er != nil {
		return er
	}
	tplVarsMap := make(map[string]models.Variable, len(tplVars))
	for _, v := range tplVars {
		tplVarsMap[fmt.Sprintf("%s/%s", v.Type, v.Name)] = v
	}

	vars := t.Data.Templates[i].Variables
	for _, iVar := range vars {
		v, err := t.getVarFromExportData(t.Data.Templates[i], iVar)
		if err != nil {
			return e.AutoNew(err, e.InternalError)
		}

		if !tplIdDuplicate || t.WhenIdDuplicate == "copy" {
			if er := models.Create(tx, v); er != nil {
				return e.AutoNew(er, e.DBError)
			}
			t.addCount("created", v)
		} else {
			// update
			dbVar, varDuplicate := tplVarsMap[fmt.Sprintf("%s/%s", v.Type, v.Name)]
			if varDuplicate {
				v.Id = dbVar.Id
				if _, err := models.UpdateModelAll(tx, v); err != nil {
					return e.AutoNew(err, e.DBError)
				}
				t.addCount("updated", v)
			} else {
				if er := models.Create(tx, v); er != nil {
					return e.AutoNew(er, e.DBError)
				}
				t.addCount("created", v)
			}
		}
	}
	return nil
}

func (t *TplImporter) ImportVcs(tx *db.Session) e.Error {
	vcsList := t.Data.Vcs
	for _, iVcs := range vcsList {
		vcs, err := t.getVcsFromExportData(iVcs)
		if err != nil {
			return e.AutoNew(err, e.InternalError)
		}

		if er := t.renameVcsIf(tx, vcs); er != nil {
			return er
		}

		dbVcs := models.Vcs{}
		if err := QueryVcsSample(tx.Unscoped().Where("id = ?", vcs.Id)).Find(&dbVcs); err != nil {
			return e.AutoNew(err, e.DBError)
		}

		vcsIdDuplicate := false
		if dbVcs.Id != "" {
			vcsIdDuplicate = true
			if t.WhenIdDuplicate == "update" && t.OrgId != dbVcs.OrgId {
				return e.New(e.ImportUpdateOrgId)
			}
		}

		if er := t.doImport(tx, vcs.Id, vcs, vcsIdDuplicate); er != nil {
			return e.AutoNew(er, e.DBError)
		}
	}
	return nil
}

func (t *TplImporter) ImportVarGroups(tx *db.Session) e.Error {
	for i := range t.Data.VarGroups {
		vg, err := t.getVarGroupFromExportData(t.Data.VarGroups[i])
		if err != nil {
			return e.AutoNew(err, e.InternalError)
		}

		if er := t.renameVarGroupIf(tx, vg); er != nil {
			return er
		}

		dbVg := models.VariableGroup{}
		if err := QueryVarGroup(tx.Unscoped().Where("id = ?", vg.Id)).Find(&dbVg); err != nil {
			return e.AutoNew(err, e.DBError)
		}

		vgIdDuplicate := false
		if dbVg.Id != "" {
			vgIdDuplicate = true
			if t.WhenIdDuplicate == "update" && t.OrgId != dbVg.OrgId {
				return e.New(e.ImportUpdateOrgId)
			}
		}

		// 注意：变量组及其变量是一个整体，当进行 update 时变量组的所有变量会被完整替换为导入的值
		if er := t.doImport(tx, vg.Id, vg, vgIdDuplicate); er != nil {
			return er
		}
	}
	return nil
}

func (t *TplImporter) addCount(op string, item models.Modeler) {
	opMap := map[string]*tplImportCount{
		"created":   &t.result.Created,
		"updated":   &t.result.Updated,
		"copied":    &t.result.Copied,
		"renamed":   &t.result.Renamed,
		"skipped":   &t.result.Skipped,
		"duplicate": &t.result.Duplicate,
	}

	countOp, ok := opMap[op]
	if !ok {
		panic(fmt.Errorf("unknown count op: %s", op))
	}

	switch o := item.(type) {
	case *models.Variable:
		t.Logger.Infof("%s %s", op, o.Id)
	case *models.Template:
		t.Logger.Infof("%s %s", op, o.Id)
		countItem := tplImportCountItem{Id: o.Id.String(), Name: o.Name}
		countOp.Templates = append(countOp.Templates, countItem)
	case *models.Vcs:
		t.Logger.Infof("%s %s", op, o.Id)
		countItem := tplImportCountItem{Id: o.Id.String(), Name: o.Name}
		countOp.Vcs = append(countOp.Vcs, countItem)
	case *models.VariableGroup:
		t.Logger.Infof("%s %s", op, o.Id)
		countItem := tplImportCountItem{Id: o.Id.String(), Name: o.Name}
		countOp.VarGroups = append(countOp.VarGroups, countItem)
	default:
		panic(fmt.Errorf("unknonw item type: %T", item))
	}
}

type queryRecordByName func(name string) *db.Session

func (t *TplImporter) getNewNameIf(tx *db.Session, origId models.Id, origName string, queryRecord queryRecordByName) (string, e.Error) {
	name := origName
	for i := 0; i < 5; i++ {
		var ids []models.Id
		err := queryRecord(name).Pluck("id", &ids)
		if err != nil {
			return "", e.AutoNew(err, e.DBError)
		}

		if len(ids) == 0 { // 不存在同名数据
			return name, nil
		} else if origId == ids[0] { // 存在同名数据且 id 相同，只有使用 copy 策略时才需要重命名
			if utils.StrInArray(t.WhenIdDuplicate, "copy") {
				name = t.importRename(origName, name)
			} else {
				return name, nil
			}
		} else {
			// 其他情况全部需要重命名
			name = t.importRename(origName, name)
		}
	}
	return "", e.New(e.TooManyRetries, fmt.Errorf("import template rename"))
}

// 检查模板名称是否重名，若重名则按规则添加后缀，直到不重名
func (t *TplImporter) renameTplIf(tx *db.Session, tpl *models.Template) e.Error {
	newName, er := t.getNewNameIf(tx, tpl.Id, tpl.Name, func(name string) *db.Session {
		return QueryTemplate(tx.Where("org_id = ? AND name = ?", t.OrgId, name))
	})
	if er != nil {
		return er
	}

	if newName != tpl.Name {
		tpl.Name = newName
		t.addCount("renamed", tpl)
	}
	return nil
}

func (t *TplImporter) renameVcsIf(tx *db.Session, vcs *models.Vcs) e.Error {
	newName, err := t.getNewNameIf(tx, vcs.Id, vcs.Name, func(name string) *db.Session {
		return QueryVcsSample(tx.Where("org_id = ? AND name = ?", t.OrgId, name))
	})
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}

	if newName != vcs.Name {
		vcs.Name = newName
		t.addCount("renamed", vcs)
	}
	return nil
}

func (t *TplImporter) renameVarGroupIf(tx *db.Session, vg *models.VariableGroup) e.Error {
	newName, err := t.getNewNameIf(tx, vg.Id, vg.Name, func(name string) *db.Session {
		return QueryVarGroup(tx.Where("org_id = ? AND `type` = ? AND name = ?", t.OrgId, vg.Type, name))
	})
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}

	if newName != vg.Name {
		vg.Name = newName
		t.addCount("renamed", vg)
	}
	return nil
}

func (t TplImporter) importRename(origName, currName string) string {
	now := time.Now()
	if origName == currName {
		return fmt.Sprintf("%s import-%s", currName, now.Format("0601021504"))
	} else {
		return fmt.Sprintf("%s-%d.%d", currName, now.Second(), now.Nanosecond()/int(time.Millisecond))
	}
}

func (t *TplImporter) getTplFromExportData(tpl exportedTpl) (*models.Template, error) {
	newTpl := models.Template{
		OrgId:          t.OrgId,
		CreatorId:      t.CreatorId,
		Name:           tpl.Name,
		TplType:        tpl.TplType,
		Description:    models.Text(tpl.Description),
		VcsId:          t.getImportedId(tpl.VcsId),
		RepoId:         tpl.RepoId,
		RepoAddr:       tpl.RepoAddr,
		RepoToken:      "",
		RepoRevision:   tpl.RepoRevision,
		Status:         tpl.Status,
		Workdir:        tpl.Workdir,
		TfVarsFile:     tpl.TfVarsFile,
		Playbook:       tpl.Playbook,
		PlayVarsFile:   tpl.PlayVarsFile,
		LastScanTaskId: "",
		TfVersion:      tpl.TfVersion,
	}
	newTpl.Id = models.Id(tpl.Id)

	var err error
	newTpl.RepoToken, err = ImportSecretStr(tpl.RepoToken, true)
	if err != nil {
		return nil, err
	}
	return &newTpl, nil
}

func (t *TplImporter) getVarFromExportData(tpl exportedTpl, v exportedTplVar) (*models.Variable, error) {
	value, err := ImportVariableValue(v.Value, v.Sensitive)
	if err != nil {
		return nil, err
	}

	mVar := models.Variable{
		OrgId:     t.OrgId,
		ProjectId: "",
		TplId:     t.getImportedId(tpl.Id),
		EnvId:     "",
		VariableBody: models.VariableBody{
			Scope:       v.Scope,
			Type:        v.Type,
			Name:        v.Name,
			Value:       models.Text(value),
			Options:     v.Options,
			Sensitive:   v.Sensitive,
			Description: models.Text(v.Description),
		},
	}
	// 变量以名称唯一标识，导入变量时总是生成一个新 id
	mVar.Id = mVar.NewId()
	return &mVar, nil
}

func (t *TplImporter) getVcsFromExportData(vcs exportedVcs) (*models.Vcs, error) {
	newVcs := models.Vcs{
		OrgId:    t.OrgId,
		Name:     vcs.Name,
		Status:   vcs.Status,
		VcsType:  vcs.VcsType,
		Address:  vcs.Address,
		VcsToken: "",
	}

	newVcs.Id = models.Id(vcs.Id)
	token, err := ImportSecretStr(vcs.VcsToken, true)
	if err != nil {
		return nil, err
	}
	newVcs.VcsToken = token

	return &newVcs, nil
}

func (t *TplImporter) getVarGroupFromExportData(vg exportedVarGroup) (*models.VariableGroup, error) {
	newVg := models.VariableGroup{
		OrgId:     t.OrgId,
		CreatorId: t.CreatorId,
		Name:      vg.Name,
		Type:      vg.Type,
		Variables: models.VarGroupVariables{},
	}

	var err error
	for _, v := range vg.Variables {
		v.Value, err = ImportVariableValue(v.Value, v.Sensitive)
		if err != nil {
			return nil, err
		}
		newVg.Variables = append(newVg.Variables, v)
	}

	newVg.Id = models.Id(vg.Id)
	return &newVg, nil
}

func (t *TplImporter) doImport(tx *db.Session, id models.Id, importObj models.Modeler, hasDuplicate bool) e.Error {
	setNewId := func() {
		newId := importObj.(models.ModelIdGenerator).NewId()
		importObj.(models.ModelIdSetter).SetId(newId)
		t.newIdMap[id] = newId
	}

	if !hasDuplicate {
		if err := models.Create(tx, importObj); err != nil {
			return e.AutoNew(err, e.DBError)
		}
		t.addCount("created", importObj)
		return nil
	}

	// id 有重复
	switch t.WhenIdDuplicate {
	case "update":
		if _, err := models.UpdateModelAll(tx, importObj); err != nil {
			return e.AutoNew(err, e.DBError)
		}
		t.addCount("updated", importObj)
	case "copy":
		setNewId()
		if err := models.Create(tx, importObj); err != nil {
			return e.AutoNew(err, e.DBError)
		}
		t.addCount("copied", importObj)
	case "skip":
		t.addCount("skipped", importObj)
	case "abort":
		t.addCount("duplicate", importObj)
		return e.New(e.ImportIdDuplicate)
	default:
		return e.New(e.ImportError, fmt.Errorf("unknonw duplicate policy '%s'", t.WhenIdDuplicate))
	}
	return nil
}

func (t *TplImporter) getImportedId(id string) models.Id {
	newId, ok := t.newIdMap[models.Id(id)]
	if ok {
		return newId
	}
	return models.Id(id)
}
