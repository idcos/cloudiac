package services

import (
	"testing"
	"encoding/json"

	"github.com/stretchr/testify/assert"

	"cloudiac/portal/models"
)

func TestName(t *testing.T) {
	want := map[string]models.Variable{}
	t.Log(want)
	var rels []VarGroupRel
	t.Log(rels)
}

func TestGetVariableGroupVar(t *testing.T) {
	type args struct {
		vgs  []VarGroupRel
		vars map[string]models.Variable
	}
	tests := []struct {
		name string
		args args
		want map[string]models.Variable
	}{
		// TODO: Add test cases.
		{
			name: "组织引用资源账号，环境内定义相同key变量",
			args: args{
				//变量组变量
				vgs: []VarGroupRel{
					{
						VariableGroupRel: models.VariableGroupRel{
							VarGroupId: "vg-c90mgoasahsk60ulahi0",
							ObjectType: "org",
							ObjectId:   "org-c90hl2qsahsj54g0s8ug",
						},
						VariableGroup: models.VariableGroup{
							models.TimedModel{},
							"测试",
							"environment",
							"u-c90hkoqsahsj54g0s8qg",
							"org-c90hl2qsahsj54g0s8ug", models.VarGroupVariables{
								models.VarGroupVariable{
									Id:          "44ac4d00-881c-4ab3-a3e1-284a4a2998ca",
									Name:        "TF_VAR_ak",
									Value:       "资源账号变量",
									Sensitive:   false,
									Description: "",
								},
							},
						},
						Overwrites: []VarGroupRel{},
					},
				},
				//普通变量
				vars: map[string]models.Variable{
					"TF_VAR_akenvironment": {
						BaseModel: models.BaseModel{
							AbstractModel: models.AbstractModel{},
							Id:            "var-c92290qsahsoh3q8gmk0", //variables_id
						},
						VariableBody: models.VariableBody{
							Scope:       "env",
							Type:        "environment",
							Name:        "TF_VAR_ak",
							Value:       "环境变量",
							Sensitive:   false,
							Description: "",
							Options:     nil,
						},
						OrgId:     "org-c902qsahsj54g0s8ug",
						ProjectId: "p-c90hl7qsahsj54g0s90g",
						TplId:     "tpl-c9183bisahsld18h0hog",
						EnvId:     "env-c9183hasahsld18h0hpg",
					},
				},
			},
			want: map[string]models.Variable{
				"TF_VAR_akenvironment": {
					BaseModel: models.BaseModel{
						models.AbstractModel{},
						"var-c92290qsahsoh3q8gmk0",
					},
					VariableBody: models.VariableBody{
						Scope:       "env",
						Type:        "environment",
						Name:        "TF_VAR_ak",
						Value:       "环境变量",
						Sensitive:   false,
						Description: "",
						Options:     nil,
					}, OrgId: "org-c902qsahsj54g0s8ug",
					ProjectId: "p-c90hl7qsahsj54g0s90g",
					TplId:     "tpl-c9183bisahsld18h0hog",
					EnvId:     "env-c9183hasahsld18h0hpg",
				},
			},
		},
		{
			name: "组织引用资源账号，组织内定义相同key变量",
			args: args{
				//变量组变量
				vgs: []VarGroupRel{
					{
						VariableGroupRel: models.VariableGroupRel{
							VarGroupId: "vg-c90mgoasahsk60ulahi0",
							ObjectType: "project",
							ObjectId:   "p-c90hl7qsahsj54g0s90g",
						},
						VariableGroup: models.VariableGroup{
							models.TimedModel{},
							"测试",
							"environment",
							"u-c90hkoqsahsj54g0s8qg",
							"org-c90hl2qsahsj54g0s8ug", models.VarGroupVariables{
								models.VarGroupVariable{
									Id:          "44ac4d00-881c-4ab3-a3e1-284a4a2998ca",
									Name:        "TF_VAR_ak",
									Value:       "资源账号变量",
									Sensitive:   false,
									Description: "",
								},
							},
						},
						Overwrites: []VarGroupRel{},
					},
				},
				//普通变量
				vars: map[string]models.Variable{
					"TF_VAR_akenvironment": {
						BaseModel: models.BaseModel{
							AbstractModel: models.AbstractModel{},
							Id:            "p-c90hl7qsahsj54g0s90g", //variables_id
						},
						VariableBody: models.VariableBody{
							"project",
							"environment",
							"TF_VAR_ak",
							"项目变量",
							false,
							"",
							nil,
						},
						OrgId:     "org-c902qsahsj54g0s8ug",
						ProjectId: "p-c90hl7qsahsj54g0s90g",
						TplId:     "tpl-c9183bisahsld18h0hog",
						EnvId:     "env-c9183hasahsld18h0hpg",
					},
				},
			},
			want: map[string]models.Variable{
				"TF_VAR_akenvironment": {
					models.BaseModel{
						models.AbstractModel{},
						"p-c90hl7qsahsj54g0s90g",
					},
					models.VariableBody{
						"project",
						"environment",
						"TF_VAR_ak",
						"项目变量",
						false,
						"",
						nil,
					}, "org-c902qsahsj54g0s8ug",
					"p-c90hl7qsahsj54g0s90g",
					"tpl-c9183bisahsld18h0hog",
					"env-c9183hasahsld18h0hpg",
				},
			},
		},
		{
			name: "环境引用资源账号，组织内定义相同key变量",
			args: args{
				//变量组变量
				vgs: []VarGroupRel{
					{
						VariableGroupRel: models.VariableGroupRel{
							VarGroupId: "vg-c90mgoasahsk60ulahi0",
							ObjectType: "env",
							ObjectId:   "env-c92tftisahsqn2avare0",
						},
						VariableGroup: models.VariableGroup{
							TimedModel: models.TimedModel{
								BaseModel: models.BaseModel{
									Id: "vg-c90mgoasahsk60ulahi0",
								},
							},
							Name:      "测试",
							Type:      "environment",
							CreatorId: "u-c90hkoqsahsj54g0s8qg",
							OrgId:     "org-c90hl2qsahsj54g0s8ug",
							Variables: models.VarGroupVariables{
								models.VarGroupVariable{
									Id:          "44ac4d00-881c-4ab3-a3e1-284a4a2998ca",
									Name:        "TF_VAR_ak",
									Value:       "资源账号变量",
									Sensitive:   false,
									Description: "测试",
								},
							},
						},
						Overwrites: []VarGroupRel{},
					},
				},
				//普通变量
				vars: map[string]models.Variable{
					"TF_VAR_akenvironment": {
						BaseModel: models.BaseModel{
							AbstractModel: models.AbstractModel{},
							Id:            "var-c92t94asahsqk2defmn0", //variables_id
						},
						VariableBody: models.VariableBody{
							"org",
							"environment",
							"TF_VAR_ak",
							"组织变量设定",
							false,
							"",
							nil,
						},
						OrgId: "org-2qsahsj54g0s8ug",
					},
				},
			},
			want: map[string]models.Variable{
				"TF_VAR_akenvironment": {
					BaseModel: models.BaseModel{},
					VariableBody: models.VariableBody{
						"env",
						"environment",
						"TF_VAR_ak",
						"资源账号变量",
						false,
						"测试",
						nil,
					},
				},
			},
		},
		{
			name: "环境引用资源账号，项目内定义相同key变量",
			args: args{
				//变量组变量
				vgs: []VarGroupRel{
					{
						VariableGroupRel: models.VariableGroupRel{
							VarGroupId: "vg-c90mgoasahsk60ulahi0",
							ObjectType: "env",
							ObjectId:   "env-c92tftisahsqn2avare0",
						},
						VariableGroup: models.VariableGroup{
							TimedModel: models.TimedModel{
								BaseModel: models.BaseModel{
									Id: "vg-c90mgoasahsk60ulahi0",
								},
							},
							Name:      "测试",
							Type:      "environment",
							CreatorId: "u-c90hkoqsahsj54g0s8qg",
							OrgId:     "org-c90hl2qsahsj54g0s8ug",
							Variables: models.VarGroupVariables{
								models.VarGroupVariable{
									Id:          "44ac4d00-881c-4ab3-a3e1-284a4a2998ca",
									Name:        "TF_VAR_ak",
									Value:       "资源账号变量",
									Sensitive:   false,
									Description: "测试",
								},
							},
						},
						Overwrites: []VarGroupRel{},
					},
				},
				//普通变量
				vars: map[string]models.Variable{
					"TF_VAR_akenvironment": {
						BaseModel: models.BaseModel{
							AbstractModel: models.AbstractModel{},
							Id:            "var-c9361l2sahsqqvgqmv6g", //variables_id
						},
						VariableBody: models.VariableBody{
							"project",
							"environment",
							"TF_VAR_ak",
							"项目变量设定",
							false,
							"",
							nil,
						},
						ProjectId: "p-c90hl7qsahsj54g0s90g",
					},
				},
			},
			want: map[string]models.Variable{
				"TF_VAR_akenvironment": {
					BaseModel: models.BaseModel{},
					VariableBody: models.VariableBody{
						"env",
						"environment",
						"TF_VAR_ak",
						"资源账号变量",
						false,
						"测试",
						nil,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetVariableGroupVar(tt.args.vgs, tt.args.vars), "GetVariableGroupVar(%v, %v)", tt.args.vgs, tt.args.vars)
		})
	}
}

func TestVarGroupRelMarshal(t *testing.T) {
	vgr := &VarGroupRel{}
	vgr.VariableGroupRel.ObjectId = "xxxxxx"

	vgv := models.VarGroupVariable{
		Id:          "vgv-id",
		Name:        "vgv-name",
		Value:       "vgv-value",
		Sensitive:   true,
		Description: "",
	}
	vgr.VariableGroup.Variables = append(vgr.VariableGroup.Variables, vgv)

	bs, err := json.Marshal(vgr)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}

