package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Device holds the schema definition for the Device entity.
type Device struct {
	ent.Schema
}

// Fields of the Device.
func (Device) Fields() []ent.Field {
	return []ent.Field{
		field.String("public_key").NotEmpty().Unique().Immutable(),
		field.String("certificate").NotEmpty().Unique(),
	}
}

// Edges of the Device.
func (Device) Edges() []ent.Edge {
	return nil
}
