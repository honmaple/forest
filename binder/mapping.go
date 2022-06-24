package binder

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func bindData(value interface{}, dst map[string][]string, tagName string) error {
	val := reflect.ValueOf(value)

	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Map {
		for k, v := range dst {
			val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
		}
		return nil
	}

	if val.Kind() != reflect.Struct {
		return errors.New("bind must be a struct")
	}

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		inline := false
		omitempty := false

		field := t.Field(i)
		tag := field.Tag.Get(tagName)
		if field.Anonymous {
			if tag == "-" {
				continue
			}
			if tag != "" {
				return fmt.Errorf("anonymous struct field: %s  are not allowed set tag", field.Name)
			}
			inline = true
		} else {
			if tag == "-" {
				continue
			}
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
		}
		vfield := val.Field(i)
		if !vfield.CanSet() || (omitempty && vfield.IsZero()) {
			continue
		}

		fieldKind := field.Type.Kind()
		if inline && fieldKind == reflect.Struct {
			if err := bindData(vfield.Addr().Interface(), dst, tagName); err != nil {
				return err
			}
			continue
		}
		if tag == "" {
			tag = field.Name
		}
		if v, ok := dst[tag]; ok {
			if err := setField(fieldKind, vfield, v[0]); err != nil {
				return err
			}
		}
	}
	return nil
}

// This function is stolen from echo
func unmarshalField(kind reflect.Kind, field reflect.Value, value string) (bool, error) {
	switch kind {
	case reflect.Ptr:
		if field.IsNil() {
			// Initialize the pointer to a nil value
			field.Set(reflect.New(field.Type().Elem()))
		}
		return unmarshalField(reflect.Struct, field.Elem(), value)
	default:
		fieldValue := field.Addr().Interface()
		if unmarshaler, ok := fieldValue.(encoding.TextUnmarshaler); ok {
			return true, unmarshaler.UnmarshalText([]byte(value))
		}
		if unmarshaler, ok := fieldValue.(json.Unmarshaler); ok {
			return true, unmarshaler.UnmarshalJSON([]byte(value))
		}
		return false, nil
	}
}

func setField(kind reflect.Kind, field reflect.Value, value string) error {
	if ok, err := unmarshalField(kind, field, value); ok {
		return err
	}

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
