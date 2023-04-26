// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

const (
	version    = "v1.3.6"
	webVersion = "v1.3.6"
)

type FileReplacer struct {
	filepath string
	rules    []ReplaceRule
	inRule   ReplaceRule
}

type ReplaceRule interface {
	IsBegin(token string) bool
	IsEnd(token string) bool
	Replace(src string) []byte
}

type BlockReplaceRule struct {
	begin   string
	end     string
	content []byte
}

func (r *BlockReplaceRule) IsBegin(token string) bool {
	return r.begin == token
}

func (r *BlockReplaceRule) IsEnd(token string) bool {
	return r.end == token
}

func (r *BlockReplaceRule) Replace(string) []byte {
	log.Printf("block replace: %s", r.begin)
	rs := make([]byte, 0, len(r.content))
	rs = append(rs, []byte(r.begin)...)
	rs = append(rs, '\n')
	rs = append(rs, r.content...)
	rs = append(rs, []byte(r.end)...)
	rs = append(rs, '\n')
	return rs
}

func NewBlockReplaceRuleFromFile(begin string, end string, filepath string) *BlockReplaceRule {
	content, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	return &BlockReplaceRule{
		begin:   begin,
		end:     end,
		content: content,
	}
}

type LineStartReplaceRule struct {
	token   string
	content string
}

func (r *LineStartReplaceRule) IsBegin(line string) bool {
	return strings.HasPrefix(line, r.token)
}

func (r *LineStartReplaceRule) IsEnd(line string) bool {
	return true
}

func (r *LineStartReplaceRule) Replace(line string) []byte {
	log.Printf("line start replace: %s", line)
	return []byte(r.content)
}

type LineRegexReplaceRule struct {
	expr string
	repl string

	re *regexp.Regexp
}

func (r *LineRegexReplaceRule) IsBegin(line string) bool {
	if r.re == nil {
		r.re = regexp.MustCompile(r.expr)
	}
	return r.re.MatchString(line)
}

func (r *LineRegexReplaceRule) IsEnd(line string) bool {
	return true
}

func (r *LineRegexReplaceRule) Replace(line string) []byte {
	log.Printf("line regex replace: %s", line)
	return r.re.ReplaceAll([]byte(line), []byte(r.repl))
}

func NewReplace(filepath string) *FileReplacer {
	return &FileReplacer{
		filepath: filepath,
		rules:    nil,
	}
}

func (r *FileReplacer) AddRule(rule ReplaceRule) *FileReplacer {
	r.rules = append(r.rules, rule)
	return r
}

func matchInRuleBegin(inRule ReplaceRule, rules []ReplaceRule, trimLine string) ReplaceRule {
	if inRule != nil {
		return inRule
	}

	for _, rule := range rules {
		// log.Println("rule", rule)
		if rule.IsBegin(trimLine) {
			// log.Println("is being", trimLine)
			return rule
		}
	}

	return nil
}

func getRuleLineBytes(inRule ReplaceRule, line, trimLine string) []byte {
	if inRule == nil {
		return []byte(line)
	}

	if !inRule.IsEnd(trimLine) {
		return nil
	}

	return inRule.Replace(line)
}

func (r *FileReplacer) Run() error {
	log.Printf("run file replace: %s", r.filepath)
	content, err := os.ReadFile(r.filepath)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(nil)
	scanner := bufio.NewScanner(bytes.NewBuffer(content))
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		trimLine := strings.TrimRightFunc(line, unicode.IsSpace)
		r.inRule = matchInRuleBegin(r.inRule, r.rules, trimLine)

		lineBytes := getRuleLineBytes(r.inRule, line, trimLine)
		if lineBytes == nil {
			continue
		}

		if _, err := buffer.Write(lineBytes); err != nil {
			return err
		}
		r.inRule = nil
	}

	return os.WriteFile(r.filepath, buffer.Bytes(), os.ModeType)
}

type changeLog struct {
	Version string
	Date    string
	Content map[string][]string
}

const releaseNoteTemplate = `# Releases

{{ range . -}}
------
## v{{.Version}} {{.Date}}
{{ range $type, $notes := .Content -}}
**{{$type}}**

{{ range $notes -}}
- {{ . }}
{{ end }}
{{ end }}

**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v{{.Version}}](https://github.com/idcos/cloudiac/releases/tag/v{{.Version}})


{{end}}
`

func generateReleaseNotes() error {
	changelogFile := "changelog.json"
	releaseNoteFile := "docs/mkdocs/releases.md"

	rp, err := os.Open(changelogFile)
	if err != nil {
		return err
	}
	defer rp.Close()

	changeLogs := make([]changeLog, 0)
	if err := json.NewDecoder(rp).Decode(&changeLogs); err != nil {
		return err
	}

	wp, err := os.OpenFile(releaseNoteFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer wp.Close()

	return template.Must(template.New("").Parse(releaseNoteTemplate)).Execute(wp, changeLogs)
}

func main() {
	if version == "" || webVersion == "" {
		log.Fatalln("version or webVersion is blank")
	}

	if err := generateReleaseNotes(); err != nil {
		log.Fatalln(err)
	}

	configs := []struct {
		file  string
		rules []ReplaceRule
	}{
		{
			"./docs/mkdocs/product-deploy/container.md",
			[]ReplaceRule{
				NewBlockReplaceRuleFromFile(
					"# auto-replace-from: docker/docker-compose.yml",
					"```",
					"./docker/docker-compose.yml",
				),
				NewBlockReplaceRuleFromFile(
					"# auto-replace-from: configs/dotenv.sample",
					"```",
					"./configs/dotenv.sample",
				),
			},
		},
		{
			"./docs/mkdocs/product-deploy/container.md",
			[]ReplaceRule{
				&LineRegexReplaceRule{expr: `image: "([^/]*cloudiac/iac-portal):latest"`, repl: fmt.Sprintf(`image: "$1:%s"`, version)},
				&LineRegexReplaceRule{expr: `image: "([^/]*cloudiac/ct-runner):latest"`, repl: fmt.Sprintf(`image: "$1:%s"`, version)},
				&LineRegexReplaceRule{expr: `image: "([^/]*cloudiac/iac-web):latest"`, repl: fmt.Sprintf(`image: "$1:%s"`, webVersion)},
			},
		},
		{
			"./docs/mkdocs/product-deploy/host.md",
			[]ReplaceRule{
				&LineStartReplaceRule{"VERSION=v", fmt.Sprintf("VERSION=%s\n", version)},
				&LineStartReplaceRule{"WEB_VERSION=v", fmt.Sprintf("WEB_VERSION=%s\n", webVersion)},
			},
		},
	}

	for _, c := range configs {
		r := NewReplace(c.file)
		for _, rule := range c.rules {
			r.AddRule(rule)
		}
		if err := r.Run(); err != nil {
			log.Fatalln(err)
		}
	}
}
