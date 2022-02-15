package policy

import (
	"bufio"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/go-playground/locales/en"
	translator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translation "github.com/go-playground/validator/v10/translations/en"
	"github.com/hashicorp/hcl"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/version"

	"github.com/pkg/errors"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()

	MSG_TEMPLATE_INVALID   = red("Error:\t") + "id: {{.RuleId}}, detail: {{.Error}}"
	MSG_TEMPLATE_ERROR     = red("Error: \t") + "group: {{.Category}}, name: {{.RuleName}}, id: {{.RuleId}}, severity: {{.Severity}}\ndetail: {{.Error}}"
	MSG_TEMPLATE_PASSED    = green("Passed: \t") + "group: {{.Category}}, name: {{.RuleName}}, id: {{.RuleId}}, severity: {{.Severity}}"
	MSG_TEMPLATE_VIOLATED  = red("Violated: \t") + "group: {{.Category}}, name: {{.RuleName}}, id: {{.RuleId}}, resource_id : {{.ResourceName}}, severity: {{.Severity}}"
	MSG_TEMPLATE_SUPRESSED = yellow("Suppressed: \t") + "group: {{.Category}}, name: {{.RuleName}}, id: {{.RuleId}}, severity: {{.Severity}}"
)

type Parser struct {
}

type Policy struct {
	Id   string `json:"Id"`
	Meta Meta   `json:"meta"`
	Rego string `json:"rego"`
}

type Meta struct {
	Category      string `json:"category"`                                           // 分组
	Root          string `json:"root" validate:"required"`                           // 根目录
	File          string `json:"file" validate:"required"`                           // 文件名
	Id            string `json:"id" validate:"required"`                             // 策略id
	Name          string `json:"name" validate:"required"`                           // 策略名称
	Label         string `json:"label"`                                              // 策略标签
	PolicyType    string `json:"policy_type" binding:"required"`                     // 策略类型
	ReferenceId   string `json:"reference_id"`                                       // 引用策略id
	ResourceType  string `json:"resource_type" binding:"required"`                   // 资源类型
	Severity      string `json:"severity" validate:"required,oneof=low medium high"` // 严重程度
	Version       int    `json:"version"`                                            // 策略版本
	FixSuggestion string `json:"fix_suggestion"`                                     // 修复建议
	Description   string `json:"description"`                                        // 描述
}

type Resource struct {
	ResourceType string `json:"resourceType" enums:"local,remote"`
	RepoAddr     string // repo 远程地址或者本地目录路径
	Token        string
	Revision     string
	SubDir       string

	InputFile string
	MapFile   string

	StopOnViolation bool
	codeDir         string
}

type OutputResult struct {
	Results Results `json:"results"`
}

type Results struct {
	PassedRules []Rule      `json:"passed_rules"`
	Violations  []Violation `json:"violations"`
	Count       TsCount     `json:"count"`
}

type Rule struct {
	RuleName    string `json:"rule_name"`
	Description string `json:"description"`
	RuleId      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
}

type Violation struct {
	RuleName     string `json:"rule_name"`
	Description  string `json:"description"`
	RuleId       string `json:"rule_id"`
	Severity     string `json:"severity"`
	Category     string `json:"category"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Comment      string `json:"skip_comment,omitempty"`
	ModuleName   string `json:"module_name,omitempty"`
	PlanRoot     string `json:"plan_root,omitempty"`
	Source       string `json:"source,omitempty"`
}

type TsCount struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
	Total  int `json:"total"`
}

type ScanError struct {
	IacType     string `json:"iac_type"`
	Directory   string `json:"directory"`
	ErrMsg      string `json:"errMsg"`
	RuleName    string `json:"rule_name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	RuleId      string `json:"rule_id"`
	File        string `json:"file"`
	Error       error  `json:"-"`
}

func UnmarshalOutputResult(bs []byte) (*OutputResult, error) {
	js := OutputResult{}
	err := json.Unmarshal(bs, &js)
	return &js, err
}

