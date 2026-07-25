<template>
  <Teleport to="body">
    <Transition name="reward-dialog">
      <div
        v-if="visible"
        class="reward-dialog-overlay"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="claimedResult ? 'recharge-lottery-result-title' : 'recharge-lottery-ready-title'"
        aria-describedby="recharge-lottery-dialog-description"
        data-testid="recharge-lottery-dialog"
        @click.self="closeDialog"
      >
        <section ref="dialogRef" class="reward-dialog-panel">
          <button
            type="button"
            class="reward-dialog-close"
            :disabled="claiming"
            :aria-label="t('common.close')"
            @click="closeDialog"
          >
            <Icon name="x" size="md" />
          </button>

          <div v-if="claimedResult" class="claimed-content">
            <span :class="rarityClass(claimedResult.rarity)" class="result-rarity">
              <Icon name="sparkles" size="sm" />
              {{ t(`payment.lottery.rarity.${claimedResult.rarity}`) }}
            </span>
            <p class="dialog-eyebrow">{{ t('payment.lottery.claimedEyebrow') }}</p>
            <h2 id="recharge-lottery-result-title">{{ t('payment.lottery.claimedTitle') }}</h2>
            <strong class="reward-amount" data-testid="reward-amount">
              +${{ claimedResult.reward_amount.toFixed(2) }}
            </strong>

            <dl class="amount-breakdown">
              <div>
                <dt>{{ t('payment.lottery.rechargeCredited') }}</dt>
                <dd data-testid="recharge-amount">${{ claimedResult.recharge_amount.toFixed(2) }}</dd>
              </div>
              <div>
                <dt>{{ t('payment.lottery.blindBoxReward') }}</dt>
                <dd class="reward-value">+${{ claimedResult.reward_amount.toFixed(2) }}</dd>
              </div>
            </dl>

            <p id="recharge-lottery-dialog-description" class="dialog-description">
              {{ t('payment.lottery.claimedDescription') }}
            </p>
            <button type="button" class="dialog-primary-button" @click="closeDialog">
              <Icon name="check" size="sm" />
              {{ t('payment.lottery.done') }}
            </button>
          </div>

          <div v-else-if="currentOpportunity" class="ready-content">
            <p class="dialog-eyebrow">{{ t('payment.lottery.readyEyebrow') }}</p>
            <h2 id="recharge-lottery-ready-title">
              {{ t('payment.lottery.rechargeCreditedAmount', { amount: currentOpportunity.recharge_amount.toFixed(2) }) }}
            </h2>
            <p id="recharge-lottery-dialog-description" class="dialog-description">
              {{ t('payment.lottery.claimDescription') }}
            </p>

            <button
              type="button"
              class="gift-claim-button"
              :disabled="claiming"
              :aria-label="t('payment.lottery.claimNow')"
              data-testid="gift-claim-button"
              @click="claimReward"
            >
              <span class="gift-stage" aria-hidden="true">
                <span class="gift-visual">
                  <Icon name="gift" size="xl" />
                </span>
              </span>
              <span class="gift-action-copy">
                <strong>{{ claiming ? t('payment.lottery.claiming') : t('payment.lottery.tapGift') }}</strong>
                <span :class="rarityClass(currentOpportunity.max_rarity)" class="rarity-cap">
                  {{ t('payment.lottery.rarityCap', { rarity: t(`payment.lottery.rarity.${currentOpportunity.max_rarity}`) }) }}
                </span>
              </span>
              <Icon name="arrowRight" size="sm" class="gift-action-arrow" aria-hidden="true" />
            </button>

            <p v-if="claimError" class="claim-error" role="alert">{{ claimError }}</p>

            <div class="dialog-footer-actions">
              <router-link :to="{ path: '/docs', query: { cat: 'activities', page: 'recharge-lottery' } }">
                <Icon name="book" size="sm" />
                {{ t('payment.lottery.fullRules') }}
              </router-link>
              <button type="button" :disabled="claiming" @click="closeDialog">
                {{ t('payment.lottery.claimLater') }}
              </button>
            </div>
          </div>
        </section>

        <div v-if="claimedResult" class="confetti" aria-hidden="true" data-testid="confetti">
          <span
            v-for="piece in confettiPieces"
            :key="piece.id"
            :style="piece.style"
          ></span>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore } from '@/stores/auth'
