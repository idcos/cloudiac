package services

import (
	"testing"
)

func TestCheckPasswordFormat(t *testing.T) {
	pwCase := []struct {
		password string
		expect   bool
	}{
		{"abc123", true},
		{"abc12", false},
		{"ABC12", false},
		{"ABC!@", false},
		{"abc!@", false},
		{"123!@", false},
		{"abc12abc12abc12abc12abcABC!@#c121", false},
		{"abc12abc12abc12abc12abc12abc12", true},
		{"abc12abc12abc12!@#$%c12ABC12", true},
		{"!@#123", true},
		{"!@#abc", true},
		{"abcABC", false},
		{"#$ABa@", true},
		{"ab12!@AB", true},
		{"ab12!@AB", true},
	}
	for _, c := range pwCase {
		err := CheckPasswordFormat(c.password)
		if c.expect {
			if err != nil {
				t.Errorf("password %s, err %v", c.password, err)
			}
		}else {
			if err == nil {
				t.Errorf("password %s, err %v", c.password, err)
			}
		}
	}
}
