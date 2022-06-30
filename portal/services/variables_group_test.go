package services

import (
	"cloudiac/portal/models"
	"encoding/json"
	"testing"
)

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
