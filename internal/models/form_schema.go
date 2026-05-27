package models

// LocalizedString holds translated text keyed by language code (e.g. "vi", "en").
type LocalizedString map[string]string

// Resolve returns the text for lang, falling back to "vi", then any available value.
func (ls LocalizedString) Resolve(lang string) string {
	if v, ok := ls[lang]; ok && v != "" {
		return v
	}
	if v, ok := ls["vi"]; ok && v != "" {
		return v
	}
	for _, v := range ls {
		if v != "" {
			return v
		}
	}
	return ""
}

// FormFieldType is the input type of a form field.
type FormFieldType string

const (
	FormFieldText     FormFieldType = "text"
	FormFieldTextarea FormFieldType = "textarea"
	FormFieldNumber   FormFieldType = "number"
	FormFieldDate     FormFieldType = "date"
	FormFieldSelect   FormFieldType = "select"
	FormFieldCheckbox FormFieldType = "checkbox"
)

// FormFieldOption is one choice for a select field.
type FormFieldOption struct {
	Value string          `json:"value"`
	Label LocalizedString `json:"label"`
}

// FormField describes a single input in a service type's application form.
type FormField struct {
	Key         string            `json:"key"`
	Type        FormFieldType     `json:"type"`
	Required    bool              `json:"required"`
	Title       LocalizedString   `json:"title"`
	Placeholder LocalizedString   `json:"placeholder,omitempty"`
	HelpText    LocalizedString   `json:"help_text,omitempty"`
	Options     []FormFieldOption `json:"options,omitempty"`
}

// FormSchemaV2 is the structured form schema stored in service_types.form_schema.
// Each field carries its own localized labels, replacing the legacy flat-string format.
type FormSchemaV2 struct {
	Fields []FormField `json:"fields"`
}
