package policy

import "testing"

func Test_getGitUrl1(t *testing.T) {
	type args struct {
		repoAddr string
		token    string
		version  string
		subDir   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"template local",
			args{"http://10.0.2.135/repos/cloudiac/terraform-alicloud-disk.git", "", "master", ""},
			"http://10.0.2.135/repos/cloudiac/terraform-alicloud-disk.git?ref=master",
		},
		{
			"template",
			args{"http://gitlab.idcos.com/iacsample/cloudiac-example.git", "the_token", "master", ""},
			"http://token:the_token@gitlab.idcos.com/iacsample/cloudiac-example.git?ref=master",
		},
		{
			"env",
			args{"http://token:the_token@gitlab.idcos.com/iacsample/cloudiac-example.git", "", "master", "ansible"},
			"http://token:the_token@gitlab.idcos.com/iacsample/cloudiac-example.git//ansible?ref=master",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGitUrl(tt.args.repoAddr, tt.args.token, tt.args.version, tt.args.subDir); got != tt.want {
				t.Errorf("getGitUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
