package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
	return []ent.Edge{
		edge.To("configs", HostConfig.Type),
	}
}

// Indexes of the User.
func (Device) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("public_key"),
	}
}
