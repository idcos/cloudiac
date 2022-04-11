// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

func getRepo(vcsId models.Id, query *db.Session, repoId string) (*vcsrv.Projects, error) {
	vcs, err := services.QueryVcsByVcsId(vcsId, query)
	if err != nil {
		return nil, err
	}
	vcsIface, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, er
	}
	repo, er := vcsIface.GetRepo(repoId)
	if er != nil {
		return nil, er
	}

	p, err := repo.FormatRepoSearch()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func CreateTemplate(c *ctx.ServiceContext, form *forms.CreateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create template %s", form.Name))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	template, err := services.CreateTemplate(tx, models.Template{
		Name:         form.Name,
		OrgId:        c.OrgId,
		Description:  form.Description,
		VcsId:        form.VcsId,
		RepoId:       form.RepoId,
		RepoFullName: form.RepoFullName,
		// 云模板的 repoAddr 和 repoToken 可以为空，若为空则会在创建任务时自动查询 vcs 获取相应值
		RepoAddr:     "",
		RepoToken:    "",
		RepoRevision: form.RepoRevision,
		CreatorId:    c.UserId,
		Workdir:      form.Workdir,
		Playbook:     form.Playbook,
		PlayVarsFile: form.PlayVarsFile,
		TfVarsFile:   form.TfVarsFile,
		TfVersion:    form.TfVersion,
		PolicyEnable: form.PolicyEnable,
		Triggers:     form.TplTriggers,
		KeyId:        form.KeyId,
	})

	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create template, err %s", err)
		if err.Code() == e.TemplateAlreadyExists {
			return nil, e.New(err.Code(), err.Err(), http.StatusBadRequest)
		}
		return nil, err
	}

	// 绑定云模版和策略组的关系
	if len(form.PolicyGroup) > 0 {
		policyForm := &forms.UpdatePolicyRelForm{
			Id:             template.Id,
			Scope:          consts.ScopeTemplate,
			PolicyGroupIds: form.PolicyGroup,
		}
		if _, err = services.UpdatePolicyRel(tx, policyForm); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	// 创建模板与项目的关系
	if err := services.CreateTemplateProject(tx, form.ProjectId, template.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	{
		updateVarsForm := forms.UpdateObjectVarsForm{
			Scope:     consts.ScopeTemplate,
			ObjectId:  template.Id,
			Variables: form.Variables,
		}
		if _, er := updateObjectVars(c, tx, &updateVarsForm); er != nil {
			_ = tx.Rollback()
			return nil, er
		}
	}

	// 创建变量组与实例的关系
	if err := services.BatchUpdateRelationship(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeTemplate, template.Id.String()); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit create template, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	if form.PolicyEnable {
		scanForm := &forms.ScanTemplateForm{
			Id: template.Id,
		}
		go func() {
			_, err := ScanTemplateOrEnv(c, scanForm, "")
			if err != nil {
				c.Logger().Errorf("open tpl policy scan err: %v, tpl id: %s", err, template.Id)
			}
		}()
	}

	// 设置 webhook
	if err := setVcsRepoWebhook(c, template.VcsId, template.RepoId, form.TplTriggers); err != nil {
		c.Logger().Errorf("set webhook err :%v", err)
	}
	return template, nil
}

func setAttrsByFormKeys(attrs models.Attrs, form *forms.UpdateTemplateForm) {
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}
	if form.HasKey("playbook") {
		attrs["playbook"] = form.Playbook
	}
	if form.HasKey("status") {
		attrs["status"] = form.Status
	}
	if form.HasKey("workdir") {
		attrs["workdir"] = form.Workdir
	}
	if form.HasKey("tfVarsFile") {
		attrs["tfVarsFile"] = form.TfVarsFile
	}
	if form.HasKey("playVarsFile") {
		attrs["playVarsFile"] = form.PlayVarsFile
	}
	if form.HasKey("tfVersion") {
		attrs["tfVersion"] = form.TfVersion
	}
	if form.HasKey("repoRevision") {
		attrs["repoRevision"] = form.RepoRevision
	}
	if form.HasKey("policyEnable") {
		attrs["policyEnable"] = form.PolicyEnable
	}
	if form.HasKey("tplTriggers") {
		attrs["triggers"] = pq.StringArray(form.TplTriggers)
	}
	if form.HasKey("keyId") {
		attrs["keyId"] = form.KeyId
	}
}

