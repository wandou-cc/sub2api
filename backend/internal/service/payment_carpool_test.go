//go:build unit

package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateCarpoolOrderRejectsUnsupportedPlanAndMethod(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	svc := &PaymentService{entClient: client}

	_, err := svc.validateCarpoolOrder(ctx, CreateOrderRequest{
		UserID:        1,
		PaymentType:   payment.TypeAlipay,
		CarpoolPlanID: 0,
	})
	require.Equal(t, "INVALID_CARPOOL_PLAN", infraerrors.Reason(err))

	plan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	_, err = svc.validateCarpoolOrder(ctx, CreateOrderRequest{
		UserID:        1,
		PaymentType:   payment.TypeWxpay,
		CarpoolPlanID: plan.ID,
	})
	require.Equal(t, "CARPOOL_ALIPAY_ONLY", infraerrors.Reason(err))
}

func TestCreateCarpoolOrderTransactionRejectsUnassignedPaidOrdersIncludingFailedFulfillment(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	user := createCarpoolTestUser(t, ctx, client, "processing", 0)
	plan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	order := createPaidCarpoolTestOrder(t, ctx, client, user, "processing", plan, 800)
	svc := &PaymentService{entClient: client}
	req := CreateOrderRequest{
		UserID:        user.ID,
		PaymentType:   payment.TypeAlipay,
		OrderType:     payment.OrderTypeCarpool,
		CarpoolPlanID: plan.ID,
		ClientIP:      "127.0.0.1",
		SrcHost:       "app.example.com",
	}
	serviceUser := &User{ID: user.ID, Email: user.Email, Username: user.Username}

	_, err := svc.createOrderInTx(ctx, req, serviceUser, nil, plan, &PaymentConfig{MaxPendingOrders: 10, OrderTimeoutMin: 15}, 800, 800, 0, 800, nil)
	require.Equal(t, "CARPOOL_PENDING_ORDER_EXISTS", infraerrors.Reason(err))

	client.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusFailed).SaveX(ctx)
	_, err = svc.createOrderInTx(ctx, req, serviceUser, nil, plan, &PaymentConfig{MaxPendingOrders: 10, OrderTimeoutMin: 15}, 800, 800, 0, 800, nil)
	require.Equal(t, "CARPOOL_PENDING_ORDER_EXISTS", infraerrors.Reason(err))
}

func TestCreateCarpoolOrderRejectsNonCNYAlipayChannel(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	configService := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{
		values: map[string]string{SettingPaymentEnabled: "true"},
	}, nil)
	plan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	svc := &PaymentService{
		entClient:     client,
		configService: configService,
		loadBalancer:  carpoolCurrencyLoadBalancerStub{},
		userRepo: carpoolUserRepositoryStub{
			user: &User{ID: 99, Email: "carpool-currency@example.com", Username: "currency", Status: payment.EntityStatusActive},
		},
	}

	_, err := svc.CreateOrder(ctx, CreateOrderRequest{
		UserID:        99,
		Amount:        1,
		PaymentType:   payment.TypeAlipay,
		OrderType:     payment.OrderTypeCarpool,
		CarpoolPlanID: plan.ID,
	})
	require.Equal(t, "CARPOOL_CNY_ONLY", infraerrors.Reason(err))
}

func TestExecuteCarpoolFulfillmentRejectsTamperedPrice(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	user := createCarpoolTestUser(t, ctx, client, "tampered", 37)
	plan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	order := createPaidCarpoolTestOrder(t, ctx, client, user, "tampered", plan, 799)
	svc := &PaymentService{entClient: client}

	err := svc.ExecuteCarpoolFulfillment(ctx, order.ID)
	require.Equal(t, "INVALID_CARPOOL_PLAN", infraerrors.Reason(err))
	require.Equal(t, 0, client.CarpoolGroup.Query().CountX(ctx))
	require.Equal(t, 37.0, client.User.GetX(ctx, user.ID).Balance)
}