func (r Resource) GetUrl(task *models.Task) string {
	u := getGitUrl(task.RepoAddr, "", task.CommitId, task.Workdir)
	return u
}

func (p Parser) Parse(filePath string) error {
	//logrus.Errorf("parse \n")

	byt, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	//logrus.Errorf("byt %s\n", byt)

	a, err := hcl.ParseBytes(byt)
	if err != nil {
		return err
	}
	//logrus.Errorf("ast %s\n", a)

	js, err := json.Marshal(a)
	if err != nil {
		return err
	}
	fmt.Printf("%s", js)
	return nil
}

func PopulateViolateSource(scanner *Scanner, res Resource, task *models.ScanTask, resultJson *TsResultJson) (*TsResultJson, error) {
	updated := false
	tfmap := models.TfParse{}

	if res.MapFile != "" {
		tfmapContent, _ := ioutil.ReadFile(res.MapFile)
		if len(tfmapContent) > 0 {
			_ = json.Unmarshal(tfmapContent, &tfmap)
		}
	}
	for idx, policyResult := range resultJson.Results.Violations {
		resLineNo := policyResult.Line
		srcFile := policyResult.File

		if (resLineNo == 0 || srcFile == "") && tfmap != nil {
			resLineNo, srcFile = findLineNoFromMap(tfmap, policyResult.ResourceName)
			resultJson.Results.Violations[idx].Line = resLineNo
			resultJson.Results.Violations[idx].File = srcFile
		}

		srcFp, err := os.Open(filepath.Join(srcFile))

		if err != nil {
			// "open src fail
			continue
		}
		reader := bufio.NewReader(srcFp)
		srcLines := ""
		for lineNo := 1; ; lineNo++ {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			srcLines = srcLines + string(line)
			if lineNo < resLineNo {
				continue
			} else if lineNo >= resLineNo {
				resultJson.Results.Violations[idx].Source += string(line)

				// 查找源码结束符
				if strings.Contains(string(line), "}") {
					break
				}

				if lineNo-resLineNo > 100 {
					// 超长源码截断
					resultJson.Results.Violations[idx].Source += string(line) + "  //...\n}"
					break
				}
			}

			updated = true
		}

		_ = srcFp.Close()
	}

	if updated {
		if js, err := json.MarshalIndent(resultJson, "", "  "); err == nil {
			err := os.WriteFile(scanner.GetResultPath(res), js, 0644)
			if err != nil {
				return nil, err
			}
		}
	}

	return resultJson, nil
}

func findLineNoFromMap(tfmap models.TfParse, resourceName string) (int, string) {
	for _, resources := range tfmap {
		for _, resource := range resources {
			if resource.Id == resourceName {
				return resource.Line, resource.Source
			}
		}
	}
	return 0, ""
}

// genPolicyFiles 将策略文件写入策略目录
func genPolicyFiles(policyDir string, policies []Policy) error {
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		return err
	}

	for _, policy := range policies {
		if err := os.MkdirAll(filepath.Join(policyDir, policy.Id), 0755); err != nil {
			return err
		}
		js, _ := json.Marshal(policy.Meta)

		if err := os.WriteFile(filepath.Join(policyDir, policy.Id, policy.Meta.Name+".json"), js, 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(policyDir, policy.Id, policy.Meta.Name+".rego"), []byte(policy.Rego), 0644); err != nil {
			return err
		}
	}
	return nil
}

// getGitUrl 获取 go-getter git 下载地址
// version 可以是 branch、tag 名称 或者 commit hash
// 同时可以通过 subDir 指定只下载某个子目录
// example: git::http://token:the_token@git.example.com/git_sample/repo-example.git//dev?ref=v1.0.0
func getGitUrl(repoAddr, token, version, subDir string) string {
	u, err := url.Parse(repoAddr)
	if err != nil {
		return ""
	}

	// gitlab http token 认证
	if token != "" {
		u.User = url.UserPassword("token", token)
	}

	query := url.Values{}
	if version != "" {
		query.Add("ref", version)
	}
	u.RawQuery = query.Encode()

	// get-getter 子目录使用双斜杠
	if subDir != "" {
		if subDir[0:1] != "/" {
			subDir = "/" + subDir
		}
		subDir = "/" + subDir
		u.Path = u.Path + subDir
	}

	return u.String()
}

