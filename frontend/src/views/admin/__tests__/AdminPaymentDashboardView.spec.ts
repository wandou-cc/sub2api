import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminPaymentDashboardView from '../orders/AdminPaymentDashboardView.vue'

const { getDashboard } = vi.hoisted(() => ({
  getDashboard: vi.fn()
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: { getDashboard },
  default: { getDashboard }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn() })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

describe('admin income management', () => {
  it('loads the selected custom date range', async () => {
    getDashboard.mockResolvedValue({
      data: {
        today_amount: 0,
        total_amount: 30,
        today_count: 0,
        total_count: 2,
        avg_amount: 15,
        user_count: 2,
        daily_series: [],
        payment_methods: [],
        top_users: []
      }
    })

    const wrapper = mount(AdminPaymentDashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          OrderStatsCards: true,
          DailyRevenueChart: true,
          DateRangePicker: {
            template: '<button class="custom-range" @click="$emit(\'update:startDate\', \'2026-07-01\'); $emit(\'update:endDate\', \'2026-07-15\'); $emit(\'change\')">range</button>'
          }
        }
      }
    })

    await flushPromises()
    getDashboard.mockClear()
    await wrapper.get('.custom-range').trigger('click')
    await flushPromises()

    expect(getDashboard).toHaveBeenCalledWith('2026-07-01', '2026-07-15')
  })
})