func TestExecuteCarpoolFulfillmentFillsSharedGroupWithoutCreditingBalance(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	firstUser := createCarpoolTestUser(t, ctx, client, "first", 25)
	secondUser := createCarpoolTestUser(t, ctx, client, "second", 40)
	plan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	firstOrder := createPaidCarpoolTestOrder(t, ctx, client, firstUser, "first", plan, 800)
	secondOrder := createPaidCarpoolTestOrder(t, ctx, client, secondUser, "second", plan, 800)
	svc := &PaymentService{entClient: client}
	startedAt := time.Now()

	require.NoError(t, svc.ExecuteCarpoolFulfillment(ctx, firstOrder.ID))
	firstOrder = client.PaymentOrder.GetX(ctx, firstOrder.ID)
	require.NotNil(t, firstOrder.CarpoolGroupID)
	group := client.CarpoolGroup.GetX(ctx, *firstOrder.CarpoolGroupID)
	require.Equal(t, CarpoolStatusWaiting, group.Status)
	require.Equal(t, 1, group.MemberCount)
	require.NotNil(t, group.DeadlineAt)
	require.WithinDuration(t, startedAt.Add(carpoolFormationWindow), *group.DeadlineAt, 2*time.Second)
	require.Equal(t, 25.0, client.User.GetX(ctx, firstUser.ID).Balance)

	_, err := svc.createOrderInTx(ctx, CreateOrderRequest{
		UserID:        firstUser.ID,
		PaymentType:   payment.TypeAlipay,
		OrderType:     payment.OrderTypeCarpool,
		CarpoolPlanID: plan.ID,
		ClientIP:      "127.0.0.1",
		SrcHost:       "app.example.com",
	}, &User{ID: firstUser.ID, Email: firstUser.Email, Username: firstUser.Username}, nil, plan, &PaymentConfig{MaxPendingOrders: 10, OrderTimeoutMin: 15}, 800, 800, 0, 800, nil)
	require.Equal(t, "CARPOOL_ALREADY_JOINED", infraerrors.Reason(err))

	require.NoError(t, svc.ExecuteCarpoolFulfillment(ctx, secondOrder.ID))
	secondOrder = client.PaymentOrder.GetX(ctx, secondOrder.ID)
	require.Equal(t, firstOrder.CarpoolGroupID, secondOrder.CarpoolGroupID)
	group = client.CarpoolGroup.GetX(ctx, *firstOrder.CarpoolGroupID)
	require.Equal(t, CarpoolStatusPurchasing, group.Status)
	require.Equal(t, 2, group.MemberCount)
	require.NotNil(t, group.FormedAt)
	require.Nil(t, group.OpenKey)
	require.Equal(t, 40.0, client.User.GetX(ctx, secondUser.ID).Balance)
}

func TestExecuteCarpoolFulfillmentFormsSinglePlanImmediately(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	user := createCarpoolTestUser(t, ctx, client, "single", 60)
	plan := createCarpoolTestPlan(t, ctx, client, 1400, 1, "single account rules")
	order := createPaidCarpoolTestOrder(t, ctx, client, user, "single", plan, 1400)
	svc := &PaymentService{entClient: client}

	require.NoError(t, svc.ExecuteCarpoolFulfillment(ctx, order.ID))
	order = client.PaymentOrder.GetX(ctx, order.ID)
	require.Equal(t, OrderStatusCompleted, order.Status)
	require.NotNil(t, order.CarpoolGroupID)
	group := client.CarpoolGroup.GetX(ctx, *order.CarpoolGroupID)
	require.Equal(t, CarpoolStatusPurchasing, group.Status)
	require.Equal(t, 1, group.MemberCount)
	require.NotNil(t, group.FormedAt)
	require.Equal(t, 60.0, client.User.GetX(ctx, user.ID).Balance)
}

