<template>
  <AppLayout>
    <main class="arena-page">
      <header class="arena-heading">
        <p class="arena-kicker">ARENA · {{ t('speedRank.limitedEvent') }}</p>
        <h1>{{ t('speedRank.title') }}</h1>
        <p>{{ t('speedRank.description') }}</p>
      </header>

      <section class="round-strip" :style="{ '--countdown-progress': countdownProgress }">
        <div class="round-clock" aria-hidden="true">
          <Icon name="clock" size="lg" />
        </div>
        <div class="round-time">
          <span>{{ t('speedRank.countdownLabel') }}</span>
          <strong aria-live="polite">{{ countdownText }}</strong>
        </div>
        <p class="round-note">{{ t('speedRank.dailyOpen') }}</p>
        <button
          type="button"
          class="refresh-button"
          :title="t('common.refresh')"
          :aria-label="t('common.refresh')"
          :disabled="loading"
          @click="loadRank"
        >
          <Icon name="refresh" size="md" :class="{ 'refresh-icon-spinning': loading }" />
        </button>
      </section>

      <section class="tide-board" :aria-label="t('speedRank.todayBoard')">
        <img class="tide-board-background" :src="tideBeach" alt="" aria-hidden="true" />
        <span class="tide-glint" aria-hidden="true"></span>

        <header class="tide-board-heading">
          <div>
            <p>{{ rankingDate }}</p>
            <h2>{{ t('speedRank.todayBoard') }}</h2>
          </div>
          <span>{{ t('speedRank.dataRefreshNote') }}</span>
        </header>

        <div v-if="loading" class="tide-state">
          <LoadingSpinner size="lg" />
        </div>
        <div v-else-if="entries.length === 0" class="tide-state">
          {{ t('speedRank.empty') }}
        </div>
        <ol v-else class="tide-podium">
          <li
            v-for="entry in entries"
            :key="entry.user_id"
            class="rank-entry"
            :class="`rank-entry-${entry.rank}`"
            :style="{ '--rank-delay': `${(entry.rank - 1) * 110}ms` }"
          >
            <div class="contender-label">
              <span class="rank-medal"><small>#</small>{{ entry.rank }}</span>
              <div>
                <strong>{{ displayName(entry) }}</strong>
                <span :title="formatNumber(entry.total_tokens)">
                  {{ t('speedRank.totalTokens') }} · {{ formatCompactToken(entry.total_tokens) }}
                </span>
              </div>
            </div>

            <div class="castle-stage" aria-hidden="true">
              <img
                class="castle-image"
                :src="rankCastles[entry.rank - 1]"
                alt=""
                draggable="false"
                decoding="async"
              />
            </div>

            <dl class="rank-stats">
              <div>
                <dt>{{ t('speedRank.inputTokens') }}</dt>
                <dd>{{ formatCompactToken(entry.input_tokens) }}</dd>
              </div>
              <div>
                <dt>{{ t('speedRank.outputTokens') }}</dt>
                <dd>{{ formatCompactToken(entry.output_tokens) }}</dd>
              </div>
              <div class="rank-reward">
                <dt>{{ t('speedRank.reward') }}</dt>
                <dd>+{{ formatReward(entry.reward) }}</dd>
              </div>
            </dl>
          </li>
        </ol>
      </section>

      <section class="almanac">
        <div class="almanac-heading">
          <div>
            <p class="arena-kicker">ALMANAC · {{ t('speedRank.historyEyebrow') }}</p>
            <h2>{{ t('speedRank.historyTitle') }}</h2>
          </div>
          <span>{{ t('speedRank.activityRules') }}</span>
        </div>

        <div v-if="history.length === 0" class="history-empty">
          {{ t('speedRank.historyEmpty') }}
        </div>
        <ol v-else class="history-list">
          <li
            v-for="(entry, index) in paginatedHistory"
            :key="entry.reward_date"
            class="history-entry"
            :style="{ '--history-delay': `${index * 45}ms` }"
          >
            <time>{{ entry.reward_date }}</time>
            <div>
              <strong>{{ displayName(entry) }}</strong>
              <span>{{ formatNumber(entry.total_tokens) }} Token</span>
            </div>
            <b>+{{ formatReward(entry.reward) }}</b>
          </li>
        </ol>
        <Pagination
          v-if="history.length > historyPageSize"
          class="history-pagination"
          :page="historyPage"
          :page-size="historyPageSize"
          :total="history.length"
          :show-page-size-selector="false"
          @update:page="historyPage = $event"
        />
      </section>

      <p class="rules-note">
        <Icon name="infoCircle" size="sm" aria-hidden="true" />
        <span>
          {{ t('speedRank.ruleRank') }} {{ t('speedRank.ruleReward') }}
          {{ t('speedRank.ruleReset') }}
        </span>
      </p>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import { usageAPI, type SpeedRankEntry } from '@/api/usage'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import tideBeach from '@/assets/icons/token-tide-beach.png'
