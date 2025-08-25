package dbflat

import (
	"gopkg.in/yaml.v3"
)

// FieldSchema describes a single field in a schema-driven struct
// Length: -1 means variable length, >0 means fixed length
// CompFlags: compression/array flags for the field
// Name: optional, for mapping to struct fields (not required for encoding)
type FieldSchema struct {
	Name      string `yaml:"name"`      // optional: struct field name
	Tag       uint16 `yaml:"tag"`       // tag number
	CompFlags uint16 `yaml:"compflags"` // compression/array flags
	Length    int    `yaml:"length"`    // -1 for variable, >0 for fixed
}

// SchemaDef describes the schema for a struct or record
type SchemaDef struct {
	Fields   []FieldSchema  `yaml:"fields"`
	TagToLen map[uint16]int `yaml:"-"` // tag -> length (for fixed-length fields)
}

// NewSchemaDef builds a SchemaDef and populates TagToLen for fixed-length fields
func NewSchemaDef(fields []FieldSchema) *SchemaDef {
	tagToLen := make(map[uint16]int)
	for _, f := range fields {
		if f.Length > 0 {
			tagToLen[f.Tag] = f.Length
		}
	}
	return &SchemaDef{
		Fields:   fields,
		TagToLen: tagToLen,
	}
}

// GetFieldByTag returns the FieldSchema for a given tag, or nil if not found
func (s *SchemaDef) GetFieldByTag(tag uint16) *FieldSchema {
	for i := range s.Fields {
		if s.Fields[i].Tag == tag {
			return &s.Fields[i]
		}
	}
	return nil
}

// LoadSchemaYAML loads a SchemaDef from YAML data (file contents or string)
func LoadSchemaYAML(data []byte) (*SchemaDef, error) {
	var s SchemaDef
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	// Populate TagToLen for fixed-length fields
	s.TagToLen = make(map[uint16]int)
	for _, f := range s.Fields {
		if f.Length > 0 {
			s.TagToLen[f.Tag] = f.Length
		}
	}
	return &s, nil
}
