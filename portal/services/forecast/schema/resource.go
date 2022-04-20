package schema

type ResourceFunc func(*ResourceData) *Resource

type Resource struct {
	Name        string
	PriceType   string
	PriceCode   string
	RequestData []*PriceRequest
}


type PriceRequest struct {
	Name  string
	Value string
}
