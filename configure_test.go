package configure

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	expectedDistrict   = "abc123"
	expectedCollection = "schools"
)

var (
	errMissingDistrictField = fmt.Errorf(missingValuesErrTemplate, []string{"district_id"})
)

func TestConfigure(t *testing.T) {
	for _, spec := range []struct {
		context    string
		args       []string
		err        error
		district   string
		collection string
	}{
		{
			context:  "normal case w/ flags",
			args:     []string{"-district_id=abc123"},
			district: expectedDistrict,
		},
		{
			context: "missing required field",
			err:     errMissingDistrictField,
		},
		{
			context: "given other field but not required field",
			args:    []string{"-collection=schools"},
			err:     errMissingDistrictField,
		},
		{
			context:  "normal case w/ json",
			args:     []string{`{"district_id":"abc123"}`},
			district: expectedDistrict,
		},
		{
			context:    "json w/ all fields",
			args:       []string{`{"district_id":"abc123","collection":"schools"}`},
			district:   expectedDistrict,
			collection: expectedCollection,
		},
		{
			context: "empty JSON blob",
			err:     errMissingDistrictField,
		},
		{
			context: "only evaluates flags if provided first",
			args:    []string{"-collection=schools", `{"district_id":"abc123"}`},
			err:     errMissingDistrictField,
		},
		{
			context: "only evaluates flags if provided first",
			args:    []string{"-collection=schools", `{"district_id":"abc123"}`},
			err:     errMissingDistrictField,
		},
		{
			context: "fails with non-declared flags",
			args:    []string{"-district_id=abc123", "-random-test-flag=X"},
			err:     errors.New("flag provided but not defined: -random-test-flag"),
		},
	} {
		// NOTE: we override both the os.Args and flag.Commandline variables to allow
		// repeated calls to the flag library.
		os.Args = append([]string{"test"}, spec.args...)
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		var config struct {
			DistrictID string `config:"district_id,required"`
			Collection string `config:"collection"`
		}
		if spec.err == nil && assert.NoError(t, Configure(&config), "Case '%s'", spec.context) {
			assert.Equal(t, spec.district, config.DistrictID, "Case '%s'", spec.context)
			assert.Equal(t, spec.collection, config.Collection, "Case '%s'", spec.context)
		} else {
			assert.Equal(t, spec.err, Configure(&config), "Case '%s'", spec.context)
		}
	}
}