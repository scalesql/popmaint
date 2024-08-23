package lx

import "fmt"

type Field struct {
	K string
	V any
}

// 0 1 2 3 4 5
// a			(len=1)
// a A 			(len=2)
// a A b B c C 	(len=6)
// a A b B     	(len=4)
// a A b       	(len=3)
func args2map(args ...any) map[string]any {
	m := make(map[string]any)
	if len(args) == 0 {
		return m
	}
	i := 0
	for i < len(args) {
		str, ok := args[i].(string)
		if !ok { // this isn't a string, move on
			m[fmt.Sprintf("!BADKEY-%d", i)] = args[i]
			i++
			continue
		}
		if i+1 >= len(args) { // no more args
			m[fmt.Sprintf("!BADKEY-%d", i)] = args[i]
			return m
		}
		m[str] = args[i+1]
		i = i + 2
	}
	return m
}

func map2args(m map[string]any) []any {
	arr := make([]any, 0)
	for k, v := range m {
		arr = append(arr, k)
		arr = append(arr, v)
	}
	return arr
}
