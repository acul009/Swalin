package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Revocation struct {
	ent.Schema
}

func (Revocation) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("revocation"),
		field.String("hash").NotEmpty().Unique(),
		field.Uint64("hasher"),
	}
}

func (Revocation) Indices() []ent.Index {
	return []ent.Index{
		index.Fields("hash", "hasher"),
	}
}
