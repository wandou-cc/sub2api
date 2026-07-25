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

// RechargeLotteryDraw holds one blind-box opportunity and its immutable result.
type RechargeLotteryDraw struct {
	ent.Schema
}

// Annotations maps the entity to the production table name.
func (RechargeLotteryDraw) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "recharge_lottery_draws"},
	}
}

// Fields defines the opportunity snapshot and the result written on claim.
func (RechargeLotteryDraw) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("order_id").
			Unique(),
		field.Float("recharge_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.String("max_rarity").
			MaxLen(20),
		field.String("rarity").
			MaxLen(20).
			Default(""),
		field.Float("reward_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.Float("balance_after").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Optional().
			Nillable(),
		field.Time("claimed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

// Edges links each opportunity to exactly one user and one recharge order.
func (RechargeLotteryDraw) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("recharge_lottery_draws").
			Field("user_id").
			Unique().
			Required(),
		edge.From("order", PaymentOrder.Type).
			Ref("recharge_lottery_draw").
			Field("order_id").
			Unique().
			Required(),
	}
}

// Indexes support pending-opportunity and recent-result queries by user.
func (RechargeLotteryDraw) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "claimed_at"),
		index.Fields("created_at"),
	}
}
