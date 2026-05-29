package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Entitlement struct {
	ent.Schema
}

func (Entitlement) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Unique(),
		field.String("source"),
		field.Bool("is_active"),
		field.Time("expires_at"),
		field.Time("last_changed_at").Optional().Nillable(),
		field.String("reason"),
	}
}

func (Entitlement) Edges() []ent.Edge {
	return nil
}

func (Entitlement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source"),
	}
}
