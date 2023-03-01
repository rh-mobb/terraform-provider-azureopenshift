package validate

import (
	"fmt"

	"golang.org/x/exp/slices"
)

func SDNType(v interface{}, _ string) (warnings []string, errors []error) {
	value := slices.Contains(getSupportSDNTypes(), v.(string))
	if value != true {
		errors = append(errors, fmt.Errorf(
			"The `sdn_type` must be `OVNKubernetes` or `OpenShiftSDN`"))
	}
	return warnings, errors
}

func getSupportSDNTypes() []string {
	return []string{"OpenShiftSDN", "OVNKubernetes"}
}
