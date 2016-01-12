package configure

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

const (
	structTagKey   = "config"
	requiredTagKey = "required"
)

var (
	ErrStringsOnly            = errors.New("Only string values are allowed in a config struct.")
	ErrNotReference           = errors.New("The config struct must be a pointer.")
	ErrStructOnly             = errors.New("Config object must be a struct.")
	ErrNoTagValue             = errors.New("Config object attributes must have a 'config' tag value.")
	ErrFlagParsed             = errors.New("Flags should not be used in conjuction, do not flag.Parse() before Configure")
	ErrInvalidJSON            = errors.New("Invalid JSON found in arguments.")
	ErrStructTagInvalidOption = errors.New("Only 'required' is a config option.")
)

func parseTagKey(tag string) (key string, required bool, err error) {
	s := strings.Split(tag, ",")
	switch len(s) {
	case 2:
		if s[1] != "required" {
			return "", false, ErrStructTagInvalidOption
		}
		return s[0], true, nil
	case 1:
		return s[0], false, nil
	default:
		return "", false, ErrNoTagValue
	}
}

func Configure(config interface{}) error {
	if flag.Parsed() {
		return ErrFlagParsed
	}

	val2 := reflect.ValueOf(config)
	if val2.Kind() != reflect.Ptr {
		return ErrStructOnly
	}

	values := map[string]*string{}
	flagFound := false
	requiredFields := false
	missingRequiredFields := []string{}

	val := val2.Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		if !valueField.CanSet() {
			return ErrNotReference
		}

		typeField := val.Type().Field(i)
		if typeField.Type.Name() != "string" {
			return ErrStringsOnly
		}

		tagVal, required, err := parseTagKey(typeField.Tag.Get(structTagKey))
		if err != nil {
			return err
		} else if required {
			requiredFields = true
		}

		values[tagVal] = flag.String(tagVal, "", "generated field")
	}
	flag.Parse()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		tagVal, _, err := parseTagKey(val.Type().Field(i).Tag.Get(structTagKey))
		if err != nil {
			return err
		}

		if *values[tagVal] != "" {
			flagFound = true
			valueField.SetString(*values[tagVal])
		}
	}

	if !flagFound && flag.Arg(0) != "" {
		jsonValues := map[string]string{}
		if err := json.NewDecoder(bytes.NewBufferString(flag.Arg(0))).Decode(&jsonValues); err != nil {
			return ErrInvalidJSON
		}

		for i := 0; i < val.NumField(); i++ {
			valueField := val.Field(i)
			tagVal, _, err := parseTagKey(val.Type().Field(i).Tag.Get(structTagKey))
			if err != nil {
				return err
			}

			if jsonValues[tagVal] != "" {
				valueField.SetString(jsonValues[tagVal])
			}
		}
	}

	// validate that all required fields were set
	if requiredFields {
		for i := 0; i < val.NumField(); i++ {
			valueField := val.Field(i)
			typeField := val.Type().Field(i)
			if typeField.Type.Name() != "string" {
				return ErrStringsOnly
			}

			tagKey, required, err := parseTagKey(typeField.Tag.Get(structTagKey))
			if err != nil {
				return err
			} else if required && valueField.String() == "" {
				missingRequiredFields = append(missingRequiredFields, tagKey)
			}
		}
		if len(missingRequiredFields) > 0 {
			return fmt.Errorf("Missing required fields: %s", missingRequiredFields)
		}
	}

	return nil
}
