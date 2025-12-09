package flagparser

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

type fieldValue struct {
	IntValue     *int
	BooleanValue *bool
	UintValue    *uint
	UintValue64  *uint64
	IntValue64   *int64
	StringValue  *string
	IsInt        bool
	Value        reflect.Value
}

func parseFields(v reflect.Value, fieldMap map[string]fieldValue) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fValue := v.Field(i)
		flagName := field.Tag.Get("flag")
		defaultValue := field.Tag.Get("default")
		usage := field.Tag.Get("usage")

		switch field.Type.Kind() {
		case reflect.String:
			var ptr = new(string)
			flag.StringVar(ptr, flagName, defaultValue, usage)
			fieldMap[flagName] = fieldValue{StringValue: ptr, IsInt: false, Value: fValue}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			defaultInt, err := strconv.Atoi(defaultValue)
			if err != nil {
				return fmt.Errorf("invalid default value for %s: %v", field.Name, err)
			}
			var ptr = new(int)
			flag.IntVar(ptr, flagName, defaultInt, usage)
			fieldMap[flagName] = fieldValue{IntValue: ptr, IsInt: true, Value: fValue}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			defaultInt, err := strconv.ParseUint(defaultValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid default value for %s: %v", field.Name, err)
			}
			var ptr = new(uint)
			flag.UintVar(ptr, flagName, uint(defaultInt), usage)
			fieldMap[flagName] = fieldValue{UintValue: ptr, IsInt: true, Value: fValue}
		case reflect.Uint64:
			defaultInt, err := strconv.ParseUint(defaultValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid default value for %s: %v", field.Name, err)
			}
			var ptr = new(uint64)
			flag.Uint64Var(ptr, flagName, defaultInt, usage)
			fieldMap[flagName] = fieldValue{UintValue64: ptr, IsInt: true, Value: fValue}
		case reflect.Bool:
			var ptr = new(bool)
			flag.BoolVar(ptr, flagName, defaultValue == "true", usage)
			fieldMap[flagName] = fieldValue{BooleanValue: ptr, IsInt: false, Value: fValue}
		case reflect.Struct:
			err := parseFields(fValue, fieldMap)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported type %s for field %s", field.Type.Kind(), field.Name)
		}
	}

	return nil
}

func ParseFlags(config interface{}) error {
	v := reflect.ValueOf(config).Elem()
	fieldMap := make(map[string]fieldValue)

	err := parseFields(v, fieldMap)
	if err != nil {
		return err
	}

	flag.Parse()

	for _, s := range fieldMap {
		switch s.Value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			s.Value.SetInt(int64(*s.IntValue))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			s.Value.SetUint(uint64(*s.UintValue))
		case reflect.Bool:
			s.Value.SetBool(*s.BooleanValue)
		case reflect.String:
			s.Value.SetString(*s.StringValue)
		default:
			return fmt.Errorf("unsupported type %s", s.Value.Kind())
		}
	}

	return nil
}
