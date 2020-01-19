package validators

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
	yaml "gopkg.in/yaml.v2"
)

// EveryContainsValidator validate whether value of Path is an array and contains any Content
type EveryContainsValidator struct {
	Path    string
	Regex   bool
	Content interface{}
}

func (v EveryContainsValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}
	containsFailFormat := `
Path:%s
Expected` + notAnnotation + ` to contain:
%s
Actual:
%s
`
	return splitInfof(
		containsFailFormat,
		v.Path,
		common.TrustedMarshalYAML([]interface{}{v.Content}),
		common.TrustedMarshalYAML(actual),
	)
}

// Validate implement Validatable
func (v EveryContainsValidator) Validate(context *ValidateContext) (bool, []string) {
	manifest, err := context.getManifest()
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	if actual, ok := actual.([]interface{}); ok {
		var contains = false
		for _, ele := range actual {
			if keyValueMatch(ele, v.Content, v.Regex, context.Negative) {
				contains = true
			} else {
				return false, v.failInfo(ele, context.Negative)
			}
		}
		if contains {
			return true, []string{}
		}
		return false, v.failInfo(actual, context.Negative)
	}

	actualYAML, _ := yaml.Marshal(actual)
	return false, splitInfof(errorFormat, fmt.Sprintf(
		"expect '%s' to be an array, got:\n%s",
		v.Path,
		string(actualYAML),
	))
}

func keyValueMatch(actual interface{}, expected interface{}, regex bool, negative bool) bool {
	act := reflect.ValueOf(actual)
	exp := reflect.ValueOf(expected)
	var match = true
	for _, k := range exp.MapKeys() {
		var expectedString = fmt.Sprintf("%v", exp.MapIndex(k))
		var actualString = fmt.Sprintf("%v", act.MapIndex(k))
		if regex {
			p, _ := regexp.Compile(expectedString)
			if !p.MatchString(actualString) != negative {
				match = false
			}
		} else {
			if (expectedString != actualString) != negative {
				match = false
			}
		}
	}
	return match
}
