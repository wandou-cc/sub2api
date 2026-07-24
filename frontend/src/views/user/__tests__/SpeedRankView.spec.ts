import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SpeedRankView from '../SpeedRankView.vue'

const { getSpeedRank, showError } = vi.hoisted(() => ({
  getSpeedRank: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/usage', () => ({
  usageAPI: { getSpeedRank }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const entries = [
  {
    rank: 1,
    user_id: 11,
    email: 'f***@example.com',
    username: 'first-user',
    input_tokens: 900_000,
    output_tokens: 100_000,
    total_tokens: 1_000_000,
    reward: 3
  },
  {
    rank: 2,
    user_id: 22,
    email: 's***@example.com',
    username: 'second-user',
    input_tokens: 700_000,
    output_tokens: 100_000,
    total_tokens: 800_000,
    reward: 2
  },
  {
    rank: 3,
    user_id: 33,
    email: 't***@example.com',
    username: '',
    input_tokens: 500_000,
    output_tokens: 100_000,
    total_tokens: 600_000,
    reward: 1
  }
]

const history = Array.from({ length: 8 }, (_, index) => ({
  ...entries[0],
  user_id: index + 1,
  reward_date: `2026-07-${String(22 - index).padStart(2, '0')}`
}))

// mountSpeedRank 挂载排行榜页面并隔离全局布局组件。
function mountSpeedRank() {
  return mount(SpeedRankView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Pagination: { template: '<div class="pagination-stub" />' }
      }
    }
  })
}

describe('SpeedRankView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-07-23T12:00:00+08:00'))
    getSpeedRank.mockReset()
    showError.mockReset()
    getSpeedRank.mockResolvedValue({
      entries,
      history,
      next_reward_at: '2026-07-24T00:00:00+08:00',
      generated_at: '2026-07-23T12:00:00+08:00',
      ranking_date: '2026-07-23',
      reward_amounts: { 1: 3, 2: 2, 3: 1 }
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders the beach podium with the three rank-specific castles', async () => {
    const wrapper = mountSpeedRank()
    await flushPromises()

    expect(getSpeedRank).toHaveBeenCalledTimes(1)
    expect(wrapper.find('.tide-board-background').attributes('src')).toContain('token-tide-beach')
    expect(wrapper.findAll('.rank-entry')).toHaveLength(3)
    expect(wrapper.findAll('.castle-image').map((image) => image.attributes('src'))).toEqual([
      expect.stringContaining('token-sandcastle-first'),
      expect.stringContaining('token-sandcastle-second'),
      expect.stringContaining('token-sandcastle-third')
    ])
    expect(wrapper.text()).toContain('first-user')
    expect(wrapper.text()).toContain('t***@example.com')
    expect(wrapper.find('.round-time strong').text()).toBe('12:00:00')

    wrapper.unmount()
  })

  it('keeps history paginated and reloads from the refresh control', async () => {
    const wrapper = mountSpeedRank()
    await flushPromises()

    expect(wrapper.findAll('.history-entry')).toHaveLength(7)
    expect(wrapper.find('.pagination-stub').exists()).toBe(true)

    await wrapper.get('.refresh-button').trigger('click')
    await flushPromises()

    expect(getSpeedRank).toHaveBeenCalledTimes(2)

    wrapper.unmount()
  })
})
