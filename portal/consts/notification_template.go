package consts

var IacUserInvitationsTpl = `
<html>
<body>
<p>尊敬的 {{.Name}}：</p>
<br />
<p>	CloudIaC的管理员 【{{.Inviter}}】 邀请您体验CloudIaC平台并加入 【{{.Organization}}】组织，您可以通过以下方式登录CloudIaC平台：</p>
<br />
<p>	平台地址：<a href="{{.Addr}}">{{.Addr}}</a></p>
<p>	用户名：{{.Email}}</p>
{{if .IsNewUser}}
<p>	初始密码：{{.InitPass}}</p>
{{end}}
<br />	
{{if .IsNewUser}}
<p>	为保障您的帐户安全，请立即登录平台并及时修改初始密码，祝您使用愉快！</p>
{{else}}
<p>	请使用您的账号登陆 CloudIaC 平台使用服务，祝您使用愉快！</p>
{{end}}
<br />	
<p>	-----该邮件由系统自动发出，请勿回复-----</p>
</body>
</html>
`
