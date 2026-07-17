<template>
  <AppLayout>
    <div class="speed-rank-page">
      <main class="speed-rank-main">
        <section class="speed-rank-countdown" :style="{ '--countdown-progress': countdownProgress }">
          <div class="speed-rank-countdown-main">
            <div class="speed-rank-progress" aria-hidden="true"></div>
            <div>
              <p class="speed-rank-eyebrow">{{ t('speedRank.countdownLabel') }}</p>
              <div class="speed-rank-countdown-values" aria-live="polite">
                <strong>{{ countdownParts.hours }}</strong>
                <span>{{ t('speedRank.hours') }}</span>
                <strong>{{ countdownParts.minutes }}</strong>
                <span>{{ t('speedRank.minutes') }}</span>
                <strong>{{ countdownParts.seconds }}</strong>
                <span>{{ t('speedRank.seconds') }}</span>
              </div>
              <p class="speed-rank-round-rule">{{ t('speedRank.dailyOpen') }}</p>
            </div>
          </div>
          <div class="speed-rank-trophy" aria-hidden="true">
            <span></span>
          </div>
        </section>

        <div v-if="loading" class="speed-rank-state">
          <LoadingSpinner />
        </div>

        <template v-else>
          <section class="speed-rank-board">
            <div class="speed-rank-board-header">
              <div>
                <p class="speed-rank-eyebrow">{{ rankingDate }}</p>
                <h2>{{ t('speedRank.todayBoard') }}</h2>
              </div>
              <button type="button" class="speed-rank-refresh" :disabled="loading" @click="loadRank">
                {{ t('common.refresh') }}
              </button>
            </div>

            <div v-if="entries.length === 0" class="speed-rank-empty">
              {{ t('speedRank.empty') }}
            </div>

            <div v-else class="speed-rank-table">
              <div class="speed-rank-row speed-rank-row-head">
                <span>{{ t('speedRank.rank') }}</span>
                <span>{{ t('speedRank.player') }}</span>
                <span>{{ t('speedRank.totalTokens') }}</span>
                <span>{{ t('speedRank.rewardPreview') }}</span>
              </div>

              <article
                v-for="entry in entries"
                :key="entry.user_id"
                class="speed-rank-row"
                :class="`speed-rank-row-${entry.rank}`"
              >
                <div class="speed-rank-place">
                  <span :class="`speed-rank-place-${entry.rank}`">{{ entry.rank }}</span>
                </div>
                <div class="speed-rank-player">
                  <div>
                    <h3>{{ displayName(entry) }}</h3>
                    <p>{{ entry.email }}</p>
                  </div>
                </div>
                <div class="speed-rank-consume">
                  <strong>{{ formatNumber(entry.total_tokens) }}</strong>
                  <span>{{ t('speedRank.inputTokens') }} {{ formatCompactToken(entry.input_tokens) }} / {{ t('speedRank.outputTokens') }} {{ formatCompactToken(entry.output_tokens) }}</span>
                </div>
                <div class="speed-rank-reward">
                  <strong>+{{ formatReward(entry.reward) }}</strong>
                </div>
              </article>

              <div class="speed-rank-table-footer">
                <span>{{ t('speedRank.dataRefreshNote') }}</span>
              </div>
            </div>
          </section>

          <section class="speed-rank-history">
            <div class="speed-rank-board-header">
              <div>
                <p class="speed-rank-eyebrow">{{ t('speedRank.historyEyebrow') }}</p>
                <h2>{{ t('speedRank.historyTitle') }}</h2>
              </div>
            </div>
            <div v-if="history.length === 0" class="speed-rank-empty speed-rank-history-empty">
              {{ t('speedRank.historyEmpty') }}
            </div>
            <div v-else class="speed-rank-history-list">
              <article v-for="entry in paginatedHistory" :key="entry.reward_date" class="speed-rank-history-row">
                <div class="speed-rank-player">
                  <div>
                    <h3>{{ displayName(entry) }}</h3>
                    <p>{{ entry.reward_date }}</p>
                  </div>
                </div>
                <div class="speed-rank-consume">
                  <strong>{{ formatNumber(entry.total_tokens) }}</strong>
                  <span>{{ entry.email }}</span>
                </div>
                <div class="speed-rank-reward">
                  <strong>+{{ formatReward(entry.reward) }}</strong>
                </div>
              </article>
            </div>
            <Pagination
              v-if="history.length > historyPageSize"
              class="speed-rank-history-pagination"
              :page="historyPage"
              :page-size="historyPageSize"
              :total="history.length"
              :show-page-size-selector="false"
              @update:page="historyPage = $event"
            />
          </section>
        </template>

        <p class="speed-rank-rules-note">
          {{ t('speedRank.activityRules') }}：{{ t('speedRank.ruleRank') }} {{ t('speedRank.ruleReward') }} {{ t('speedRank.ruleReset') }}
        </p>
      </main>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import { usageAPI, type SpeedRankEntry } from '@/api/usage'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const entries = ref<SpeedRankEntry[]>([])
