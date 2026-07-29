package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/carpoolgroup"
	"github.com/Wei-Shaw/sub2api/ent/carpoolplan"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	CarpoolStatusWaiting       = "waiting"
	CarpoolStatusPurchasing    = "purchasing"
	CarpoolStatusActive        = "active"
	CarpoolStatusRefundPending = "refund_pending"
	CarpoolStatusRefunded      = "refunded"
	CarpoolStatusExpired       = "expired"

	carpoolFormationWindow = 48 * time.Hour
	carpoolServiceDuration = 30 * 24 * time.Hour
)

type CarpoolPlanOverview struct {
	ID               int64      `json:"id"`
	TotalAmount      float64    `json:"total_amount"`
	Size             int        `json:"size"`
	Price            float64    `json:"price"`
	Note             string     `json:"note"`
	CurrentMembers   int        `json:"current_members"`
	RemainingMembers int        `json:"remaining_members"`
	DeadlineAt       *time.Time `json:"deadline_at,omitempty"`
}

type CarpoolGroupOverview struct {
	ID             int64      `json:"id"`
	OrderID        int64      `json:"order_id,omitempty"`
	CarpoolPlanID  int64      `json:"carpool_plan_id"`
	TargetMembers  int        `json:"target_members"`
	TotalAmount    float64    `json:"total_amount"`
	PricePerMember float64    `json:"price_per_member"`
	PlanNote       string     `json:"plan_note"`
	CurrentMembers int        `json:"current_members"`
	Status         string     `json:"status"`
	StatusReason   *string    `json:"status_reason,omitempty"`
	DeadlineAt     *time.Time `json:"deadline_at,omitempty"`
	FormedAt       *time.Time `json:"formed_at,omitempty"`
	OpenedAt       *time.Time `json:"opened_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CarpoolOverview struct {
	Plans    []CarpoolPlanOverview  `json:"plans"`
	MyGroups []CarpoolGroupOverview `json:"my_groups"`
}

type CarpoolPlanInput struct {
	TotalAmount   float64 `json:"total_amount"`
	TargetMembers int     `json:"target_members"`
	Note          string  `json:"note"`
}

type carpoolPlanSnapshot struct {
	planID         int64
	revision       int
	targetMembers  int
	totalAmount    float64
	pricePerMember float64
	note           string
}

// carpoolPricePerMember validates the CNY total and rounds each member's share to cents.
func carpoolPricePerMember(totalAmount float64, targetMembers int) (float64, error) {
	if math.IsNaN(totalAmount) || math.IsInf(totalAmount, 0) || totalAmount <= 0 {
		return 0, infraerrors.BadRequest("INVALID_CARPOOL_PLAN_AMOUNT", "carpool plan total amount must be positive")
	}
	if targetMembers <= 0 {
		return 0, infraerrors.BadRequest("INVALID_CARPOOL_PLAN_MEMBERS", "carpool plan target members must be positive")
	}
	amount := decimal.NewFromFloat(totalAmount)
	if !amount.Equal(amount.Round(2)) {
		return 0, infraerrors.BadRequest("INVALID_CARPOOL_PLAN_AMOUNT", "carpool plan total amount supports at most two decimal places")
	}
	return amount.Div(decimal.NewFromInt(int64(targetMembers))).Round(2).InexactFloat64(), nil
}

// ListCarpoolPlans returns every configured plan in descending group-size order.
func (s *PaymentService) ListCarpoolPlans(ctx context.Context) ([]*dbent.CarpoolPlan, error) {
	plans, err := s.entClient.CarpoolPlan.Query().
		Order(dbent.Desc(carpoolplan.FieldTargetMembers), dbent.Asc(carpoolplan.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list carpool plans: %w", err)
	}
	return plans, nil
}

// CreateCarpoolPlan validates and persists the three administrator-configurable fields.
func (s *PaymentService) CreateCarpoolPlan(ctx context.Context, req CarpoolPlanInput) (*dbent.CarpoolPlan, error) {
	if _, err := carpoolPricePerMember(req.TotalAmount, req.TargetMembers); err != nil {
		return nil, err
	}
	note := strings.TrimSpace(req.Note)
	if note == "" {
		return nil, infraerrors.BadRequest("INVALID_CARPOOL_PLAN_NOTE", "carpool plan note is required")
	}
	plan, err := s.entClient.CarpoolPlan.Create().
		SetTotalAmount(req.TotalAmount).
		SetTargetMembers(req.TargetMembers).
		SetNote(note).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create carpool plan: %w", err)
	}
	return plan, nil
}

// UpdateCarpoolPlan locks the plan and rejects edits that could change an active purchase or waiting group.
func (s *PaymentService) UpdateCarpoolPlan(ctx context.Context, id int64, req CarpoolPlanInput) (*dbent.CarpoolPlan, error) {
	if _, err := carpoolPricePerMember(req.TotalAmount, req.TargetMembers); err != nil {
		return nil, err
	}
	note := strings.TrimSpace(req.Note)
	if note == "" {
		return nil, infraerrors.BadRequest("INVALID_CARPOOL_PLAN_NOTE", "carpool plan note is required")
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin carpool plan update transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	plan, err := tx.CarpoolPlan.Query().Where(carpoolplan.IDEQ(id)).ForUpdate().Only(ctx)
	if dbent.IsNotFound(err) {
		return nil, infraerrors.NotFound("CARPOOL_PLAN_NOT_FOUND", "carpool plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("lock carpool plan: %w", err)
	}
	unfinishedOrder, err := tx.PaymentOrder.Query().Where(
		paymentorder.OrderTypeEQ(payment.OrderTypeCarpool),
		paymentorder.CarpoolPlanIDEQ(id),
		paymentorder.CarpoolGroupIDIsNil(),
		paymentorder.Or(
			paymentorder.StatusIn(OrderStatusPending, OrderStatusPaid, OrderStatusRecharging),
			paymentorder.And(
				paymentorder.StatusEQ(OrderStatusFailed),
				paymentorder.PaidAtNotNil(),
			),
		),
	).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check carpool plan orders: %w", err)
	}
	waitingGroup, err := tx.CarpoolGroup.Query().Where(
		carpoolgroup.CarpoolPlanIDEQ(id),
		carpoolgroup.StatusEQ(CarpoolStatusWaiting),
	).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check carpool plan groups: %w", err)
	}
	if unfinishedOrder || waitingGroup {
		return nil, infraerrors.Conflict("CARPOOL_PLAN_IN_USE", "carpool plan has an active order or waiting group")
	}
	updated, err := tx.CarpoolPlan.UpdateOneID(plan.ID).
		SetTotalAmount(req.TotalAmount).
		SetTargetMembers(req.TargetMembers).
		SetNote(note).
		AddRevision(1).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update carpool plan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit carpool plan update: %w", err)
	}
	return updated, nil
}

// DeleteCarpoolPlan removes only unused configuration while historical snapshots remain intact.
func (s *PaymentService) DeleteCarpoolPlan(ctx context.Context, id int64) error {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin carpool plan delete transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.CarpoolPlan.Query().Where(carpoolplan.IDEQ(id)).ForUpdate().Only(ctx); dbent.IsNotFound(err) {
		return infraerrors.NotFound("CARPOOL_PLAN_NOT_FOUND", "carpool plan not found")
	} else if err != nil {
		return fmt.Errorf("lock carpool plan: %w", err)
	}
	unfinishedOrder, err := tx.PaymentOrder.Query().Where(
		paymentorder.OrderTypeEQ(payment.OrderTypeCarpool),
		paymentorder.CarpoolPlanIDEQ(id),
		paymentorder.CarpoolGroupIDIsNil(),
		paymentorder.Or(
			paymentorder.StatusIn(OrderStatusPending, OrderStatusPaid, OrderStatusRecharging),
			paymentorder.And(
				paymentorder.StatusEQ(OrderStatusFailed),
				paymentorder.PaidAtNotNil(),
			),
		),
	).Exist(ctx)
	if err != nil {
		return fmt.Errorf("check carpool plan orders: %w", err)
	}
	waitingGroup, err := tx.CarpoolGroup.Query().Where(
		carpoolgroup.CarpoolPlanIDEQ(id),
		carpoolgroup.StatusEQ(CarpoolStatusWaiting),
	).Exist(ctx)
	if err != nil {
		return fmt.Errorf("check carpool plan groups: %w", err)
	}
	if unfinishedOrder || waitingGroup {
		return infraerrors.Conflict("CARPOOL_PLAN_IN_USE", "carpool plan has an active order or waiting group")
	}
	if err := tx.CarpoolPlan.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete carpool plan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit carpool plan delete: %w", err)
	}
	return nil
}

// validateCarpoolOrder resolves the server-owned plan before payment calculations begin.
func (s *PaymentService) validateCarpoolOrder(ctx context.Context, req CreateOrderRequest) (*dbent.CarpoolPlan, error) {
	if req.CarpoolPlanID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CARPOOL_PLAN", "carpool order requires a plan")
	}
	if payment.GetBasePaymentType(req.PaymentType) != payment.TypeAlipay {
		return nil, infraerrors.BadRequest("CARPOOL_ALIPAY_ONLY", "carpool subscriptions only support Alipay")
	}
	if _, err := s.ExpireCarpoolGroups(ctx); err != nil {
		return nil, err
	}
	plan, err := s.entClient.CarpoolPlan.Get(ctx, req.CarpoolPlanID)
	if dbent.IsNotFound(err) {
		return nil, infraerrors.NotFound("INVALID_CARPOOL_PLAN", "carpool plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get carpool plan: %w", err)
	}
	if _, err := carpoolPricePerMember(plan.TotalAmount, plan.TargetMembers); err != nil || plan.Revision <= 0 || strings.TrimSpace(plan.Note) == "" {
		return nil, infraerrors.BadRequest("INVALID_CARPOOL_PLAN", "carpool plan configuration is invalid")
	}
	return plan, nil
}

// ExecuteCarpoolFulfillment assigns a paid order to one group without crediting user balance.
func (s *PaymentService) ExecuteCarpoolFulfillment(ctx context.Context, orderID int64) error {
	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if order.Status == OrderStatusCompleted {
		return nil
	}
	if psIsRefundStatus(order.Status) {
		return infraerrors.BadRequest("INVALID_STATUS", "refund-related order cannot fulfill")
	}
	if order.Status != OrderStatusPaid && order.Status != OrderStatusFailed && order.Status != OrderStatusRecharging {
		return infraerrors.BadRequest("INVALID_STATUS", "order cannot fulfill in status "+order.Status)
	}
	snapshot, err := carpoolSnapshotFromOrder(order)
	if err != nil {
		return err
	}
	lease, err := s.acquirePaymentFulfillmentLease(ctx, order)
	if err != nil {
		return err
	}
	if lease == nil {
		return nil
	}
	if err := s.assignCarpoolOrder(ctx, orderID, snapshot, lease); err != nil {
		s.markFailed(ctx, orderID, lease, err)
		return err
	}
	return nil
}

// carpoolSnapshotFromOrder validates the immutable plan fields captured when the order was created.
func carpoolSnapshotFromOrder(order *dbent.PaymentOrder) (carpoolPlanSnapshot, error) {
	if order.CarpoolPlanID == nil || order.CarpoolPlanRevision == nil || order.CarpoolSize == nil || order.CarpoolTotalAmount == nil || order.CarpoolPlanNote == nil {
		return carpoolPlanSnapshot{}, infraerrors.BadRequest("INVALID_CARPOOL_PLAN", "carpool order is missing its plan snapshot")
	}
	price, err := carpoolPricePerMember(*order.CarpoolTotalAmount, *order.CarpoolSize)
	if err != nil || *order.CarpoolPlanRevision <= 0 || strings.TrimSpace(*order.CarpoolPlanNote) == "" || order.Amount != price {
		return carpoolPlanSnapshot{}, infraerrors.BadRequest("INVALID_CARPOOL_PLAN", "carpool order plan snapshot is invalid")
	}
	return carpoolPlanSnapshot{
		planID:         *order.CarpoolPlanID,
		revision:       *order.CarpoolPlanRevision,
		targetMembers:  *order.CarpoolSize,
		totalAmount:    *order.CarpoolTotalAmount,
		pricePerMember: price,
		note:           *order.CarpoolPlanNote,
	}, nil
}

// assignCarpoolOrder serializes membership changes and completes the paid order in one transaction.
func (s *PaymentService) assignCarpoolOrder(ctx context.Context, orderID int64, snapshot carpoolPlanSnapshot, lease *paymentFulfillmentLease) error {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin carpool fulfillment transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	order, err := tx.PaymentOrder.Query().Where(
		paymentorder.IDEQ(orderID),
		paymentorder.StatusEQ(OrderStatusRecharging),
		paymentorder.UpdatedAtEQ(lease.version),
	).ForUpdate().Only(ctx)
	if err != nil {
		return fmt.Errorf("lock carpool order: %w", err)
	}
	now := time.Now()
	if order.CarpoolGroupID != nil {
		if _, err := tx.PaymentOrder.UpdateOneID(orderID).SetStatus(OrderStatusCompleted).SetCompletedAt(now).Save(ctx); err != nil {
			return fmt.Errorf("complete assigned carpool order: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit assigned carpool order: %w", err)
		}
		return nil
	}
	plan, err := tx.CarpoolPlan.Query().Where(carpoolplan.IDEQ(snapshot.planID)).ForUpdate().Only(ctx)
	if dbent.IsNotFound(err) {
		return infraerrors.Conflict("CARPOOL_PLAN_CHANGED", "carpool plan changed before fulfillment")
	}
	if err != nil {
		return fmt.Errorf("lock carpool plan for fulfillment: %w", err)
	}
	if plan.Revision != snapshot.revision || plan.TargetMembers != snapshot.targetMembers || plan.TotalAmount != snapshot.totalAmount || plan.Note != snapshot.note {
		return infraerrors.Conflict("CARPOOL_PLAN_CHANGED", "carpool plan changed before fulfillment")
	}

	group, err := s.getOrCreateOpenCarpoolGroup(ctx, tx, snapshot, now)
	if err != nil {
		return err
	}
	alreadyJoined, err := group.QueryOrders().Where(paymentorder.UserIDEQ(order.UserID)).Exist(ctx)
	if err != nil {
		return fmt.Errorf("check carpool member: %w", err)
	}
	if alreadyJoined {
		return infraerrors.Conflict("CARPOOL_ALREADY_JOINED", "the user already has a seat in the current carpool group")
	}

	nextCount := group.MemberCount + 1
	groupUpdate := tx.CarpoolGroup.UpdateOneID(group.ID).SetMemberCount(nextCount)
	if nextCount == group.TargetMembers {
		groupUpdate.SetStatus(CarpoolStatusPurchasing).SetFormedAt(now).ClearOpenKey()
	}
	if _, err := groupUpdate.Save(ctx); err != nil {
		return fmt.Errorf("update carpool progress: %w", err)
	}
	if _, err := tx.PaymentOrder.UpdateOneID(orderID).
		SetCarpoolGroupID(group.ID).
		SetStatus(OrderStatusCompleted).
		SetCompletedAt(now).
		ClearFailedAt().
		ClearFailedReason().
		Save(ctx); err != nil {
		return fmt.Errorf("complete carpool order: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit carpool fulfillment: %w", err)
	}
	s.writeAuditLog(ctx, orderID, "CARPOOL_JOINED", "system", map[string]any{
		"carpoolGroupID": group.ID,
		"carpoolPlanID":  snapshot.planID,
		"targetMembers":  snapshot.targetMembers,
		"currentMembers": nextCount,
	})
	return nil
}

// getOrCreateOpenCarpoolGroup returns the single current group for a plan under a row lock.
func (s *PaymentService) getOrCreateOpenCarpoolGroup(ctx context.Context, tx *dbent.Tx, snapshot carpoolPlanSnapshot, now time.Time) (*dbent.CarpoolGroup, error) {
	if snapshot.targetMembers == 1 {
		group, err := tx.CarpoolGroup.Create().
			SetCarpoolPlanID(snapshot.planID).
			SetCarpoolPlanRevision(snapshot.revision).
			SetTargetMembers(1).
			SetTotalAmount(snapshot.totalAmount).
			SetPricePerMember(snapshot.pricePerMember).
			SetPlanNote(snapshot.note).
			SetMemberCount(0).
			SetStatus(CarpoolStatusWaiting).
			SetDeadlineAt(now).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("create single-member carpool group: %w", err)
		}
		return group, nil
	}

	openKey := strconv.FormatInt(snapshot.planID, 10)
	for attempt := 0; attempt < 2; attempt++ {
		groupID, err := tx.CarpoolGroup.Create().
			SetCarpoolPlanID(snapshot.planID).
			SetCarpoolPlanRevision(snapshot.revision).
			SetTargetMembers(snapshot.targetMembers).
			SetTotalAmount(snapshot.totalAmount).
			SetPricePerMember(snapshot.pricePerMember).
			SetPlanNote(snapshot.note).
			SetMemberCount(0).
			SetStatus(CarpoolStatusWaiting).
			SetOpenKey(openKey).
			SetDeadlineAt(now.Add(carpoolFormationWindow)).
			OnConflictColumns(carpoolgroup.FieldOpenKey).
			SetOpenKey(openKey).
			ID(ctx)
		if err != nil {
			return nil, fmt.Errorf("create current carpool group: %w", err)
		}
		group, err := tx.CarpoolGroup.Query().Where(carpoolgroup.IDEQ(groupID)).ForUpdate().Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("lock current carpool group: %w", err)
		}
		if group.Status == CarpoolStatusWaiting && group.DeadlineAt != nil && group.DeadlineAt.After(now) {
			if group.CarpoolPlanID != snapshot.planID || group.CarpoolPlanRevision != snapshot.revision || group.TargetMembers != snapshot.targetMembers || group.TotalAmount != snapshot.totalAmount || group.PricePerMember != snapshot.pricePerMember || group.PlanNote != snapshot.note {
				return nil, infraerrors.Conflict("CARPOOL_PLAN_CHANGED", "current carpool group plan snapshot does not match the order")
			}
			return group, nil
		}
		if group.Status == CarpoolStatusWaiting {
			if _, err := tx.CarpoolGroup.UpdateOneID(group.ID).
				SetStatus(CarpoolStatusRefundPending).
				SetStatusReason("not_formed").
				ClearOpenKey().
				Save(ctx); err != nil {
				return nil, fmt.Errorf("expire current carpool group: %w", err)
			}
			continue
		}
		return nil, infraerrors.Conflict("CARPOOL_GROUP_CHANGED", "current carpool group changed while joining")
	}
	return nil, infraerrors.Conflict("CARPOOL_GROUP_CHANGED", "failed to allocate a current carpool group")
}

// ExpireCarpoolGroups advances formation failures and completed service periods.
func (s *PaymentService) ExpireCarpoolGroups(ctx context.Context) (int, error) {
	now := time.Now()
	waiting, err := s.entClient.CarpoolGroup.Update().Where(
		carpoolgroup.StatusEQ(CarpoolStatusWaiting),
		carpoolgroup.DeadlineAtLTE(now),
	).SetStatus(CarpoolStatusRefundPending).SetStatusReason("not_formed").ClearOpenKey().Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("expire waiting carpool groups: %w", err)
	}
	active, err := s.entClient.CarpoolGroup.Update().Where(
		carpoolgroup.StatusEQ(CarpoolStatusActive),
		carpoolgroup.ExpiresAtLTE(now),
	).SetStatus(CarpoolStatusExpired).Save(ctx)
	if err != nil {
		return waiting, fmt.Errorf("expire active carpool groups: %w", err)
	}
	if err := s.syncCarpoolRefundedGroups(ctx); err != nil {
		return waiting + active, err
	}
	return waiting + active, nil
}

// GetCarpoolOverview returns configured plans, live progress, and the current user's records.
func (s *PaymentService) GetCarpoolOverview(ctx context.Context, userID int64) (*CarpoolOverview, error) {
	if _, err := s.ExpireCarpoolGroups(ctx); err != nil {
		return nil, err
	}
	openGroups, err := s.entClient.CarpoolGroup.Query().Where(carpoolgroup.StatusEQ(CarpoolStatusWaiting)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list current carpool groups: %w", err)
	}
	openByPlan := make(map[int64]*dbent.CarpoolGroup, len(openGroups))
	for _, group := range openGroups {
		openByPlan[group.CarpoolPlanID] = group
	}
	configuredPlans, err := s.ListCarpoolPlans(ctx)
	if err != nil {
		return nil, err
	}
	plans := make([]CarpoolPlanOverview, 0, len(configuredPlans))
	for _, configuredPlan := range configuredPlans {
		price, err := carpoolPricePerMember(configuredPlan.TotalAmount, configuredPlan.TargetMembers)
		if err != nil || configuredPlan.Revision <= 0 || strings.TrimSpace(configuredPlan.Note) == "" {
			return nil, fmt.Errorf("carpool plan %d has invalid configuration", configuredPlan.ID)
		}
		plan := CarpoolPlanOverview{
			ID:               configuredPlan.ID,
			TotalAmount:      configuredPlan.TotalAmount,
			Size:             configuredPlan.TargetMembers,
			Price:            price,
			Note:             configuredPlan.Note,
			RemainingMembers: configuredPlan.TargetMembers,
		}
		if group := openByPlan[configuredPlan.ID]; group != nil {
			plan.CurrentMembers = group.MemberCount
			plan.RemainingMembers = configuredPlan.TargetMembers - group.MemberCount
			plan.DeadlineAt = group.DeadlineAt
		}
		plans = append(plans, plan)
	}
	groups, err := s.entClient.CarpoolGroup.Query().Where(
		carpoolgroup.HasOrdersWith(paymentorder.UserIDEQ(userID)),
	).WithOrders(func(query *dbent.PaymentOrderQuery) {
		query.Where(paymentorder.UserIDEQ(userID))
	}).Order(dbent.Desc(carpoolgroup.FieldCreatedAt)).Limit(20).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list user carpool groups: %w", err)
	}
	myGroups := make([]CarpoolGroupOverview, 0, len(groups))
	for _, group := range groups {
		item := CarpoolGroupOverview{
			ID:             group.ID,
			CarpoolPlanID:  group.CarpoolPlanID,
			TargetMembers:  group.TargetMembers,
			TotalAmount:    group.TotalAmount,
			PricePerMember: group.PricePerMember,
			PlanNote:       group.PlanNote,
			CurrentMembers: group.MemberCount,
			Status:         group.Status,
			StatusReason:   group.StatusReason,
			DeadlineAt:     group.DeadlineAt,
			FormedAt:       group.FormedAt,
			OpenedAt:       group.OpenedAt,
			ExpiresAt:      group.ExpiresAt,
			CreatedAt:      group.CreatedAt,
		}
		if len(group.Edges.Orders) == 1 {
			item.OrderID = group.Edges.Orders[0].ID
		}
		myGroups = append(myGroups, item)
	}
	return &CarpoolOverview{Plans: plans, MyGroups: myGroups}, nil
}

// ListCarpoolGroups returns current or completed groups with their paid member orders.
func (s *PaymentService) ListCarpoolGroups(ctx context.Context, history bool) ([]*dbent.CarpoolGroup, error) {
	if _, err := s.ExpireCarpoolGroups(ctx); err != nil {
		return nil, err
	}
	finalStatuses := []string{CarpoolStatusRefunded, CarpoolStatusExpired}
	query := s.entClient.CarpoolGroup.Query()
	if history {
		query.Where(carpoolgroup.StatusIn(finalStatuses...))
	} else {
		query.Where(carpoolgroup.StatusNotIn(finalStatuses...))
	}
	groups, err := query.WithOrders(func(orders *dbent.PaymentOrderQuery) {
		orders.Order(dbent.Asc(paymentorder.FieldCreatedAt))
	}).Order(dbent.Desc(carpoolgroup.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list carpool groups: %w", err)
	}
	return groups, nil
}

// OpenCarpoolGroup records delivery at confirmation time and applies the fixed 30-day service period.
func (s *PaymentService) OpenCarpoolGroup(ctx context.Context, groupID int64) error {
	openedAt := time.Now()
	expiresAt := openedAt.Add(carpoolServiceDuration)
	updated, err := s.entClient.CarpoolGroup.Update().Where(
		carpoolgroup.IDEQ(groupID),
		carpoolgroup.StatusEQ(CarpoolStatusPurchasing),
	).SetStatus(CarpoolStatusActive).SetOpenedAt(openedAt).SetExpiresAt(expiresAt).Save(ctx)
	if err != nil {
		return fmt.Errorf("open carpool group: %w", err)
	}
	if updated == 0 {
		return infraerrors.Conflict("INVALID_CARPOOL_STATUS", "only purchasing carpool groups can be opened")
	}
	return nil
}

// MarkCarpoolRefundPending closes a group and exposes manual per-order refunds to administrators.
func (s *PaymentService) MarkCarpoolRefundPending(ctx context.Context, groupID int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "refund reason is required")
	}
	updated, err := s.entClient.CarpoolGroup.Update().Where(
		carpoolgroup.IDEQ(groupID),
		carpoolgroup.StatusIn(CarpoolStatusWaiting, CarpoolStatusPurchasing, CarpoolStatusActive),
	).SetStatus(CarpoolStatusRefundPending).SetStatusReason(reason).ClearOpenKey().Save(ctx)
	if err != nil {
		return fmt.Errorf("mark carpool refund pending: %w", err)
	}
	if updated == 0 {
		return infraerrors.Conflict("INVALID_CARPOOL_STATUS", "carpool group cannot enter refund pending from its current status")
	}
	return nil
}

// syncCarpoolRefundedGroups closes groups after every member order has a recorded refund.
func (s *PaymentService) syncCarpoolRefundedGroups(ctx context.Context) error {
	groups, err := s.entClient.CarpoolGroup.Query().Where(carpoolgroup.StatusEQ(CarpoolStatusRefundPending)).WithOrders().All(ctx)
	if err != nil {
		return fmt.Errorf("list refund-pending carpool groups: %w", err)
	}
	for _, group := range groups {
		if len(group.Edges.Orders) == 0 {
			continue
		}
		allRefunded := true
		for _, order := range group.Edges.Orders {
			if order.Status != OrderStatusRefunded && order.Status != OrderStatusPartiallyRefunded {
				allRefunded = false
				break
			}
		}
		if !allRefunded {
			continue
		}
		if _, err := s.entClient.CarpoolGroup.UpdateOneID(group.ID).SetStatus(CarpoolStatusRefunded).Save(ctx); err != nil {
			return fmt.Errorf("complete carpool refunds: %w", err)
		}
	}
	return nil
}