import { useRechargeLotteryStore } from '@/stores/rechargeLottery'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { RechargeLotteryOpportunity, RechargeLotteryRarity } from '@/types/payment'

const { t } = useI18n()
const route = useRoute()
const authStore = useAuthStore()
const lotteryStore = useRechargeLotteryStore()
const dialogRef = ref<HTMLElement | null>(null)
const claimedResult = ref<RechargeLotteryOpportunity | null>(null)
const claimError = ref('')
let previousActiveElement: HTMLElement | null = null

const confettiColors = ['#e6a51f', '#2f9d75', '#3b82c4', '#d4534d', '#8b5bb5']
const confettiPieces = Array.from({ length: 72 }, (_, index) => {
  const startX = index % 2 === 0 ? 47 : 53
  const targetX = 3 + ((index * 47) % 95)
  const targetY = 4 + ((index * 31) % 52)
  const finalX = targetX + ((index * 17) % 15) - 7
  const rotation = 180 + ((index * 73) % 540)

  return {
    id: index,
    style: {
      '--confetti-start-x': `${startX}vw`,
      '--confetti-burst-x': `${targetX - startX}vw`,
      '--confetti-burst-y': `${targetY - 58}vh`,
      '--confetti-fall-x': `${finalX - startX}vw`,
      '--confetti-fall-y': `${54 + (index % 7)}vh`,
      '--confetti-rotation': `${index % 2 === 0 ? rotation : -rotation}deg`,
      '--confetti-finish-rotation': `${index % 2 === 0 ? rotation * 2 : rotation * -2}deg`,
      '--confetti-width': `${3 + (index % 3)}px`,
      '--confetti-height': `${5 + ((index * 5) % 4)}px`,
      '--confetti-radius': index % 5 === 0 ? '50%' : '1px',
      '--confetti-delay': `${(index % 12) * 22}ms`,
      '--confetti-duration': `${1900 + (index % 8) * 85}ms`,
      backgroundColor: confettiColors[index % confettiColors.length],
    },
  }
})

const currentOpportunity = computed(() => {
  return lotteryStore.overview?.opportunities.find(item => item.order_id === lotteryStore.dialogOrderId) || null
})

const visible = computed(() => {
  return authStore.isAuthenticated && (currentOpportunity.value !== null || claimedResult.value !== null)
})

const claiming = computed(() => {
  return currentOpportunity.value !== null && lotteryStore.claimingOrderId === currentOpportunity.value.order_id
})

// 登录、路由切换或页面重新可见时刷新机会，确保全局弹窗能发现新充值。
async function refreshOverview(): Promise<void> {
  try {
    await lotteryStore.fetchOverview()
  } catch (cause: unknown) {
    console.error('[recharge-lottery] Failed to refresh opportunities:', cause)
  }
}

// 点击礼盒后才调用领取接口，成功结果由服务端金额驱动。
async function claimReward(): Promise<void> {
  if (!currentOpportunity.value || claiming.value) return

  const orderId = currentOpportunity.value.order_id
  const userId = authStore.user?.id ?? null
  claimError.value = ''
  try {
    const result = await lotteryStore.claim(orderId)
    if (lotteryStore.dialogOrderId === orderId && userId === (authStore.user?.id ?? null)) {
      claimedResult.value = result
    }
  } catch (cause: unknown) {
    if (lotteryStore.dialogOrderId === orderId && userId === (authStore.user?.id ?? null)) {
      claimError.value = extractApiErrorMessage(cause, t('payment.lottery.drawFailed'))
    }
  }
}

// 关闭只隐藏本次机会，不触发领取；活动页仍可重新打开。
function closeDialog(): void {
  if (claiming.value) return

  const orderId = claimedResult.value?.order_id || currentOpportunity.value?.order_id
  if (!orderId) return

  lotteryStore.dismissDialog(orderId)
  claimedResult.value = null
  claimError.value = ''
}

// 按稀有度返回结果标识颜色。
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

// 页面重新回到前台时检查充值完成后新产生的机会。
function onVisibilityChange(): void {
  if (document.visibilityState === 'visible' && authStore.isAuthenticated) {
    void refreshOverview()
  }
}