const history = ref<SpeedRankEntry[]>([])
const historyPage = ref(1)
const historyPageSize = 7
const nextRewardAt = ref<Date | null>(null)
const rankingDate = ref('')
const now = ref(Date.now())
let timer: number | null = null

const countdownParts = computed(() => {
  const target = nextRewardAt.value?.getTime() ?? now.value
  const totalSeconds = Math.max(0, Math.floor((target - now.value) / 1000))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  return {
    hours: String(hours).padStart(2, '0'),
    minutes: String(minutes).padStart(2, '0'),
    seconds: String(seconds).padStart(2, '0'),
  }
})

const countdownProgress = computed(() => {
  if (!nextRewardAt.value) {
    return '0deg'
  }
  const remainingMs = Math.max(0, nextRewardAt.value.getTime() - now.value)
  const progress = Math.max(0, Math.min(1, 1 - remainingMs / 86400000))
  return `${Math.round(progress * 360)}deg`
})

const paginatedHistory = computed(() => {
  const start = (historyPage.value - 1) * historyPageSize
  return history.value.slice(start, start + historyPageSize)
})

// loadRank 加载今日排行榜和下一次发奖时间。
async function loadRank() {
  loading.value = true
  try {
    const data = await usageAPI.getSpeedRank()
    entries.value = data.entries
    history.value = data.history
    historyPage.value = 1
    nextRewardAt.value = new Date(data.next_reward_at)
    rankingDate.value = data.ranking_date
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('speedRank.loadFailed')))
  } finally {
    loading.value = false
  }
}

// displayName 返回后端已脱敏的用户名，未设置用户名时展示脱敏邮箱。
function displayName(entry: SpeedRankEntry) {
  return entry.username || entry.email
}

// formatCompactToken 将 Token 数压缩成适合场景标签展示的短数字。
function formatCompactToken(value: number) {
  return new Intl.NumberFormat(undefined, {
    notation: 'compact',
    maximumFractionDigits: 1,
  }).format(value)
}

// formatNumber 格式化表格里的完整 Token 数字。
function formatNumber(value: number) {
  return new Intl.NumberFormat().format(value)
}

// formatReward 格式化奖励金额。
function formatReward(value: number) {
  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(value)
}

onMounted(() => {
  loadRank()
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (timer !== null) {
    window.clearInterval(timer)
  }
})
</script>

<style scoped>
.speed-rank-page {
  min-width: 0;
}

.speed-rank-main {
  display: grid;
  align-content: start;
  gap: 1.25rem;
  min-width: 0;
}

.speed-rank-countdown,
.speed-rank-board,
.speed-rank-history {
  border: 1px solid color-mix(in srgb, var(--app-line) 72%, transparent);
  border-radius: 0.5rem;
  background:
    linear-gradient(135deg, color-mix(in srgb, var(--app-surface) 88%, var(--app-accent) 12%), var(--app-surface)),
    var(--app-surface);
  box-shadow: 0 18px 46px color-mix(in srgb, var(--app-ink) 7%, transparent);
}

.speed-rank-board,
.speed-rank-history {
  overflow: hidden;
}

.speed-rank-countdown {
  position: relative;
  overflow: hidden;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 13rem;
  align-items: center;
  min-height: 7.25rem;
  padding: 1.35rem 1.65rem;
}

.speed-rank-countdown::after {
  position: absolute;
  top: -2.2rem;
  right: 1.7rem;
  width: 11rem;
  height: 11rem;
  border-radius: 999px;
  background: radial-gradient(circle at 36% 36%, rgba(255, 255, 255, 0.86), color-mix(in srgb, var(--app-accent) 20%, transparent) 58%, transparent 60%);
  content: '';
}

