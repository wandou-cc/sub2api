<template>
  <section class="lottery-page">
    <header class="lottery-header">
      <div class="lottery-header-bar">
        <span class="lottery-eyebrow">BLINDBOX · {{ t('payment.lottery.activityLabel') }}</span>
        <div class="lottery-header-actions">
          <router-link to="/purchase" class="lottery-recharge-link">
            <Icon name="creditCard" size="sm" />
            {{ t('payment.lottery.recharge') }}
          </router-link>
          <router-link
            :to="{ path: '/docs', query: { cat: 'activities', page: 'recharge-lottery' } }"
            class="lottery-rules-link"
          >
            {{ t('payment.lottery.fullRules') }}
            <Icon name="arrowRight" size="sm" />
          </router-link>
        </div>
      </div>
      <h1>{{ t('payment.lottery.title') }}</h1>
      <p>{{ t('payment.lottery.description') }}</p>
    </header>

    <div v-if="loading" class="lottery-state lottery-panel">
      <span class="lottery-spinner"></span>
      <span>{{ t('common.loading') }}</span>
    </div>

    <div v-else-if="errorMessage" class="lottery-state lottery-panel lottery-error">
      <Icon name="exclamationCircle" size="lg" />
      <span>{{ errorMessage }}</span>
      <button type="button" class="lottery-retry" @click="loadOverview">
        <Icon name="refresh" size="sm" />
        {{ t('common.refresh') }}
      </button>
    </div>

    <template v-else-if="overview">
      <section class="lottery-summary" :aria-label="t('payment.lottery.title')">
        <div class="summary-item summary-count" aria-live="polite">
          <span>{{ t('payment.lottery.pendingLabel') }}</span>
          <div>
            <strong>{{ overview.pending_count }}</strong>
            <small>{{ t('payment.lottery.boxUnit') }}</small>
          </div>
          <p>{{ t('payment.lottery.summaryPendingHint') }}</p>
        </div>
        <div class="summary-item">
          <span>{{ t('payment.lottery.summaryTrigger') }}</span>
          <strong>{{ t('payment.lottery.summaryTriggerValue') }}</strong>
        </div>
        <div class="summary-item">
          <span>{{ t('payment.lottery.summaryReward') }}</span>
          <strong>{{ t('payment.lottery.summaryRewardValue') }}</strong>
        </div>
      </section>

      <section class="lottery-panel lottery-pending" :aria-label="t('payment.lottery.pendingTitle')">
        <div class="section-heading">
          <div>
            <span>{{ t('payment.lottery.unopenedLabel') }}</span>
            <h2>{{ t('payment.lottery.pendingTitle') }}</h2>
          </div>
          <strong>{{ t('payment.lottery.pendingCount', { count: overview.pending_count }) }}</strong>
        </div>

        <div v-if="pendingOpportunities.length" class="blind-box-list">
          <article v-for="opportunity in pendingOpportunities" :key="opportunity.order_id" class="blind-box-row">
            <div class="blind-box-visual">
              <span class="blind-box-icon">
                <Icon name="gift" size="xl" />
              </span>
              <small>#{{ opportunity.order_id }}</small>
            </div>
            <div class="blind-box-details">
              <div class="blind-box-title">
                <strong>{{ t('payment.lottery.balanceBox') }}</strong>
              </div>
              <dl>
                <div>
                  <dt>{{ t('payment.lottery.credited') }}</dt>
                  <dd>${{ opportunity.recharge_amount.toFixed(2) }}</dd>
                </div>
                <div>
                  <dt>{{ t('payment.lottery.maxRarity') }}</dt>
                  <dd>{{ t(`payment.lottery.rarity.${opportunity.max_rarity}`) }}</dd>
                </div>
              </dl>
            </div>
            <button
              type="button"
              class="lottery-open-button"
              @click="lotteryStore.openDialog(opportunity.order_id)"
            >
              <Icon name="gift" size="sm" />
              {{ t('payment.lottery.open') }}
            </button>
          </article>
        </div>

        <div v-else class="lottery-empty">
          <span><Icon name="gift" size="xl" /></span>
          <div>
            <h3>{{ t('payment.lottery.emptyTitle') }}</h3>
            <p>{{ t('payment.lottery.noPending') }}</p>
          </div>
          <router-link to="/purchase" class="lottery-recharge-link lottery-empty-action">
            <Icon name="creditCard" size="sm" />
            {{ t('payment.lottery.recharge') }}
          </router-link>
        </div>
      </section>

      <section class="lottery-panel lottery-history" :aria-label="t('payment.lottery.recentResults')">
        <div class="section-heading history-heading">
          <div>
            <span>{{ t('payment.lottery.rewardLog') }}</span>
            <h2>{{ t('payment.lottery.recentResults') }}</h2>
          </div>
        </div>

        <div v-if="displayedResults.length" class="history-list">
          <article v-for="result in displayedResults" :key="result.order_id" class="history-row">
            <span :class="rarityClass(result.rarity)" class="history-rarity">
              {{ t(`payment.lottery.rarity.${result.rarity}`) }}
            </span>
            <div class="history-reward">
              <strong>+${{ result.reward_amount.toFixed(2) }}</strong>
              <small>{{ formatDateTime(result.claimed_at) }}</small>
            </div>
            <div class="history-context">
              <span>{{ t('payment.lottery.historyRecharge', { amount: result.recharge_amount.toFixed(2) }) }}</span>
              <small>{{ t('payment.lottery.order', { id: result.order_id }) }}</small>
            </div>
          </article>
        </div>
        <p v-else class="history-empty">{{ t('payment.lottery.noResults') }}</p>
      </section>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRechargeLotteryStore } from '@/stores/rechargeLottery'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'