// Escape 与遮罩关闭行为一致，领取请求进行中不可中断结果展示。
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && visible.value) {
    closeDialog()
  }
}

watch(
  [() => authStore.isAuthenticated, () => route.fullPath],
  ([isAuthenticated]) => {
    if (!isAuthenticated) {
      claimedResult.value = null
      lotteryStore.reset()
      return
    }
    void refreshOverview()
  },
  { immediate: true },
)

watch(
  [
    () => authStore.isAuthenticated,
    () => lotteryStore.overview,
    () => lotteryStore.dialogOrderId,
    () => lotteryStore.dismissedOrderIds,
  ],
  ([isAuthenticated]) => {
    if (!isAuthenticated || claimedResult.value || lotteryStore.dialogOrderId !== 0) return

    const nextOpportunity = lotteryStore.overview?.opportunities.find(item => (
      !item.claimed && !lotteryStore.dismissedOrderIds.includes(item.order_id)
    ))
    if (nextOpportunity) {
      lotteryStore.openDialog(nextOpportunity.order_id)
    }
  },
  { deep: true, immediate: true },
)

watch(visible, async (isVisible) => {
  if (isVisible) {
    previousActiveElement = document.activeElement as HTMLElement
    document.body.classList.add('modal-open')
    await nextTick()
    dialogRef.value?.querySelector<HTMLElement>('.gift-claim-button, .dialog-primary-button')?.focus()
    return
  }

  document.body.classList.remove('modal-open')
  previousActiveElement?.focus()
  previousActiveElement = null
})

onMounted(() => {
  document.addEventListener('visibilitychange', onVisibilityChange)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
  document.removeEventListener('keydown', onKeydown)
  document.body.classList.remove('modal-open')
})
</script>

<style scoped>
.reward-dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: 110;
  display: grid;
  place-items: center;
  overflow-y: auto;
  padding: 1.5rem;
  background: rgb(24 24 22 / 46%);
  backdrop-filter: blur(4px);
}

.reward-dialog-panel {
  --dialog-accent: #bd4d57;
  --dialog-accent-soft: #f6e5e6;
  --dialog-ink: #191a18;
  --dialog-muted: #6f706b;
  --dialog-line: #deded9;
  position: relative;
  width: min(100%, 35rem);
  min-height: 30rem;
  overflow: hidden;
  border: 1px solid #d4d4cf;
  border-radius: 0.5rem;
  background: #fbfbf9;
  box-shadow: 0 1.75rem 5rem rgb(0 0 0 / 24%);
  color: var(--dialog-ink);
  color-scheme: light;
}

.reward-dialog-close {
  position: absolute;
  top: 1.15rem;
  right: 1.15rem;
  z-index: 4;
  display: grid;
  width: 2.25rem;
  height: 2.25rem;
  place-items: center;
  border: 1px solid var(--dialog-line);
  border-radius: 50%;
  background: transparent;
  color: var(--dialog-muted);
  cursor: pointer;
  transition: border-color 160ms ease, color 160ms ease, transform 160ms ease;
}

.reward-dialog-close:hover:not(:disabled),
.reward-dialog-close:focus-visible {
  border-color: var(--dialog-accent);
  color: var(--dialog-accent);
  transform: rotate(4deg);
}

.reward-dialog-close:disabled {
  cursor: wait;
  opacity: 0.5;
}

.ready-content,
.claimed-content {
  position: relative;
  z-index: 1;
  display: flex;
  min-height: 30rem;
  flex-direction: column;
  padding: 3rem 2.5rem 1.75rem;
}

.ready-content {
  align-items: flex-start;
  text-align: left;
}

