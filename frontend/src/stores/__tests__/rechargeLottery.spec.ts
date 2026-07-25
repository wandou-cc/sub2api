import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const getRechargeLottery = vi.hoisted(() => vi.fn())
const drawRechargeLottery = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getRechargeLottery,
    drawRechargeLottery,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ refreshUser }),
}))

import { useRechargeLotteryStore } from '@/stores/rechargeLottery'
import type { RechargeLotteryOverview } from '@/types/payment'

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

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

const anotherPendingOpportunity = {
  ...pendingOpportunity,
  order_id: 84,
  recharge_amount: 100,
  max_rarity: 'legendary' as const,
}

describe('useRechargeLotteryStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    getRechargeLottery.mockReset().mockResolvedValue({
      data: { pending_count: 1, opportunities: [pendingOpportunity] },
    })
    drawRechargeLottery.mockReset().mockResolvedValue({ data: claimedOpportunity })
    refreshUser.mockReset().mockResolvedValue(undefined)
  })

  it('fetches the current opportunities', async () => {
    const store = useRechargeLotteryStore()

    await store.fetchOverview()

    expect(getRechargeLottery).toHaveBeenCalledTimes(1)
    expect(store.overview?.pending_count).toBe(1)
    expect(store.loading).toBe(false)
  })

  it('ignores an older overview response that finishes last', async () => {
    const firstResponse = deferred<{ data: RechargeLotteryOverview }>()
    const secondResponse = deferred<{ data: RechargeLotteryOverview }>()
    getRechargeLottery.mockReset()
      .mockReturnValueOnce(firstResponse.promise)
      .mockReturnValueOnce(secondResponse.promise)
    const store = useRechargeLotteryStore()

    const firstFetch = store.fetchOverview()
    const secondFetch = store.fetchOverview()
    secondResponse.resolve({ data: { pending_count: 1, opportunities: [pendingOpportunity] } })
    await secondFetch
    firstResponse.resolve({ data: { pending_count: 0, opportunities: [] } })
    await firstFetch

    expect(store.overview).toEqual({ pending_count: 1, opportunities: [pendingOpportunity] })
    expect(store.loading).toBe(false)
  })

  it('claims once, replaces the pending item, and refreshes the user balance', async () => {
    const store = useRechargeLotteryStore()
    await store.fetchOverview()

    const result = await store.claim(42)

    expect(result).toEqual(claimedOpportunity)
    expect(drawRechargeLottery).toHaveBeenCalledTimes(1)
    expect(drawRechargeLottery).toHaveBeenCalledWith(42)
    expect(store.overview?.pending_count).toBe(0)
    expect(store.overview?.opportunities[0]).toEqual(claimedOpportunity)
    expect(refreshUser).toHaveBeenCalledTimes(1)

    await expect(store.claim(42)).rejects.toThrow('is not available')
    expect(drawRechargeLottery).toHaveBeenCalledTimes(1)
  })

  it('merges a claim into the latest overview without dropping a newly fetched opportunity', async () => {
    const claimResponse = deferred<{ data: typeof claimedOpportunity }>()
    drawRechargeLottery.mockReset().mockReturnValueOnce(claimResponse.promise)
    const store = useRechargeLotteryStore()
    await store.fetchOverview()

    const claimPromise = store.claim(42)
    getRechargeLottery.mockResolvedValueOnce({
      data: { pending_count: 2, opportunities: [pendingOpportunity, anotherPendingOpportunity] },
    })
    await store.fetchOverview()
    claimResponse.resolve({ data: claimedOpportunity })
    await claimPromise

    expect(store.overview?.pending_count).toBe(1)
    expect(store.overview?.opportunities).toEqual([claimedOpportunity, anotherPendingOpportunity])
  })

  it('does not let a fetch started before a successful claim restore pending state', async () => {
    const staleResponse = deferred<{ data: RechargeLotteryOverview }>()
    const store = useRechargeLotteryStore()
    await store.fetchOverview()
    getRechargeLottery.mockReset().mockReturnValueOnce(staleResponse.promise)

    const staleFetch = store.fetchOverview()
    await store.claim(42)
    staleResponse.resolve({ data: { pending_count: 1, opportunities: [pendingOpportunity] } })
    await staleFetch

    expect(store.overview?.pending_count).toBe(0)
    expect(store.overview?.opportunities).toEqual([claimedOpportunity])
  })

  it('does not repopulate state when an old fetch finishes after reset', async () => {
    const response = deferred<{ data: RechargeLotteryOverview }>()
    getRechargeLottery.mockReset().mockReturnValueOnce(response.promise)
    const store = useRechargeLotteryStore()

    const fetchPromise = store.fetchOverview()
    store.reset()
    response.resolve({ data: { pending_count: 1, opportunities: [pendingOpportunity] } })
    await fetchPromise

    expect(store.overview).toBeNull()
    expect(store.loading).toBe(false)
  })

  it('dismisses without claiming and can reopen the same opportunity', async () => {
    const store = useRechargeLotteryStore()
    await store.fetchOverview()

    store.openDialog(42)
    store.dismissDialog(42)

    expect(store.dialogOrderId).toBe(0)
    expect(store.dismissedOrderIds).toEqual([42])
    expect(drawRechargeLottery).not.toHaveBeenCalled()

    store.openDialog(42)
    expect(store.dialogOrderId).toBe(42)
    expect(store.dismissedOrderIds).toEqual([])
  })
})
