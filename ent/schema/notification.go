package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Notification struct {
	ent.Schema
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty(),
		field.String("type"),
		field.Time("expires_at"),
		field.Time("scheduled_for"),
		field.Time("sent_at").Optional().Nillable(),
	}
}

func (Notification) Edges() []ent.Edge {
	return nil
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "type", "expires_at").Unique(),
	}
}