import firstCastle from '@/assets/icons/token-sandcastle-first.png'
import secondCastle from '@/assets/icons/token-sandcastle-second.png'
import thirdCastle from '@/assets/icons/token-sandcastle-third.png'

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
const rankCastles = [firstCastle, secondCastle, thirdCastle]
let timer: number | null = null

const countdownParts = computed(() => {
  const target = nextRewardAt.value?.getTime() ?? now.value
  const totalSeconds = Math.max(0, Math.floor((target - now.value) / 1000))
  return {
    hours: String(Math.floor(totalSeconds / 3600)).padStart(2, '0'),
    minutes: String(Math.floor((totalSeconds % 3600) / 60)).padStart(2, '0'),
    seconds: String(totalSeconds % 60).padStart(2, '0')
  }
})

const countdownText = computed(
  () => `${countdownParts.value.hours}:${countdownParts.value.minutes}:${countdownParts.value.seconds}`
)

const countdownProgress = computed(() => {
  if (!nextRewardAt.value) return '0deg'

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
    maximumFractionDigits: 1
  }).format(value)
}

// formatNumber 格式化历史记录里的完整 Token 数字。
function formatNumber(value: number) {
  return new Intl.NumberFormat().format(value)
}

// formatReward 格式化奖励金额。
function formatReward(value: number) {
  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2
  }).format(value)
}

onMounted(() => {
  loadRank()
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (timer !== null) window.clearInterval(timer)
})
</script>

<style scoped>
.arena-page {
  display: grid;
  width: 100%;
  max-width: 90rem;
  min-width: 0;
  margin: 0 auto;
  gap: 1.5rem;
  color: var(--app-ink);
}

.arena-heading {
  max-width: 48rem;
}

.arena-heading h1,
.almanac-heading h2,
.tide-board-heading h2 {
  margin: 0;
  font-family: var(--app-font-display);
  font-weight: 900;
  letter-spacing: 0;
}

.arena-heading h1 {
  font-size: 2.25rem;
  line-height: 1.15;
}

.arena-heading > p:last-child {
  margin: 0.45rem 0 0;
  color: var(--app-muted);
  font-size: 0.95rem;
}

.arena-kicker {
  margin: 0 0 0.55rem;
  color: var(--app-soft);
  font-family: var(--app-font-mono);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.12em;
}

.round-strip {
  display: grid;
  grid-template-columns: auto minmax(12rem, 1fr) auto auto;
  align-items: center;
  gap: 1.1rem;
  min-height: 6.25rem;
  padding: 1rem 1.25rem;
  border: 1px solid color-mix(in srgb, var(--app-line) 76%, transparent);
  border-radius: 0.5rem;
  background: var(--app-surface);
}

.round-clock {
  display: grid;
  width: 3.8rem;
  height: 3.8rem;
  place-items: center;
  border-radius: 50%;
  background:
    radial-gradient(circle, var(--app-surface) 57%, transparent 59%),
    conic-gradient(var(--app-accent-strong) var(--countdown-progress), var(--app-line) 0);
  color: var(--app-muted);
}

.round-time {
  display: grid;
  gap: 0.15rem;
}

.round-time span,
.round-note {
  margin: 0;
  color: var(--app-soft);
  font-size: 0.8rem;
  font-weight: 700;
}

.round-time strong {
  font-family: var(--app-font-mono);
  font-size: 1.75rem;
  line-height: 1;
}

.refresh-button {
  display: grid;
  width: 2.5rem;
  height: 2.5rem;
  place-items: center;
  border: 1px solid var(--app-line);
  border-radius: 50%;
  background: var(--app-surface);
  color: var(--app-ink);
  cursor: pointer;
  transition: border-color 160ms ease, background-color 160ms ease, transform 160ms ease;
}

.refresh-button:hover:not(:disabled) {
  border-color: var(--app-soft);
  background: var(--app-surface-muted);
  transform: translateY(-1px);
}

.refresh-button:focus-visible {
  outline: 2px solid var(--app-accent);
  outline-offset: 2px;
}

.refresh-button:disabled {
  cursor: wait;
  opacity: 0.6;
}

.refresh-icon-spinning {
  animation: refresh-spin 800ms linear infinite;
}