func setAttrsVcsInfoByForm(attrs models.Attrs, form *forms.UpdateTemplateForm) {
	if form.HasKey("vcsId") && form.HasKey("repoId") && form.HasKey("repoFullName") {
		attrs["vcsId"] = form.VcsId
		attrs["repoId"] = form.RepoId
		attrs["repoFullName"] = form.RepoFullName
		if form.VcsId != "" {
			// 当云模板关联了 vcs 时需要清空 repoAddr，这样才能支持 vcs 更新。
			attrs["repoAddr"] = ""
		}
	}
}

func updatetplByFormKey(c *ctx.ServiceContext, tx *db.Session, tpl *models.Template, form *forms.UpdateTemplateForm) e.Error {
	var err e.Error
	// 更新和策略组的绑定关系
	if form.HasKey("policyGroup") {
		policyForm := &forms.UpdatePolicyRelForm{
			Id:             tpl.Id,
			Scope:          consts.ScopeTemplate,
			PolicyGroupIds: form.PolicyGroup,
		}
		_, err = services.UpdatePolicyRel(tx, policyForm)
	}
	if err != nil {
		return err
	}

	if form.HasKey("projectId") {
		if err = services.DeleteTemplateProject(tx, form.Id); err != nil {
			return err
		}
		err = services.CreateTemplateProject(tx, form.ProjectId, form.Id)
	}
	if err != nil {
		return err
	}

	if form.HasKey("variables") {
		updateVarsForm := forms.UpdateObjectVarsForm{
			Scope:     consts.ScopeTemplate,
			ObjectId:  form.Id,
			Variables: form.Variables,
		}
		_, err = updateObjectVars(c, tx, &updateVarsForm)
	}
	if err != nil {
		return err
	}

	if form.HasKey("varGroupIds") || form.HasKey("delVarGroupIds") {
		// 创建变量组与实例的关系
		err = services.BatchUpdateRelationship(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeTemplate, form.Id.String())
	}
	return err
}

func UpdateTemplate(c *ctx.ServiceContext, form *forms.UpdateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %s", form.Id))

	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.TemplateNotExists, err, http.StatusBadRequest)
	}

	// 根据云模板ID, 组织ID查询该云模板是否属于该组织
	if tpl.OrgId != c.OrgId {
		return nil, e.New(e.TemplateNotExists, http.StatusForbidden, fmt.Errorf("the organization does not have permission to delete the current template"))
	}
	attrs := models.Attrs{}
	setAttrsByFormKeys(attrs, form)
	setAttrsVcsInfoByForm(attrs, form)

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	if tpl, err = services.UpdateTemplate(tx, form.Id, attrs); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 更新和策略组的绑定关系
	err = updatetplByFormKey(c, tx, tpl, form)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit update template, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	// 自动触发一次检测
	if form.PolicyEnable {
		tplScanForm := &forms.ScanTemplateForm{
			Id: tpl.Id,
		}
		go func() {
			_, err := ScanTemplateOrEnv(c, tplScanForm, "")
			if err != nil {
				c.Logger().Errorf("open tpl policy scan err: %v, tpl id: %s", err, tpl.Id)
			}
		}()
	}

	// 设置 webhook
	if err := setVcsRepoWebhook(c, tpl.VcsId, tpl.RepoId, tpl.Triggers); err != nil {
		c.Logger().Errorf("set webhook err :%v", err)
	}

	return tpl, err
}

func DeleteTemplate(c *ctx.ServiceContext, form *forms.DeleteTemplateForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete template %s", form.Id))
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	// 根据ID 查询云模板是否存在
	tpl, err := services.GetTemplateById(tx, form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %v", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	// 根据云模板ID, 组织ID查询该云模板是否属于该组织
	if tpl.OrgId != c.OrgId {
		return nil, e.New(e.TemplateNotExists, http.StatusForbidden, fmt.Errorf("The organization does not have permission to delete the current template"))
	}

	// 查询模板是否有活跃环境
	if ok, err := services.QueryActiveEnv(tx.Where("tpl_id = ?", form.Id)).Exists(); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	} else if ok {
		return nil, e.New(e.TemplateActiveEnvExists, http.StatusMethodNotAllowed,
			fmt.Errorf("The cloud template cannot be deleted because there is an active environment"))
	}
	// 删除策略组关系
	if err := services.DeletePolicyGroupRel(tx, form.Id, consts.ScopeTemplate); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 根据ID 删除云模板
	if err := services.DeleteTemplate(tx, tpl.Id); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 删除 webhook
	if err := delVcsRepoWebhook(c, tpl.VcsId, tpl.RepoId); err != nil {
		c.Logger().Errorf("delete webhook err :%v", err)
	}
	return nil, nil
}

