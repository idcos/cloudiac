// Copyright 2021 CloudJ Company Limited. All rights reserved.

package consts

var IacUserInvitationsTpl = `
<html>
<body>
尊敬的 {{.Name}}：
<br>
<br>&nbsp;&nbsp;&nbsp;&nbsp;CloudIaC的管理员 【{{.Inviter}}】 邀请您体验CloudIaC平台并加入 【{{.Organization}}】组织，您可以通过以下方式登录CloudIaC平台：
<br>
<br>&nbsp;&nbsp;&nbsp;&nbsp;平台地址：<a href="{{.Addr}}">{{.Addr}}</a>
<br>&nbsp;&nbsp;&nbsp;&nbsp;帐户：{{.Email}}
{{if .IsNewUser}}
<br>&nbsp;&nbsp;&nbsp;&nbsp;初始密码：{{.InitPass}}
{{end}}
<br>
{{if .IsNewUser}}
<br>&nbsp;&nbsp;&nbsp;&nbsp;为保障您的帐户安全，请立即登录平台并及时修改初始密码，祝您使用愉快！
<br>
{{else}}
<br>&nbsp;&nbsp;&nbsp;&nbsp;请使用您的帐户登陆 CloudIaC 平台使用服务，祝您使用愉快！
<br>
{{end}}
<br>-----该邮件由系统自动发出，请勿回复-----
</body>
</html>
`

var IacTaskRunning = `
<html>
<body>
<p>尊敬的CloudIaC用户：</p>
<br />
<p>	【{{.Creator}}】在CloudIaC平台发起了部署任务，详情如下：</p> 
<br />	
<p>	所属组织：{{.OrgName}}</p>
<p>	所属项目：{{.ProjectName}}</p>
<p>	云模板：{{.TemplateName}}</p>
<p>	分支/tag：{{.Revision}}</p>
<p>	环境名称：{{.EnvName}}</p>
<br />	
<p>	更多详情请点击：{{.Addr}}</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`

var IacTaskCompleteTpl = `
<html>
<body>
<p>尊敬的CloudIaC用户：</p>
<br />
<p>	【{{.Creator}}】在CloudIaC平台发起的部署任务已执行完成，详情如下：</p> 
<br />	
<p>	所属组织：{{.OrgName}}</p>
<p>	所属项目：{{.ProjectName}}</p>
<p>	云模板：{{.TemplateName}}</p>
<p>	分支/tag：{{.Revision}}</p>
<p>	环境名称：{{.EnvName}}</p>
<p>	执行结果：成功</p>
<p>	资源数量：{{.ResAdded}}+ {{.ResChanged}}~ {{.ResDestroyed}}-</p>
<br />	
<p>	更多详情请点击：{{.Addr}}</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`

var IacTaskFailedTpl = `
<html>
<body>
<p>尊敬的CloudIaC用户：</p>
<br />
<p>	【{{.Creator}}】在CloudIaC平台发起的部署任务执行失败，详情如下：</p> 
<br />	
<p>	所属组织：{{.OrgName}}</p>
<p>	所属项目：{{.ProjectName}}</p>
<p>	云模板：{{.TemplateName}}</p>
<p>	分支/tag：{{.Revision}}</p>
<p>	环境名称：{{.EnvName}}</p>
<p>	执行结果：失败</p>
<p>	失败原因：{{.Message}}</p>
<br />	
<p>	更多详情请点击：{{.Addr}}</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`

var IacTaskApprovingTpl = `
<html>
<body>
<p>尊敬的CloudIaC用户：</p>
<br />
<p>	【{{.Creator}}】在CloudIaC平台发起的部署任务等待审批中，详情如下：</p> 
<br />	
<p>	所属组织：{{.OrgName}}</p>
<p>	所属项目：{{.ProjectName}}</p>
<p>	云模板：{{.TemplateName}}</p>
<p>	分支/tag：{{.Revision}}</p>
<p>	环境名称：{{.EnvName}}</p>
<p>	执行结果：审批中</p>
<p>	失败原因：{{.Message}}</p>
<br />	
<p>	更多详情请点击：{{.Addr}}</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`

const (
	IacTaskRunningMarkdown = `
尊敬的CloudIaC用户：

	【{{.Creator}}】在CloudIaC平台发起了部署任务，详情如下：

	所属组织：{{.OrgName}}

	所属项目：{{.ProjectName}}

	云模板：{{.TemplateName}}

	分支/tag：{{.Revision}}

	环境名称：{{.EnvName}}

	更多详情请点击：{{.Addr}}

	-----该消息由系统自动发出，请勿回复-----

`
	IacTaskApprovingMarkdown = `
尊敬的CloudIaC用户：

	【{{.Creator}}】在CloudIaC平台发起的部署任务等待审批中，详情如下：

	所属组织：{{.OrgName}}

	所属项目：{{.ProjectName}}

	云模板：{{.TemplateName}}

	分支/tag：{{.Revision}}

	环境名称：{{.EnvName}}

	执行结果：审批中

	失败原因：{{.Message}}

	更多详情请点击：{{.Addr}}

	-----该消息由系统自动发出，请勿回复-----
`
	IacTaskFailedMarkdown = `

尊敬的CloudIaC用户：

	【{{.Creator}}】在CloudIaC平台发起的部署任务执行失败，详情如下：

	所属组织：{{.OrgName}}

	所属项目：{{.ProjectName}}

	云模板：{{.TemplateName}}

	分支/tag：{{.Revision}}

	环境名称：{{.EnvName}}

	执行结果：失败

	失败原因：{{.Message}}

	更多详情请点击：{{.Addr}}

	-----该消息由系统自动发出，请勿回复-----

`
	IacTaskCompleteMarkdown = `
尊敬的CloudIaC用户：

	【{{.Creator}}】在CloudIaC平台发起的部署任务已执行完成，详情如下：

	所属组织：{{.OrgName}}

	所属项目：{{.ProjectName}}

	云模板：{{.TemplateName}}

	分支/tag：{{.Revision}}

	环境名称：{{.EnvName}}

	执行结果：成功

	资源数量：{{.ResAdded}}+ {{.ResChanged}}~ {{.ResDestroyed}}-

	更多详情请点击：{{.Addr}}

	-----该消息由系统自动发出，请勿回复-----
`
)
