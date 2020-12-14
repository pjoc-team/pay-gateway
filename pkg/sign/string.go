package sign

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pjoc-team/tracing/logger"
	"reflect"
	"sort"
	"strings"
	"time"
)

// ParamsCompacter compacter fields to form
type ParamsCompacter struct {
	IgnoreKeys                  []string
	IgnoreEmptyValue            bool
	entityFieldNames            []string
	SortedKeyFieldNames         []string
	PairsDelimiter              string
	KeyValueDelimiter           string
	FieldTag                    string
	fieldTagNameAndFieldNameMap map[string]nameAndValueFunc
}

type nameAndValueFunc struct {
	fieldName string
	value     fieldValueFunc
}

type fieldValueFunc func(value interface{}) (string, error)

// NewParamsCompacter new
func NewParamsCompacter(
	entityDemoInstance interface{}, fieldTag string, ignoreKeys []string, ignoreEmptyValue bool,
	pairsDelimiter string, keyValueDelimiter string,
) ParamsCompacter {
	log := logger.Log()
	if fieldTag == "" {
		fieldTag = "json"
	}

	fieldNames := make([]string, 0)
	fieldTagNameAndFieldNameMap := make(map[string]nameAndValueFunc)
	instance := structs.New(entityDemoInstance)
	entityFieldNames := instance.Names()
l:
	for _, k := range entityFieldNames {
		field := instance.Field(k)
		tag := field.Tag(fieldTag)
		name, tagOptions := parseTag(tag)
		log.Infof("field: %v with options: %v", field, tagOptions)
		if tagOptions.Contains("-") || name == "-" || name == "" {
			log.Infof("ignore field: %v by tag: %v", k, name)
			continue
		}

		nvf := nameAndValueFunc{
			fieldName: field.Name(),
			value: func(value interface{}) (string, error) {
				return fmt.Sprintf("%v", value), nil
			},
		}

		tp := reflect.TypeOf(field)
		switch tp {
		case reflect.TypeOf(&timestamp.Timestamp{}):
			nvf.value = func(value interface{}) (string, error) {
				ts := value.(*timestamp.Timestamp)
				return ts.AsTime().Format(time.RFC3339Nano), nil
			}
		}

		fieldTagNameAndFieldNameMap[name] = nvf
		for _, ignore := range ignoreKeys {
			if ignore == name {
				continue l
			}
		}
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	p := ParamsCompacter{}
	p.entityFieldNames = entityFieldNames
	p.SortedKeyFieldNames = fieldNames
	p.IgnoreEmptyValue = ignoreEmptyValue
	p.PairsDelimiter = pairsDelimiter
	p.KeyValueDelimiter = keyValueDelimiter
	p.fieldTagNameAndFieldNameMap = fieldTagNameAndFieldNameMap

	return p
}

func valueFunc() {

}

// ParamsToString convert to string
func (p ParamsCompacter) ParamsToString(instance interface{}) string {
	defer func() {
		if message := recover(); message != nil {
			logger.Log().Errorf(
				"failed to convert instance: %#v to form, error: %v", instance, message,
			)
		}
	}()
	params := make(map[string]string)
	s := structs.New(instance)
	for _, tagFieldName := range p.SortedKeyFieldNames {
		nameValueFunc := p.fieldTagNameAndFieldNameMap[tagFieldName]
		field := s.Field(nameValueFunc.fieldName)
		value := field.Value()
		if p.IgnoreEmptyValue && (value == nil || value == "") {
			continue
		}
		fieldValue, err := nameValueFunc.value(value)
		if err != nil {
			return ""
		}
		params[tagFieldName] = fieldValue
	}

	return p.BuildMapToString(params)
}

// BuildMapToString build map to string
func (p ParamsCompacter) BuildMapToString(params map[string]string) string {
	builder := strings.Builder{}
	delimiter := ""
	for _, key := range p.SortedKeyFieldNames {
		value := params[key]
		if p.IgnoreEmptyValue && value == "" {
			continue
		}
		builder.WriteString(delimiter)
		builder.WriteString(key)
		builder.WriteString(p.KeyValueDelimiter)
		builder.WriteString(value)
		delimiter = p.PairsDelimiter
	}
	return builder.String()
}
