package request

import (
	"reflect"
	"strings"
)

// NormalizeStrings applies normalization rules to string and *string fields.
// Supported rules: trim, lower, upper, singlespace.
// Currently it does not support nested struct.
func NormalizeStrings(s interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("normalize")
		if tag == "" || !field.CanSet() {
			continue
		}

		// Get string value (supports string and *string)
		var str string
		isPtr := field.Kind() == reflect.Ptr
		if isPtr {
			if field.IsNil() || field.Elem().Kind() != reflect.String {
				continue
			}
			str = field.Elem().String()
		} else if field.Kind() == reflect.String {
			str = field.String()
		} else {
			continue
		}

		// Apply rules
		for _, rule := range strings.Split(tag, ",") {
			switch strings.TrimSpace(rule) {
			case "trim":
				str = strings.TrimSpace(str)
			case "lower":
				str = strings.ToLower(str)
			case "upper":
				str = strings.ToUpper(str)
			case "singlespace":
				str = strings.Join(strings.Fields(str), " ")
			}
		}

		// Set value back
		if isPtr {
			field.Elem().SetString(str)
		} else {
			field.SetString(str)
		}
	}
}

