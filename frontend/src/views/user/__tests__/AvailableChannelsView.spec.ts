import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AvailableChannelsView from '../AvailableChannelsView.vue'

const { getAvailable, showError } = vi.hoisted(() => ({
  getAvailable: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api/channels', () => ({
  default: { getAvailable },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: { count?: number }) =>
        params?.count === undefined ? key : `${key}:${params.count}`,
    }),
  }
})

const channels = [
  {
    name: 'Alpha Channel',
    description: 'Primary coding channel',
    platforms: [
      {
        platform: 'anthropic',
        groups: [
          {
            id: 1,
            name: 'Claude Group',
            platform: 'anthropic',
            subscription_type: 'standard',
            rate_multiplier: 0.5,
            peak_rate_enabled: false,
            peak_start: '',
            peak_end: '',
            peak_rate_multiplier: 0,
            is_exclusive: false,
          },
        ],
        supported_models: [
          { name: 'claude-sonnet', platform: 'anthropic', pricing: null },
        ],
      },
      {
        platform: 'openai',
        groups: [
          {
            id: 2,
            name: 'GPT Group',
            platform: 'openai',
            subscription_type: 'standard',
            rate_multiplier: 0.3,
            peak_rate_enabled: false,
            peak_start: '',
            peak_end: '',
            peak_rate_multiplier: 0,
            is_exclusive: true,
          },
        ],
        supported_models: [
          { name: 'gpt-codex', platform: 'openai', pricing: null },
        ],
      },
    ],
  },
  {
    name: 'Beta Channel',
    description: 'Image channel',
    platforms: [
      {
        platform: 'gemini',
        groups: [
          {
            id: 3,
            name: 'Gemini Group',
            platform: 'gemini',
            subscription_type: 'standard',
            rate_multiplier: 0.4,
            peak_rate_enabled: false,
            peak_start: '',
            peak_end: '',
            peak_rate_multiplier: 0,
            is_exclusive: false,
          },
        ],
        supported_models: [
          { name: 'gemini-image', platform: 'gemini', pricing: null },
        ],
      },
    ],
  },
]

/** 挂载渠道页并用最小替身保留可观测的模型和分组文本。 */
function mountAvailableChannels() {
  return mount(AvailableChannelsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
        PlatformIcon: true,
        GroupBadge: {
          props: ['name'],
          template: '<span class="group-badge-stub">{{ name }}</span>',
        },
        SupportedModelChip: {
          props: ['model'],
          template: '<span class="model-chip-stub">{{ model.name }}</span>',
        },
      },
    },
  })
}

describe('AvailableChannelsView', () => {
  beforeEach(() => {
    getAvailable.mockReset().mockResolvedValue(channels)
    showError.mockReset()
  })

  it('uses channel navigation and switches the detail panel without a table', async () => {
    const wrapper = mountAvailableChannels()
    await flushPromises()

    expect(getAvailable).toHaveBeenCalledTimes(1)
    expect(wrapper.find('table').exists()).toBe(false)
    expect(wrapper.findAll('[data-testid="channel-option"]')).toHaveLength(2)
    expect(wrapper.get('[data-testid="channel-detail"]').text()).toContain('Alpha Channel')
    expect(wrapper.get('[data-testid="channel-detail"]').text()).toContain('claude-sonnet')

    await wrapper.findAll('[data-testid="platform-tab"]')[1].trigger('click')
    expect(wrapper.get('[data-testid="channel-detail"]').text()).toContain('gpt-codex')
    expect(wrapper.get('[data-testid="channel-detail"]').text()).not.toContain('claude-sonnet')

    await wrapper.findAll('[data-testid="channel-option"]')[1].trigger('click')
    expect(wrapper.get('[data-testid="channel-detail"]').text()).toContain('Beta Channel')
    expect(wrapper.get('[data-testid="channel-detail"]').text()).toContain('gemini-image')
  })

  it('narrows both the channel navigation and detail panel from a model search', async () => {
    const wrapper = mountAvailableChannels()
    await flushPromises()

    await wrapper.get('.channel-search').setValue('gemini-image')

    expect(wrapper.findAll('[data-testid="channel-option"]')).toHaveLength(1)
    expect(wrapper.get('[data-testid="channel-detail"]').text()).toContain('Beta Channel')
    expect(wrapper.get('[data-testid="channel-detail"]').text()).toContain('gemini-image')
  })
})
