import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { useRechargeLotteryStore } from '@/stores/rechargeLottery'
import type { RechargeLotteryOverview } from '@/types/payment'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

import RechargeLotteryPanel from '../RechargeLotteryPanel.vue'

const routerLinkStub = {
  props: ['to'],
  template: '<a :data-path="typeof to === \'string\' ? to : to.path" :data-cat="to.query?.cat" :data-page="to.query?.page"><slot /></a>',
}

function lotteryOverview(): RechargeLotteryOverview {
  return {
    pending_count: 1,
    opportunities: [
      {
        order_id: 42,
        recharge_amount: 20,
        max_rarity: 'epic_plus',
        claimed: false,
        rarity: '',
        reward_amount: 0,
        created_at: '2026-07-24T00:00:00Z',
      },
      {
        order_id: 41,
        recharge_amount: 20,
        max_rarity: 'epic_plus',
        claimed: true,
        rarity: 'epic',
        reward_amount: 5.25,
        balance_after: 45.25,
        created_at: '2026-07-23T00:00:00Z',
        claimed_at: '2026-07-23T06:56:05Z',
      },
    ],
  }
}

describe('RechargeLotteryPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    useRechargeLotteryStore().overview = lotteryOverview()
  })

  it('shows pending opportunities without inline odds and links to the full rules', () => {
    const wrapper = mount(RechargeLotteryPanel, {
      global: { stubs: { RouterLink: routerLinkStub } },
    })

    expect(wrapper.text()).toContain('payment.lottery.open')
    expect(wrapper.text()).not.toContain('80%')
    expect(wrapper.text()).not.toContain('20%')
    expect(wrapper.find('.rarity-badge').exists()).toBe(false)
    expect(wrapper.get('.lottery-summary').exists()).toBe(true)
    const rulesLink = wrapper.get('a[data-path="/docs"]')
    expect(rulesLink.attributes('data-cat')).toBe('activities')
    expect(rulesLink.attributes('data-page')).toBe('recharge-lottery')
  })

  it('opens the shared global dialog instead of claiming directly', async () => {
    const store = useRechargeLotteryStore()
    const wrapper = mount(RechargeLotteryPanel, {
      global: { stubs: { RouterLink: routerLinkStub } },
    })

    await wrapper.get('.lottery-open-button').trigger('click')

    expect(store.dialogOrderId).toBe(42)
    expect(store.overview?.opportunities[0].claimed).toBe(false)
  })

  it('renders claimed rewards as history rows', () => {
    const wrapper = mount(RechargeLotteryPanel, {
      global: { stubs: { RouterLink: routerLinkStub } },
    })

    expect(wrapper.get('.history-reward strong').text()).toBe('+$5.25')
    expect(wrapper.get('.history-reward small').text()).not.toBe('')
    expect(wrapper.get('.history-rarity').text()).toBe('payment.lottery.rarity.epic')
  })
})
