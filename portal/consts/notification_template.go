package consts

var IacUserInvitationsTpl = `
<html>
<body>
<span>尊敬的 {{.Name}}：</span>
<br />
<br />
<span>&nbsp;&nbsp;&nbsp;&nbsp;CloudIaC的管理员 【{{.Inviter}}】 邀请您体验CloudIaC平台并加入 【{{.Organization}}】组织，您可以通过以下方式登录CloudIaC平台：</span>
<br />
<br />
<span>&nbsp;&nbsp;&nbsp;&nbsp;平台地址：<a href="{{.Addr}}">{{.Addr}}</a></span>
<br />
<span>&nbsp;&nbsp;&nbsp;&nbsp;帐户：{{.Email}}</span>
<br />
{{if .IsNewUser}}
<span>&nbsp;&nbsp;&nbsp;&nbsp;初始密码：{{.InitPass}}</span>
<br />
{{end}}
{{if .IsNewUser}}
<br />
<span>&nbsp;&nbsp;&nbsp;&nbsp;为保障您的帐户安全，请立即登录平台并及时修改初始密码，祝您使用愉快！</span>
<br />
{{else}}
<br />
<span>&nbsp;&nbsp;&nbsp;&nbsp;请使用您的帐户登陆 CloudIaC 平台使用服务，祝您使用愉快！</span>
<br />
{{end}}
<br />
<span>	-----该邮件由系统自动发出，请勿回复-----</span>
</body>
</html>
`
