<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div class="relative w-full sm:max-w-md">
          <Icon
            name="search"
            size="md"
            class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
          />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('availableChannels.searchPlaceholder')"
            class="channel-search input pl-10"
          />
        </div>

        <button
          type="button"
          class="btn btn-secondary self-end sm:self-auto"
          :disabled="loading"
          :title="t('common.refresh', 'Refresh')"
          @click="loadChannels"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div
        v-if="loading"
        class="flex min-h-[420px] items-center justify-center rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800"
      >
        <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
      </div>

      <div
        v-else-if="filteredChannels.length === 0"
        class="flex min-h-[420px] flex-col items-center justify-center rounded-lg border border-gray-200 bg-white px-6 text-center dark:border-dark-700 dark:bg-dark-800"
      >
        <Icon name="inbox" size="xl" class="mb-3 text-gray-400" />
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ searchQuery.trim() ? t('availableChannels.noSearchResults') : t('availableChannels.empty') }}
        </p>
      </div>

      <div
        v-else
        class="grid min-h-[540px] gap-4 lg:h-[calc(100vh-10rem)] lg:grid-cols-[280px_minmax(0,1fr)]"
      >
        <aside
          class="flex min-h-0 flex-col overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800"
        >
          <div class="flex items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-dark-700">
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('availableChannels.channelDirectory') }}
            </h2>
            <span class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('availableChannels.channelCount', { count: filteredChannels.length }) }}
            </span>
          </div>

          <nav class="max-h-80 overflow-y-auto lg:max-h-none lg:flex-1" :aria-label="t('availableChannels.channelDirectory')">
            <button
              v-for="channel in filteredChannels"
              :key="channel.name"
              type="button"
              data-testid="channel-option"
              class="relative block w-full border-b border-gray-100 px-4 py-3 text-left transition-colors last:border-b-0 hover:bg-gray-50 dark:border-dark-700/70 dark:hover:bg-dark-700/50"
              :class="
                selectedChannel?.name === channel.name
                  ? 'bg-gray-50 before:absolute before:inset-y-0 before:left-0 before:w-0.5 before:bg-primary-500 dark:bg-dark-700/60'
                  : ''
              "
              :aria-pressed="selectedChannel?.name === channel.name"
              @click="selectChannel(channel)"
            >
              <span class="block break-words text-sm font-medium text-gray-900 dark:text-white">
                {{ channel.name }}
              </span>
              <span class="mt-2 flex flex-wrap items-center gap-1.5">
                <span
                  v-for="section in channel.platforms"
                  :key="section.platform"
                  :class="[
                    'inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-[10px] font-medium uppercase',
                    platformBadgeClass(section.platform),
                  ]"
                >
                  <PlatformIcon :platform="section.platform as GroupPlatform" size="xs" />
                  {{ platformLabel(section.platform) }}
                </span>
                <span class="ml-auto text-[11px] text-gray-400 dark:text-gray-500">
                  {{ t('availableChannels.modelCount', { count: channelModelCount(channel) }) }}
                </span>
              </span>
            </button>
          </nav>
        </aside>

        <article
          v-if="selectedChannel && selectedPlatform"
          data-testid="channel-detail"
          class="min-h-0 overflow-y-auto rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800"
        >
          <header class="border-b border-gray-200 px-5 py-5 md:px-6 dark:border-dark-700">
            <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
              <div class="min-w-0">
                <h2 class="break-words text-xl font-semibold text-gray-950 dark:text-white">
                  {{ selectedChannel.name }}
                </h2>
                <p
                  v-if="selectedChannel.description"
                  class="mt-3 max-w-4xl border-l-2 border-primary-400 bg-gray-50 px-3 py-2 text-sm leading-6 text-gray-600 dark:bg-dark-700/40 dark:text-gray-300"
                >
                  {{ selectedChannel.description }}
                </p>
              </div>

              <div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
                <span>{{ t('availableChannels.groupCount', { count: selectedPlatform.groups.length }) }}</span>
                <span>{{ t('availableChannels.modelCount', { count: selectedPlatform.supported_models.length }) }}</span>
              </div>
            </div>

            <div
              class="mt-5 flex flex-wrap gap-2"
              role="tablist"
              :aria-label="t('availableChannels.platforms')"
            >
              <button
                v-for="section in selectedChannel.platforms"
                :key="section.platform"
                type="button"
                role="tab"
                data-testid="platform-tab"
                class="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors"
                :class="[
                  platformBadgeClass(section.platform),
                  selectedPlatform.platform === section.platform
                    ? 'ring-1 ring-current ring-offset-1 dark:ring-offset-dark-800'
                    : 'opacity-65 hover:opacity-100',
                ]"
                :aria-selected="selectedPlatform.platform === section.platform"
                @click="selectedPlatformName = section.platform"
              >
                <PlatformIcon :platform="section.platform as GroupPlatform" size="xs" />
                {{ platformLabel(section.platform) }}
              </button>
            </div>
          </header>

          <div class="px-5 py-6 md:px-6">
            <section aria-labelledby="available-channel-groups">
              <div class="flex items-center justify-between gap-3">
                <h3 id="available-channel-groups" class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('availableChannels.availableGroups') }}
                </h3>
                <span class="text-xs text-gray-400 dark:text-gray-500">
                  {{ t('availableChannels.groupCount', { count: selectedPlatform.groups.length }) }}
                </span>
              </div>

              <div class="mt-3 divide-y divide-gray-100 border-y border-gray-200 dark:divide-dark-700 dark:border-dark-700">
                <div
                  v-for="group in selectedPlatform.groups"
                  :key="group.id"
                  class="flex flex-col gap-2 py-3 sm:flex-row sm:items-center sm:justify-between"
                >
                  <GroupBadge
                    :name="group.name"
                    :platform="group.platform as GroupPlatform"
                    :subscription-type="group.subscription_type as SubscriptionType"
                    :rate-multiplier="group.rate_multiplier"
                    :peak-rate-enabled="group.peak_rate_enabled"
                    :peak-start="group.peak_start"
                    :peak-end="group.peak_end"
                    :peak-rate-multiplier="group.peak_rate_multiplier"
                    always-show-rate
                  />
                  <span
                    class="inline-flex w-fit items-center gap-1 text-[11px] font-medium uppercase"
                    :class="
                      group.is_exclusive
                        ? 'text-purple-600 dark:text-purple-400'
                        : 'text-gray-500 dark:text-gray-400'
                    "
                    :title="
                      group.is_exclusive
                        ? t('availableChannels.exclusiveTooltip')
                        : t('availableChannels.publicTooltip')
                    "
                  >
                    <Icon :name="group.is_exclusive ? 'shield' : 'globe'" size="xs" />
                    {{ group.is_exclusive ? t('availableChannels.exclusive') : t('availableChannels.public') }}
                  </span>
                </div>
              </div>
            </section>

            <section class="mt-7 border-t border-gray-200 pt-6 dark:border-dark-700" aria-labelledby="available-channel-models">
              <div class="flex items-center justify-between gap-3">
                <h3 id="available-channel-models" class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('availableChannels.supportedModels') }}
                </h3>
                <span class="text-xs text-gray-400 dark:text-gray-500">
                  {{ t('availableChannels.modelCount', { count: selectedPlatform.supported_models.length }) }}
                </span>
              </div>

              <div v-if="selectedPlatform.supported_models.length > 0" class="mt-3 flex flex-wrap gap-2">
                <SupportedModelChip
                  v-for="model in selectedPlatform.supported_models"
                  :key="`${selectedPlatform.platform}-${model.name}`"
                  :model="model"
                  pricing-key-prefix="availableChannels.pricing"
                  :no-pricing-label="t('availableChannels.noPricing')"
                  :show-platform="false"
                  :platform-hint="selectedPlatform.platform"
                />
              </div>
              <p v-else class="mt-3 text-sm text-gray-400">
                {{ t('availableChannels.noModels') }}
              </p>
            </section>
          </div>
        </article>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import SupportedModelChip from '@/components/channels/SupportedModelChip.vue'
