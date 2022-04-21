package terraform

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
	"github.com/tidwall/gjson"
	"os"
	"reflect"
	"testing"
)

func TestBuildResource(t *testing.T) {
	type args struct {
		resource     []*schema.Resource
		registryMap  *ResourceRegistryMap
		t            string
		providerName string
		address      string
		rawValues    gjson.Result
	}
	tests := []struct {
		name string
		args args
		want []*schema.Resource
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildResource(tt.args.resource, tt.args.registryMap, tt.args.t, tt.args.providerName, tt.args.address, tt.args.rawValues); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParserPlanJson(t *testing.T) {
	type args struct {
		b []byte
	}
	b, err := os.ReadFile("./plan.json")
	fmt.Println(err, "1111")
	tests := []struct {
		name                     string
		args                     args
		wantCreateResource       []*schema.Resource
		wantDeleteResource       []*schema.Resource
		wantUpdateBeforeResource []*schema.Resource
	}{
		{
			name: "test-01",
			args: args{
				b: b,
			},
			wantCreateResource: []*schema.Resource{
				&schema.Resource{
					Name:      "alicloud_instance.instance1",
					PriceType: "",
					PriceCode: "ecs",
					RequestData: []*schema.PriceRequest{
						&schema.PriceRequest{Name: "", Value: ""},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreateResource, gotDeleteResource, gotUpdateBeforeResource := ParserPlanJson(tt.args.b)
			for _, Create := range gotCreateResource {
				for _, value := range Create.RequestData {
					fmt.Println(value)
				}
			}
			t.Log(gotCreateResource, "gotCreateResource")
			t.Log(gotDeleteResource, "gotDeleteResource")
			t.Log(gotUpdateBeforeResource, "gotUpdateBeforeResource")
			//if !reflect.DeepEqual(gotCreateResource, tt.wantCreateResource) {
			//	t.Errorf("ParserPlanJson() gotCreateResource = %v, want %v", gotCreateResource, tt.wantCreateResource)
			//}
			//if !reflect.DeepEqual(gotDeleteResource, tt.wantDeleteResource) {
			//	t.Errorf("ParserPlanJson() gotDeleteResource = %v, want %v", gotDeleteResource, tt.wantDeleteResource)
			//}
			//if !reflect.DeepEqual(gotUpdateBeforeResource, tt.wantUpdateBeforeResource) {
			//	t.Errorf("ParserPlanJson() gotUpdateBeforeResource = %v, want %v", gotUpdateBeforeResource, tt.wantUpdateBeforeResource)
			//}
		})
	}
}