.dialog-eyebrow {
  margin: 0;
  color: var(--dialog-accent);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.ready-content h2,
.claimed-content h2 {
  margin: 0.65rem 0 0;
  font-family: Georgia, 'Songti SC', 'STSong', serif;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.22;
}

.dialog-description {
  max-width: 28rem;
  margin: 0.8rem 0 0;
  color: var(--dialog-muted);
  font-size: 0.86rem;
  line-height: 1.7;
}

.rarity-cap,
.result-rarity {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-radius: 999px;
  padding: 0.3rem 0.7rem;
  font-size: 0.7rem;
  font-weight: 800;
}

.rarity-cap {
  justify-self: start;
}

.gift-claim-button {
  display: grid;
  width: 100%;
  grid-template-columns: 7rem minmax(0, 1fr) auto;
  align-items: center;
  gap: 1.25rem;
  margin: 1.65rem 0 0;
  border: 0;
  border-top: 1px solid var(--dialog-line);
  border-bottom: 1px solid var(--dialog-line);
  padding: 1.15rem 0.25rem;
  background: transparent;
  color: var(--dialog-ink);
  text-align: left;
  cursor: pointer;
}

.gift-stage {
  position: relative;
  display: grid;
  width: 7rem;
  height: 7rem;
  place-items: center;
  animation: gift-float 2.8s ease-in-out infinite;
}

.gift-stage::before,
.gift-stage::after {
  position: absolute;
  width: 0.42rem;
  height: 0.42rem;
  border: 1px solid var(--dialog-accent);
  content: '';
  transform: rotate(45deg) scale(0.45);
  animation: gift-spark 2.2s ease-in-out infinite;
}

.gift-stage::before {
  top: 0.65rem;
  left: 0.25rem;
}

.gift-stage::after {
  top: 1.8rem;
  right: 0.2rem;
  animation-delay: -1.1s;
}

.gift-visual {
  position: relative;
  display: grid;
  width: 6rem;
  height: 6rem;
  place-items: center;
  overflow: hidden;
  border: 1px solid #dba8ad;
  border-radius: 0.5rem;
  background: var(--dialog-accent-soft);
  color: #9f3741;
  box-shadow: 0 0.85rem 1.9rem rgb(116 48 55 / 12%);
  transform-origin: 50% 88%;
  transition: transform 180ms ease, border-color 180ms ease, background-color 180ms ease, box-shadow 180ms ease;
}

.gift-visual::before {
  position: absolute;
  top: -25%;
  left: -35%;
  width: 0.85rem;
  height: 150%;
  background: rgb(255 255 255 / 42%);
  content: '';
  opacity: 0;
  transform: rotate(18deg);
  animation: gift-sheen 3.2s ease-in-out infinite;
}

.gift-visual :deep(svg) {
  position: relative;
  z-index: 1;
  width: 3.5rem;
  height: 3.5rem;
  stroke-width: 1.25;
}

.gift-claim-button:hover:not(:disabled) .gift-visual,
.gift-claim-button:focus-visible .gift-visual {
  border-color: var(--dialog-accent);
  background: #f8dfe1;
  box-shadow: 0 0.9rem 2rem rgb(116 48 55 / 16%);
  transform: translateY(-0.2rem);
}

.gift-claim-button:focus-visible {
  outline: none;
}

.gift-claim-button:focus-visible .gift-visual {
  box-shadow: 0 0 0 2px #fbfbf9, 0 0 0 4px var(--dialog-accent), 0 0.9rem 2rem rgb(116 48 55 / 16%);
}

.gift-action-copy {
  display: grid;
  min-width: 0;
  justify-items: start;
  gap: 0.55rem;
}

.gift-action-copy strong {
  overflow-wrap: anywhere;
  font-family: Georgia, 'Songti SC', 'STSong', serif;
  font-size: 1.08rem;
  font-weight: 700;
}

.gift-action-arrow {
  color: var(--dialog-accent);
  transition: transform 160ms ease;
}

.gift-claim-button:hover:not(:disabled) .gift-action-arrow {
  transform: translateX(0.2rem);
}

.gift-claim-button:disabled {
  cursor: wait;
  opacity: 0.72;
}

.gift-claim-button:disabled .gift-stage {
  animation-play-state: paused;
}

.gift-claim-button:disabled .gift-stage::before,
.gift-claim-button:disabled .gift-stage::after,
.gift-claim-button:disabled .gift-visual::before {
  animation-play-state: paused;
}

.gift-claim-button:disabled .gift-visual {
  animation: gift-opening 720ms ease-in-out both;
}

.claim-error {
  margin: 0.75rem 0 0;
  color: #b42318;
  font-size: 0.78rem;
}

.dialog-footer-actions {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: auto;
  padding-top: 1.25rem;
  border-top: 1px solid var(--dialog-line);
}

.dialog-footer-actions a,
.dialog-footer-actions button {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  border: 0;
  background: transparent;
  color: var(--dialog-muted);
  font-size: 0.76rem;
  text-decoration: none;
  cursor: pointer;
}

.dialog-footer-actions a:hover,
.dialog-footer-actions button:hover:not(:disabled) {
  color: var(--dialog-accent);
}

.claimed-content {
  align-items: center;
  justify-content: center;
  text-align: center;
}

.claimed-content .result-rarity,
.claimed-content .dialog-eyebrow,
.claimed-content h2,
.claimed-content .amount-breakdown,
.claimed-content .dialog-description,
.claimed-content .dialog-primary-button {
  animation: result-rise 420ms ease-out both;
}

.claimed-content .dialog-eyebrow {
  animation-delay: 70ms;
}

.claimed-content h2 {
  animation-delay: 120ms;
}

.result-rarity {
  position: relative;
  z-index: 2;
  margin-bottom: 0.8rem;
}

.reward-amount {
  position: relative;
  z-index: 2;
  margin-top: 0.8rem;
  color: var(--dialog-accent);
  font-family: var(--app-font-mono);
  font-size: 3.1rem;
  line-height: 1;
  animation: reward-pop 560ms cubic-bezier(0.2, 0.85, 0.25, 1.25) 180ms both;
}

.amount-breakdown {
  position: relative;
  z-index: 2;
  display: grid;
  width: 100%;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 1.6rem 0 0;
  border-top: 1px solid var(--dialog-line);
  border-bottom: 1px solid var(--dialog-line);
  animation-delay: 280ms;
}

.amount-breakdown > div {
  min-width: 0;
  padding: 1rem;
}

.amount-breakdown > div + div {
  border-left: 1px solid var(--dialog-line);
}

.amount-breakdown dt {
  color: var(--dialog-muted);
  font-size: 0.72rem;
}

.amount-breakdown dd {
  margin: 0.3rem 0 0;
  overflow-wrap: anywhere;
  font-family: var(--app-font-mono);
  font-size: 1rem;
  font-weight: 800;
}

.amount-breakdown .reward-value {
  color: var(--dialog-accent);
}

.claimed-content .dialog-description {
  animation-delay: 350ms;
}

.dialog-primary-button {
  position: relative;
  z-index: 2;
  display: inline-flex;
  min-width: 10rem;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  margin-top: 1.35rem;
  border: 1px solid var(--dialog-ink);
  border-radius: 0.5rem;
  background: var(--dialog-ink);
  color: #fbfbf9;
  font-size: 0.84rem;
  font-weight: 800;
  cursor: pointer;
  animation-delay: 410ms;
  transition: background-color 160ms ease, transform 160ms ease;
}

.dialog-primary-button:hover,
.dialog-primary-button:focus-visible {
  background: #343531;
  transform: translateY(-1px);
}

.dialog-primary-button:focus-visible {
  outline: 2px solid var(--dialog-accent);
  outline-offset: 3px;
}

.confetti {
  position: fixed;
  inset: 0;
  z-index: 3;
  overflow: hidden;
  background: transparent;
  pointer-events: none;
}

.confetti span {
  position: absolute;
  top: 58vh;
  left: var(--confetti-start-x);
  width: var(--confetti-width);
  height: var(--confetti-height);
  border-radius: var(--confetti-radius);
  box-shadow: 0 0 0 0.5px rgb(255 255 255 / 16%);
  opacity: 0;
  transform-origin: center;
  animation: confetti-burst var(--confetti-duration) cubic-bezier(0.18, 0.72, 0.26, 1) var(--confetti-delay) forwards;
}

.rarity-common {
  background: #f0f0ed;
  color: #555650;
}

.rarity-rare {
  background: #e7f4ee;
  color: #08734f;
}

.rarity-epic {
  background: #f1eafa;
  color: #6d31a5;
}

.rarity-epic-plus {
  background: #fff1d2;
  color: #835600;
}

.rarity-legendary {
  background: #f9e3e4;
  color: #a93a44;
}

.reward-dialog-enter-active,
.reward-dialog-leave-active {
  transition: opacity 180ms ease;
}

.reward-dialog-enter-active .reward-dialog-panel,
.reward-dialog-leave-active .reward-dialog-panel {
  transition: transform 320ms cubic-bezier(0.2, 0.8, 0.25, 1), opacity 220ms ease;
}

.reward-dialog-enter-from,
.reward-dialog-leave-to {
  opacity: 0;
}

.reward-dialog-enter-from .reward-dialog-panel,
.reward-dialog-leave-to .reward-dialog-panel {
  opacity: 0;
  transform: translateY(1.25rem) scale(0.96);
}

@keyframes gift-float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-0.25rem);
  }
}

