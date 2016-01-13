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
	ErrNotReference           = errors.New("The config struct must be a pointer to a struct.")
	ErrStructOnly             = errors.New("Config object must be a struct.")
	ErrNoTagValue             = errors.New("Config object attributes must have a 'config' tag value.")
	ErrTooManyTagValues       = errors.New("Config object attributes can only have a key and optional required attribute.")
	ErrFlagParsed             = errors.New("The flag library cannot be used in conjunction with configure")
	ErrInvalidJSON            = errors.New("Invalid JSON found in arguments.")
	ErrStructTagInvalidOption = errors.New("Only 'required' is a config option.")
)

// parseTagKey parses the values in a tag.
func parseTagKey(tag string) (key string, required bool, err error) {
	if tag == "" {
		return "", false, ErrNoTagValue
	}

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
		return "", false, ErrTooManyTagValues
	}
}

// Configure takes a reference to an interface that has 'config' tags on all atttributes of
// the struct. Configure first tries to find values for these attributes through command line
// flags, then will attempt to parse the first argument as a JSON blob.
// An attribute can be required by appending ',required' to the config key.
func Configure(configStruct interface{}) error {
	if flag.Parsed() {
		return ErrFlagParsed
	}

	reflectConfig := reflect.ValueOf(configStruct)
	if reflectConfig.Kind() != reflect.Ptr {
		return ErrStructOnly
	}

	var (
		configFlags  = flag.NewFlagSet("configure", flag.ContinueOnError)
		flagValueMap = map[string]*string{} // holds references to attribute flags
		flagFound    = false                // notes if any flags are found, JSON parsing is skipped if so
		config       = reflectConfig.Elem()
	)

	// this block creates flags for every attribute
	for i := 0; i < config.NumField(); i++ {
		valueField := config.Field(i)
		if !valueField.CanSet() {
			return ErrNotReference
		}

		// currently we only support strings
		typedAttr := config.Type().Field(i)
		if typedAttr.Type.Kind() != reflect.String {
			return ErrStringsOnly
		}

		// get the name of the value and create a flag
		tagVal, _, err := parseTagKey(typedAttr.Tag.Get(structTagKey))
		if err != nil {
			return err
		}
		flagValueMap[tagVal] = configFlags.String(tagVal, "", "generated field")
	}
	if err := configFlags.Parse(os.Args[1:]); err != nil {
		return err
	}

	// grab values from flag map
	for i := 0; i < config.NumField(); i++ {
		valueField := config.Field(i)
		tagVal, _, err := parseTagKey(config.Type().Field(i).Tag.Get(structTagKey))
		if err != nil {
			return err
		}

		if *flagValueMap[tagVal] != "" {
			flagFound = true
			valueField.SetString(*flagValueMap[tagVal])
		}
	}

	// if no flags were found and we have a value in the first arg, we try to parse JSON from it.
	if !flagFound && configFlags.Arg(0) != "" {
		jsonValues := map[string]string{}
		if err := json.NewDecoder(bytes.NewBufferString(configFlags.Arg(0))).Decode(&jsonValues); err != nil {
			return ErrInvalidJSON
		}

		for i := 0; i < config.NumField(); i++ {
			valueField := config.Field(i)
			tagVal, _, err := parseTagKey(config.Type().Field(i).Tag.Get(structTagKey))
			if err != nil {
				return err
			} else if jsonValues[tagVal] != "" {
				valueField.SetString(jsonValues[tagVal])
			}
		}
	}

	// validate that all required fields were set
	missingRequiredFields := []string{}
	for i := 0; i < config.NumField(); i++ {
		tagKey, required, err := parseTagKey(config.Type().Field(i).Tag.Get(structTagKey))
		if err != nil {
			return err
		} else if required && config.Field(i).String() == "" {
			missingRequiredFields = append(missingRequiredFields, tagKey)
		}
	}
	if len(missingRequiredFields) > 0 {
		return fmt.Errorf(missingValuesErrTemplate, missingRequiredFields)
	}

	return nil
}