func TestExpireCarpoolGroupsAdvancesFormationAndServiceExpiry(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	now := time.Now()
	waitingPlan := createCarpoolTestPlan(t, ctx, client, 1600, 4, "four-person rules")
	waiting := client.CarpoolGroup.Create().
		SetCarpoolPlanID(waitingPlan.ID).
		SetCarpoolPlanRevision(waitingPlan.Revision).
		SetTargetMembers(4).
		SetTotalAmount(waitingPlan.TotalAmount).
		SetPricePerMember(400).
		SetPlanNote(waitingPlan.Note).
		SetMemberCount(1).
		SetStatus(CarpoolStatusWaiting).
		SetOpenKey(strconv.FormatInt(waitingPlan.ID, 10)).
		SetDeadlineAt(now.Add(-time.Minute)).
		SaveX(ctx)
	activePlan := createCarpoolTestPlan(t, ctx, client, 1400, 1, "single account rules")
	active := client.CarpoolGroup.Create().
		SetCarpoolPlanID(activePlan.ID).
		SetCarpoolPlanRevision(activePlan.Revision).
		SetTargetMembers(1).
		SetTotalAmount(activePlan.TotalAmount).
		SetPricePerMember(1400).
		SetPlanNote(activePlan.Note).
		SetMemberCount(1).
		SetStatus(CarpoolStatusActive).
		SetOpenedAt(now.Add(-31 * 24 * time.Hour)).
		SetExpiresAt(now.Add(-24 * time.Hour)).
		SaveX(ctx)
	svc := &PaymentService{entClient: client}

	advanced, err := svc.ExpireCarpoolGroups(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, advanced)
	waiting = client.CarpoolGroup.GetX(ctx, waiting.ID)
	require.Equal(t, CarpoolStatusRefundPending, waiting.Status)
	require.Equal(t, "not_formed", *waiting.StatusReason)
	require.Nil(t, waiting.OpenKey)
	require.Equal(t, CarpoolStatusExpired, client.CarpoolGroup.GetX(ctx, active.ID).Status)
}

func TestOpenCarpoolGroupAppliesFixedThirtyDayPeriod(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	plan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	group := client.CarpoolGroup.Create().
		SetCarpoolPlanID(plan.ID).
		SetCarpoolPlanRevision(plan.Revision).
		SetTargetMembers(2).
		SetTotalAmount(plan.TotalAmount).
		SetPricePerMember(800).
		SetPlanNote(plan.Note).
		SetMemberCount(2).
		SetStatus(CarpoolStatusPurchasing).
		SetFormedAt(time.Now()).
		SaveX(ctx)
	svc := &PaymentService{entClient: client}
	beforeOpen := time.Now()

	require.NoError(t, svc.OpenCarpoolGroup(ctx, group.ID))
	afterOpen := time.Now()
	group = client.CarpoolGroup.GetX(ctx, group.ID)
	require.Equal(t, CarpoolStatusActive, group.Status)
	require.False(t, group.OpenedAt.Before(beforeOpen))
	require.False(t, group.OpenedAt.After(afterOpen))
	require.Equal(t, group.OpenedAt.Add(30*24*time.Hour), *group.ExpiresAt)
}