import type { RechargeLotteryRarity } from '@/types/payment'

const { t } = useI18n()
const lotteryStore = useRechargeLotteryStore()
const loading = computed(() => lotteryStore.loading)
const overview = computed(() => lotteryStore.overview)
const errorMessage = computed(() => {
  return lotteryStore.error
    ? extractApiErrorMessage(lotteryStore.error, t('payment.lottery.loadFailed'))
    : ''
})

// 活动页只展示尚未领取的资格。
const pendingOpportunities = computed(() => {
  return overview.value ? overview.value.opportunities.filter(item => !item.claimed) : []
})

// 活动页展示最近六条领取结果。
const displayedResults = computed(() => {
  return overview.value ? overview.value.opportunities.filter(item => item.claimed).slice(0, 6) : []
})

// 重试加载失败的盲盒机会，错误继续由共享状态展示。
async function loadOverview(): Promise<void> {
  try {
    await lotteryStore.fetchOverview()
  } catch (cause: unknown) {
    console.error('[recharge-lottery] Failed to retry opportunities:', cause)
  }
}

// 按稀有度返回统一的结果标识颜色。
function rarityClass(rarity: RechargeLotteryRarity | ''): string {
  switch (rarity) {
    case 'rare':
      return 'rarity-rare'
    case 'epic':
      return 'rarity-epic'
    case 'epic_plus':
      return 'rarity-epic-plus'
    case 'legendary':
      return 'rarity-legendary'
    default:
      return 'rarity-common'
  }
}
</script>

<style scoped>
.lottery-page {
  width: 100%;
  max-width: 80rem;
  min-width: 0;
  margin: 0 auto;
  padding: 0.5rem 0 4rem;
  color: var(--app-ink);
}

.lottery-header {
  padding: 0.5rem 0 2.5rem;
}

.lottery-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
}