.tide-board {
  position: relative;
  isolation: isolate;
  min-height: clamp(36rem, 54vw, 43rem);
  overflow: hidden;
  border: 1px solid #c6aa75;
  border-radius: 0.5rem;
  background: #decba3;
  box-shadow: 0 1.15rem 2.75rem rgba(71, 55, 28, 0.14);
}

.tide-board::before {
  position: absolute;
  z-index: 1;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(255, 252, 239, 0.1) 0%, transparent 32%),
    linear-gradient(0deg, rgba(79, 54, 27, 0.2) 0%, transparent 32%);
  content: '';
  pointer-events: none;
}

.tide-board::after {
  position: absolute;
  z-index: 5;
  inset: 0;
  border: 1px solid rgba(255, 255, 255, 0.3);
  content: '';
  pointer-events: none;
}

.tide-board-background {
  position: absolute;
  z-index: 0;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  user-select: none;
}

.tide-glint {
  position: absolute;
  z-index: 2;
  top: 39%;
  left: -8%;
  width: 116%;
  height: 0.75rem;
  border-top: 2px solid rgba(255, 255, 255, 0.66);
  border-bottom: 1px solid rgba(225, 250, 250, 0.44);
  filter: drop-shadow(0 0.25rem 0.3rem rgba(255, 255, 255, 0.42));
  opacity: 0.7;
  transform: rotate(-0.5deg);
  animation: tide-drift 9s ease-in-out infinite alternate;
  pointer-events: none;
}

.tide-board-heading {
  position: absolute;
  z-index: 4;
  top: 1.2rem;
  right: 1.2rem;
  left: 1.2rem;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  color: #263c3c;
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.72);
}

.tide-board-heading p {
  margin: 0 0 0.2rem;
  font-family: var(--app-font-mono);
  font-size: 0.7rem;
  font-weight: 900;
}

.tide-board-heading h2 {
  font-size: 1.55rem;
}

.tide-board-heading > span {
  max-width: 17rem;
  padding: 0.4rem 0.65rem;
  border: 1px solid rgba(57, 69, 62, 0.18);
  border-radius: 999px;
  background: rgba(255, 253, 243, 0.76);
  font-size: 0.7rem;
  font-weight: 800;
  text-align: right;
  backdrop-filter: blur(0.4rem);
}

.tide-state {
  position: absolute;
  z-index: 4;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #4a402f;
  font-weight: 800;
}

.tide-podium {
  position: absolute;
  z-index: 3;
  inset: 7.2rem 1.5rem 1.15rem;
  display: grid;
  grid-template-areas: 'second first third';
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.12fr) minmax(0, 1fr);
  align-items: end;
  gap: clamp(0.6rem, 2vw, 1.6rem);
  margin: 0;
  padding: 0;
  list-style: none;
}

.rank-entry {
  display: grid;
  min-width: 0;
  align-content: end;
  justify-items: center;
  animation: castle-arrive 680ms cubic-bezier(0.2, 0.75, 0.25, 1) both;
  animation-delay: var(--rank-delay);
}

.rank-entry-1 {
  z-index: 3;
  grid-area: first;
}

.rank-entry-2 {
  z-index: 2;
  grid-area: second;
}

.rank-entry-3 {
  z-index: 1;
  grid-area: third;
}

.contender-label {
  display: flex;
  width: min(100%, 18rem);
  min-width: 0;
  min-height: 3.1rem;
  align-items: center;
  gap: 0.6rem;
  padding: 0.42rem 0.72rem 0.42rem 0.42rem;
  border: 1px solid rgba(73, 59, 36, 0.32);
  border-radius: 999px;
  background: rgba(255, 253, 245, 0.88);
  box-shadow: 0 0.3rem 0.8rem rgba(89, 65, 31, 0.13);
  color: #2f2b23;
  backdrop-filter: blur(0.45rem);
}

.rank-medal {
  position: relative;
  display: grid;
  width: 2.15rem;
  height: 2.15rem;
  flex: 0 0 auto;
  grid-auto-flow: column;
  place-content: center;
  align-items: baseline;
  border: 1px solid #b77d16;
  border-radius: 50%;
  background: #f4c95d;
  color: #4c330c;
  font-family: var(--app-font-mono);
  font-size: 0.88rem;
  font-weight: 900;
}

.rank-medal small {
  font-size: 0.55rem;
}

.rank-entry-1 .rank-medal {
  animation: medal-float 3.2s ease-in-out 1s infinite;
}

.rank-entry-2 .rank-medal {
  border-color: #879099;
  background: #e3e6e8;
  color: #3c4449;
}

