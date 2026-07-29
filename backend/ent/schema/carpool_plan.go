package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CarpoolPlan stores an administrator-configured carpool product.
type CarpoolPlan struct {
	ent.Schema
}

// Annotations maps the entity to the existing SQL naming convention.
func (CarpoolPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "carpool_plans"},
	}
}

// Fields defines the configurable fields and the internal snapshot revision.
func (CarpoolPlan) Fields() []ent.Field {
	return []ent.Field{
		field.Float("total_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Int("target_members"),
		field.String("note").
			NotEmpty().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Int("revision").
			Default(1),
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

// Indexes supports stable plan ordering and administrative lookups by group size.
func (CarpoolPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("target_members"),
		index.Fields("created_at"),
	}
}