import userChannelsAPI, {
  type UserAvailableChannel,
  type UserChannelPlatformSection,
} from '@/api/channels'
import type { GroupPlatform, SubscriptionType } from '@/types'
import { platformBadgeClass, platformLabel } from '@/utils/platformColors'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const channels = ref<UserAvailableChannel[]>([])
const loading = ref(false)
const searchQuery = ref('')
const selectedChannelName = ref('')
const selectedPlatformName = ref('')

/**
 * 搜索命中渠道名或描述时保留完整渠道；命中平台、分组或模型时只保留相关平台，
 * 让右侧详情与用户的搜索目标一致。
 */
const filteredChannels = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return channels.value

  return channels.value
    .map((channel) => {
      const channelMatched =
        channel.name.toLowerCase().includes(query) ||
        channel.description.toLowerCase().includes(query)
      if (channelMatched) return channel

      const platforms = channel.platforms.filter(
        (section) =>
          section.platform.toLowerCase().includes(query) ||
          section.groups.some((group) => group.name.toLowerCase().includes(query)) ||
          section.supported_models.some((model) => model.name.toLowerCase().includes(query)),
      )
      if (platforms.length === 0) return null
      return { ...channel, platforms }
    })
    .filter((channel): channel is UserAvailableChannel => channel !== null)
})

/** 当前选择不在搜索结果中时展示第一项，避免详情与左侧结果脱节。 */
const selectedChannel = computed<UserAvailableChannel | null>(() => {
  const matched = filteredChannels.value.find((channel) => channel.name === selectedChannelName.value)
  if (matched) return matched
  return filteredChannels.value[0] || null
})

/** 渠道切换或搜索缩小平台范围后，详情展示当前结果中的第一个有效平台。 */
const selectedPlatform = computed<UserChannelPlatformSection | null>(() => {
  if (!selectedChannel.value) return null
  const matched = selectedChannel.value.platforms.find(
    (section) => section.platform === selectedPlatformName.value,
  )
  if (matched) return matched
  return selectedChannel.value.platforms[0] || null
})

/** 统计渠道各平台的展示模型总数。 */
function channelModelCount(channel: UserAvailableChannel): number {
  return channel.platforms.reduce((count, section) => count + section.supported_models.length, 0)
}

/** 切换渠道时同步选择它的首个平台。 */
function selectChannel(channel: UserAvailableChannel) {
  selectedChannelName.value = channel.name
  selectedPlatformName.value = channel.platforms[0].platform
}

/** 重新读取可用渠道；请求失败沿用页面统一错误提示。 */
async function loadChannels() {
  loading.value = true
  try {
    channels.value = await userChannelsAPI.getAvailable()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

onMounted(loadChannels)
</script>
