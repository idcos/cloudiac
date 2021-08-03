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