func NewScannerFromLocalDir(srcPath string, policyDir string, inputFile string, mapFile string) (*Scanner, error) {
	res := Resource{
		ResourceType: "local",
		RepoAddr:     srcPath,
		codeDir:      srcPath,
		InputFile:    inputFile,
		MapFile:      mapFile,
	}
	scanner, err := NewScanner([]Resource{res})
	scanner.PolicyDir = policyDir

	return scanner, err
}

func EngineScan(regoFile string, configFile string) (interface{}, error) {
	configString, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read configFile: %w", err)
	}

	var input interface{}
	err = json.Unmarshal(configString, &input)
	if err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// 初始化 rego 引擎
	ctx := context.Background()
	obj := ast.NewObject()
	env := ast.NewObject()

	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.NullTerm())
		} else if len(parts) > 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.StringTerm(parts[1]))
		}
	}

	obj.Insert(ast.StringTerm("env"), ast.NewTerm(env))
	obj.Insert(ast.StringTerm("version"), ast.StringTerm(version.Version))
	obj.Insert(ast.StringTerm("commit"), ast.StringTerm(version.Vcs))

	info := ast.NewTerm(obj)

	regoArgs := []func(*rego.Rego){
		rego.Input(input),
		rego.Query("data"),
		rego.Load([]string{regoFile}, nil),
		rego.Runtime(info),
	}

	// 执行规则检查
	r := rego.New(regoArgs...)
	resultSet, err := r.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("evaluating policy: %w", err)
	}

	// 获取结果
	var result interface{}
	if len(resultSet) > 0 && len(resultSet[0].Expressions) > 0 {
		result = resultSet[0].Expressions[0].Value
	}

	return result, nil
}

type Rego struct {
	filePath string
	content  string
	pkg      string
	rule     string
	query    string
	rules    []string

	compiler *ast.Compiler
}

func (r *Rego) LoadRego() (string, error) {
	content, err := ioutil.ReadFile(r.filePath)
	if err != nil {
		return "", nil
	}
	return string(content), nil
}

func (r *Rego) Compile() (*ast.Compiler, error) {
	compiler, err := ast.CompileModules(map[string]string{
		r.filePath: r.content,
	})
	return compiler, err
}

func (r *Rego) ParsePackage() (string, error) {
	return strings.TrimPrefix(r.compiler.Modules[r.filePath].Package.String(), "package "), nil
}

func (r *Rego) ParseRules() ([]string, error) {
	var rules []string
	for _, r := range r.compiler.Modules[r.filePath].Rules {
		rules = append(rules, r.Head.Name.String())
	}

	return rules, nil
}

func (r *Rego) ParseResource(result []interface{}) []string {
	resMap := make(map[string]bool)
	for _, v := range result {
		var resId string
		switch res := v.(type) {
		// terrascan 自定义结果 返回
		case map[string]interface{}:
			_, ok := res["Id"]
			if !ok {
				// custom return id not found
				continue
			}

			resId, ok = res["Id"].(string)
			if !ok {
				//invalid custom return id
				continue
			}
			resId = res["Id"].(string)
		case string:
			resId = res
		default:
			// violate id not found
			continue
		}
		// remove array index from id
		if strings.LastIndex(resId, "[") != -1 {
			resId = resId[:strings.LastIndex(resId, "[")]
		}
		resMap[resId] = true
	}
	var resources []string
	for k, _ := range resMap {
		resources = append(resources, k)
	}

	return resources
}

func (r *Rego) String() string {
	str := fmt.Sprintf("file: %s\n", r.filePath)
	str += fmt.Sprintf("package: %s\n", r.pkg)
	for i, rule := range r.rules {
		str += fmt.Sprintf("rule[%d]: %s\n", i, rule)
	}
	return str
}

