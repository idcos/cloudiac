package schema

type ReferenceIDFunc func(d *ResourceData) []string

type RegistryItem struct {
	Name                string
	Notes               []string
	RFunc               ResourceFunc
	ReferenceAttributes []string
	CustomRefIDFunc     ReferenceIDFunc
	NoPrice             bool
}