func TestPrepareCarpoolRefundRequiresGroupTransitionAndNeverDeductsBalance(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	user := createCarpoolTestUser(t, ctx, client, "refund", 90)
	provider := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("Carpool Alipay").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetRefundEnabled(true).
		SaveX(ctx)
	carpoolPlan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	group := client.CarpoolGroup.Create().
		SetCarpoolPlanID(carpoolPlan.ID).
		SetCarpoolPlanRevision(carpoolPlan.Revision).
		SetTargetMembers(2).
		SetTotalAmount(carpoolPlan.TotalAmount).
		SetPricePerMember(800).
		SetPlanNote(carpoolPlan.Note).
		SetMemberCount(2).
		SetStatus(CarpoolStatusActive).
		SetOpenedAt(time.Now()).
		SetExpiresAt(time.Now().Add(20 * 24 * time.Hour)).
		SaveX(ctx)
	order := createPaidCarpoolTestOrder(t, ctx, client, user, "refund", carpoolPlan, 800)
	order = client.PaymentOrder.UpdateOneID(order.ID).
		SetStatus(OrderStatusCompleted).
		SetPayAmount(824).
		SetCarpoolGroupID(group.ID).
		SetProviderInstanceID(strconv.FormatInt(provider.ID, 10)).
		SaveX(ctx)
	svc := &PaymentService{entClient: client}

	plan, result, err := svc.PrepareRefund(ctx, order.ID, 400, "official refund approved", false, true)
	require.Nil(t, plan)
	require.Nil(t, result)
	require.Equal(t, "INVALID_CARPOOL_STATUS", infraerrors.Reason(err))

	client.CarpoolGroup.UpdateOneID(group.ID).
		SetStatus(CarpoolStatusRefundPending).
		SetStatusReason("official refund approved").
		SaveX(ctx)
	plan, result, err = svc.PrepareRefund(ctx, order.ID, 400, "official refund approved", false, true)
	require.NoError(t, err)
	require.Nil(t, result)
	require.False(t, plan.DeductBalance)
	require.Equal(t, payment.DeductionTypeNone, plan.DeductionType)
	require.Zero(t, plan.BalanceToDeduct)
	require.InDelta(t, 412, plan.GatewayAmount, 0.001)

	client.CarpoolGroup.UpdateOneID(group.ID).
		ClearOpenedAt().
		ClearExpiresAt().
		SaveX(ctx)
	plan, result, err = svc.PrepareRefund(ctx, order.ID, 400, "service was not delivered", false, false)
	require.Nil(t, plan)
	require.Nil(t, result)
	require.Equal(t, "CARPOOL_FULL_REFUND_REQUIRED", infraerrors.Reason(err))

	plan, result, err = svc.PrepareRefund(ctx, order.ID, 800, "service was not delivered", false, false)
	require.NoError(t, err)
	require.Nil(t, result)
	require.Equal(t, 800.0, plan.RefundAmount)
	require.InDelta(t, 824, plan.GatewayAmount, 0.001)
}

func TestCarpoolPricePerMemberRequiresExactCentDivision(t *testing.T) {
	price, err := carpoolPricePerMember(1600, 4)
	require.NoError(t, err)
	require.Equal(t, 400.0, price)

	price, err = carpoolPricePerMember(100.02, 3)
	require.NoError(t, err)
	require.Equal(t, 33.34, price)

	_, err = carpoolPricePerMember(100, 3)
	require.Equal(t, "CARPOOL_PLAN_AMOUNT_NOT_DIVISIBLE", infraerrors.Reason(err))

	_, err = carpoolPricePerMember(100.001, 2)
	require.Equal(t, "INVALID_CARPOOL_PLAN_AMOUNT", infraerrors.Reason(err))
}

func TestCreateCarpoolOrderPersistsPlanSnapshot(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	user := createCarpoolTestUser(t, ctx, client, "snapshot", 0)
	plan := createCarpoolTestPlan(t, ctx, client, 1600, 4, "four-person rules")
	svc := &PaymentService{entClient: client}

	order, err := svc.createOrderInTx(ctx, CreateOrderRequest{
		UserID:        user.ID,
		PaymentType:   payment.TypeAlipay,
		OrderType:     payment.OrderTypeCarpool,
		CarpoolPlanID: plan.ID,
		ClientIP:      "127.0.0.1",
		SrcHost:       "app.example.com",
	}, &User{ID: user.ID, Email: user.Email, Username: user.Username}, nil, plan, &PaymentConfig{MaxPendingOrders: 10, OrderTimeoutMin: 15}, 400, 400, 0, 400, nil)
	require.NoError(t, err)
	require.Equal(t, plan.ID, *order.CarpoolPlanID)
	require.Equal(t, plan.Revision, *order.CarpoolPlanRevision)
	require.Equal(t, plan.TargetMembers, *order.CarpoolSize)
	require.Equal(t, plan.TotalAmount, *order.CarpoolTotalAmount)
	require.Equal(t, plan.Note, *order.CarpoolPlanNote)
}

