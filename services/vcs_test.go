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
		{`variable "key" {
			default = 1 
			description = "description"
			type = list(string)
			sensitive = true
		}`},
		{`variable "key" {
			default = 1 
			description = "description"
			type = list(string)
			sensitive = true
  			validation {
    			condition     = length(var.image_id) > 4 && substr(var.image_id, 0, 4) == "ami-"
    			error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  			}
		}`},
	}

	for _, c := range cases {
		_, err := ParseTfVariables("test_variables.tf", []byte(c.content))
		assert.NoError(err)
	}
}
