package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"mime/multipart"
)

type TplExportForm struct {
	forms.BaseForm

	Ids      []models.Id `json:"ids" form:"ids" binding:required` // 待导出的云模板 id 列表
	Download bool        `json:"download" form:"download"`        // download 模式(直接返回导出数据 ，并触发浏览器下载)
}

func TemplateExport(c *ctx.ServiceContext, form *TplExportForm) (interface{}, e.Error) {
	return services.ExportTemplates(c.DB(), c.OrgId, form.Ids)
}

type TplImportForm struct {
	forms.BaseForm

	IdDuplicate string      `json:"idDuplicate" form:"idDuplicate" bind:"required"` // id 重复时的处理方式, enum('update','skip','copy','abort')
	Projects    []models.Id `json:"projects" form:"projects"`                       // 关联项目 id 列表

	Data services.TplExportedData `json:"data"  swaggerignore:"true"` // 待导入数据(JSON 格式，与 file 参数二选一)

	File *multipart.FileHeader `form:"file" swaggerignore:"true"` // 待导入文件(与 data 参数二选一)
}

func TemplateImport(c *ctx.ServiceContext, form *TplImportForm) (result *services.TplImportResult, er e.Error) {
	importer := services.TplImporter{
		Logger:          c.Logger().WithField("action", "ImportTemplate"),
		OrgId:           c.OrgId,
		CreatorId:       consts.SysUserId,
		ProjectIds:      form.Projects,
		Data:            form.Data,
		WhenIdDuplicate: form.IdDuplicate,
	}
	// return importer.Import(c.DB())

	_ = c.DB().Transaction(func(tx *db.Session) error {
		result, er = importer.Import(tx)
		return er
	})
	return result, er
}
