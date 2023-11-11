package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Device holds the schema definition for the Device entity.
type HostConfig struct {
	ent.Schema
}

func (HostConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("config").NotEmpty(),
		field.String("type").NotEmpty(),
	}
}

func (HostConfig) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("device", Device.Type).Ref("configs").Unique().Required(),
	}
}

func (HostConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type").Edges("device").Unique(),
	}
}