type ScanResult struct {
	FileFolder string    `json:"file/folder"`
	IacType    string    `json:"iac_type"`
	ScannedAt  string    `json:"scanned_at"`
	Status     string    `json:"status"`
	Priority   string    `json:"priority"`
	Error      ScanError `json:"error"`
}

func RegoParse(regoFile string, inputFile string, ruleName ...string) ([]interface{}, error) {
	reg := Rego{
		filePath: regoFile,
	}
	var err error

	reg.content, err = reg.LoadRego()
	if err != nil {
		// error load rego file
		return nil, err
	}

	reg.compiler, err = reg.Compile()
	if err != nil {
		// error compiling rego
		return nil, err
	}

	reg.pkg, err = reg.ParsePackage()
	if err != nil {
		// error parse package
		return nil, err
	}

	reg.rules, err = reg.ParseRules()
	if err != nil {
		// error parse rules
		return nil, err
	}

	// 规则名称：
	// 1. 使用 meta 定义的 name
	// 2. 使用 文件名 对应的 rule name
	// 3. 使用 @rule 标记的规则
	// 4. 使用第一条 rule
	if len(ruleName) > 0 {
		reg.rule = ruleName[0]
	} else {
		found := false
		for _, r := range reg.rules {
			if r == utils.FileNameWithoutExt(reg.filePath) {
				reg.rule = r
				found = true
				break
			}
		}
		if !found {
			ruleReg := "(?m)\\s*#+\\s*@rule.*\\n\\s*([^\\s]*)\\s*{"
			regex := regexp.MustCompile(ruleReg)
			match := regex.FindStringSubmatch(reg.content)

			if len(match) == 2 {
				reg.rule = strings.TrimSpace(match[1])
			} else {
				reg.rule = reg.rules[0]
			}
		}
	}
	reg.query = fmt.Sprintf("data.%s.%s", reg.pkg, reg.rule)

	// 读取待执行的输入文件
	inputBuf, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("read configFile: %w", err)
	}

	var input interface{}
	err = json.Unmarshal(inputBuf, &input)
	if err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// 初始化 rego 引擎
	ctx := context.Background()
	obj := ast.NewObject()
	env := ast.NewObject()

	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.NullTerm())
		} else if len(parts) > 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.StringTerm(parts[1]))
		}
	}

	obj.Insert(ast.StringTerm("env"), ast.NewTerm(env))
	obj.Insert(ast.StringTerm("version"), ast.StringTerm(version.Version))
	obj.Insert(ast.StringTerm("commit"), ast.StringTerm(version.Vcs))

	info := ast.NewTerm(obj)

	regoArgs := []func(*rego.Rego){
		rego.Input(input),
		rego.Query(reg.query),
		rego.Load([]string{regoFile}, nil),
		rego.Runtime(info),
	}

	// 执行规则检查
	r := rego.New(regoArgs...)
	resultSet, err := r.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("evaluating policy: %w", err)
	}

	// 获取结果
	var result []interface{}
	if len(resultSet) > 0 && len(resultSet[0].Expressions) > 0 {
		result = resultSet[0].Expressions[0].Value.([]interface{})
	}

	return result, nil
}

type TsResult struct {
	ScanErrors        []ScanError `json:"scan_errors,omitempty"`
	PassedRules       []Rule      `json:"passed_rules,omitempty"`
	Violations        []Violation `json:"violations"`
	SuppressedRules   []Rule      `json:"suppressed_rules"`
	SkippedViolations []Violation `json:"skipped_violations"`
	ScanSummary       ScanSummary `json:"scan_summary"`
}

type ScanSummary struct {
	FileFolder         string `json:"file/folder"`
	IacType            string `json:"iac_type"`
	ScannedAt          string `json:"scanned_at"`
	PoliciesValidated  int    `json:"policies_validated"`
	ViolatedPolicies   int    `json:"violated_policies"`
	PoliciesSuppressed int    `json:"policies_suppressed"`
	PoliciesError      int    `json:"policies_error"`
	Low                int    `json:"low"`
	Medium             int    `json:"medium"`
	High               int    `json:"high"`
}

