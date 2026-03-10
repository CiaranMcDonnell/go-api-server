package utils

import (
	"reflect"
	"strings"
)

// SanitizeStruct trims whitespace from all string fields and normalizes
// email fields (lowercased) on a struct pointer. It handles both string
// and *string fields, and recurses into embedded structs.
func SanitizeStruct(obj interface{}) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	sanitizeValue(v.Elem())
}

func sanitizeValue(v reflect.Value) {
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("json")
		jsonName := strings.SplitN(tag, ",", 2)[0]
		isEmail := strings.Contains(fieldType.Tag.Get("validate"), "email")

		switch field.Kind() {
		case reflect.String:
			s := strings.TrimSpace(field.String())
			if isEmail || jsonName == "email" {
				s = strings.ToLower(s)
			}
			field.SetString(s)
		case reflect.Ptr:
			if field.IsNil() {
				continue
			}
			elem := field.Elem()
			if elem.Kind() == reflect.String {
				s := strings.TrimSpace(elem.String())
				if isEmail || jsonName == "email" {
					s = strings.ToLower(s)
				}
				elem.SetString(s)
			}
		case reflect.Struct:
			sanitizeValue(field)
		}
	}
}
