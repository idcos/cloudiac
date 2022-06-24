package resps

import (
	"encoding/json"
	"testing"
)

func TestTaskDetailRespJSON(t *testing.T) {
	resp := TaskDetailResp{}
	bs, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}