func TemplateDetail(c *ctx.ServiceContext, form *forms.DetailTemplateForm) (*resps.TemplateDetailResp, e.Error) {
	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	project_ids, err := services.QueryProjectByTplId(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	varialbeList, err := services.SearchVariableByTemplateId(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	if tpl.RepoFullName == "" {
		repo, err := getRepo(tpl.VcsId, c.DB(), tpl.RepoId)
		if err != nil {
			return nil, e.New(e.VcsError, err)
		}
		tpl.RepoFullName = repo.FullName
	}
	temp, err := services.GetPolicyRels(c.DB(), tpl.Id, consts.ScopeTemplate)
	if err != nil {
		return nil, err
	}
	policyGroups := []string{}
	for _, v := range temp {
		policyGroups = append(policyGroups, v.PolicyGroupId)
	}

	tplDetail := &resps.TemplateDetailResp{
		Template:    tpl,
		Variables:   varialbeList,
		ProjectList: project_ids,
		PolicyGroup: policyGroups,
	}
	return tplDetail, nil

}

func getTplIdList(db *db.Session, projectId models.Id) ([]models.Id, e.Error) {
	tplIdList := make([]models.Id, 0)
	if projectId == "" {
		return tplIdList, nil
	}

	return services.QueryTplByProjectId(db, projectId)
}

func updateTaskAndPolicyStatus(db *db.Session, templates []*resps.SearchTemplateResp) ([]string, e.Error) {
	vcsIds := make([]string, 0)
	for _, v := range templates {
		if v.RepoAddr == "" {
			vcsIds = append(vcsIds, v.VcsId)
		}
		var scanTaskStatus string
		// 如果开启
		scanTask, err := services.GetTplLastScanTask(db, v.Id)
		if err != nil {
			scanTaskStatus = ""
			if !e.IsRecordNotFound(err) {
				return vcsIds, e.New(e.DBError, err)
			}
		} else {
			scanTaskStatus = scanTask.PolicyStatus
		}
		v.PolicyStatus = models.PolicyStatusConversion(scanTaskStatus, v.PolicyEnable)
	}
	return vcsIds, nil
}

func updateTmplRepoAddr(templates []*resps.SearchTemplateResp, vcsAttr map[string]models.Vcs) {

	portAddr := configs.Get().Portal.Address
	for _, tpl := range templates {
		if tpl.RepoAddr == "" && tpl.RepoFullName != "" {
			if vcsAttr[tpl.VcsId].VcsType == consts.GitTypeLocal {
				tpl.RepoAddr = utils.JoinURL(portAddr, vcsAttr[tpl.VcsId].Address, tpl.RepoId)
			} else {
				tpl.RepoAddr = utils.JoinURL(vcsAttr[tpl.VcsId].Address, tpl.RepoFullName)
			}
		}
	}
}

func SearchTemplate(c *ctx.ServiceContext, form *forms.SearchTemplateForm) (tpl interface{}, err e.Error) {
	tplIdList, err := getTplIdList(c.DB(), c.ProjectId)
	if err != nil {
		return nil, err
	}
	if c.ProjectId != "" && len(tplIdList) == 0 {
		return getEmptyListResult(form)
	}

	query := services.QueryTemplateByOrgId(c.DB(), form.Q, c.OrgId, tplIdList, c.ProjectId)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	templates := make([]*resps.SearchTemplateResp, 0)
	if err := p.Scan(&templates); err != nil {
		return nil, e.New(e.DBError, err)
	}

	vcsIds, err := updateTaskAndPolicyStatus(c.DB(), templates)
	if err != nil {
		return nil, err
	}

	vcsList, err := services.GetVcsListByIds(c.DB(), vcsIds)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	vcsAttr := make(map[string]models.Vcs)
	for _, v := range vcsList {
		vcsAttr[v.Id.String()] = v
	}
	updateTmplRepoAddr(templates, vcsAttr)

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     templates,
	}, nil
}

type TplCheckResult struct {
	TfVars   TplCheckResultItem
	Playbook TplCheckResultItem
}
type TplCheckResultItem struct {
	Error string
}

func TemplateChecks(c *ctx.ServiceContext, form *forms.TemplateChecksForm) (interface{}, e.Error) {

	// 如果云模版名称传入，校验名称是否重复.
	if form.Name != "" {
		tpl, err := services.QueryTemplateByName(c.DB().Where("id != ?", form.TemplateId), form.Name, c.OrgId)
		if tpl != nil {
			return nil, e.New(e.TemplateNameRepeat, err)
		}
		// 数据库相关错误
		if err != nil && err.Code() != e.TemplateNotExists {
			return nil, err
		}
	}
	if form.Workdir != "" {
		// 检查工作目录下.tf 文件是否存在
		searchForm := &forms.RepoFileSearchForm{
			RepoId:       form.RepoId,
			RepoRevision: form.RepoRevision,
			VcsId:        form.VcsId,
			Workdir:      form.Workdir,
		}
		results, err := VcsRepoFileSearch(c, searchForm, "", consts.TfFileMatch)
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, e.New(e.TemplateWorkdirError, fmt.Errorf("no '%s' files", consts.TfFileMatch))
		}
	}

	err, checkResult := CheckTemplateOrEnvConfig(c, form.TfVarsFile, form.Playbook, form.RepoId, form.RepoRevision, form.Workdir, form.VcsId)
	if err != nil {
		return nil, err
	}
	if checkResult.Playbook.Error != "" || checkResult.TfVars.Error != "" {
		return checkResult, e.New(e.BadParam)
	}
	return resps.TemplateChecksResp{
		CheckResult: consts.TplTfCheckSuccess,
	}, nil
}

