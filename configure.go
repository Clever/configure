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
	ErrStringAndBoolOnly      = errors.New("Only string/bool values are allowed in a config struct.")
	ErrBoolCannotBeRequired   = errors.New("Boolean attributes cannot be required")
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
		configFlags        = flag.NewFlagSet("configure", flag.ContinueOnError)
		flagStringValueMap = map[string]*string{} // holds references to attribute string flags
		flagBoolValueMap   = map[string]*bool{}   // holds references to attribute bool flags
		flagFound          = false                // notes if any flags are found, JSON parsing is skipped if so
		config             = reflectConfig.Elem()
	)

	// this block creates flags for every attribute
	for i := 0; i < config.NumField(); i++ {
		valueField := config.Field(i)
		if !valueField.CanSet() {
			return ErrNotReference
		}

		// currently we only support strings and bools
		typedAttr := config.Type().Field(i)
		if typedAttr.Type.Kind() != reflect.String && typedAttr.Type.Kind() != reflect.Bool {
			return ErrStringAndBoolOnly
		}

		// get the name of the value and create a flag
		tagVal, _, err := parseTagKey(typedAttr.Tag.Get(structTagKey))
		if err != nil {
			return err
		}
		switch typedAttr.Type.Kind() {
		case reflect.String:
			flagStringValueMap[tagVal] = configFlags.String(tagVal, "", "generated field")
		case reflect.Bool:
			// set the default to the value passed in
			flagBoolValueMap[tagVal] = configFlags.Bool(tagVal, config.Field(i).Bool(), "generated field")
		}
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

		typedAttr := config.Type().Field(i)
		switch typedAttr.Type.Kind() {
		case reflect.String:
			if *flagStringValueMap[tagVal] != "" {
				flagFound = true
				valueField.SetString(*flagStringValueMap[tagVal])
			}
		case reflect.Bool:
			// we can only know if a bool flag was set if the default was changed
			if *flagBoolValueMap[tagVal] != config.Field(i).Bool() {
				flagFound = true
			}
			valueField.SetBool(*flagBoolValueMap[tagVal]) // always set from flags
		}
	}

	// if no flags were found and we have a value in the first arg, we try to parse JSON from it.
	if !flagFound && configFlags.Arg(0) != "" {
		jsonValues := map[string]interface{}{}
		if err := json.NewDecoder(bytes.NewBufferString(configFlags.Arg(0))).Decode(&jsonValues); err != nil {
			return ErrInvalidJSON
		}

		for i := 0; i < config.NumField(); i++ {
			valueField := config.Field(i)
			tagVal, _, err := parseTagKey(config.Type().Field(i).Tag.Get(structTagKey))
			if err != nil {
				return err
			} else if _, ok := jsonValues[tagVal]; ok {
				typedAttr := config.Type().Field(i)
				switch typedAttr.Type.Kind() {
				case reflect.String:
					valueField.SetString(jsonValues[tagVal].(string))
				case reflect.Bool:
					valueField.SetBool(jsonValues[tagVal].(bool))
				}
			}
		}
	}

	// validate that all required fields were set
	missingRequiredFields := []string{}
	for i := 0; i < config.NumField(); i++ {
		tagKey, required, err := parseTagKey(config.Type().Field(i).Tag.Get(structTagKey))
		if err != nil {
			return err
		} else if required {
			switch config.Field(i).Type().Kind() {
			case reflect.String:
				if config.Field(i).String() == "" {
					missingRequiredFields = append(missingRequiredFields, tagKey)
				}
			case reflect.Bool:
				return ErrBoolCannotBeRequired
			}
		}
	}
	if len(missingRequiredFields) > 0 {
		return fmt.Errorf(missingValuesErrTemplate, missingRequiredFields)
	}

	return nil
}