type TsResultJson struct {
	Results TsResult `json:"results"`
}

func UnmarshalTfResultJson(bs []byte) (*TsResultJson, error) {
	js := TsResultJson{}
	err := json.Unmarshal(bs, &js)
	return &js, err
}

type PolicyWithMeta struct {
	Id   string `json:"Id"`
	Meta Meta   `json:"meta"`
	Rego string `json:"rego"`
}

func ParsePolicyGroup(dirname string) ([]*PolicyWithMeta, e.Error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, e.New(e.InternalError, err, http.StatusInternalServerError)
	}

	otherFiles := append(files)
	var regoFiles []RegoFile
	// 遍历当前目录
	for _, f := range files {
		// 优先处理 json 的 meta 及对应的 rego 文件
		if filepath.Ext(f.Name()) == ".json" {
			regoFileName := utils.FileNameWithoutExt(f.Name()) + ".rego"
			regoFilePath := filepath.Join(dirname, regoFileName)
			if utils.FileExist(regoFilePath) {
				regoFiles = append(regoFiles, RegoFile{
					MetaFile: filepath.Join(dirname, f.Name()),
					RegoFile: filepath.Join(dirname, regoFileName),
				})
				// 将已经处理的 rego 排除
				for i, rf := range otherFiles {
					if rf.Name() == regoFileName {
						otherFiles[i] = otherFiles[len(otherFiles)-1]
						otherFiles = otherFiles[:len(otherFiles)-1]
						break
					}
				}
			}
		}
	}

	// 遍历其他没有 json meta 的 rego 文件
	for _, f := range otherFiles {
		if filepath.Ext(f.Name()) == ".rego" {
			regoFiles = append(regoFiles, RegoFile{
				RegoFile: filepath.Join(dirname, f.Name()),
			})
		}
	}

	// 解析 rego 元信息
	var policies []*PolicyWithMeta
	for _, r := range regoFiles {
		p, err := ParseMeta(r.RegoFile, r.MetaFile)
		if err != nil {
			regoPath, _ := filepath.Rel(dirname, r.RegoFile)
			if r.MetaFile != "" {
				metaPath, _ := filepath.Rel(dirname, r.MetaFile)
				return nil, e.New(e.BadRequest,
					errors.Wrapf(err, "parse policy(%s,%s) error: %v", metaPath, regoPath, err),
					http.StatusBadRequest)
			}

			return nil, e.New(e.BadRequest,
				errors.Wrapf(err, "parse policy (%s) error: %v", regoPath, err),
				http.StatusBadRequest)
		}
		policies = append(policies, p)
	}

	return policies, nil
}

type RegoFile struct {
	MetaFile string
	RegoFile string
}

