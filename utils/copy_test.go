package utils

import "testing"

// copier 深拷贝时如果未传入 option IgnoreEmpty=true 则会导致 nil []byte 拷贝为 make([]byte, 0)，
// 这里做一个测试用例覆盖
func TestDeepCopyNilByteSlice(t *testing.T) {
	type A struct {
		JSON []byte
	}

	a := A{}
	ca := A{}
	DeepCopy(&ca, &a)
	if ca.JSON != nil {
		t.Fatalf("deep copy empty []byte failed, except nil, got '%v'", ca.JSON)
	}
}
