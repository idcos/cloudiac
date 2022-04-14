// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"fmt"
	"sort"
	"time"
)

func CreateOrganization(tx *db.Session, org models.Organization) (*models.Organization, e.Error) {
	if org.Id == "" {
		org.Id = models.NewId("org")
	}
	if err := models.Create(tx, &org); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.OrganizationAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &org, nil
}

func UpdateOrganization(tx *db.Session, id models.Id, attrs models.Attrs) (org *models.Organization, re e.Error) {
	org = &models.Organization{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Organization{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.OrganizationAlreadyExists)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update org error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(org); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query org error: %v", err))
	}
	return
}

func DeleteOrganization(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Organization{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete org error: %v", err))
	}
	return nil
}

func GetOrganizationById(tx *db.Session, id models.Id) (*models.Organization, e.Error) {
	o := models.Organization{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.OrganizationNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func GetOrganizationNotExistsByName(tx *db.Session, name string) (*models.Organization, error) {
	o := models.Organization{}
	if err := tx.Where("name = ?", name).First(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func GetUserByAlias(tx *db.Session, alias string) (*models.Organization, error) {
	o := models.Organization{}
	if err := tx.Where("alias = ?", alias).First(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func FindOrganization(query *db.Session) (orgs []*models.Organization, err error) {
	err = query.Find(&orgs)
	return
}

func QueryOrganization(query *db.Session) *db.Session {
	query = query.Model(&models.Organization{})
	// 创建人名称
	query = query.Joins("left join iac_user as u on u.id = iac_org.creator_id").
		LazySelectAppend("u.name as creator,iac_org.*")
	return query
}

func CreateUserOrgRel(tx *db.Session, userOrg models.UserOrg) (*models.UserOrg, e.Error) {
	if err := models.Create(tx, &userOrg); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.UserAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &userOrg, nil
}

func DeleteUserOrgRel(tx *db.Session, userId models.Id, orgId models.Id) e.Error {
	if _, err := tx.Where("user_id = ? AND org_id = ?", userId, orgId).Delete(&models.UserOrg{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete user %v for org %v error: %v", userId, orgId, err))
	}
	return nil
}

func UpdateUserOrgRel(tx *db.Session, userOrg models.UserOrg) e.Error {
	attrs := models.Attrs{"role": userOrg.Role}
	if _, err := models.UpdateAttr(tx.Where("user_id = ? and org_id = ?", userOrg.UserId, userOrg.OrgId), &models.UserOrg{}, attrs); err != nil {
		return e.New(e.DBError, fmt.Errorf("update user org error: %v", err))
	}
	return nil
}

func FindUsersOrgRel(query *db.Session, userId models.Id, orgId models.Id) (userOrgRel []*models.UserOrg, err error) {
	if err := query.Where("user_id = ? AND org_id = ?", userId, orgId).Find(&userOrgRel); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return
}

func GetOrgIdsByUser(query *db.Session, userId models.Id) (orgIds []models.Id, err e.Error) {
	var userOrgRel []*models.UserOrg
	if err := query.Where("user_id = ?", userId).Find(&userOrgRel); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	for _, o := range userOrgRel {
		orgIds = append(orgIds, o.OrgId)
	}
	return
}

func GetUserIdsByOrg(query *db.Session, orgId models.Id) (userIds []models.Id, err e.Error) {
	var userOrgRel []*models.UserOrg
	if err := query.Where("org_id = ?", orgId).Find(&userOrgRel); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	for _, o := range userOrgRel {
		userIds = append(userIds, o.UserId)
	}
	return
}

func GetDemoOrganization(tx *db.Session) (*models.Organization, e.Error) {
	o := models.Organization{}
	if err := tx.Where("is_demo = 1").First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.OrganizationNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func TryAddDemoRelation(tx *db.Session, userId models.Id) (err e.Error) {
	if common.DemoOrgId == "" {
		return
	}
	demoProject, _ := GetDemoProject(tx, models.Id(common.DemoOrgId))
	// 用户加入演示组织
	_, err = CreateUserOrgRel(tx, models.UserOrg{OrgId: models.Id(common.DemoOrgId), UserId: userId, Role: consts.OrgRoleAdmin})
	if err != nil {
		return
	}
	// 用户加入演示项目
	_, err = CreateProjectUser(tx, models.UserProject{
		Role:      consts.ProjectRoleManager,
		UserId:    userId,
		ProjectId: demoProject.Id,
	})
	return
}

func GetOrgResourcesQuery(tx *db.Session, searchStr string, orgId, userId models.Id, isSuperAdmin bool) *db.Session {
	query := tx.Joins("inner join iac_env on iac_env.last_res_task_id = iac_resource.task_id left join " +
		"iac_project on iac_resource.project_id = iac_project.id").
		LazySelectAppend("iac_project.name as project_name, iac_env.name as env_name, iac_resource.id as resource_id," +
			"iac_resource.name as resource_name, iac_resource.task_id, iac_resource.project_id as project_id, iac_resource.attrs as attrs," +
			"iac_resource.env_id as env_id, iac_resource.provider as provider, iac_resource.type, iac_resource.module")
	query = query.Where("iac_env.org_id = ?", orgId)
	if searchStr != "" {
		query = query.Where("iac_resource.name like ? OR iac_resource.type like ? OR iac_resource.attrs like ?", fmt.Sprintf("%%%s%%", searchStr),
			fmt.Sprintf("%%%s%%", searchStr), fmt.Sprintf("%%%s%%", searchStr))
	}
	if !isSuperAdmin {
		// 查一下当前用户属于哪些项目
		query = query.Joins("left join iac_user_project on iac_user_project.project_id = iac_resource.project_id").
			LazySelectAppend("iac_user_project.user_id")
		query = query.Where("iac_user_project.user_id = ?", userId)
	}
	return query

}

type EnvStatResult struct {
	MyStatus string
	Id       models.Id
	Name     string
	Count    int
}

func GetOrgProjectsEnvStat(tx *db.Session, orgId models.Id, projectIds []string) ([]resps.EnvStatResp, e.Error) {
	/* sample sql:
	select
		t.status as my_status,
		iac_project.id as id,
		iac_project.name as name,
		count(*) as count
	from
		(
		select
			if(task_status = '',
			status,
			task_status) as status,
			project_id
		from
			iac_env
		where
			archived = 0
			and org_id = 'org-c8gg9fosm56injdlb85g'
			and project_id in ('p-c8gg9josm56injdlb86g', 'p-c8kmkngsm56jqosq6bkg')
	) as t
	JOIN iac_project ON
		t.project_id = iac_project.id
	group by
		t.status, iac_project.id;
	*/

	subQuery := tx.Model(&models.Env{}).Select(`if(task_status = '', status, task_status) as status, project_id`)
	subQuery = subQuery.Where("archived = ?", 0).Where("org_id = ?", orgId)

	if len(projectIds) > 0 {
		subQuery = subQuery.Where("project_id in ?", projectIds)
	}

	query := tx.Table("(?) as t", subQuery.Expr()).Select(`t.status as my_status, iac_project.id as id, iac_project.name as name, count(*) as count`)

	query = query.Joins(`JOIN iac_project ON t.project_id = iac_project.id`)
	query = query.Group("t.status, iac_project.id")

	var dbResults []EnvStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2EnvStatResp(dbResults), nil
}

func dbResult2EnvStatResp(dbResults []EnvStatResult) []resps.EnvStatResp {
	var m = make(map[string][]EnvStatResult)
	var mTotalCount = make(map[string]int)
	for _, result := range dbResults {
		if _, ok := m[result.MyStatus]; !ok {
			m[result.MyStatus] = make([]EnvStatResult, 0)
			mTotalCount[result.MyStatus] = 0
		}
		m[result.MyStatus] = append(m[result.MyStatus], result)
		mTotalCount[result.MyStatus] += result.Count
	}

	var results = make([]resps.EnvStatResp, 0)
	for k, v := range m {
		data := resps.EnvStatResp{
			Status:  k,
			Count:   mTotalCount[k],
			Details: make([]resps.DetailStatResp, 0),
		}

		for _, p := range v {
			data.Details = append(data.Details, resps.DetailStatResp{
				Id:    p.Id,
				Name:  p.Name,
				Count: p.Count,
			})
		}
		results = append(results, data)
	}

	return results
}

type ResStatResult struct {
	ResType string
	Id      models.Id
	Name    string
	Count   int
}

func GetOrgProjectsResStat(tx *db.Session, orgId models.Id, projectIds []string, limit int) ([]resps.ResStatResp, e.Error) {
	/* sample sql
	select
		iac_resource.type as res_type,
		iac_project.id as id,
		iac_project.name as name,
		count(*) as count
	from
		iac_resource
	join iac_env on
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	join iac_project on
		iac_project.id = iac_resource.project_id
	where
		iac_env.org_id = 'org-c8gg9fosm56injdlb85g'
		and iac_env.project_id in ('p-c8gg9josm56injdlb86g', 'p-c8kmkngsm56jqosq6bkg')
	group by
		iac_resource.type, iac_project.id
	order by
		count desc
	limit 10;
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.type as res_type, iac_project.id as id, iac_project.name as name, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Joins(`join iac_project on iac_project.id = iac_resource.project_id`)
	query = query.Where(`iac_env.org_id = ?`, orgId)

	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}

	query = query.Group("iac_resource.type, iac_project.id").Order("count desc")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbResults []ResStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2ResStatResp(dbResults), nil
}

func dbResult2ResStatResp(dbResults []ResStatResult) []resps.ResStatResp {

	var m = make(map[string][]ResStatResult)
	var mTotalCount = make(map[string]int)
	for _, result := range dbResults {
		if _, ok := m[result.ResType]; !ok {
			m[result.ResType] = make([]ResStatResult, 0)
			mTotalCount[result.ResType] = 0
		}
		m[result.ResType] = append(m[result.ResType], result)
		mTotalCount[result.ResType] += result.Count
	}

	var results []resps.ResStatResp
	for k, v := range m {
		data := resps.ResStatResp{
			ResType: k,
			Count:   mTotalCount[k],
			Details: make([]resps.DetailStatResp, 0),
		}

		for _, p := range v {
			data.Details = append(data.Details, resps.DetailStatResp{
				Id:    p.Id,
				Name:  p.Name,
				Count: p.Count,
			})
		}
		results = append(results, data)
	}

	return results
}

type ProjectStatResult struct {
	ResType string
	Date    string
	Id      models.Id
	Name    string
	Count   int
}

func GetOrgProjectStat(tx *db.Session, orgId models.Id, projectIds []string, limit int) ([]resps.ProjOrEnvResStatResp, e.Error) {
	/* sample sql:
	select
		iac_resource.project_id as id,
		iac_project.name as name,
		iac_resource.type as res_type,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m") as date,
		count(*) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	JOIN iac_project ON
		iac_project.id = iac_resource.project_id
	where
		iac_env.org_id = 'org-c8gg9fosm56injdlb85g'
		AND iac_env.project_id IN ('p-c8gg9josm56injdlb86g', 'aaa')
		AND (DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(CURDATE(), "%Y-%m")
			OR
		DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m"))
	group by
		date,
		iac_resource.type,
		iac_resource.project_id
	limit 10;
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.project_id as id, iac_project.name as name, iac_resource.type as res_type, DATE_FORMAT(iac_resource.applied_at, "%Y-%m") as date, count(*) as count`)

	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Joins("JOIN iac_project ON iac_project.id = iac_resource.project_id")
	query = query.Where("iac_env.org_id = ?", orgId)
	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(CURDATE(), "%Y-%m") OR DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m")`)

	query = query.Group("date,iac_resource.type,iac_resource.project_id")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbResults []ProjectStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2ProjectResStatResp(dbResults), nil
}

func dbResult2ProjectResStatResp(dbResults []ProjectStatResult) []resps.ProjOrEnvResStatResp {
	// date -> resType -> data
	now := time.Now()
	curMonth := now.Format("2006-01")
	lastMonth := now.AddDate(0, -1, 0).Format("2006-01")

	m, mResTypeCount, mProjectCount := splitProjectResStatDataByMonth(dbResults)

	var results = make([]resps.ProjOrEnvResStatResp, 2)
	results[0].Date = lastMonth
	results[0].ResTypes = getProjectResStatDataByMonth(m[lastMonth], mResTypeCount, mProjectCount, lastMonth)

	results[1].Date = curMonth
	results[1].ResTypes = getProjectResStatDataByMonth(m[curMonth], mResTypeCount, mProjectCount, curMonth)

	// 计算增长数量
	for i := range results[1].ResTypes {
		// 某资源类型下各个项目增长数量总和
		resKey := [2]string{lastMonth, results[1].ResTypes[i].ResType}
		results[1].ResTypes[i].Up = results[1].ResTypes[i].Count
		if _, ok := mResTypeCount[resKey]; ok {
			results[1].ResTypes[i].Up -= mResTypeCount[resKey]
		}

		// 某资源类型下某个项目增长数量
		for j := range results[1].ResTypes[i].Details {
			projectKey := [3]string{lastMonth, results[1].ResTypes[i].ResType, results[1].ResTypes[i].Details[j].Id.String()}
			results[1].ResTypes[i].Details[j].Up = results[1].ResTypes[i].Details[j].Count
			if _, ok := mProjectCount[projectKey]; ok {
				results[1].ResTypes[i].Details[j].Up -= mProjectCount[projectKey]
			}
		}
	}

	return results
}

func getProjectResStatDataByMonth(resTypeData map[string][]ProjectStatResult, mResTypeCount map[[2]string]int, mProjectCount map[[3]string]int, month string) []resps.ResTypeDetailStatWithUpResp {
	var results = make([]resps.ResTypeDetailStatWithUpResp, 0)

	for resType, data := range resTypeData {
		details := make([]resps.DetailStatWithUpResp, 0)
		for _, d := range data {
			details = append(details, resps.DetailStatWithUpResp{
				Id:    d.Id,
				Name:  d.Name,
				Count: mProjectCount[[3]string{month, resType, d.Id.String()}],
			})
		}
		results = append(results, resps.ResTypeDetailStatWithUpResp{
			ResType: resType,
			Count:   mResTypeCount[[2]string{month, resType}],
			Details: details,
		})
	}
	return results
}

func splitProjectResStatDataByMonth(dbResults []ProjectStatResult) (map[string]map[string][]ProjectStatResult, map[[2]string]int, map[[3]string]int) {

	// date -> resType -> data
	m := make(map[string]map[string][]ProjectStatResult)
	now := time.Now()
	curMonth := now.Format("2006-01")
	lastMonth := now.AddDate(0, -1, 0).Format("2006-01")

	m[curMonth] = make(map[string][]ProjectStatResult)
	m[lastMonth] = make(map[string][]ProjectStatResult)

	// 计算数量
	mResTypeCount := make(map[[2]string]int) // date+resType
	mProjectCount := make(map[[3]string]int) // date+resType+projectid

	for _, result := range dbResults {
		switch result.Date {
		case curMonth:
			if m[curMonth][result.ResType] == nil {
				m[curMonth][result.ResType] = make([]ProjectStatResult, 0)
			}
			m[curMonth][result.ResType] = append(m[curMonth][result.ResType], result)
		case lastMonth:
			if m[lastMonth][result.ResType] == nil {
				m[lastMonth][result.ResType] = make([]ProjectStatResult, 0)
			}
			m[lastMonth][result.ResType] = append(m[lastMonth][result.ResType], result)
		}

		resTypeKey := [2]string{result.Date, result.ResType}
		if _, ok := mResTypeCount[resTypeKey]; !ok {
			mResTypeCount[resTypeKey] = 0
		}
		mResTypeCount[resTypeKey] += 1

		projectKey := [3]string{result.Date, result.ResType, string(result.Id)}
		if _, ok := mProjectCount[projectKey]; !ok {
			mProjectCount[projectKey] = 0
		}
		mProjectCount[projectKey] += 1
	}

	// 补全当前月缺失的资源类型
	for resType := range m[lastMonth] {
		if _, ok := m[curMonth][resType]; !ok {
			m[curMonth][resType] = make([]ProjectStatResult, 0)
		}
	}

	// 补全上个月缺失的资源类型
	for resType := range m[curMonth] {
		if _, ok := m[lastMonth][resType]; !ok {
			m[lastMonth][resType] = make([]ProjectStatResult, 0)
		}
	}

	return m, mResTypeCount, mProjectCount
}

func GetOrgResGrowTrend(tx *db.Session, orgId models.Id, projectIds []string, days int) ([]resps.ResGrowTrendResp, e.Error) {
	/* sample sql
	select
		iac_resource.project_id as id,
		iac_project.name as name,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date,
		count(*) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	JOIN iac_project ON
		iac_project.id = iac_resource.project_id
	where
		iac_env.org_id = 'org-c8gg9fosm56injdlb85g'
		and iac_env.project_id in ('p-c8gg9josm56injdlb86g', 'aaa')
		and DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY), "%Y-%m-%d")
	group by
		date,
		iac_resource.project_id
	order by
		date
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.project_id as id, iac_project.name as name, DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Joins("JOIN iac_project ON iac_project.id = iac_resource.project_id")

	query = query.Where("iac_env.org_id = ?", orgId)
	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}

	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), "%Y-%m-%d")`, days)

	query = query.Group("date, iac_resource.project_id").Order("date")

	var dbResults []ProjectStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -1*days)
	endDate := now

	return getResGrowTrendByDays(startDate, endDate, dbResults, days), nil
}

func getResGrowTrendByDays(startDate, endDate time.Time, dbResults []ProjectStatResult, days int) []resps.ResGrowTrendResp {

	// date -> detail(project or env)
	var m = make(map[string][]ProjectStatResult)
	var mDateCount = make(map[string]int)
	var mDetailCount = make(map[[2]string]int)

	for i := 0; i < days; i++ {
		startDate = startDate.AddDate(0, 0, 1)
		if startDate.Format("2006-01-02") > endDate.Format("2006-01-02") {
			break
		}
		m[startDate.Format("2006-01-02")] = make([]ProjectStatResult, 0)
	}

	for _, data := range dbResults {
		if _, ok := m[data.Date]; !ok {
			continue
		}

		m[data.Date] = append(m[data.Date], data)

		if _, ok := mDateCount[data.Date]; !ok {
			mDateCount[data.Date] = 0
		}
		mDateCount[data.Date] += data.Count

		detailKey := [2]string{data.Date, data.Id.String()}
		mDetailCount[detailKey] = data.Count
	}

	return dbResults2ResGrowTrendResp(m, mDateCount)
}

func dbResults2ResGrowTrendResp(m map[string][]ProjectStatResult, mDateCount map[string]int) []resps.ResGrowTrendResp {

	var results = make([]resps.ResGrowTrendResp, 0)
	for date, v := range m {
		details := make([]resps.DetailStatResp, 0)
		for _, d := range v {
			details = append(details, resps.DetailStatResp{
				Id:    d.Id,
				Name:  d.Name,
				Count: d.Count,
			})
		}

		results = append(results, resps.ResGrowTrendResp{
			Date:    date,
			Count:   mDateCount[date],
			Details: details,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Date < results[j].Date
	})

	return results
}
