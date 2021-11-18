package consts


var PrCommentTpl = `
🤖&nbsp;&nbsp;PR Plan for CloudIac environment <a href="{{.Addr}}">{{.Name}}</a><br>
`+"```Plan {{.Status}}```"+`
<details>
<summary>Plan Details</summary>
<pre><code>
{{.Content}}
</code></pre>
</details>
`
