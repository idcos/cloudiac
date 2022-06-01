// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"fmt"
	"path"
	"sort"
	"strings"
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

func GetOrgOrProjectResourcesQuery(tx *db.Session, searchStr string, orgId, projectId, userId models.Id, isSuperAdmin bool) *db.Session {
	query := tx.Joins("inner join iac_env on iac_env.last_res_task_id = iac_resource.task_id left join " +
		"iac_project on iac_resource.project_id = iac_project.id").
		LazySelectAppend("iac_project.name as project_name, iac_env.name as env_name, iac_resource.id as resource_id," +
			"iac_resource.name as resource_name, iac_resource.task_id, iac_resource.project_id as project_id, iac_resource.attrs as attrs," +
			"iac_resource.env_id as env_id, iac_resource.provider as provider, iac_resource.type, iac_resource.module")

	if orgId != "" {
		query = query.Where("iac_env.org_id = ?", orgId)
	}

	if orgId != "" && projectId != "" {
		query = query.Where("iac_env.project_id = ?", projectId)
	}

	if searchStr != "" {
		query = query.Where("iac_resource.name like ? OR iac_resource.type like ? OR iac_resource.attrs like ?", fmt.Sprintf("%%%s%%", searchStr),
			fmt.Sprintf("%%%s%%", searchStr), fmt.Sprintf("%%%s%%", searchStr))
	}

	if !isSuperAdmin && !UserHasOrgRole(userId, orgId, consts.OrgRoleAdmin) {
		// 查一下当前用户属于哪些项目
		query = query.Joins("left join iac_user_project on iac_user_project.project_id = iac_resource.project_id").
			LazySelectAppend("iac_user_project.user_id")
		query = query.Where("iac_user_project.user_id = ?", userId)
	}

	return query
}

func GetProviderQuery(providers string, query *db.Session) *db.Session {
	if len(providers) != 0 {
		var tempSql []string
		var tempList []interface{}
		for _, v := range strings.Split(providers, ",") {
			tempSql = append(tempSql, "iac_resource.provider like ?")
			tempList = append(tempList, strings.Join([]string{"%/", v}, ""))
		}
		query = query.Where(strings.Join(tempSql, " OR "), tempList...)
	}
	return query
}

func GetOrgOrProjectResourcesResp(currentPage, pageSize int, query *db.Session) (*[]resps.OrgOrProjectResourcesResp, *page.Paginator, e.Error) {
	rs := make([]resps.OrgOrProjectResourcesResp, 0)
	p := page.New(currentPage, pageSize, query)
	if err := p.Scan(&rs); err != nil {
		return nil, nil, e.New(e.DBError, err)
	}
	for i := range rs {
		rs[i].Provider = path.Base(rs[i].Provider)
	}
	return &rs, p, nil
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
	where
		iac_project.status = 'enable'
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
	query = query.Where(`iac_project.status = 'enable'`)
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
		and iac_project.status = 'enable'
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
	query = query.Where(`iac_project.status = 'enable'`)

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

type ProjectOrEnvStatResult struct {
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
		and iac_project.status = 'enable'
	group by
		iac_resource.type,
		iac_resource.project_id
	limit 10;
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.project_id as id, iac_project.name as name, iac_resource.type as res_type, count(*) as count`)

	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Joins("JOIN iac_project ON iac_project.id = iac_resource.project_id")
	query = query.Where("iac_env.org_id = ?", orgId)
	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}
	query = query.Where(`iac_project.status = 'enable'`)

	query = query.Group("iac_resource.type,iac_resource.project_id")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbResults []ProjectOrEnvStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2ProjectOrEnvResStatResp(dbResults), nil
}

func dbResult2ProjectOrEnvResStatResp(dbResults []ProjectOrEnvStatResult) []resps.ProjOrEnvResStatResp {
	uniqDetails := make(map[models.Id]ProjectOrEnvStatResult)
	for _, result := range dbResults {
		uniqDetails[result.Id] = result
	}

	// resType -> projectId -> data
	mCount := make(map[string]int)
	m := make(map[string][]resps.DetailStatResp)

	// 统计每个资源类型下，每个项目包含此资源类型的数量
	for _, result := range dbResults {
		if _, ok := mCount[result.ResType]; !ok {
			mCount[result.ResType] = 0
		}
		mCount[result.ResType] += result.Count

		if _, ok := m[result.ResType]; !ok {
			m[result.ResType] = getFullDetails(uniqDetails)
		}
		setStatDetail(result, m)
	}

	var results = make([]resps.ProjOrEnvResStatResp, 0)
	for k, v := range m {
		results = append(results, resps.ProjOrEnvResStatResp{
			ResType: k,
			Count:   mCount[k],
			Details: v,
		})
	}

	return results
}

func getFullDetails(uniqDetails map[models.Id]ProjectOrEnvStatResult) []resps.DetailStatResp {
	fullDetails := make([]resps.DetailStatResp, 0)
	for k, v := range uniqDetails {
		fullDetails = append(fullDetails, resps.DetailStatResp{
			Id:    k,
			Name:  v.Name,
			Count: 0,
		})
	}

	return fullDetails
}

func setStatDetail(result ProjectOrEnvStatResult, m map[string][]resps.DetailStatResp) {
	for i, detail := range m[result.ResType] {
		if detail.Id != result.Id {
			continue
		}
		m[result.ResType][i].Count = result.Count
	}
}

func GetOrgResGrowTrend(tx *db.Session, orgId models.Id, projectIds []string, days int) ([]resps.ResGrowTrendResp, e.Error) {
	/* sample sql
	select
		iac_resource.project_id as id,
		iac_project.name as name,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date,
		count(DISTINCT iac_resource.env_id, iac_resource.address) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.id = iac_resource.env_id
	JOIN iac_project ON
		iac_project.id = iac_resource.project_id
	where
		iac_env.org_id = 'org-c8gg9fosm56injdlb85g'
		and iac_env.project_id in ('p-c8gg9josm56injdlb86g', 'aaa')
		and iac_project.status = 'enable'
		and DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY), "%Y-%m-%d")
	group by
		date,
		iac_resource.project_id
	order by
		date
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.project_id as id, iac_project.name as name, DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date, count(DISTINCT iac_resource.env_id, iac_resource.address) as count`)
	query = query.Joins(`join iac_env on iac_env.id = iac_resource.env_id`)
	query = query.Joins("JOIN iac_project ON iac_project.id = iac_resource.project_id")

	query = query.Where("iac_env.org_id = ?", orgId)
	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}
	query = query.Where(`iac_project.status = 'enable'`)

	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), "%Y-%m-%d")`, days)

	query = query.Group("date, iac_resource.project_id").Order("date")

	var dbResults []ProjectOrEnvStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -1*days)
	endDate := now

	return getResGrowTrendByDays(startDate, endDate, dbResults, days), nil
}

func getResGrowTrendByDays(startDate, endDate time.Time, dbResults []ProjectOrEnvStatResult, days int) []resps.ResGrowTrendResp {

	// date -> detail(project or env)
	var m = make(map[string][]ProjectOrEnvStatResult)
	var mDateCount = make(map[string]int)
	var mDetailCount = make(map[[2]string]int)

	for i := 0; i < days; i++ {
		startDate = startDate.AddDate(0, 0, 1)
		if startDate.Format("2006-01-02") > endDate.Format("2006-01-02") {
			break
		}
		m[startDate.Format("2006-01-02")] = make([]ProjectOrEnvStatResult, 0)
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

func dbResults2ResGrowTrendResp(m map[string][]ProjectOrEnvStatResult, mDateCount map[string]int) []resps.ResGrowTrendResp {

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