.lottery-eyebrow {
  color: var(--app-soft);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.lottery-header-actions {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.lottery-recharge-link,
.lottery-rules-link,
.lottery-open-button,
.lottery-retry {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  font-size: 0.82rem;
  font-weight: 750;
  text-decoration: none;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease, transform 160ms ease;
}

.lottery-recharge-link {
  border: 1px solid var(--app-ink);
  border-radius: 0.5rem;
  padding: 0.55rem 0.85rem;
  background: var(--app-ink);
  color: var(--app-surface);
}

.lottery-recharge-link:hover {
  transform: translateY(-1px);
}

.lottery-rules-link {
  color: var(--app-muted);
}

.lottery-rules-link:hover {
  color: var(--app-accent-strong);
}

.lottery-header h1 {
  max-width: 45rem;
  margin: 1.75rem 0 0;
  font-family: Georgia, 'Songti SC', 'STSong', serif;
  font-size: 2.75rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.15;
}

.lottery-header p {
  max-width: 55rem;
  margin: 0.65rem 0 0;
  color: var(--app-muted);
  font-size: 0.95rem;
  line-height: 1.7;
}

.lottery-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid var(--app-line);
  border-radius: 0.5rem;
  background: var(--app-surface);
}

.summary-item {
  display: grid;
  min-width: 0;
  align-content: start;
  gap: 0.45rem;
  padding: 1.4rem 1.55rem;
}

.summary-item + .summary-item {
  border-left: 1px solid var(--app-line);
}

.summary-item > span,
.section-heading span {
  color: var(--app-soft);
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.summary-item > strong {
  overflow-wrap: anywhere;
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.55;
}

.summary-count > div {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.summary-count strong {
  color: var(--app-accent-strong);
  font-family: var(--app-font-mono);
  font-size: 1.85rem;
  line-height: 1;
}

.summary-count small,
.summary-count p {
  color: var(--app-muted);
  font-size: 0.75rem;
}

.summary-count p {
  margin: 0;
}

.lottery-state {
  display: flex;
  min-height: 13rem;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  margin-top: 2rem;
  color: var(--app-muted);
  font-size: 0.9rem;
}

.lottery-error {
  flex-direction: column;
  color: #dc2626;
}

.lottery-spinner {
  width: 1.1rem;
  height: 1.1rem;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: lottery-spin 700ms linear infinite;
}

.lottery-retry {
  border: 1px solid var(--app-line);
  border-radius: 0.5rem;
  padding: 0.55rem 0.85rem;
  background: var(--app-surface-muted);
  color: var(--app-ink);
  cursor: pointer;
}

.lottery-panel {
  overflow: hidden;
  border: 1px solid var(--app-line);
  border-radius: 0.5rem;
  background: var(--app-surface);
}

.lottery-pending,
.lottery-history {
  margin-top: 2rem;
  padding: 1.65rem 1.9rem;
}

.section-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 1.1rem;
  border-bottom: 1px solid var(--app-line);
}

.section-heading h2 {
  margin: 0.25rem 0 0;
  font-family: Georgia, 'Songti SC', 'STSong', serif;
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: 0;
}

.section-heading > strong {
  color: var(--app-muted);
  font-size: 0.78rem;
}

.blind-box-row {
  display: grid;
  grid-template-columns: 5rem minmax(0, 1fr) auto;
  min-height: 8.25rem;
  align-items: center;
  gap: 1.25rem;
  padding: 1rem 0;
  border-bottom: 1px solid var(--app-line);
}

.blind-box-visual {
  display: grid;
  min-width: 0;
  justify-items: center;
  gap: 0.45rem;
}

.blind-box-icon {
  display: grid;
  width: 4rem;
  height: 4rem;
  place-items: center;
  border: 1px solid var(--app-ink);
  border-radius: 0.5rem;
  background: var(--app-ink);
  color: var(--app-surface);
  transition: transform 160ms ease;
}

.blind-box-row:hover .blind-box-icon {
  transform: translateY(-2px);
}

.blind-box-visual small {
  color: var(--app-muted);
  font-family: var(--app-font-mono);
  font-size: 0.68rem;
  font-weight: 700;
}

.blind-box-details {
  display: grid;
  min-width: 0;
  gap: 0.75rem;
}

.blind-box-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.blind-box-title > strong {
  min-width: 0;
  font-size: 0.95rem;
}

.rarity-common {
  background: #f3f4f6;
  color: #4b5563;
}

.rarity-rare {
  background: #e8f5ef;
  color: #08734f;
}

.rarity-epic {
  background: #f1eafa;
  color: #6d31a5;
}

.rarity-epic-plus {
  background: #fff3d6;
  color: #8a5900;
}

.rarity-legendary {
  background: #fde8e7;
  color: #b42318;
}

.blind-box-details dl {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem 1.5rem;
  margin: 0;
}

.blind-box-details dl > div {
  display: flex;
  min-width: 8rem;
  align-items: baseline;
  gap: 0.45rem;
}

.blind-box-details dt {
  color: var(--app-muted);
  font-size: 0.72rem;
}

.blind-box-details dd {
  margin: 0;
  overflow-wrap: anywhere;
  font-size: 0.8rem;
  font-weight: 750;
}

.lottery-open-button {
  min-width: 8.25rem;
  border: 1px solid var(--app-ink);
  border-radius: 0.5rem;
  padding: 0.6rem 0.9rem;
  background: var(--app-ink);
  color: var(--app-surface);
  cursor: pointer;
}

.lottery-open-button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.lottery-open-button:disabled {
  cursor: wait;
  opacity: 0.65;
}

.lottery-empty {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 1rem;
  min-height: 10rem;
  padding: 1.25rem 0;
}

.lottery-empty > span {
  display: grid;
  width: 3.5rem;
  height: 3.5rem;
  place-items: center;
  border-radius: 0.5rem;
  background: var(--app-surface-muted);
  color: var(--app-muted);
}

.lottery-empty h3,
.lottery-empty p {
  margin: 0;
}

.lottery-empty h3 {
  font-size: 0.95rem;
}

.lottery-empty p {
  margin-top: 0.3rem;
  color: var(--app-muted);
  font-size: 0.8rem;
}

.lottery-empty-action {
  min-width: 7rem;
}

.history-row {
  display: grid;
  grid-template-columns: 5rem minmax(0, 1fr) minmax(10rem, auto);
  align-items: center;
  gap: 1.25rem;
  min-height: 5.75rem;
  padding: 0.9rem 0.35rem;
  border-bottom: 1px solid var(--app-line);
}

.history-rarity {
  justify-self: start;
  border-radius: 999px;
  padding: 0.3rem 0.65rem;
  font-size: 0.7rem;
  font-weight: 800;
}

.history-reward,
.history-context {
  display: grid;
  min-width: 0;
  gap: 0.2rem;
}

.history-reward strong,
.history-reward small,
.history-context span,
.history-context small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-reward strong {
  color: var(--app-success);
  font-family: var(--app-font-mono);
  font-size: 1.05rem;
}

.history-reward small,
.history-context small {
  color: var(--app-muted);
  font-size: 0.72rem;
}

.history-context {
  justify-items: end;
  text-align: right;
}

.history-context span {
  font-size: 0.8rem;
  font-weight: 700;
}

.history-empty {
  margin: 0;
  padding: 2.5rem 0 1rem;
  color: var(--app-muted);
  font-size: 0.82rem;
}

:global(.dark) .rarity-common {
  background: #30312f;
  color: #d2d3cf;
}

:global(.dark) .rarity-rare {
  background: #123d31;
  color: #77d7b5;
}

:global(.dark) .rarity-epic {
  background: #342149;
  color: #d7b4f3;
}

:global(.dark) .rarity-epic-plus {
  background: #4a3512;
  color: #f3c86c;
}

:global(.dark) .rarity-legendary {
  background: #4c201f;
  color: #f2a6a1;
}

@media (max-width: 800px) {
  .lottery-summary {
    grid-template-columns: 1fr;
  }

  .summary-item + .summary-item {
    border-top: 1px solid var(--app-line);
    border-left: 0;
  }

  .blind-box-row {
    grid-template-columns: 4.5rem minmax(0, 1fr);
  }

  .lottery-open-button {
    grid-column: 2;
    justify-self: start;
  }
}

@media (max-width: 560px) {
  .lottery-page {
    padding-top: 0;
  }

  .lottery-header {
    padding-bottom: 2rem;
  }

  .lottery-header-bar {
    align-items: flex-start;
    flex-direction: column;
  }

  .lottery-header-actions {
    width: 100%;
    justify-content: space-between;
  }

  .lottery-header h1 {
    margin-top: 1.5rem;
    font-size: 2.15rem;
  }

  .summary-item {
    padding: 1.2rem;
  }

  .lottery-pending,
  .lottery-history {
    margin-top: 1.25rem;
    padding: 1.25rem 1rem;
  }

  .section-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.55rem;
  }

  .blind-box-row {
    grid-template-columns: 4rem minmax(0, 1fr);
    align-items: start;
    gap: 0.85rem;
  }

  .blind-box-icon {
    width: 3.5rem;
    height: 3.5rem;
  }

  .blind-box-title {
    align-items: flex-start;
  }

  .lottery-open-button {
    grid-column: 1 / -1;
    width: 100%;
  }

  .history-row {
    grid-template-columns: auto minmax(0, 1fr);
    gap: 0.75rem 1rem;
    padding: 1rem 0;
  }

  .history-context {
    grid-column: 2;
    justify-items: start;
    text-align: left;
  }

  .lottery-empty {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .lottery-empty-action {
    grid-column: 1 / -1;
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .lottery-recharge-link,
  .lottery-rules-link,
  .lottery-open-button,
  .lottery-retry,
  .blind-box-icon {
    transition: none;
  }
}

@keyframes lottery-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
