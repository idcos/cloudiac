package models

type Stack struct {
	ID               string `json:"id"`
	Hostname         string `json:"hostname"`
	Name             string `json:"name"`
	Title            string `json:"title"`
	Namespace        string `json:"namespace"`
	Description      string `json:"description"`
	StackKey         string `json:"stackKey"`
	SourceType       string `json:"sourceType"`
	ExternalSource   string `json:"externalSource"`
	ExchangeRepoPath string `json:"exchangeRepoPath"`
	ExchangeRepoAddr string `json:"exchangeRepoAddr"`
}

type StackMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	Workdir      string `yaml:"workdir"`
	TfVarsFile   string `yaml:"tfVarsFile"`
	Playbook     string `yaml:"playbook"`
	PlayVarsFile string `yaml:"playVarsFile"`

	TfVersion string `yaml:"tfVersion"`
}
