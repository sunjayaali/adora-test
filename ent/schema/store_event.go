package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StoreEvent struct {
	ent.Schema
}

func (StoreEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_id").Unique(),
		field.String("user_id"),
		field.String("type"),
		field.Int64("event_time_ms"),
		field.String("product_id"),
		field.Time("received_at").Default(time.Now),
	}
}

func (StoreEvent) Edges() []ent.Edge {
	return nil
}

func (StoreEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
