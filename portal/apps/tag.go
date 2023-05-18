// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
)

func CreateTag(c *ctx.ServiceContext, form *forms.CreateTagForm) (*resps.RespTag, e.Error) {
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	var source string = "api"
	if c.ApiTokenId == "" {
		source = "user"
	}

	// 查询key是否存在
	tagKeyList, err := services.FindTagKeyByName(tx, form.Key, c.OrgId)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 不存在则 创建tagKey
	if len(tagKeyList) == 0 {
		// 创建tagKey
		newTagKey, err := services.CreateTagKey(tx, models.TagKey{
			OrgId: c.OrgId,
			Key:   form.Key,
		})

		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		// key不存在value一定不存在 创建tagValue
		tagValue, err := services.CreateTagValue(tx, models.TagValue{
			OrgId: c.OrgId,
			KeyId: newTagKey.Id,
			Value: form.Value,
		})

		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		// 创建tag与环境的管理关系
		if _, err := services.CreateTagRel(tx, models.TagRel{
			OrgId:      c.OrgId,
			TagKeyId:   newTagKey.Id,
			TagValueId: tagValue.Id,
			ObjectId:   form.ObjectId,
			ObjectType: form.ObjectType,
			Source:     source,
		}); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		// key存在，查询value是否存在
		tagValueList, err := services.FindTagValueByName(tx, form.Value, tagKeyList[0].Id, c.OrgId)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if len(tagValueList) == 0 {
			// value不存在，创建value
			tagValue, err := services.CreateTagValue(tx, models.TagValue{
				OrgId: c.OrgId,
				KeyId: tagKeyList[0].Id,
				Value: form.Value,
			})

			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}

			// 创建tag与环境的管理关系
			if _, err := services.CreateTagRel(tx, models.TagRel{
				OrgId:      c.OrgId,
				TagKeyId:   tagKeyList[0].Id,
				TagValueId: tagValue.Id,
				ObjectId:   form.ObjectId,
				ObjectType: form.ObjectType,
				Source:     source,
			}); err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		} else {
			// key value同时存在，查询关联关系是否存在
			tagRelList, err := services.FindTagRelById(tx, tagKeyList[0].Id, tagValueList[0].Id, c.OrgId, form.ObjectId, form.ObjectType)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			// 存在则不做操作直接退出，不存在则创建关联关系
			if len(tagRelList) <= 0 {
				// 创建tag与环境的管理关系
				if _, err := services.CreateTagRel(tx, models.TagRel{
					OrgId:      c.OrgId,
					TagKeyId:   tagKeyList[0].Id,
					TagValueId: tagValueList[0].Id,
					ObjectId:   form.ObjectId,
					ObjectType: form.ObjectType,
					Source:     source,
				}); err != nil {
					_ = tx.Rollback()
					return nil, err
				}
			}
		}

	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

func SearchTag(c *ctx.ServiceContext, form *forms.SearchTagsForm) (interface{}, e.Error) {
	tag := make([]*resps.RespTag, 0)
	query := services.SearchTag(c.DB(), c.OrgId, form.ObjectId, form.ObjectType)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&tag); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     tag,
	}, nil
}

func DeleteTag(c *ctx.ServiceContext, form *forms.DeleteTagsForm) (result interface{}, err e.Error) {
	if err = services.DeleteTagRel(c.DB(), form.KeyId, form.ValueId, c.OrgId, form.ObjectId, form.ObjectType); err != nil {
		return nil, err
	}
	return
}

func UpdateTag(c *ctx.ServiceContext, form *forms.UpdateTagsForm) (tag *resps.RespTag, err e.Error) {
	// 删除原有标签关联关系
	if _, err := DeleteTag(c, &forms.DeleteTagsForm{
		ObjectType: form.ObjectType,
		ObjectId:   form.ObjectId,
		KeyId:      form.KeyId,
		ValueId:    form.ValueId,
	}); err != nil {
		return nil, err
	}

	// 创建标签
	tag, err = CreateTag(c, &forms.CreateTagForm{
		Key:        form.Key,
		Value:      form.Value,
		ObjectType: form.ObjectType,
		ObjectId:   form.ObjectId,
	})
	if err != nil {
		return nil, err
	}

	return tag, err
}
