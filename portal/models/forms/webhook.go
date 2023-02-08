// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

type WebhooksApiHandler struct {
	BaseForm
	VcsType          string           `uri:"vcsType" binding:"required,oneof=gitlab github gitea gitee" swaggerignore:"true"` //url参数
	VcsId            string           `uri:"vcsId" binding:"required,max=32" swaggerignore:"true"`                            //url参数
	ObjectKind       string           `json:"object_kind"`                                                                    // gitlab事件对象类型（push/merge_request）
	Ref              string           `json:"ref"`                                                                            // push分支
	UserId           uint             `json:"user_id"`                                                                        // 用户id
	UserName         string           `json:"user_name"`                                                                      // 用户名称
	Project          Project          `json:"project"`                                                                        // 项目信息
	ObjectAttributes ObjectAttributes `json:"object_attributes"`                                                              // 返回值信息
	User             User             `json:"user"`                                                                           // 用户信息
	PullRequest      PullRequest      `json:"pull_request"`                                                                   //gitea
	Action           string           `json:"action"`                                                                         // gitea pr状态，示例：open
	Before           string           `json:"before"`                                                                         //gitea push时回调的commitid
	After            string           `json:"after"`                                                                          //gitea push时回调的commitid
	Repository       Repository       `json:"repository"`                                                                     //gitea pr回调仓库信息
}

type Project struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type ObjectAttributes struct {
	SourceBranch string `json:"source_branch"`   // 源分支
	TargetBranch string `json:"target_branch"`   // 目标分支
	State        string `json:"state"`           // mr/pr动作(open、close)
	Iid          int    `json:"iid" form:"iid" ` // prId
}

type User struct {
	Name string `json:"name"`
}

//PullRequest gitea
type PullRequest struct {
	Id     int  `json:"id"`
	Base   Base `json:"base"`
	Head   Head `json:"head"`
	Number int  `json:"number"`
}

//Base gitea
type Base struct {
	Ref string `json:"ref"`
}

//Head gitea
type Head struct {
	Ref string `json:"ref"`
}

type Repository struct {
	Id       int    `json:"id"`
	FullName string `json:"full_name"`
}
