package services

import (
	"encoding/json"
	"testing"
)

func TestRegistryRespUnmarshal(t *testing.T) {
	jsonData := []byte("{\"code\":0,\"message\":\"\",\"result\":{\"list\":[{\"id\":\"07SXgVdATmWUGgkqCBjfpA\",\"creatorId\":\"debuguser\",\"namespace\":\"aliyun\",\"name\":\"iacregistry-example\",\"description\":\"\",\"label\":\"\",\"icon\":\"\",\"repoName\":\"iacregistry-example-07SXgVdATmWUGgkqCBjfpA\",\"vcsId\":\"\",\"repoId\":\"\",\"repoAddr\":\"https://gitee.com/jxinging/iacregistry-example.git\",\"repoPath\":\"policies/aliyun/iacregistry-example-07SXgVdATmWUGgkqCBjfpA\"}],\"page\":1,\"pageSize\":10,\"total\":1}}")
	resp := registryResp{}
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Result) == 0 {
		t.Fatalf("unmarshal registry result failed")
	} else {
		t.Logf("response.result: %s", resp.Result)
	}
}