.rank-entry-3 .rank-medal {
  border-color: #a56139;
  background: #dca176;
  color: #512c19;
}

.contender-label > div {
  display: grid;
  min-width: 0;
  gap: 0.08rem;
}

.contender-label strong,
.contender-label span:last-child {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contender-label strong {
  font-size: 0.82rem;
}

.contender-label span:last-child {
  color: #735e3f;
  font-size: 0.67rem;
  font-weight: 800;
}

.castle-stage {
  display: flex;
  width: 100%;
  min-height: 14rem;
  align-items: flex-end;
  justify-content: center;
}

.rank-entry-1 .castle-stage {
  min-height: clamp(19rem, 26vw, 23.5rem);
}

.rank-entry-2 .castle-stage {
  min-height: clamp(13rem, 20vw, 17.5rem);
}

.rank-entry-3 .castle-stage {
  min-height: clamp(13rem, 19vw, 16.5rem);
}

.castle-image {
  display: block;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  transform-origin: center bottom;
  user-select: none;
  transition: transform 240ms ease, filter 240ms ease;
}

.rank-entry-1 .castle-image {
  width: min(100%, 18rem);
  filter: drop-shadow(0 1rem 1.25rem rgba(82, 55, 24, 0.24));
}

.rank-entry-2 .castle-image {
  width: min(100%, 21rem);
  filter: drop-shadow(0 0.8rem 1rem rgba(82, 55, 24, 0.2));
}

.rank-entry-3 .castle-image {
  width: min(100%, 15.5rem);
  filter: drop-shadow(0 0.75rem 1rem rgba(82, 55, 24, 0.18));
}

.rank-entry:hover .castle-image {
  filter: drop-shadow(0 1.1rem 1.35rem rgba(82, 55, 24, 0.28));
  transform: translateY(-0.3rem) scale(1.015);
}

.rank-stats {
  display: grid;
  width: min(100%, 18rem);
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: -0.15rem 0 0;
  overflow: hidden;
  border: 1px solid rgba(73, 59, 36, 0.25);
  border-radius: 0.4rem;
  background: rgba(255, 252, 241, 0.82);
  color: #433a2c;
  backdrop-filter: blur(0.4rem);
}

.rank-stats div {
  display: grid;
  min-width: 0;
  gap: 0.08rem;
  padding: 0.45rem 0.38rem;
  text-align: center;
}

.rank-stats div + div {
  border-left: 1px solid rgba(73, 59, 36, 0.17);
}

.rank-stats dt {
  overflow: hidden;
  color: #7a694f;
  font-size: 0.6rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rank-stats dd {
  margin: 0;
  overflow: hidden;
  font-family: var(--app-font-mono);
  font-size: 0.75rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rank-stats .rank-reward dd {
  color: #ad493c;
}

.almanac {
  padding: 1.5rem 0 0;
  border-top: 1px solid var(--app-line);
}

.almanac-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.almanac-heading h2 {
  font-size: 1.5rem;
}

.almanac-heading > span {
  color: var(--app-soft);
  font-size: 0.75rem;
  font-weight: 800;
}

.history-list {
  display: grid;
  margin: 0;
  padding: 0;
  border-top: 1px solid var(--app-line);
  list-style: none;
}

.history-entry {
  display: grid;
  grid-template-columns: 8rem minmax(0, 1fr) auto;
  align-items: center;
  gap: 1rem;
  min-height: 4.4rem;
  border-bottom: 1px solid var(--app-line);
  animation: history-reveal 420ms ease-out both;
  animation-delay: var(--history-delay);
}

.history-entry time {
  color: var(--app-soft);
  font-family: var(--app-font-mono);
  font-size: 0.78rem;
  font-weight: 800;
}

.history-entry div {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: 0.8rem;
}

.history-entry strong,
.history-entry span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-entry span {
  color: var(--app-soft);
  font-size: 0.78rem;
}

.history-entry > b {
  color: var(--app-accent-strong);
  font-family: var(--app-font-mono);
}

.history-empty {
  padding: 3rem 0;
  color: var(--app-soft);
  text-align: center;
}

.history-pagination {
  background: transparent !important;
}

.rules-note {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
  margin: -0.5rem 0 0;
  color: var(--app-soft);
  font-size: 0.72rem;
  line-height: 1.7;
}

.rules-note svg {
  flex: 0 0 auto;
  margin-top: 0.2rem;
}

@keyframes castle-arrive {
  from {
    opacity: 0;
    transform: translateY(1.4rem) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes tide-drift {
  from { transform: translateX(-2%) rotate(-0.5deg); }
  to { transform: translateX(2%) rotate(0.25deg); }
}

@keyframes medal-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-0.18rem); }
}

@keyframes history-reveal {
  from {
    opacity: 0;
    transform: translateY(0.4rem);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes refresh-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 900px) {
  .tide-board {
    min-height: 39rem;
  }

  .tide-podium {
    inset-inline: 0.9rem;
    gap: 0.55rem;
  }

  .castle-stage {
    min-height: 12rem;
  }

  .rank-entry-1 .castle-stage {
    min-height: 20rem;
  }

  .rank-entry-2 .castle-stage,
  .rank-entry-3 .castle-stage {
    min-height: 14.5rem;
  }

  .contender-label {
    padding-right: 0.5rem;
  }
}

@media (max-width: 760px) {
  .arena-heading h1 {
    font-size: 1.8rem;
  }

  .round-strip {
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 0.8rem;
    min-height: 5.5rem;
    padding: 0.8rem;
  }

  .round-note {
    display: none;
  }

  .round-clock {
    width: 3.25rem;
    height: 3.25rem;
  }

  .round-time strong {
    font-size: 1.4rem;
  }

  .tide-board {
    min-height: 45rem;
  }

  .tide-board-background {
    object-position: 54% center;
  }

  .tide-board-heading {
    top: 1rem;
    right: 1rem;
    left: 1rem;
  }

  .tide-board-heading > span {
    display: none;
  }

  .tide-glint {
    top: 18%;
  }

  .tide-podium {
    position: relative;
    inset: auto;
    grid-template-areas: none;
    grid-template-columns: minmax(0, 1fr);
    align-items: stretch;
    gap: 0.8rem;
    padding: 7.4rem 1rem 1rem;
  }

  .rank-entry {
    display: grid;
    grid-template-areas:
      'castle contender'
      'castle stats';
    grid-template-columns: 7.5rem minmax(0, 1fr);
    align-content: center;
    align-items: center;
    gap: 0.45rem 0.75rem;
    min-height: 11.2rem;
  }

  .rank-entry-1,
  .rank-entry-2,
  .rank-entry-3 {
    grid-area: auto;
  }

  .rank-entry-1 { order: 1; }
  .rank-entry-2 { order: 2; }
  .rank-entry-3 { order: 3; }

  .contender-label {
    width: 100%;
    grid-area: contender;
  }

  .castle-stage,
  .rank-entry-1 .castle-stage,
  .rank-entry-2 .castle-stage,
  .rank-entry-3 .castle-stage {
    width: 7.5rem;
    min-height: 10.5rem;
    grid-area: castle;
  }

  .rank-entry-2 .castle-stage {
    min-height: 8rem;
  }

  .rank-entry-2 .castle-image {
    width: 8.5rem;
    max-width: 8.5rem;
  }

  .rank-stats {
    width: 100%;
    grid-area: stats;
  }

  .almanac-heading > span {
    display: none;
  }

  .history-entry {
    grid-template-columns: 5.5rem minmax(0, 1fr) auto;
    gap: 0.65rem;
  }

  .history-entry div {
    display: grid;
    gap: 0.15rem;
  }
}

@media (max-width: 460px) {
  .arena-page {
    gap: 1.15rem;
  }

  .round-clock {
    display: none;
  }

  .round-strip {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .tide-board {
    min-height: 42rem;
  }

  .tide-podium {
    padding-inline: 0.7rem;
  }

  .rank-entry {
    grid-template-columns: 6.3rem minmax(0, 1fr);
    gap: 0.35rem 0.55rem;
    min-height: 10.25rem;
  }

  .castle-stage,
  .rank-entry-1 .castle-stage,
  .rank-entry-2 .castle-stage,
  .rank-entry-3 .castle-stage {
    width: 6.3rem;
    min-height: 9.5rem;
  }

  .rank-entry-2 .castle-image {
    width: 7rem;
    max-width: 7rem;
  }

  .rank-medal {
    width: 1.9rem;
    height: 1.9rem;
    font-size: 0.8rem;
  }

  .contender-label {
    min-height: 2.8rem;
    gap: 0.45rem;
  }

  .rank-stats div {
    padding-inline: 0.2rem;
  }

  .rank-stats dt {
    font-size: 0.55rem;
  }

  .rank-stats dd {
    font-size: 0.68rem;
  }

  .history-entry {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .history-entry time {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .rank-entry,
  .rank-entry-1 .rank-medal,
  .history-entry,
  .tide-glint,
  .refresh-icon-spinning {
    animation: none;
  }

  .castle-image,
  .refresh-button {
    transition: none;
  }
}
</style>
