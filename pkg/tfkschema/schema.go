package tfkschema

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-kubernetes/kubernetes"
)

var ErrAttrNotFound = fmt.Errorf("could not find attribute in resource schema")

// ResourceSchema returns the named Terraform Provider Resource schema
// as defined in the `terraform-provider-kubernetes` package
func ResourceSchema(name string) *schema.Resource {
	prov := kubernetes.Provider().(*schema.Provider)

	if res, ok := prov.ResourcesMap[name]; ok {
		return res
	}

	return nil
}

// IsKubernetesKindSupported returns true if a matching resource is found in the Terraform provider
func IsKubernetesKindSupported(obj runtime.Object) bool {
	name := ToTerraformResourceType(obj)

	res := ResourceSchema(name)
	if res != nil {
		return true
	}

	return false
}

// IsAttributeSupported scans the Terraform resource to determine if the named
// attribute is supported by the Kubernetes provider.
func IsAttributeSupported(attrName string) bool {
	attrParts := strings.Split(attrName, ".")
	res := ResourceSchema(attrParts[0])
	if res == nil {
		return false
	}
	schemaMap := res.Schema

	attr := search(schemaMap, attrParts[1:])
	if attr != nil {
		return true
	}
	return false
}

// IsAttributeRequired scans the Terraform resource to determine if the named
// attribute is required by the Kubernetes provider.
func IsAttributeRequired(attrName string) bool {
	attrParts := strings.Split(attrName, ".")
	res := ResourceSchema(attrParts[0])
	if res == nil {
		return false
	}
	schemaMap := res.Schema

	attr := search(schemaMap, attrParts[1:])
	if attr != nil {
		return attr.Required
	}

	return false
}

func search(m map[string]*schema.Schema, attrParts []string) *schema.Schema {
	searchKey := attrParts[0]
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	if v, ok := m[searchKey]; ok {
		if len(attrParts) == 1 {
			// we hit the bottom of our search and found the attribute
			return v
		}

		if v.Elem != nil {
			switch v.Elem.(type) {
			case *schema.Resource:
				res := v.Elem.(*schema.Resource)
				return search(res.Schema, attrParts[1:])
			}
		}

	}

	return nil
}
