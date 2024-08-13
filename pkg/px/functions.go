package px

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func (px *PX) applyFuncs(m map[string]any) (map[string]any, []error) {
	errs := make([]error, 0)
	for _, fld := range px.Functions {
		val, src, fn, err := parseFunc(fld.V)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if val != nil {
			m[fld.K] = val
			continue
		}
		// TODO apply the functions
		switch fn {
		case "hostname()":
			hn, err := os.Hostname()
			if err != nil {
				errs = append(errs, fmt.Errorf("os.hostname: %w", err))
				continue
			}
			m[fld.K] = hn
		case "exename()":
			val, err := os.Executable()
			if err != nil {
				errs = append(errs, fmt.Errorf("os.exename: %w", err))
				continue
			}
			val = filepath.Base(val)
			m[fld.K] = val
		case "delete()":
			delete(m, fld.K)
		case "copy()":
			m[fld.K] = m[src]
		case "move()":
			m[fld.K] = m[src]
			delete(m, src)
		case "jobid()":
			if px.JobID != "" {
				m[fld.K] = px.JobID
			}
		default:
			if px.Constants != nil {
				cons, ok := px.Constants[fn]
				if ok {
					m[fld.K] = cons
					continue
				}
			}
			errs = append(errs, fmt.Errorf("unknown function: %s", fn))
		}
	}

	return m, errs
}

// parseFieldValue
// function vs. value

// parseFunc checks for a value and/or function
func parseFunc(value any) (any, string, string, error) {
	// if this isn't a string, just return the value
	str, ok := value.(string)
	if !ok {
		return value, "", "", nil
	}
	// it is a string, so we parse it into field/value and function
	function := ""
	ff := strings.Split(str, ".")
	// count functions
	fnCount := 0
	ptr := 0
	for i, x := range ff {
		if strings.HasSuffix(x, "()") {
			ptr = i
			fnCount++
		}
	}
	if fnCount == 0 { // no functions, return the string as a value
		return str, "", "", nil
	}
	if fnCount > 1 { // too many functions
		return nil, "", "", fmt.Errorf("%s: multiple functions", str)
	}

	// make sure only last one is function
	if ptr != len(ff)-1 {
		return nil, "", "", fmt.Errorf("%s: function must be last", str)
	}

	// if the last is a funtion, pull it out
	if strings.HasSuffix(ff[len(ff)-1], "()") {
		function = ff[len(ff)-1]
		ff = ff[:len(ff)-1] // remove the function we just extracted
	}

	field := strings.Join(ff, ".")
	return nil, field, function, nil
}

func dotted2nested(dotted map[string]any) (map[string]any, error) {
	// groups := make(map[string]bool)
	nested := make(map[string]any)
	for mapkey, mapvalue := range dotted {
		//println("mapkey:", mapkey)
		parentkey := make([]string, 0)
		keys := strings.Split(mapkey, ".")
		m := nested
		for i, k := range keys { // 0, 1, 2
			parentkey = append(parentkey, k)
			//println("k:", k, parentkey)
			if i == len(keys)-1 { // value
				// if this already exists, it is an error
				// because it can only be a map
				_, ok := m[k]
				if ok {
					return nil, fmt.Errorf("%s: already exists", strings.Join(parentkey, "."))
				}
				m[k] = mapvalue
			} else {
				v, ok := m[k] //.(map[string]any)
				//fmt.Printf("key: %s Type: %T\n", k, v)
				// if ok, and not map, return error
				if ok && reflect.ValueOf(v).Kind() != reflect.Map {
					//println("error")
					return nil, fmt.Errorf("%s: parent and value", strings.Join(parentkey, "."))
				}
				// if not ok, add it
				if !ok {
					//println("not ok")
					child := make(map[string]any)
					m[k] = child
					m = child
					continue
				}
				// it must be Ok and a map, set m to this and keep going down
				if ok && reflect.ValueOf(v).Kind() == reflect.Map {
					//println("ok and map")
					child := v.(map[string]any)
					m = child
					continue
				}
				// we fell through.  This is bad.
				return nil, fmt.Errorf("%s: fallthrough", keys)
			}
			//fmt.Printf("MAP: %v\n", nested)
		}
	}
	// str, err := map2json(nested)
	// if err != nil {
	// 	println(err)
	// }
	// println(str)
	return nested, nil
}

// func map2json(m map[string]any) (string, error) {
// 	bb, err := json.Marshal(m)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(bb), nil
// }

func anys2map(payload string, args ...any) map[string]any {
	// fmt.Println(payload)
	// fmt.Println(args...)
	// fmt.Println(len(args))
	m := make(map[string]any)
	if len(args) == 0 {
		return m
	}
	if len(args) == 1 {
		key, ok := args[0].(string)
		if !ok {
			panic(fmt.Sprintf("not string: %v", args[0]))
		}
		m[key] = "!BADKEY"
		return m
	}

	for i := 0; i < len(args)-1; i += 2 {
		if i+1 >= len(args) { // need two values
			break
		}
		key, ok := args[i].(string)
		if !ok {
			panic(fmt.Sprintf("not string: %v", args[i]))
		}
		if payload != "" {
			key = payload + "." + key
		}
		value := args[i+1]
		m[key] = value
	}
	// TODO if an odd number, add the last key with "!BADKEY" as the value
	if len(args)%2 == 1 {
		// sl[len(sl)-1]
		key, ok := args[len(args)-1].(string)
		if !ok {
			panic(fmt.Sprintf("not string: %v", args[0]))
		}
		m[key] = "!BADKEY"
		return m
	}
	return m
}
