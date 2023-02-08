// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
)

var TfListVersions []string
var m sync.RWMutex

type tfVersionList struct {
	tflist []string
}

func AutoGetTfVersion(c *ctx.ServiceContext, form *forms.TemplateTfVersionSearchForm) (interface{}, e.Error) {
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.DB())
	if err != nil {
		return nil, err
	}
	repo, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	repoDetail, er := repo.GetRepo(form.RepoId)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	content, er := repoDetail.ReadFileContent(form.VcsBranch, filepath.Join(form.Workdir, "versions.tf"))
	// 没有找到versions.tf 文件，使用默认版本，不报错
	if er != nil {
		return consts.DefaultTerraformVersion, nil
	}
	tfconstraint := GetUserTfVersion(content)
	// 如果用户versions.tf 中没有制定terraform 版本，使用我们默认版本
	if tfconstraint == "" {
		return consts.DefaultTerraformVersion, nil
	}
	// 查看内置版本中有无满足用户约束条件的版本
	tfVersion, tferr := GetDetailTfVersion(common.TerraformVersions, tfconstraint)
	if tferr != nil {
		return nil, e.New(e.InvalidTfVersion, tferr)
	}
	if tfVersion != "" {
		return tfVersion, nil
	} else {
		// 如果内置版本中没有满足用户版本，则从官方提供所有版本中查找
		tflist := getTfVersions()
		if len(tflist) > 0 {
			tfVersion, tferr = GetDetailTfVersion(tflist, tfconstraint)
			// 官方提供所有版本没有找到，则抛错认定用户指定版本不存在
			if tferr != nil || tfVersion == "" {
				return nil, e.New(e.InvalidTfVersion, tferr)
			}
			return tfVersion, nil
		}
	}

	return nil, e.New(e.VcsError, fmt.Errorf("Illegal terrain version number, please enter after verification"))
}

// tflist: 提供的terraform版本约束列表
// tfconstraint: 用户versions.tf中指定的版本约束范围
func GetDetailTfVersion(tflist []string, tfconstraint string) (string, error) {
	constrains, er := semver.NewConstraint(tfconstraint)
	if er != nil {
		return "", e.New(e.InvalidTfVersion, er)
	}
	versions := make([]*semver.Version, len(tflist))
	for i, tfvals := range tflist {
		version, err := semver.NewVersion(tfvals)
		if err != nil {
			return "", e.New(e.InvalidTfVersion, er)
		}

		versions[i] = version
	}
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	for _, element := range versions {
		if constrains.Check(element) { // Validate a version against a constraint
			tfversion := element.String()
			return tfversion, nil
		}
	}
	// tflist中没有满足用户的版本, 不报错以及返回空字符串
	return "", nil
}

func GetUserTfVersion(f []byte) string {
	lines := strings.Split(string(f), "\n")
	for _, v := range lines {
		lineInfo := strings.Contains(v, "required_version")
		if lineInfo {
			re1, _ := regexp.Compile(`".*"`)
			if re1 == nil {
				return ""
			}
			result := re1.FindAllStringSubmatch(v, -1)
			// 去除匹配到字符串双引号
			return strings.Trim(result[0][0], "\"")
		}
	}
	return ""
}

// 获取官方提供的terraform versions 列表
func GetTFURLBody(mirrorURL string) ([]string, error) {

	hasSlash := strings.HasSuffix(mirrorURL, "/")
	if !hasSlash { //if does not have slash - append slash
		mirrorURL = fmt.Sprintf("%s/", mirrorURL)
	}
	// 设置http 请求超时时间
	cli := http.Client{Timeout: consts.HttpClientTimeout * time.Second}
	resp, errURL := cli.Get(mirrorURL)
	if errURL != nil {
		return nil, errURL
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return []string{}, nil
	}
	body, errBody := ioutil.ReadAll(resp.Body)
	if errBody != nil {
		return nil, errBody
	}
	bodyString := string(body)
	result := strings.Split(bodyString, "\n")
	return result, nil
}

func GetTFList(mirrorURL string, preRelease bool) ([]string, error) {

	result, error := GetTFURLBody(mirrorURL)
	if error != nil {
		return nil, error
	}

	var tfVersionList tfVersionList
	var semver string
	if preRelease {
		// Getting versions from body; should return match /X.X.X-@/ where X is a number,@ is a word character between a-z or A-Z
		semver = `\/(\d+\.\d+\.\d+)(-[a-zA-z]+\d*)?\/`
	} else {
		// Getting versions from body; should return match /X.X.X/ where X is a number
		semver = `\/(\d+\.\d+\.\d+)\/`
	}
	r, _ := regexp.Compile(semver)
	for i := range result {
		if r.MatchString(result[i]) {
			str := r.FindString(result[i])
			trimstr := strings.Trim(str, "/") //remove "/" from /X.X.X/
			tfVersionList.tflist = append(tfVersionList.tflist, trimstr)
		}
	}

	if len(tfVersionList.tflist) == 0 {
		fmt.Printf("Cannot get list from mirror: %s\n", mirrorURL)
	}

	return tfVersionList.tflist, nil

}

func initTfversions() {
	// 添加写锁
	m.Lock()
	defer m.Unlock()
	for {
		tfversion, err := GetTFList(consts.DefaultTfMirror, true)
		if err == nil {
			TfListVersions = tfversion
			break
		} else {
			time.Sleep(1 * time.Second)

		}
	}
}

func InitTfVersions() {
	go func() {
		initTfversions()
		for {
			time.Sleep(86400 * 7 * time.Second)
			initTfversions()
		}
	}()

}

func getTfVersions() []string {
	// 添加读锁
	m.RLock()
	defer m.RUnlock()
	return TfListVersions
}
