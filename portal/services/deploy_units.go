// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"bytes"
	"cloudiac/portal/models/forms"
	"fmt"
	"strings"
	"text/template"
)

var moduleTpl = `{{ range $key, $value := .Units }}
module "{{ $key }}" {
  source = "{{ $value.Module }}" {{ range $index, $vars := $value.Vars }} {{ range $var_name, $var_value := $vars }}
  {{ $var_name }} = {{ if hasPrefix $var_value }}  {{ replace $var_value }} {{ else }} "{{ $var_value }}" {{ end }} {{ end }} {{ end }}
}
{{ end }}`

var outputTpl = `{{ range $key, $value := .Outputs }}
output "{{ $key }}" {
  value = {{ replace $value }} 
}
{{ end }}
`

func Yaml2Hcl(units forms.DeployForm) string {
	replace := func(value string) (string, error) {
		if strings.HasPrefix(value, "${") {
			return fmt.Sprintf("models.%s.this_private_ip[0]",
				strings.Replace(strings.Replace(value, "${", "", -1), "}", "", -1)), nil
		}
		return value, nil
	}

	hasPrefix := func(value string) (bool, error) {
		if strings.HasPrefix(value, "${") || strings.HasPrefix(value, "["){
			return true, nil
		}
		return false, nil
	}

	var (
		moduleMsg bytes.Buffer
		outputMsg bytes.Buffer
	)

	_ = template.Must(template.New("").Funcs(
		template.FuncMap{
			"replace":   replace,
			"hasPrefix": hasPrefix,
		}).Parse(moduleTpl)).Execute(&moduleMsg, units)

	_ = template.Must(template.New("").Funcs(
		template.FuncMap{
			"replace": replace,
		}).Parse(outputTpl)).Execute(&outputMsg, units)

	return fmt.Sprintf("%s%s", moduleMsg.String(), outputMsg.String())
}