func TestGetCarpoolOverviewUsesConfiguredPlanAndPlanScopedProgress(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	plan := createCarpoolTestPlan(t, ctx, client, 1800, 3, "three-person rules")
	client.CarpoolGroup.Create().
		SetCarpoolPlanID(plan.ID).
		SetCarpoolPlanRevision(plan.Revision).
		SetTargetMembers(plan.TargetMembers).
		SetTotalAmount(plan.TotalAmount).
		SetPricePerMember(600).
		SetPlanNote(plan.Note).
		SetMemberCount(1).
		SetStatus(CarpoolStatusWaiting).
		SetOpenKey(strconv.FormatInt(plan.ID, 10)).
		SetDeadlineAt(time.Now().Add(time.Hour)).
		SaveX(ctx)
	svc := &PaymentService{entClient: client}

	overview, err := svc.GetCarpoolOverview(ctx, 999)
	require.NoError(t, err)
	require.Len(t, overview.Plans, 1)
	require.Equal(t, plan.ID, overview.Plans[0].ID)
	require.Equal(t, 1800.0, overview.Plans[0].TotalAmount)
	require.Equal(t, 3, overview.Plans[0].Size)
	require.Equal(t, 600.0, overview.Plans[0].Price)
	require.Equal(t, "three-person rules", overview.Plans[0].Note)
	require.Equal(t, 1, overview.Plans[0].CurrentMembers)
	require.Equal(t, 2, overview.Plans[0].RemainingMembers)
}

func TestCarpoolPlanChangesAreBlockedOnlyByActivePurchases(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	user := createCarpoolTestUser(t, ctx, client, "plan-lock", 0)
	plan := createCarpoolTestPlan(t, ctx, client, 1600, 2, "shared account rules")
	order := createPaidCarpoolTestOrder(t, ctx, client, user, "plan-lock", plan, 800)
	svc := &PaymentService{entClient: client}

	_, err := svc.UpdateCarpoolPlan(ctx, plan.ID, CarpoolPlanInput{TotalAmount: 1800, TargetMembers: 2, Note: "updated rules"})
	require.Equal(t, "CARPOOL_PLAN_IN_USE", infraerrors.Reason(err))
	require.Equal(t, "CARPOOL_PLAN_IN_USE", infraerrors.Reason(svc.DeleteCarpoolPlan(ctx, plan.ID)))

	client.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusFailed).SaveX(ctx)
	_, err = svc.UpdateCarpoolPlan(ctx, plan.ID, CarpoolPlanInput{TotalAmount: 1800, TargetMembers: 2, Note: "updated rules"})
	require.Equal(t, "CARPOOL_PLAN_IN_USE", infraerrors.Reason(err))
	require.Equal(t, "CARPOOL_PLAN_IN_USE", infraerrors.Reason(svc.DeleteCarpoolPlan(ctx, plan.ID)))

	client.PaymentOrder.DeleteOneID(order.ID).ExecX(ctx)
	waiting := client.CarpoolGroup.Create().
		SetCarpoolPlanID(plan.ID).
		SetCarpoolPlanRevision(plan.Revision).
		SetTargetMembers(plan.TargetMembers).
		SetTotalAmount(plan.TotalAmount).
		SetPricePerMember(800).
		SetPlanNote(plan.Note).
		SetMemberCount(0).
		SetStatus(CarpoolStatusWaiting).
		SetOpenKey(strconv.FormatInt(plan.ID, 10)).
		SetDeadlineAt(time.Now().Add(time.Hour)).
		SaveX(ctx)
	_, err = svc.UpdateCarpoolPlan(ctx, plan.ID, CarpoolPlanInput{TotalAmount: 1800, TargetMembers: 2, Note: "updated rules"})
	require.Equal(t, "CARPOOL_PLAN_IN_USE", infraerrors.Reason(err))

	client.CarpoolGroup.UpdateOneID(waiting.ID).
		SetStatus(CarpoolStatusPurchasing).
		ClearOpenKey().
		SaveX(ctx)
	updated, err := svc.UpdateCarpoolPlan(ctx, plan.ID, CarpoolPlanInput{TotalAmount: 1800, TargetMembers: 2, Note: "updated rules"})
	require.NoError(t, err)
	require.Equal(t, plan.Revision+1, updated.Revision)
	require.NoError(t, svc.DeleteCarpoolPlan(ctx, plan.ID))
	require.Equal(t, 1, client.CarpoolGroup.Query().CountX(ctx))
}

