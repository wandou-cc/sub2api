package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CarpoolGroup stores one manually fulfilled shared-account subscription group.
type CarpoolGroup struct {
	ent.Schema
}

func (CarpoolGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "carpool_groups"},
	}
}

func (CarpoolGroup) Fields() []ent.Field {
	return []ent.Field{
		// Plan fields are immutable snapshots so history remains valid after plan edits or deletion.
		field.Int64("carpool_plan_id"),
		field.Int("carpool_plan_revision"),
		field.Int("target_members"),
		field.Float("total_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("price_per_member").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.String("plan_note").
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Int("member_count").
			Default(0),
		field.String("status").
			MaxLen(30),
		field.String("open_key").
			Optional().
			Nillable().
			MaxLen(32).
			Unique(),
		field.String("status_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("deadline_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("formed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("opened_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (CarpoolGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("orders", PaymentOrder.Type),
	}
}

func (CarpoolGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("carpool_plan_id"),
		index.Fields("status"),
		index.Fields("deadline_at"),
		index.Fields("expires_at"),
		index.Fields("created_at"),
	}
}
