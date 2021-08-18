package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

// CreatePolicy 创建策略
func CreatePolicy(c *ctx.ServiceContext, form *forms.CreatePolicyForm) (*models.Policy, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create policy %s", form.Name))

	if form.PolicyType == "" {
		form.PolicyType = "cloudiac"
	}
	if form.Category == "" {
		form.PolicyType = "compliance"
	}

	p := models.Policy{
		Name:          form.Name,
		CreatorId:     c.UserId,
		FixSuggestion: form.FixSuggestion,
		Severity:      form.Severity,
		Rego:          form.Rego,
		Category:      form.Category,
		ResourceType:  form.ResourceType,
		PolicyType:    form.PolicyType,
	}
	refId, err := services.GetPolicyReferenceId(c.DB(), &p)
	if err != nil {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	p.ReferenceId = refId

	policy, err := services.CreatePolicy(c.DB(), &p)
	if err != nil && err.Code() == e.PolicyAlreadyExist {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating policy, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	return policy, nil
}