.speed-rank-progress {
  position: relative;
  display: grid;
  width: 4rem;
  height: 4rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  background:
    radial-gradient(circle at center, #ffffff 46%, transparent 47%),
    conic-gradient(var(--app-accent) var(--countdown-progress), color-mix(in srgb, var(--app-accent) 16%, var(--app-surface-muted)) 0);
  box-shadow: 0 10px 24px color-mix(in srgb, var(--app-accent) 24%, transparent);
}

.speed-rank-progress::before {
  position: absolute;
  width: 1.55rem;
  height: 1.55rem;
  border: 0.22rem solid var(--app-accent);
  border-radius: 999px;
  box-shadow: inset 0 -0.3rem 0 color-mix(in srgb, var(--app-accent) 16%, transparent);
  content: '';
}

.speed-rank-progress::after {
  position: absolute;
  width: 0.38rem;
  height: 1rem;
  border-radius: 999px;
  background: var(--app-accent);
  content: '';
  transform: rotate(-35deg) translate(0.22rem, -0.18rem);
}

.speed-rank-countdown-main {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 1.1rem;
  min-width: 0;
}

.speed-rank-countdown-values {
  display: grid;
  grid-template-columns: repeat(6, auto);
  align-items: end;
  gap: 0.55rem;
}

.speed-rank-countdown-values strong {
  color: var(--app-accent);
  font-family: var(--app-font-mono);
  font-size: clamp(2rem, 4vw, 2.75rem);
  font-weight: 900;
  line-height: 1;
}

.speed-rank-countdown-values span {
  padding-bottom: 0.22rem;
  color: var(--app-muted);
  font-size: 0.85rem;
  font-weight: 800;
}

.speed-rank-eyebrow {
  margin: 0 0 0.35rem;
  color: var(--app-soft);
  font-size: 0.78rem;
  font-weight: 800;
}

.speed-rank-round-rule {
  margin: 0.45rem 0 0;
  color: var(--app-soft);
  font-size: 0.82rem;
  font-weight: 700;
}

.speed-rank-trophy {
  position: relative;
  z-index: 1;
  height: 6rem;
}

.speed-rank-trophy span {
  position: absolute;
  right: 2.1rem;
  bottom: 0.35rem;
  width: 4.9rem;
  height: 4.4rem;
  border: 0.48rem solid color-mix(in srgb, var(--app-accent) 36%, transparent);
  border-radius: 0.7rem 0.7rem 1.7rem 1.7rem;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.82), color-mix(in srgb, var(--app-accent) 24%, transparent));
  transform: rotate(-10deg);
}

.speed-rank-trophy span::before,
.speed-rank-trophy span::after {
  position: absolute;
  content: '';
}

.speed-rank-trophy span::before {
  right: -1.55rem;
  top: 0.55rem;
  width: 1.75rem;
  height: 1.7rem;
  border: 0.38rem solid color-mix(in srgb, var(--app-accent) 28%, transparent);
  border-left: 0;
  border-radius: 0 999px 999px 0;
}

.speed-rank-trophy span::after {
  left: 1.25rem;
  bottom: -1.4rem;
  width: 2.7rem;
  height: 0.8rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-accent) 24%, transparent);
}

.speed-rank-state,
.speed-rank-empty {
  display: flex;
  min-height: 18rem;
  align-items: center;
  justify-content: center;
  color: var(--app-soft);
}

.speed-rank-board-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.35rem 1.55rem 1rem;
}

.speed-rank-board h2,
.speed-rank-history h2 {
  margin: 0;
  color: var(--app-ink);
  font-family: var(--app-font-display);
  font-size: 1.2rem;
  font-weight: 900;
}

.speed-rank-refresh {
  min-height: 2.15rem;
  padding: 0 0.9rem;
  border: 0;
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-accent) 12%, var(--app-surface-muted));
  color: var(--app-accent);
  font-size: 0.82rem;
  font-weight: 900;
}

.speed-rank-table {
  min-width: 0;
  overflow-x: auto;
}

.speed-rank-row {
  display: grid;
  grid-template-columns: 4.5rem minmax(12rem, 1fr) minmax(12rem, 0.95fr) 8.5rem;
  align-items: center;
  min-width: 45rem;
  min-height: 4.65rem;
  border-top: 1px solid color-mix(in srgb, var(--app-line) 72%, transparent);
  padding: 0 1.55rem;
  color: var(--app-ink);
  transition: background-color 160ms ease;
}

