package px

import (
	"fmt"
	"reflect"
	"strings"
)

func nested2dotted(m map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	for k, v := range m {
		if reflect.ValueOf(v).Kind() == reflect.Map {
			child, ok := v.(map[string]any)
			if !ok {
				return result, fmt.Errorf("invalid map: %s", k)
			}
			err := nested2dottedChild(k, result, child)
			if err != nil {
				return result, err
			}
		} else {
			result[k] = v
		}
	}
	return result, nil
}

func nested2dottedChild(path string, parent map[string]any, m map[string]any) error {
	depth := strings.Count(path, ".")
	if depth > 9 {
		return fmt.Errorf("field path too deep: %s", path)
	}
	for k, v := range m {
		childPath := fmt.Sprintf("%s.%s", path, k)
		if reflect.ValueOf(v).Kind() == reflect.Map {
			child, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid map: %s", childPath)
			}
			err := nested2dottedChild(childPath, parent, child)
			if err != nil {
				return err
			}
		} else {
			parent[childPath] = v
		}
	}
	return nil
}
