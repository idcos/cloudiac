package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

type FileReplacer struct {
	filepath string
	rules    []ReplaceRule
	inRule   ReplaceRule
}

type ReplaceRule interface {
	IsBegin(token string) bool
	IsEnd(token string) bool
	Replace(orig []byte) []byte
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

func (r *BlockReplaceRule) Replace([]byte) []byte {
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
	return strings.HasPrefix(line, r.token)
}

func (r *LineStartReplaceRule) Replace([]byte) []byte {
	log.Printf("line start replace: %s", r.token)
	return []byte(r.content)
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
		if r.inRule == nil {
			for _, rule := range r.rules {
				// log.Println("rule", rule)
				if rule.IsBegin(trimLine) {
					// log.Println("is being", trimLine)
					r.inRule = rule
					break
				}
			}
		}

		if r.inRule != nil {
			// 处于匹配规则中，或者当前行匹配到了 beging
			// 当前行匹配了 beging，我们也需要同步检查是否匹配了 end
			if r.inRule.IsEnd(trimLine) {
				// log.Println("is end", trimLine)
				if _, err := buffer.Write(r.inRule.Replace(nil)); err != nil {
					return err
				}
				r.inRule = nil
			}
		} else {
			// 未找到匹配，直接写入当前行
			_, err := buffer.WriteString(line)
			if err != nil {
				return err
			}
		}
	}

	if err := os.WriteFile(r.filepath, buffer.Bytes(), os.ModeType); err != nil {
		return err
	}
	return nil
}

func main() {
	versionBytes, err := os.ReadFile("VERSION")
	if err != nil {
		log.Fatalln(err)
	}
	version := strings.TrimSpace(string(versionBytes))

	configs := []struct {
		file  string
		rules []ReplaceRule
	}{
		{
			"./docs/mkdocs/deploy/container.md",
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
			"./docs/mkdocs/deploy/host.md",
			[]ReplaceRule{
				&LineStartReplaceRule{"VERSION=v", fmt.Sprintf("VERSION=%s\n", version)},
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
