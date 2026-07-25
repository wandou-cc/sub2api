import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { useRechargeLotteryStore } from '@/stores/rechargeLottery'

const getRechargeLottery = vi.hoisted(() => vi.fn())
const drawRechargeLottery = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const routeState = vi.hoisted(() => ({ fullPath: '/dashboard' }))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getRechargeLottery,
    drawRechargeLottery,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isAuthenticated: true,
    refreshUser,
  }),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params?.amount ? `${key}:${params.amount}` : key,
    }),
  }
})

import RechargeLotteryRewardDialog from '../RechargeLotteryRewardDialog.vue'

const pendingOpportunity = {
  order_id: 42,
  recharge_amount: 20,
  max_rarity: 'epic_plus' as const,
  claimed: false,
  rarity: '' as const,
  reward_amount: 0,
  created_at: '2026-07-24T00:00:00Z',
}

const claimedOpportunity = {
  ...pendingOpportunity,
  claimed: true,
  rarity: 'epic' as const,
  reward_amount: 5.25,
  balance_after: 31.5,
  claimed_at: '2026-07-24T00:01:00Z',
}

const routerLinkStub = {
  props: ['to'],
  template: '<a><slot /></a>',
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise
  })
  return { promise, resolve }
}

describe('RechargeLotteryRewardDialog', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    getRechargeLottery.mockReset().mockResolvedValue({
      data: { pending_count: 1, opportunities: [pendingOpportunity] },
    })
    drawRechargeLottery.mockReset().mockResolvedValue({ data: claimedOpportunity })
    refreshUser.mockReset().mockResolvedValue(undefined)
  })

  it('opens globally for a pending recharge opportunity', async () => {
    const wrapper = mount(RechargeLotteryRewardDialog, {
      global: { stubs: { Teleport: true, RouterLink: routerLinkStub } },
    })
    await flushPromises()

    expect(getRechargeLottery).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="recharge-lottery-dialog"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="gift-claim-button"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="gift-claim-button"] .gift-action-copy').exists()).toBe(true)
    expect(wrapper.get('[data-testid="gift-claim-button"] .rarity-cap').exists()).toBe(true)
    expect(wrapper.get('[data-testid="recharge-lottery-dialog"]').text()).toContain('20.00')
  })

  it('closes without issuing a reward when the gift is not clicked', async () => {
    const wrapper = mount(RechargeLotteryRewardDialog, {
      global: { stubs: { Teleport: true, RouterLink: routerLinkStub } },
    })
    await flushPromises()

    await wrapper.get('.dialog-footer-actions button').trigger('click')
    await flushPromises()

    expect(drawRechargeLottery).not.toHaveBeenCalled()
    expect(useRechargeLotteryStore().dialogOrderId).toBe(0)
    expect(useRechargeLotteryStore().dismissedOrderIds).toEqual([42])
  })

  it('claims on gift click and separates recharge from reward with confetti', async () => {
    const wrapper = mount(RechargeLotteryRewardDialog, {
      global: { stubs: { Teleport: true, RouterLink: routerLinkStub } },
    })
    await flushPromises()

    await wrapper.get('[data-testid="gift-claim-button"]').trigger('click')
    await flushPromises()

    expect(drawRechargeLottery).toHaveBeenCalledTimes(1)
    expect(drawRechargeLottery).toHaveBeenCalledWith(42)
    expect(wrapper.get('[data-testid="recharge-amount"]').text()).toBe('$20.00')
    expect(wrapper.get('[data-testid="reward-amount"]').text()).toBe('+$5.25')
    expect(wrapper.text()).toContain('payment.lottery.rechargeCredited')
    expect(wrapper.text()).toContain('payment.lottery.blindBoxReward')
    const confetti = wrapper.get('[data-testid="confetti"]')
    expect(confetti.findAll('span')).toHaveLength(72)
    expect(confetti.element.parentElement?.classList.contains('reward-dialog-overlay')).toBe(true)
  })

  it('ignores a claim response from a reset session', async () => {
    const claimResponse = deferred<{ data: typeof claimedOpportunity }>()
    drawRechargeLottery.mockReset().mockReturnValueOnce(claimResponse.promise)
    const wrapper = mount(RechargeLotteryRewardDialog, {
      global: { stubs: { Teleport: true, RouterLink: routerLinkStub } },
    })
    await flushPromises()

    await wrapper.get('[data-testid="gift-claim-button"]').trigger('click')
    const store = useRechargeLotteryStore()
    store.reset()
    claimResponse.resolve({ data: claimedOpportunity })
    await flushPromises()

    expect(wrapper.find('[data-testid="reward-amount"]').exists()).toBe(false)
    expect(store.overview).toBeNull()
    expect(store.dialogOrderId).toBe(0)
  })
})

describe('App global recharge lottery registration', () => {
  it('mounts the reward dialog at the application root', () => {
    const testDir = dirname(fileURLToPath(import.meta.url))
    const appSource = readFileSync(resolve(testDir, '../../../App.vue'), 'utf8')

    expect(appSource).toContain("import RechargeLotteryRewardDialog from '@/components/payment/RechargeLotteryRewardDialog.vue'")
    expect(appSource).toContain('<RechargeLotteryRewardDialog />')
  })
})
