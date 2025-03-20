package form_builder

import (
	"fmt"
	"reflect"
	"strings"
)

type FormField struct {
	Name 		string `json:"name,omitempty"`
	Label     	string `json:"label,omitempty"`
	Value  		string `json:"value,omitempty"`
	Type      	string `json:"type,omitempty"`// input type: "text", "password", "checkbox", etc.
	Required  	bool `json:"required,omitempty"`
	Options  	[]map[string]string `json:"options,omitempty"`
	ReadOnly  	bool `json:"readonly,omitempty"`
	SortOrder   int32 `json:"readonly,omitempty"`
}

type Form struct {
	Title  string
	Confirm string
	Action string
	Method string
	Fields []FormField
}

func parseFormTag(tag string) map[string]string {
	props := map[string]string{}
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			props[key] = value
		}
	}
	return props
}

func parseJsonTag(tag string) []string {
	props := strings.Split(tag, ",")
	return props
}

func generateFormFields(input interface{}, mode string) ([]FormField, error) {
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var fields []FormField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Get The value
		value, ok := v.Field(i).Interface().(string)
		if !ok {
			// Option 2: Convert it to a string using fmt.Sprintf, if appropriate:
			value = fmt.Sprintf("%v", value)
		}

		// Get Tags
		tagForm := field.Tag.Get("form")
		if tagForm == "" {
			continue
		}
		tagJson := field.Tag.Get("json")
		if tagJson == "" {
			continue
		}
		jprops := parseJsonTag(tagJson)
		fprops := parseFormTag(tagForm)

		// Check Req
		req := false
		if r, ok := fprops["required"]; ok && r == "true" {
			req = true
		}
		if r, ok := fprops["required-create"]; ok && r == "true" && mode == "create" {
			req = true
		}

		fields = append(fields, FormField{
			Name:      jprops[0],
			Value:     value,
			Label:     fprops["label"],
			Type:      fprops["type"],
			Required:  req,
		})
	}
	return fields, nil
}

func GenerateForm(method string, action string, mode string, input interface{}, title string) (Form, error) {
	fields, err := generateFormFields(input, mode)
	return Form{
		Title : title,
		Action: action,
		Method: method,
		Fields: fields,
	}, err
}