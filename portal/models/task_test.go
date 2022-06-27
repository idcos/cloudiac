package models

import (
	"encoding/json"
	"testing"
)

func TestTaskMarshalJSON(t *testing.T) {
	task := Task{}
	task.Variables = append(task.Variables, VariableBody{
		Sensitive: true,
		Value:     "is-sensitive-value",
	})
	bs, err := json.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bs))

	tt := Task{}
	if err := json.Unmarshal(bs, &tt); err != nil {
		t.Fatal(err)
	}
	if tt.Variables[0].Value != "is-sensitive-value" {
		t.Fatalf("unexpected value: %s", tt.Variables[0].Value)
	}
}
