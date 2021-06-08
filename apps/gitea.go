package apps

//type SearchRepoResponse struct {
//	Repos []*Repository `json:"data"`
//}
//type Repository struct {
//	ID            int64     `json:"id"`
//	Description   string    `json:"description"`
//	DefaultBranch string    `json:"default_branch"`
//	SSHURL        string    `json:"ssh_url"`
//	CloneURL      string    `json:"clone_url"`
//	Name          string    `json:"name"`
//	Updated       time.Time `json:"updated_at"`
//}
//
//func GetGiteaReadme(vcs *models.Vcs, form *forms.GetReadmeForm) (interface{}, e.Error) {
//	vcsIface, err := vcsrv.GetVcsInstance(vcs)
//	if err != nil {
//		return nil, e.New(e.GitLabError, err)
//	}
//	repoIface, err := vcsIface.GetRepo(vcsrv.VcsIfaceOptions{IdOrPath: strconv.Itoa(form.RepoId)})
//	if err != nil {
//		return nil, e.New(e.GitLabError, err)
//	}
//	b, err := repoIface.ReadFileContent(vcsrv.VcsIfaceOptions{})
//	if err != nil {
//		return nil, e.New(e.GitLabError, err)
//	}
//	res := models.FileContent{
//		Content: string(b[:]),
//	}
//	return res, nil
//}
//
//func ListGiteaRepoBranches(vcs *models.Vcs, form *forms.GetGitBranchesForm) ([]*Branches, e.Error) {
//
//	repo, err := GetGiteaRepoById(vcs, form.RepoId)
//	if err != nil {
//		return nil, err
//	}
//	path := vcs.Address + "/api/v1" + fmt.Sprintf("/repos/%s/branches?limit=0&page=0", repo)
//	request, er := http.NewRequest("GET", path, nil)
//	if er != nil {
//		return nil, e.New(e.BadRequest, er)
//	}
//	response, er := vcsrv.DoGiteaRequest(request, vcs.VcsToken)
//	defer response.Body.Close()
//	body, _ := ioutil.ReadAll(response.Body)
//	rep := []map[string]interface{}{}
//	json.Unmarshal(body, &rep)
//	branchList := []*Branches{}
//	for _, v := range rep {
//		branch := &Branches{Name: v["name"].(string)}
//		branchList = append(branchList, branch)
//	}
//	return branchList, nil
//
//}
//
//func GetGiteaRepoById(vcs *models.Vcs, repoId int) (string, e.Error) {
//
//	path := vcs.Address + fmt.Sprintf("/api/v1/repositories/%d", repoId)
//	request, err := http.NewRequest("GET", path, nil)
//	if err != nil {
//		return "", e.New(e.BadRequest, err)
//	}
//	response, err := vcsrv.DoGiteaRequest(request, vcs.VcsToken)
//	if response == nil || err != nil {
//		return "", e.New(e.BadRequest, err)
//	}
//	defer response.Body.Close()
//	repo := map[string]interface{}{}
//	body, _ := ioutil.ReadAll(response.Body)
//	json.Unmarshal(body, &repo)
//
//	if d, ok := repo["full_name"].(string); ok {
//		return d, nil
//	}
//	return "", nil
//
//}
//
//func ListGiteaOrganizationRepos(vcs *models.Vcs, form *forms.GetGitProjectsForm) (interface{}, e.Error) {
//
//	link, _ := url.Parse("/repos/search")
//	link.RawQuery = fmt.Sprintf("page=%d&limit=%d", form.CurrentPage(), form.PageSize())
//	if form.Q != "" {
//		link.RawQuery = link.RawQuery + fmt.Sprintf("&q=%s", form.Q)
//	}
//	path := vcs.Address + "/api/v1" + link.String()
//	request, err := http.NewRequest("GET", path, nil)
//	if err != nil {
//		return nil, e.New(e.BadRequest, err)
//	}
//	response, err := vcsrv.DoGiteaRequest(request, vcs.VcsToken)
//	if response == nil || err != nil {
//		return "", e.New(e.BadRequest, err)
//	}
//	defer response.Body.Close()
//	var total int64
//	if len(response.Header["X-Total-Count"]) != 0 {
//		total, _ = strconv.ParseInt(response.Header["X-Total-Count"][0], 10, 64)
//	}
//
//	body, _ := ioutil.ReadAll(response.Body)
//	rep := SearchRepoResponse{}
//	json.Unmarshal(body, &rep)
//	projectList := []*Projects{}
//	for _, v := range rep.Repos {
//		project := &Projects{
//			ID:             int(v.ID),
//			Description:    v.Description,
//			DefaultBranch:  v.DefaultBranch,
//			SSHURLToRepo:   v.SSHURL,
//			HTTPURLToRepo:  v.CloneURL,
//			Name:           v.Name,
//			LastActivityAt: &v.Updated,
//		}
//		projectList = append(projectList, project)
//	}
//
//	return page.PageResp{
//		Total:    total,
//		PageSize: form.CurrentPage(),
//		List:     projectList,
//	}, nil
//
//}
