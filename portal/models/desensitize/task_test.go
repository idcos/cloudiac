package desensitize

import (
	"cloudiac/portal/models"
	"encoding/json"
	"testing"
)

func TestDesensitizeTaskMarshalJSON(t *testing.T) {
	task := models.Task{}
	task.Variables = append(task.Variables, models.VariableBody{
		Sensitive: true,
		Value:     "is-sensitive-value",
	})
	desensitizedTask := NewTask(task)

	bs, err := json.Marshal(desensitizedTask)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bs))

	tt := Task{}
	if err := json.Unmarshal(bs, &tt); err != nil {
		t.Fatal(err)
	}
	if tt.Variables[0].Value != "" {
		t.Fatalf("unexpected value: %s", tt.Variables[0].Value)
	}
}
