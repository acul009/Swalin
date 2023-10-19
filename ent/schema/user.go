package schema

import (
	"rahnit-rmm/util"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique().NotEmpty(),
		field.JSON("password_client_hashing_options", &util.ArgonParameters{}).Sensitive(),
		field.JSON("password_server_hashing_options", &util.ArgonParameters{}).Sensitive(),
		field.String("password_double_hashed").NotEmpty().Sensitive(),
		field.String("certificate").NotEmpty().Unique(),
		field.String("public_key").NotEmpty().Unique().Immutable(),
		field.String("encrypted_private_key").NotEmpty().Sensitive(),
		field.String("totp_secret").Sensitive(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username"),
		index.Fields("public_key"),
	}
}
