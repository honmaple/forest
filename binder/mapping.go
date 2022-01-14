package binder

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func bindData(value interface{}, dst map[string][]string, tagName string) error {
	v := reflect.ValueOf(value)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("not struct")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}
		inline := false
		omitempty := false
		opts := strings.Split(tag, ",")
		if len(opts) > 1 {
			for _, flag := range opts[1:] {
				switch flag {
				case "omitempty":
					omitempty = true
				case "inline":
					inline = true
				}
			}
			tag = opts[0]
		}

		vfield := v.Field(i)
		if omitempty && vfield.IsZero() {
			continue
		}

		fieldKind := field.Type.Kind()
		if inline && fieldKind == reflect.Struct {
			// info := bindData(vfield.Interface(), tagName)
			// for _, k := range info.fields {
			//	if _, ok := values[k]; !ok {
			//		fields = append(fields, k)
			//		values[k] = info.values[k]
			//	}
			// }
			continue
		}
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}
		if v, ok := dst[tag]; ok {
			if err := setField(fieldKind, vfield, v[0]); err != nil {
				return err
			}
		}
	}
	return nil
}

func setField(kind reflect.Kind, field reflect.Value, value string) error {
	switch kind {
	case reflect.Bool:
		return setBoolField(value, field)
	case reflect.Int:
		return setIntField(value, 0, field)
	case reflect.Int8:
		return setIntField(value, 8, field)
	case reflect.Int16:
		return setIntField(value, 16, field)
	case reflect.Int32:
		return setIntField(value, 32, field)
	case reflect.Int64:
		return setIntField(value, 64, field)
	case reflect.Uint:
		return setUintField(value, 0, field)
	case reflect.Uint8:
		return setUintField(value, 8, field)
	case reflect.Uint16:
		return setUintField(value, 16, field)
	case reflect.Uint32:
		return setUintField(value, 32, field)
	case reflect.Uint64:
		return setUintField(value, 64, field)
	case reflect.Float32:
		return setFloatField(value, 32, field)
	case reflect.Float64:
		return setFloatField(value, 64, field)
	case reflect.String:
		field.SetString(value)
		return nil
	default:
		return errors.New("unknown field type")
	}
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
