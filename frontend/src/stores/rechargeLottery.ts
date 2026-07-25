import { defineStore } from 'pinia'
import { ref } from 'vue'
import { paymentAPI } from '@/api/payment'
import { useAuthStore } from '@/stores/auth'
import type { RechargeLotteryOpportunity, RechargeLotteryOverview } from '@/types/payment'

export const useRechargeLotteryStore = defineStore('rechargeLottery', () => {
  const authStore = useAuthStore()
  const overview = ref<RechargeLotteryOverview | null>(null)
  const loading = ref(false)
  const error = ref<unknown>(null)
  const claimingOrderId = ref(0)
  const dialogOrderId = ref(0)
  const dismissedOrderIds = ref<number[]>([])
  let sessionRevision = 0
  let fetchRevision = 0

  // 从服务端刷新当前用户的待领取机会和最近领取结果。
  async function fetchOverview(): Promise<void> {
    const requestRevision = ++fetchRevision
    const requestSessionRevision = sessionRevision
    const requestUserId = authStore.user?.id ?? null
    loading.value = true
    error.value = null
    try {
      const response = await paymentAPI.getRechargeLottery()
      if (
        requestRevision === fetchRevision
        && requestSessionRevision === sessionRevision
        && requestUserId === (authStore.user?.id ?? null)
      ) {
        overview.value = response.data
      }
    } catch (cause: unknown) {
      if (
        requestRevision === fetchRevision
        && requestSessionRevision === sessionRevision
        && requestUserId === (authStore.user?.id ?? null)
      ) {
        error.value = cause
      }
      throw cause
    } finally {
      if (
        requestRevision === fetchRevision
        && requestSessionRevision === sessionRevision
        && requestUserId === (authStore.user?.id ?? null)
      ) {
        loading.value = false
      }
    }
  }

  // 领取指定充值订单的盲盒奖励，并用服务端结果更新共享状态。
  async function claim(orderId: number): Promise<RechargeLotteryOpportunity> {
    const currentOverview = overview.value
    const opportunity = currentOverview?.opportunities.find(item => item.order_id === orderId)
    if (!currentOverview || !opportunity || opportunity.claimed) {
      throw new Error(`Recharge lottery opportunity ${orderId} is not available`)
    }
    if (claimingOrderId.value !== 0) {
      throw new Error(`Recharge lottery order ${claimingOrderId.value} is already being claimed`)
    }

    const claimSessionRevision = sessionRevision
    const claimUserId = authStore.user?.id ?? null
    claimingOrderId.value = orderId
    error.value = null
    try {
      const response = await paymentAPI.drawRechargeLottery(orderId)
      const result = response.data
      if (
        claimSessionRevision !== sessionRevision
        || claimUserId !== (authStore.user?.id ?? null)
      ) {
        return result
      }

      // 领取结果已经落库，所有更早发起的概览请求都不能再覆盖它。
      fetchRevision += 1
      loading.value = false
      const latestOverview = overview.value
      if (latestOverview) {
        const latestOpportunity = latestOverview.opportunities.find(item => item.order_id === orderId)
        overview.value = {
          pending_count: latestOverview.pending_count - (latestOpportunity && !latestOpportunity.claimed ? 1 : 0),
          opportunities: latestOpportunity
            ? latestOverview.opportunities.map(item => item.order_id === orderId ? result : item)
            : [result, ...latestOverview.opportunities],
        }
      }

      // 奖励已经由领取接口落账；用户资料刷新失败不能把成功领取误报为失败。
      try {
        await authStore.refreshUser()
      } catch (cause: unknown) {
        console.error('[recharge-lottery] Failed to refresh user balance after claim:', cause)
      }
      return result
    } catch (cause: unknown) {
      if (
        claimSessionRevision === sessionRevision
        && claimUserId === (authStore.user?.id ?? null)
        && claimingOrderId.value === orderId
      ) {
        error.value = cause
      }
      throw cause
    } finally {
      if (
        claimSessionRevision === sessionRevision
        && claimUserId === (authStore.user?.id ?? null)
        && claimingOrderId.value === orderId
      ) {
        claimingOrderId.value = 0
      }
    }
  }

  // 打开指定订单的全局领取弹窗；活动页可重新打开本会话中关闭过的盲盒。
  function openDialog(orderId: number): void {
    dismissedOrderIds.value = dismissedOrderIds.value.filter(id => id !== orderId)
    dialogOrderId.value = orderId
  }

  // 关闭当前弹窗但不调用领取接口，资格仍保留在服务端。
  function dismissDialog(orderId: number): void {
    if (!dismissedOrderIds.value.includes(orderId)) {
      dismissedOrderIds.value = [...dismissedOrderIds.value, orderId]
    }
    if (dialogOrderId.value === orderId) {
      dialogOrderId.value = 0
    }
  }

  // 用户退出登录时清空所有盲盒及弹窗状态。
  function reset(): void {
    sessionRevision += 1
    fetchRevision += 1
    overview.value = null
    loading.value = false
    error.value = null
    claimingOrderId.value = 0
    dialogOrderId.value = 0
    dismissedOrderIds.value = []
  }

  return {
    overview,
    loading,
    error,
    claimingOrderId,
    dialogOrderId,
    dismissedOrderIds,
    fetchOverview,
    claim,
    openDialog,
    dismissDialog,
    reset,
  }
})