//ParseMeta 解析 rego metadata，如果存在 file.json 则从 json 文件读取 metadata，否则通过头部注释读取 metadata
func ParseMeta(regoFilePath string, metaFilePath string) (*PolicyWithMeta, e.Error) {
	var meta Meta
	buf, er := os.ReadFile(regoFilePath)
	if er != nil {
		return nil, e.New(e.PolicyRegoInvalid, fmt.Errorf("read rego file: %v", er))
	}
	regoContent := string(buf)

	// 1. 如果存在 json metadata，则解析 json 文件
	if metaFilePath != "" {
		content, er := os.ReadFile(metaFilePath)
		if er != nil {
			return nil, e.New(e.PolicyMetaInvalid, fmt.Errorf("read meta file: %v", er))
		}
		er = json.Unmarshal(content, &meta)
		if er != nil {
			return nil, e.New(e.PolicyMetaInvalid, fmt.Errorf("unmarshal meta file: %v", er))
		}
		meta.File = filepath.Base(regoFilePath)
		meta.Root = filepath.Dir(regoFilePath)
	} else {
		// 2. 无 json metadata，通过头部注释解析信息
		//	## id 为策略在策略组中的唯一标识，由大小写英文字符、数字、"."、"_"、"-" 组成
		//	## 建议按`组织_云商_资源名称/分类_编号`的格式进行命名
		//	# @id: cloudiac_alicloud_security_p001
		//
		//	# @name: 策略名称A
		//	# @description: 这是策略的描述
		//
		//	## 策略类型，如 aws, k8s, github, alicloud, ...
		//	# @policy_type: alicloud
		//
		//	## 资源类型，如 aws_ami, k8s_pod, alicloud_ecs, ...
		//	# @resource_type: aliyun_ami
		//
		//	## 策略严重级别: 可选 HIGH/MEDIUM/LOW
		//	# @severity: HIGH
		//
		//	## 策略分类(或者叫标签)，多个分类使用逗号分隔
		//	# @label: cat1,cat2
		//
		//	## 策略修复建议（支持多行）
		//	# @fix_suggestion:
		//	Terraform 代码去掉`associate_public_ip_address`配置
		//	```
		//resource "aws_instance" "bar" {
		//  ...
		//- associate_public_ip_address = true
		//}
		//```
		//	# @fix_suggestion_end

		meta = Meta{
			Id:           ExtractStr("id", regoContent),
			File:         filepath.Base(regoFilePath),
			Root:         filepath.Dir(regoFilePath),
			Name:         utils.FileNameWithoutExt(regoFilePath),
			Description:  ExtractStr("description", regoContent),
			PolicyType:   ExtractStr("policy_type", regoContent),
			ResourceType: ExtractStr("resource_type", regoContent),
			Label:        ExtractStr("label", regoContent),
			Category:     ExtractStr("category", regoContent),
			ReferenceId:  ExtractStr("reference_id", regoContent),
			Severity:     ExtractStr("severity", regoContent),
		}
		ver := ExtractStr("version", regoContent)
		meta.Version, _ = strconv.Atoi(ver)
		if meta.ReferenceId == "" {
			meta.ReferenceId = ExtractStr("id", regoContent)
		}

		// 多行注释提取
		regex := regexp.MustCompile(`(?s)@fix_suggestion:\\s*(.*)\\s*#+\\s*@fix_suggestion_end`)
		match := regex.FindStringSubmatch(regoContent)
		if len(match) == 2 {
			meta.FixSuggestion = strings.TrimSpace(match[1])
		} else {
			// 单行注释提取
			meta.FixSuggestion = ExtractStr("fix_suggestion", regoContent)
		}
	}

	if meta.ResourceType == "" {
		return nil, e.New(e.PolicyRegoMissingComment, fmt.Errorf("missing resource type info"))
	}
	if meta.PolicyType == "" {
		// alicloud_instance => alicloud
		meta.PolicyType = meta.ResourceType[:strings.Index(meta.ResourceType, "_")]
	}
	if meta.Severity == "" {
		meta.Severity = consts.PolicySeverityMedium
	}

	meta.Severity = strings.ToLower(meta.Severity)

	uni := translator.New(en.New())
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()
	if err := en_translation.RegisterDefaultTranslations(validate, trans); err != nil {
		return nil, e.New(e.InternalError, fmt.Errorf("register validator translator en error: %v", err))
	}
	if err := validate.Struct(meta); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return nil, e.New(e.PolicyMetaInvalid, fmt.Errorf("invalid policy meta: %+v", fmt.Errorf(err.Translate(trans))))
		}
		return nil, e.New(e.PolicyMetaInvalid, fmt.Errorf("invalid policy meta: %+v", err))
	}

	return &PolicyWithMeta{
		Id:   meta.Id,
		Meta: meta,
		Rego: regoContent,
	}, nil
}

// ExtractStr 提取 # @keyword: xxx 格式字符串
func ExtractStr(keyword string, input string) string {
	regex := regexp.MustCompile(fmt.Sprintf("(?m)^\\s*#+\\s*@%s:\\s*(.*)$", keyword))
	match := regex.FindStringSubmatch(input)
	if len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}