@keyframes gift-spark {
  0%,
  100% {
    opacity: 0.25;
    transform: rotate(45deg) scale(0.45);
  }
  45% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }
}

@keyframes gift-sheen {
  0%,
  58% {
    left: -35%;
    opacity: 0;
  }
  66% {
    opacity: 0.75;
  }
  82%,
  100% {
    left: 115%;
    opacity: 0;
  }
}

@keyframes gift-opening {
  0%,
  100% {
    transform: rotate(0deg) scale(1);
  }
  20% {
    transform: rotate(-4deg) scale(1.03);
  }
  40% {
    transform: rotate(4deg) scale(1.03);
  }
  62% {
    transform: rotate(-2.5deg) scale(1.02);
  }
  82% {
    transform: rotate(1.5deg) scale(1.01);
  }
}

@keyframes result-rise {
  from {
    opacity: 0;
    transform: translateY(0.75rem);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes reward-pop {
  from {
    opacity: 0;
    transform: scale(0.72);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes confetti-burst {
  0% {
    opacity: 0;
    transform: translate3d(0, 0, 0) rotate(0deg) scale(0.7);
  }
  8% {
    opacity: 1;
    transform: translate3d(0, -1.5vh, 0) rotate(35deg) scale(1);
  }
  42% {
    opacity: 1;
    transform: translate3d(var(--confetti-burst-x), var(--confetti-burst-y), 0) rotate(var(--confetti-rotation)) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate3d(var(--confetti-fall-x), var(--confetti-fall-y), 0) rotate(var(--confetti-finish-rotation)) scale(0.92);
  }
}

@media (max-width: 480px) {
  .reward-dialog-overlay {
    align-items: end;
    padding: 0;
  }

  .reward-dialog-panel {
    width: 100%;
    min-height: min(31rem, 92dvh);
    border-right: 0;
    border-bottom: 0;
    border-left: 0;
    border-radius: 0.5rem 0.5rem 0 0;
  }

  .ready-content,
  .claimed-content {
    min-height: min(31rem, 92dvh);
    padding: 3.5rem 1.25rem 1.5rem;
  }

  .ready-content h2,
  .claimed-content h2 {
    font-size: 1.7rem;
  }

  .gift-claim-button {
    grid-template-columns: 5.75rem minmax(0, 1fr) auto;
    gap: 0.8rem;
    margin-top: 1.35rem;
    padding: 1rem 0;
  }

  .gift-stage {
    width: 5.75rem;
    height: 5.75rem;
  }

  .gift-visual {
    width: 5rem;
    height: 5rem;
  }

  .gift-visual :deep(svg) {
    width: 3rem;
    height: 3rem;
  }

  .gift-action-copy strong {
    font-size: 0.98rem;
  }

  .reward-amount {
    font-size: 2.55rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .reward-dialog-enter-active,
  .reward-dialog-leave-active,
  .reward-dialog-enter-active .reward-dialog-panel,
  .reward-dialog-leave-active .reward-dialog-panel,
  .reward-dialog-close,
  .gift-visual,
  .gift-action-arrow,
  .dialog-primary-button {
    transition: none;
  }

  .gift-stage,
  .gift-stage::before,
  .gift-stage::after,
  .gift-visual::before,
  .gift-claim-button:disabled .gift-visual,
  .claimed-content .result-rarity,
  .claimed-content .dialog-eyebrow,
  .claimed-content h2,
  .claimed-content .reward-amount,
  .claimed-content .amount-breakdown,
  .claimed-content .dialog-description,
  .claimed-content .dialog-primary-button {
    animation: none;
  }

  .confetti span {
    animation: none;
    display: none;
  }
}
</style>
