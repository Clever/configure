package configure

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

const (
	structTagKey             = "config"
	requiredTagKey           = "required"
	missingValuesErrTemplate = "Missing required fields: %s"
)

var (
	ErrStringsOnly            = errors.New("Only string values are allowed in a config struct.")
	ErrNotReference           = errors.New("The config struct must be a pointer.")
	ErrStructOnly             = errors.New("Config object must be a struct.")
	ErrNoTagValue             = errors.New("Config object attributes must have a 'config' tag value.")
	ErrFlagParsed             = errors.New("The flag library cannot be used in conjunction with configure")
	ErrInvalidJSON            = errors.New("Invalid JSON found in arguments.")
	ErrStructTagInvalidOption = errors.New("Only 'required' is a config option.")
)

func parseTagKey(tag string) (key string, required bool, err error) {
	s := strings.Split(tag, ",")
	switch len(s) {
	case 2:
		if s[1] != requiredTagKey {
			return "", false, ErrStructTagInvalidOption
		}
		return s[0], true, nil
	case 1:
		return s[0], false, nil
	default:
		return "", false, ErrNoTagValue
	}
}

// Configure takes a reference to an interface that has 'config' tags on all atttributes of
// the struct. Configure first tries to find values for these attributes through command line
// flags, then will attempt to parse the first argument as a JSON blob.
// An attribute can be required by appending ',required' to the config key.
func Configure(config interface{}) error {
	if flag.Parsed() {
		return ErrFlagParsed
	}

	val2 := reflect.ValueOf(config)
	if val2.Kind() != reflect.Ptr {
		return ErrStructOnly
	}

	values := map[string]*string{}
	configFlags := flag.NewFlagSet("configure", flag.ContinueOnError)
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

		values[tagVal] = configFlags.String(tagVal, "", "generated field")
	}
	if err := configFlags.Parse(os.Args[1:]); err != nil {
		return err
	}

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

	if !flagFound && configFlags.Arg(0) != "" {
		jsonValues := map[string]string{}
		if err := json.NewDecoder(bytes.NewBufferString(configFlags.Arg(0))).Decode(&jsonValues); err != nil {
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
			return fmt.Errorf(missingValuesErrTemplate, missingRequiredFields)
		}
	}

	return nil
}
