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

var IacTaskStartTpl = `
<html>
<body>
<p>尊敬的 .username：</p>
<br />
<p>	【xxx】在CloudIaC平台发起了部署任务，详情如下：</p> ？？？？
<br />	
<p>	所属组织：zzz</p>
<p>	所属项目：project-1</p>
<p>	云模板：template-1</p>
<p>	分支/tag：release/0.5.0</p>
<p>	环境名称：测试环境</p>
<br />	
<p>	更多详情请点击：.addr</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`

var IacTaskEndTpl = `
<html>
<body>
<p>尊敬的 .username：</p>
<br />
<p>	【xxx】在CloudIaC平台发起的部署任务已执行完成，详情如下：</p> ？？？？
<br />	
<p>	所属组织：zzz</p>
<p>	所属项目：project-1</p>
<p>	云模板：template-1</p>
<p>	分支/tag：release/0.5.0</p>
<p>	环境名称：测试环境</p>
<p>	执行结果：已完成</p>
<p>	资源数量：7+ 0~ 0-</p>
<br />	
<p>	更多详情请点击：.addr</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`

var IacTaskFailedTpl = `
<html>
<body>
<p>尊敬的 .username：</p>
<br />
<p>	【xxx】在CloudIaC平台发起的部署任务执行失败，详情如下：</p> ？？？？
<br />	
<p>	所属组织：zzz</p>
<p>	所属项目：project-1</p>
<p>	云模板：template-1</p>
<p>	分支/tag：release/0.5.0</p>
<p>	环境名称：测试环境</p>
<p>	执行结果：已完成</p>
<p>	失败原因：审批被驳回</p>
<p>	资源数量：7+ 0~ 0-</p>
<br />	
<p>	更多详情请点击：.addr</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`

var IacTaskApprovingTpl = `
<html>
<body>
<p>尊敬的 .username：</p>
<br />
<p>	【xxx】在CloudIaC平台发起的部署任务执行失败，详情如下：</p> ？？？？
<br />	
<p>	所属组织：zzz</p>
<p>	所属项目：project-1</p>
<p>	云模板：template-1</p>
<p>	分支/tag：release/0.5.0</p>
<p>	环境名称：测试环境</p>
<p>	执行结果：已完成</p>
<p>	失败原因：审批被驳回</p>
<p>	资源数量：7+ 0~ 0-</p>
<br />	
<p>	更多详情请点击：.addr</p>
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`
