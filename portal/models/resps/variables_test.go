// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import (
	"cloudiac/portal/models/desensitize"
	"encoding/json"
	"testing"
)

func TestVariableRespMarshalJson(t *testing.T) {
	resp := VariableResp{}
	resp.Overwrites = &desensitize.Variable{}
	resp.Overwrites.Name = "TestVariableRespMarshalJson"
	bs, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)

	r := VariableResp{}
	if err := json.Unmarshal(bs, &r); err != nil {
		t.Fatal(err)
	}
	if r.Overwrites == nil || r.Overwrites.Name != "TestVariableRespMarshalJson" {
		t.Fatal("TestVariableRespMarshalJson")
	}
}
