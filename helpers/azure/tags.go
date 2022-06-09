package azure

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TagsExpand(tagsMap map[string]interface{}) map[string]*string {
	output := make(map[string]*string, len(tagsMap))

	for i, v := range tagsMap {
		// Validate should have ignored this error already
		value, _ := TagValueToString(v)
		output[i] = &value
	}

	return output
}

func TagsFlattenAndSet(d *schema.ResourceData, tagMap map[string]*string) error {
	flattened := Flatten(tagMap)
	if err := d.Set("tags", flattened); err != nil {
		return fmt.Errorf("setting `tags`: %s", err)
	}

	return nil
}

func Flatten(tagMap map[string]*string) map[string]interface{} {
	// If tagsMap is nil, len(tagsMap) will be 0.
	output := make(map[string]interface{}, len(tagMap))

	for i, v := range tagMap {
		if v == nil {
			continue
		}

		output[i] = *v
	}

	return output
}
