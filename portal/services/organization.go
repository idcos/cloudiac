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

func GetOrgProjectsdEnvStat(tx *db.Session, orgId models.Id, projectIds []string) ([]resps.EnvStatResp, e.Error) {
	subQuery := tx.Model(&models.Env{}).Select(`if(task_status = '', status, task_status) as status`)
	subQuery = subQuery.Where("archived = ?", 0).Where("org_id = ?", orgId)

	if len(projectIds) > 0 {
		subQuery = subQuery.Where("project_id in ?", projectIds)
	}

	query := tx.Table("(?) as t", subQuery.Expr()).Select(`status, count(*) as count`).Group("status")

	var results []resps.EnvStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}

func GetOrgProjectsResStat(tx *db.Session, orgId models.Id, projectIds []string, limit int) ([]resps.ResStatResp, e.Error) {
	query := tx.Model(&models.Resource{}).Select(`iac_resource.type as res_type, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id`)
	query = query.Where(`iac_env.org_id = ?`, orgId)

	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}

	query = query.Group("res_type").Order("count desc")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var results []resps.ResStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}

func GetOrgProjectStat(tx *db.Session, orgId models.Id, projectIds []string, limit int) ([]resps.ProjectStatResp, e.Error) {
	/* sample sql:
	select
		iac_resource.project_id as project_id,
		iac_project.name as project_name,
		iac_resource.type as res_type,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m") as date,
		count(*) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.last_res_task_id = iac_resource.task_id
	JOIN iac_project ON
		iac_project.id = iac_resource.project_id
	where
		iac_env.org_id = 'org-c8gg9fosm56injdlb85g'
		AND iac_env.project_id IN ('p-c8gg9josm56injdlb86g', 'aaa')
		AND (DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(CURDATE(), "%Y-%m")
			OR
		DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m"))
	group by
		iac_resource.type,
		iac_resource.project_id,
		date
	limit 10;
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.project_id as project_id, iac_project.name as project_name, iac_resource.type as es_type, DATE_FORMAT(iac_resource.applied_at, "%Y-%m") as date, count(*) as count`)

	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id`)
	query = query.Joins("JOIN iac_project ON iac_project.id = iac_resource.project_id")
	query = query.Where("iac_env.org_id = ?", orgId)
	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(CURDATE(), "%Y-%m") OR DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m")`)

	query = query.Group("iac_resource.type,iac_resource.project_id,date")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var results []resps.ProjectStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}

func GetOrgResGrowTrend(tx *db.Session, orgId models.Id, projectIds []string, days int) ([]resps.ResGrowTrendResp, e.Error) {
	/* sample sql
	select
		DATE_FORMAT(applied_at, "%Y-%m-%d") as date,
		count(*) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.last_res_task_id = iac_resource.task_id
	where
		iac_env.org_id = 'org-c8gg9fosm56injdlb85g'
		and iac_env.project_id in ('p-c8gg9josm56injdlb86g', 'aaa')
		and (
		applied_at > DATE_SUB(CURDATE(), INTERVAL 7 DAY)
			or (applied_at > DATE_SUB(DATE_SUB(CURDATE(), INTERVAL 7 DAY), INTERVAL 1 MONTH)
				and applied_at <= DATE_SUB(CURDATE(), INTERVAL 1 MONTH)))
	group by
		date
	order by
		date
	*/

	query := tx.Model(&models.Resource{}).Select(`DATE_FORMAT(applied_at, "%Y-%m-%d") as date, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id`)

	query = query.Where("iac_env.org_id = ?", orgId)
	if len(projectIds) > 0 {
		query = query.Where(`iac_env.project_id in ?`, projectIds)
	}

	query = query.Where(`applied_at > DATE_SUB(CURDATE(), INTERVAL 7 DAY) or (applied_at > DATE_SUB(DATE_SUB(CURDATE(), INTERVAL 7 DAY), INTERVAL 1 MONTH) and applied_at <= DATE_SUB(CURDATE(), INTERVAL 1 MONTH))`)

	query = query.Group("date").Order("date")

	var results []resps.ResGrowTrendResp
	if err := query.Debug().Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}