.speed-rank-row-head {
  min-height: 3.15rem;
  border-top: 0;
  background: color-mix(in srgb, var(--app-surface-muted) 52%, transparent);
  color: var(--app-soft);
  font-size: 0.78rem;
  font-weight: 900;
}

.speed-rank-place span {
  display: grid;
  place-items: center;
  border-radius: 999px;
  font-family: var(--app-font-mono);
  font-weight: 900;
}

.speed-rank-place span {
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid color-mix(in srgb, var(--app-line) 70%, transparent);
  background: var(--app-surface-muted);
  color: var(--app-muted);
}

.speed-rank-place-1 {
  background: #fff4c7 !important;
  color: #c68515 !important;
  box-shadow: 0 8px 18px rgba(223, 158, 36, 0.18);
}

.speed-rank-place-2 {
  background: #eef2fb !important;
  color: #8b9ab1 !important;
}

.speed-rank-place-3 {
  background: #fff0e9 !important;
  color: #c66c42 !important;
}

.speed-rank-row-1 {
  background: linear-gradient(90deg, rgba(255, 212, 95, 0.13), transparent 42%);
}

.speed-rank-row-2 {
  background: linear-gradient(90deg, rgba(139, 154, 177, 0.09), transparent 42%);
}

.speed-rank-row-3 {
  background: linear-gradient(90deg, rgba(198, 108, 66, 0.08), transparent 42%);
}

.speed-rank-row:not(.speed-rank-row-head):hover {
  background-color: color-mix(in srgb, var(--app-accent) 5%, transparent);
}

.speed-rank-player {
  min-width: 0;
}

.speed-rank-player h3 {
  margin: 0;
  overflow: hidden;
  color: var(--app-ink);
  font-size: 0.95rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.speed-rank-player p {
  margin: 0.25rem 0 0;
  overflow: hidden;
  color: var(--app-soft);
  font-size: 0.78rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.speed-rank-consume {
  display: grid;
  gap: 0.22rem;
  min-width: 0;
}

.speed-rank-consume strong {
  overflow: hidden;
  color: var(--app-ink);
  font-family: var(--app-font-mono);
  font-size: 1.05rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.speed-rank-consume span {
  overflow: hidden;
  color: var(--app-soft);
  font-size: 0.74rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.speed-rank-reward {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.25rem;
  border-radius: 0.45rem;
  background: color-mix(in srgb, var(--app-accent) 10%, var(--app-surface-muted));
}

.speed-rank-reward strong {
  color: var(--app-accent-strong);
  font-family: var(--app-font-mono);
  font-size: 0.95rem;
  font-weight: 900;
}

.speed-rank-table-footer {
  display: flex;
  min-height: 3rem;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid color-mix(in srgb, var(--app-line) 72%, transparent);
  padding: 0 1.55rem;
  color: var(--app-soft);
  font-size: 0.78rem;
  font-weight: 800;
}

.speed-rank-history-empty {
  min-height: 8rem;
}

.speed-rank-history-list {
  display: grid;
  min-width: 0;
  overflow-x: auto;
  border-top: 1px solid color-mix(in srgb, var(--app-line) 72%, transparent);
}

.speed-rank-history-row {
  display: grid;
  grid-template-columns: minmax(13rem, 1fr) minmax(12rem, 0.95fr) 8.5rem;
  align-items: center;
  min-width: 40rem;
  min-height: 4.8rem;
  gap: 1rem;
  border-bottom: 1px solid color-mix(in srgb, var(--app-line) 72%, transparent);
  padding: 0 1.55rem;
}

.speed-rank-history-list .speed-rank-history-row:last-child {
  border-bottom: 0;
}

.speed-rank-history-pagination {
  background: transparent !important;
}

.speed-rank-rules-note {
  margin: -0.25rem 0 0;
  color: var(--app-soft);
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.7;
}

@media (max-width: 760px) {
  .speed-rank-countdown {
    grid-template-columns: 1fr;
    padding: 1rem;
  }

  .speed-rank-trophy {
    display: none;
  }

  .speed-rank-countdown-values {
    grid-template-columns: repeat(3, auto);
    justify-content: start;
  }

  .speed-rank-board-header {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
