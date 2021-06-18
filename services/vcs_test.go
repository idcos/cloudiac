package services

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadHCLFile(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		content string
	}{
		{"variable \"key\" {\n }\n"},
		{"variable \"key\" {\n default=1\n }\n"},
		{"variable \"key\" {\n default=1\n description=\"description\"\n }\n"},
	}

	for _, c := range cases {
		_, err := readHCLFile([]byte(c.content))
		assert.NoError(err)
	}
}
