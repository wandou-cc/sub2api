import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminCarpoolGroupsView from '../AdminCarpoolGroupsView.vue'

const getCarpoolGroups = vi.hoisted(() => vi.fn())
const openCarpoolGroup = vi.hoisted(() => vi.fn())
const markCarpoolRefundPending = vi.hoisted(() => vi.fn())
const getCarpoolPlans = vi.hoisted(() => vi.fn())
const createCarpoolPlan = vi.hoisted(() => vi.fn())
const updateCarpoolPlan = vi.hoisted(() => vi.fn())
const deleteCarpoolPlan = vi.hoisted(() => vi.fn())
const refundOrder = vi.hoisted(() => vi.fn())
const queryRefund = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getCarpoolGroups,
    openCarpoolGroup,
    markCarpoolRefundPending,
    getCarpoolPlans,
    createCarpoolPlan,
    updateCarpoolPlan,
    deleteCarpoolPlan,
    refundOrder,
    queryRefund,
  },
  default: {
    getCarpoolGroups,
    openCarpoolGroup,
    markCarpoolRefundPending,
    getCarpoolPlans,
    createCarpoolPlan,
    updateCarpoolPlan,
    deleteCarpoolPlan,
    refundOrder,
    queryRefund,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const BaseDialogStub = {
  props: ['show', 'title'],
  template: '<div v-if="show"><p>{{ title }}</p><slot /><slot name="footer" /></div>',
}

function carpoolGroupFixture() {
  return {
    id: 12,
    carpool_plan_id: 4,
    target_members: 2,
    current_members: 2,
    total_amount: 1600,
    price_per_member: 800,
    plan_note: 'two-person rules',
    status: 'purchasing' as const,
    formed_at: '2026-07-26T08:00:00Z',
    created_at: '2026-07-26T07:00:00Z',
    updated_at: '2026-07-26T08:00:00Z',
    members: [
      {
        order_id: 101,
        user_id: 9,
        user_email: 'member@example.com',
        user_name: 'member',
        amount: 800,
        pay_amount: 800,
        currency: 'CNY',
        payment_type: 'alipay',
        status: 'COMPLETED' as const,
        refund_amount: 0,
      },
    ],
  }
}

function carpoolPlanFixture() {
  return {
    id: 4,
    total_amount: 1600,
    target_members: 2,
    price_per_member: 800,
    note: 'two-person rules',
    created_at: '2026-07-26T07:00:00Z',
    updated_at: '2026-07-26T08:00:00Z',
  }
}

function mountView() {
  return mount(AdminCarpoolGroupsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: BaseDialogStub,
        Icon: true,
        OrderStatusBadge: true,
        Teleport: true,
        Transition: false,
      },
    },
  })
}

describe('AdminCarpoolGroupsView', () => {
  beforeEach(() => {
    getCarpoolGroups.mockReset().mockResolvedValue({ data: [carpoolGroupFixture()] })
    openCarpoolGroup.mockReset().mockResolvedValue({ data: {} })
    markCarpoolRefundPending.mockReset()
    getCarpoolPlans.mockReset().mockResolvedValue({ data: [carpoolPlanFixture()] })
    createCarpoolPlan.mockReset().mockResolvedValue({ data: carpoolPlanFixture() })
    updateCarpoolPlan.mockReset().mockResolvedValue({ data: carpoolPlanFixture() })
    deleteCarpoolPlan.mockReset().mockResolvedValue({ data: {} })
    refundOrder.mockReset()
    queryRefund.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads the current queue and switches to history', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(getCarpoolGroups).toHaveBeenNthCalledWith(1, false)
    expect(wrapper.text()).toContain('member@example.com')

    const historyButton = wrapper.findAll('button').find(button => button.text() === 'payment.admin.carpool.history')
    expect(historyButton).toBeDefined()
    await historyButton!.trigger('click')
    await flushPromises()

    expect(getCarpoolGroups).toHaveBeenLastCalledWith(true)
  })

  it('lets the backend own the opening time and 30-day expiry', async () => {
    const wrapper = mountView()
    await flushPromises()

    const openButton = wrapper.findAll('button').find(button => button.text().includes('payment.admin.carpool.markOpened'))
    expect(openButton).toBeDefined()
    await openButton!.trigger('click')

    expect(wrapper.text()).toContain('payment.admin.carpool.openDialogMessage')
    const confirmButton = wrapper.findAll('button').find(button => button.text() === 'common.confirm')
    expect(confirmButton).toBeDefined()
    await confirmButton!.trigger('click')
    await flushPromises()

    expect(openCarpoolGroup).toHaveBeenCalledWith(12)
  })

  it('loads plan configuration and creates a three-field package', async () => {
    const wrapper = mountView()
    await flushPromises()

    const plansButton = wrapper.findAll('button').find(button => button.text() === 'payment.admin.carpool.planConfig')
    expect(plansButton).toBeDefined()
    await plansButton!.trigger('click')
    await flushPromises()

    expect(getCarpoolPlans).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('two-person rules')

    const createButton = wrapper.findAll('button').find(button => button.text().includes('payment.admin.carpool.createPlan'))
    expect(createButton).toBeDefined()
    await createButton!.trigger('click')

    const numberInputs = wrapper.findAll('input[type="number"]')
    await numberInputs[0].setValue('1800')
    await numberInputs[1].setValue('3')
    await wrapper.get('textarea[rows="8"]').setValue('three-person rules')
    const confirmButton = wrapper.findAll('button').find(button => button.text() === 'common.confirm')
    expect(confirmButton).toBeDefined()
    await confirmButton!.trigger('click')
    await flushPromises()

    expect(createCarpoolPlan).toHaveBeenCalledWith({
      total_amount: 1800,
      target_members: 3,
      note: 'three-person rules',
    })
  })
})