// createCarpoolTestUser creates a persisted user whose balance must remain unchanged by carpool fulfillment.
func createCarpoolTestUser(t *testing.T, ctx context.Context, client *dbent.Client, suffix string, balance float64) *dbent.User {
	t.Helper()
	return client.User.Create().
		SetEmail("carpool-" + suffix + "@example.com").
		SetPasswordHash("hash").
		SetUsername("carpool-" + suffix).
		SetBalance(balance).
		SaveX(ctx)
}

// createCarpoolTestPlan creates the persisted configuration referenced by order snapshots.
func createCarpoolTestPlan(t *testing.T, ctx context.Context, client *dbent.Client, totalAmount float64, targetMembers int, note string) *dbent.CarpoolPlan {
	t.Helper()
	return client.CarpoolPlan.Create().
		SetTotalAmount(totalAmount).
		SetTargetMembers(targetMembers).
		SetNote(note).
		SaveX(ctx)
}

// createPaidCarpoolTestOrder creates the paid order snapshot consumed by carpool fulfillment.
func createPaidCarpoolTestOrder(t *testing.T, ctx context.Context, client *dbent.Client, user *dbent.User, suffix string, plan *dbent.CarpoolPlan, amount float64) *dbent.PaymentOrder {
	t.Helper()
	return client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(amount).
		SetPayAmount(amount).
		SetFeeRate(0).
		SetRechargeCode("CARPOOL-" + suffix).
		SetOutTradeNo("sub2_carpool_" + suffix).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-carpool-" + suffix).
		SetOrderType(payment.OrderTypeCarpool).
		SetCarpoolPlanID(plan.ID).
		SetCarpoolPlanRevision(plan.Revision).
		SetCarpoolSize(plan.TargetMembers).
		SetCarpoolTotalAmount(plan.TotalAmount).
		SetCarpoolPlanNote(plan.Note).
		SetStatus(OrderStatusPaid).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("app.example.com").
		SaveX(ctx)
}

type carpoolUserRepositoryStub struct {
	UserRepository
	user *User
}

// GetByID returns the active user needed before create-order currency validation.
func (s carpoolUserRepositoryStub) GetByID(context.Context, int64) (*User, error) {
	return s.user, nil
}

type carpoolCurrencyLoadBalancerStub struct{}

// GetInstanceConfig is unused because create-order selection returns the full snapshot directly.
func (carpoolCurrencyLoadBalancerStub) GetInstanceConfig(context.Context, int64) (map[string]string, error) {
	return nil, nil
}

// SelectInstance returns an Alipay-capable gateway configured to settle in USD.
func (carpoolCurrencyLoadBalancerStub) SelectInstance(context.Context, string, payment.PaymentType, payment.Strategy, float64) (*payment.InstanceSelection, error) {
	return &payment.InstanceSelection{
		ProviderKey:    payment.TypeAirwallex,
		SupportedTypes: payment.TypeAlipay,
		Config:         map[string]string{"currency": "USD"},
	}, nil
}
