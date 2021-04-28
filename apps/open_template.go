package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/services"
)

func OpenSearchTemplate(c *ctx.ServiceCtx) (interface{}, e.Error) {
	resp := make([]struct {
		Name string `json:"name"`
		Guid string `json:"guid"`
	}, 0)
	if err := services.OpenSearchTemplate(c.DB()).Scan(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return resp, nil
}
