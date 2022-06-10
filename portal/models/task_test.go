package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskMarshalJSON(t *testing.T) {
	task := Task{
		BaseTask: BaseTask{
			SoftDeleteModel: SoftDeleteModel{
				TimedModel: TimedModel{
					CreatedAt: Time(time.Now()),
					UpdatedAt: Time(time.Now()),
				},
				DeletedAtT: 0,
			},
			StartAt: nil,
			EndAt:   nil,
		},
	}
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
	if tt.Variables[0].Value != "" {
		t.Fatalf("unexpected value: %s", tt.Variables[0].Value)
	}
}
