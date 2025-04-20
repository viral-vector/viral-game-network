package form_builder

import (
	"fmt"
	"strconv"
	"reflect"
	"strings"
	dbtype "viral-game-network/src/database/type"
	"github.com/surrealdb/surrealdb.go/pkg/models"
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
	Submit string
	Confirm string
	Action string
	Method string
	Fields []FormField
	CanReset  bool
	DisableSubmit bool
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

	recordIDType := reflect.TypeOf(models.RecordID{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

        fieldValue := v.Field(i)
        var valueStr string

		if (fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Interface) && fieldValue.IsNil() {
			valueStr = ""
		} else if fieldValue.Type() == recordIDType {
			// If the field is exactly of type models.RecordID.
			ident := fieldValue.Interface().(models.RecordID)
			valueStr = ident.String()
		} else if ident, ok := fieldValue.Interface().(dbtype.Model); ok {
			// Field implements our Model interface.
			valueStr = ident.ModelID()
		} else if s, ok := fieldValue.Interface().(string); ok {
			valueStr = s
		} else {
			// Fallback: use fmt.Sprintf.
			valueStr = fmt.Sprintf("%v", fieldValue.Interface())
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
		if r, ok := fprops["required-update"]; ok && r == "true" && mode == "update" {
			req = true
		}

		readonly := false
		if r, ok := fprops["readonly"]; ok && r == "true" {
			readonly = true
		}
		if r, ok := fprops["readonly-update"]; ok && r == "true" && mode == "update" {
			readonly = true
		}

		sortorder := int32(0)
		if r, ok := fprops["sortorder"]; ok {
			if i, err := strconv.Atoi(r); err == nil {
				sortorder = int32(i)
			}
		}

		fields = append(fields, FormField{
			Name:      jprops[0],
			Value:     valueStr,
			Label:     fprops["label"],
			Type:      fprops["type"],
			Required:  req,
			ReadOnly:  readonly,
			SortOrder: sortorder,
			Options:   nil,
		})
	}
	return fields, nil
}

func GenerateForm(
	method string, 
	action string, 
	mode string, 
	input interface{}, 
	title string,
	submit string,
) (*Form, error) {
	fields, err := generateFormFields(input, mode)

	if submit == "" {
	   	submit = "Submit"
	}

	return &Form{
		Title : title,
		Submit: submit,
		Action: action,
		Method: method,
		Fields: fields,
		CanReset: true,
		Confirm: "",
		DisableSubmit: false,
	}, err
}