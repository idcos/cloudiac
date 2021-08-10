package policy

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io/ioutil"
)

type Policy struct {
	*ast.File
}

type Parser struct {
}

func (p Parser) Parse(filePath string) error {
	//logrus.Errorf("parse \n")

	byt, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	//logrus.Errorf("byt %s\n", byt)

	a, err := hcl.ParseBytes(byt)
	if err != nil {
		return err
	}
	//logrus.Errorf("ast %s\n", a)

	js, err := json.Marshal(a)
	if err != nil {
		return err
	}
	fmt.Printf("%s", js)
	return nil
}