//TemplateDeployCheck 只检测选择的目录下template是否合法
func TemplateDeployCheck(c *ctx.ServiceContext, form *forms.TemplateChecksForm) e.Error {
	// 检查工作目录下.tf 文件是否存在
	if form.Workdir != "" {
		searchForm := &forms.RepoFileSearchForm{
			RepoId:       form.RepoId,
			RepoRevision: form.RepoRevision,
			VcsId:        form.VcsId,
			Workdir:      form.Workdir,
		}
		results, err := VcsRepoFileSearch(c, searchForm, "", consts.TfFileMatch)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return e.New(e.TemplateWorkdirError, fmt.Errorf("no '%s' files", consts.TfFileMatch))
		}
	}
	return nil
}

func CheckTemplateOrEnvConfig(c *ctx.ServiceContext, tfVarsFile, playbook, repoId, reporevision, workdir string, vcsId models.Id) (e.Error, TplCheckResult) {
	checkResult := TplCheckResult{}
	if tfVarsFile != "" {
		searchForm := &forms.RepoFileSearchForm{
			RepoId:       repoId,
			RepoRevision: reporevision,
			VcsId:        vcsId,
			Workdir:      workdir,
		}
		results, err := VcsRepoFileSearch(c, searchForm, "", consts.TfVarFileMatch)
		if err != nil {
			return err, checkResult
		}
		if len(results) == 0 {
			checkResult.TfVars.Error = "Under the current working directory Tfvars end file does not exist"
		} else if !utils.ArrayIsExistsStr(results, tfVarsFile) {
			checkResult.TfVars.Error = "Under the current working directory Tfvars end file does not exist"
		}
	}
	if playbook != "" {
		searchForm := &forms.RepoFileSearchForm{
			RepoId:       repoId,
			RepoRevision: reporevision,
			VcsId:        vcsId,
			Workdir:      workdir,
		}
		results, err := VcsRepoFileSearch(c, searchForm, consts.PlaybookDir, consts.PlaybookMatch)
		if err != nil {
			return err, checkResult
		}
		if len(results) == 0 {
			checkResult.Playbook.Error = "The playbook file in the current working directory does not exist"
		}
		if !utils.ArrayIsExistsStr(results, playbook) {
			checkResult.Playbook.Error = "The playbook file in the current working directory does not exist"
		}
	}
	return nil, checkResult
}

func setVcsRepoWebhook(c *ctx.ServiceContext, vcsId models.Id, repoId string, triggers pq.StringArray) error {
	vcs, err := services.QueryVcsByVcsId(vcsId, c.DB())
	if err != nil {
		return err
	}

	// 获取token
	token, err := GetWebhookToken(c)
	if err != nil {
		return err
	}

	if err := vcsrv.SetWebhook(vcs, string(repoId), token.Key, triggers); err != nil {
		return err
	}
	return nil
}

func delVcsRepoWebhook(c *ctx.ServiceContext, vcsId models.Id, repoId string) error {
	return setVcsRepoWebhook(c, vcsId, repoId, []string{})
}
