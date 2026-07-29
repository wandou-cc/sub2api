<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 pb-4 dark:border-dark-700">
        <div class="inline-flex rounded-md border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-800">
          <button v-for="tab in tabs" :key="tab.value" class="rounded px-4 py-2 text-sm font-medium transition-colors"
            :class="activeTab === tab.value ? 'bg-gray-950 text-white dark:bg-white dark:text-gray-950' : 'text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
            @click="switchTab(tab.value)">
            {{ tab.label }}
          </button>
        </div>
        <div class="flex items-center gap-2">
          <button v-if="activeTab === 'plans'" class="btn btn-primary px-3 py-2 text-sm" @click="openCreatePlanDialog">
            <Icon name="plus" size="sm" />
            {{ t('payment.admin.carpool.createPlan') }}
          </button>
          <button class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="loadActiveTab">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <div v-if="loading" class="flex justify-center py-20">
        <span class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></span>
      </div>
      <template v-else-if="activeTab === 'plans'">
        <div v-if="plans.length === 0" class="border-y border-gray-200 py-16 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
          {{ t('payment.admin.carpool.noPlans') }}
        </div>
        <div v-else class="overflow-x-auto rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
          <table class="min-w-full divide-y divide-gray-200 text-left text-sm dark:divide-dark-700">
            <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-900/40 dark:text-gray-400">
              <tr>
                <th class="px-5 py-3 font-medium">{{ t('payment.admin.carpool.groupSize') }}</th>
                <th class="px-5 py-3 font-medium">{{ t('payment.admin.carpool.totalAmount') }}</th>
                <th class="px-5 py-3 font-medium">{{ t('payment.admin.carpool.perMemberAmount') }}</th>
                <th class="min-w-80 px-5 py-3 font-medium">{{ t('payment.admin.carpool.planNote') }}</th>
                <th class="px-5 py-3 font-medium">{{ t('payment.admin.carpool.updatedAt') }}</th>
                <th class="px-5 py-3 text-right font-medium">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="plan in plans" :key="plan.id" :data-test="`carpool-admin-plan-${plan.id}`">
                <td class="whitespace-nowrap px-5 py-4 font-medium text-gray-900 dark:text-white">{{ t('payment.carpool.planName', { count: plan.target_members }) }}</td>
                <td class="whitespace-nowrap px-5 py-4 tabular-nums text-gray-700 dark:text-gray-300">¥{{ plan.total_amount.toFixed(2) }}</td>
                <td class="whitespace-nowrap px-5 py-4 tabular-nums text-gray-700 dark:text-gray-300">¥{{ plan.price_per_member.toFixed(2) }}</td>
                <td class="whitespace-pre-line px-5 py-4 leading-6 text-gray-600 dark:text-gray-300">{{ plan.note }}</td>
                <td class="whitespace-nowrap px-5 py-4 text-xs text-gray-500 dark:text-gray-400">{{ formatDateTime(plan.updated_at) }}</td>
                <td class="whitespace-nowrap px-5 py-4 text-right">
                  <button class="rounded p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white" :title="t('common.edit')" @click="openEditPlanDialog(plan)">
                    <Icon name="edit" size="sm" />
                  </button>
                  <button class="ml-1 rounded p-2 text-red-500 hover:bg-red-50 hover:text-red-700 dark:text-red-400 dark:hover:bg-red-950" :title="t('common.delete')" @click="deletePlanTarget = plan">
                    <Icon name="trash" size="sm" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
      <div v-else-if="groups.length === 0" class="border-y border-gray-200 py-16 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
        {{ t('payment.admin.carpool.empty') }}
      </div>
      <template v-else>
        <section v-for="group in groups" :key="group.id" class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
          <header class="grid gap-4 border-b border-gray-200 px-5 py-4 dark:border-dark-700 lg:grid-cols-[1fr_auto] lg:items-center">
            <div class="flex flex-wrap items-center gap-x-5 gap-y-2">
              <div>
                <p class="text-base font-semibold text-gray-950 dark:text-white">{{ t('payment.carpool.planName', { count: group.target_members }) }} #{{ group.id }}</p>
                <p class="mt-1 text-xs text-gray-400">{{ formatDateTime(group.created_at) }}</p>
              </div>
              <span class="rounded px-2 py-1 text-xs font-medium" :class="statusClass(group.status)">{{ statusText(group.status) }}</span>
              <span class="text-sm tabular-nums text-gray-600 dark:text-gray-300">{{ group.current_members }}/{{ group.target_members }} {{ t('payment.admin.carpool.members') }}</span>
              <span class="text-sm text-gray-500 dark:text-gray-400">¥{{ group.price_per_member.toFixed(2) }} / {{ t('payment.carpool.person') }}</span>
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.carpool.totalAmount') }} ¥{{ group.total_amount.toFixed(2) }}</span>
            </div>
            <div class="flex flex-wrap gap-2 lg:justify-end">
              <button v-if="group.status === 'purchasing'" class="btn btn-primary px-3 py-2 text-sm" @click="openDeliveryDialog(group)">
                <Icon name="check" size="sm" />
                {{ t('payment.admin.carpool.markOpened') }}
              </button>
              <button v-if="group.status === 'waiting' || group.status === 'purchasing' || group.status === 'active'" class="btn btn-secondary px-3 py-2 text-sm text-red-600 dark:text-red-400" @click="openRefundPendingDialog(group)">
                <Icon name="dollar" size="sm" />
                {{ t('payment.admin.carpool.markRefundPending') }}
              </button>
            </div>
          </header>

          <div class="grid gap-x-8 gap-y-2 bg-gray-50 px-5 py-3 text-xs text-gray-500 dark:bg-dark-900/40 dark:text-gray-400 sm:grid-cols-2 lg:grid-cols-4">
            <p>{{ t('payment.admin.carpool.deadline') }}: {{ group.deadline_at ? formatDateTime(group.deadline_at) : '-' }}</p>
            <p>{{ t('payment.admin.carpool.formedAt') }}: {{ group.formed_at ? formatDateTime(group.formed_at) : '-' }}</p>
            <p>{{ t('payment.admin.carpool.openedAt') }}: {{ group.opened_at ? formatDateTime(group.opened_at) : '-' }}</p>
            <p>{{ t('payment.admin.carpool.serviceExpiresAt') }}: {{ group.expires_at ? formatDateTime(group.expires_at) : '-' }}</p>
            <p v-if="group.status_reason" class="sm:col-span-2 lg:col-span-4">{{ t('payment.admin.carpool.statusReason') }}: {{ statusReasonText(group.status_reason) }}</p>
            <p class="whitespace-pre-line sm:col-span-2 lg:col-span-4">{{ t('payment.admin.carpool.planNote') }}: {{ group.plan_note }}</p>
          </div>

          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 text-left text-sm dark:divide-dark-700">
              <thead class="text-xs text-gray-500 dark:text-gray-400">
                <tr>
                  <th class="px-5 py-3 font-medium">{{ t('payment.orders.orderId') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('payment.admin.colUser') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('payment.orders.payAmount') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('payment.orders.status') }}</th>
                  <th class="px-5 py-3 text-right font-medium">{{ t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="member in group.members" :key="member.order_id">
                  <td class="whitespace-nowrap px-5 py-3 font-mono text-gray-700 dark:text-gray-300">#{{ member.order_id }}</td>
                  <td class="px-5 py-3">
                    <p class="font-medium text-gray-900 dark:text-white">{{ member.user_email }}</p>
                    <p class="mt-0.5 text-xs text-gray-400">#{{ member.user_id }} · {{ member.user_name }}</p>
                  </td>
                  <td class="whitespace-nowrap px-5 py-3 font-medium text-gray-900 dark:text-white">{{ currencySymbol(member.currency) }}{{ member.pay_amount.toFixed(2) }}</td>
                  <td class="whitespace-nowrap px-5 py-3"><OrderStatusBadge :status="member.status" /></td>
                  <td class="whitespace-nowrap px-5 py-3 text-right">
                    <button v-if="group.status === 'refund_pending' && canRefund(member.status)" class="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950" @click="openMemberRefundDialog(group, member)">
                      <Icon name="dollar" size="sm" />
                      {{ t('payment.admin.refund') }}
                    </button>
                    <button v-else-if="member.status === 'REFUND_PENDING'" class="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium text-amber-600 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-950" :disabled="actionLoading" @click="queryRefund(member.order_id)">
                      <Icon name="refresh" size="sm" />
                      {{ t('payment.admin.queryRefundStatus') }}
                    </button>
                    <span v-else class="text-xs text-gray-400">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>
    </div>

    <BaseDialog :show="planDialogOpen" :title="editingPlan ? t('payment.admin.carpool.editPlan') : t('payment.admin.carpool.createPlan')" @close="closePlanDialog">
      <div class="space-y-4">
        <div>
          <label class="input-label">{{ t('payment.admin.carpool.totalAmount') }}</label>
          <input v-model.number="planForm.total_amount" type="number" min="0.01" step="0.01" class="input mt-1 w-full" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.carpool.groupSize') }}</label>
          <input v-model.number="planForm.target_members" type="number" min="1" step="1" class="input mt-1 w-full" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.carpool.planNote') }}</label>
          <textarea v-model="planForm.note" rows="8" class="input mt-1 w-full" :placeholder="t('payment.admin.carpool.planNotePlaceholder')"></textarea>
        </div>
        <p v-if="planPerMemberAmount !== null" class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('payment.admin.carpool.perMemberPreview', { amount: `¥${planPerMemberAmount.toFixed(2)}` }) }}
        </p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="closePlanDialog">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading || !planFormValid" @click="savePlan">{{ t('common.confirm') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!deletePlanTarget" :title="t('payment.admin.carpool.deletePlan')" width="narrow" @close="deletePlanTarget = null">
      <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">{{ t('payment.admin.carpool.deletePlanConfirm') }}</p>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="deletePlanTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmDeletePlan">{{ t('common.confirm') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!deliveryTarget" :title="t('payment.admin.carpool.openDialogTitle')" @close="deliveryTarget = null">
      <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">{{ t('payment.admin.carpool.openDialogMessage') }}</p>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="deliveryTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading" @click="confirmDelivery">{{ t('common.confirm') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!refundPendingTarget" :title="t('payment.admin.carpool.refundPendingDialogTitle')" @close="refundPendingTarget = null">
      <div>
        <label class="input-label">{{ t('payment.admin.refundReason') }}</label>
        <textarea v-model="refundPendingReason" rows="3" class="input mt-1 w-full" :placeholder="t('payment.admin.refundReasonPlaceholder')"></textarea>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="refundPendingTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading || !refundPendingReason.trim()" @click="confirmRefundPending">{{ t('common.confirm') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!refundMember" :title="t('payment.admin.refundOrder')" @close="refundMember = null">
      <div v-if="refundMember" class="space-y-4">
        <div class="grid grid-cols-2 gap-4 border-y border-gray-200 py-4 text-sm dark:border-dark-700">
          <div><p class="text-xs text-gray-500">{{ t('payment.orders.orderId') }}</p><p class="mt-1 font-mono">#{{ refundMember.order_id }}</p></div>
          <div><p class="text-xs text-gray-500">{{ t('payment.orders.payAmount') }}</p><p class="mt-1 font-medium">{{ currencySymbol(refundMember.currency) }}{{ refundMember.pay_amount.toFixed(2) }}</p></div>
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.refundAmount') }}</label>
          <input v-model.number="memberRefundForm.amount" type="number" min="0.01" :max="refundMember.amount" step="0.01" class="input mt-1 w-full" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.carpool.refundAmountHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.refundReason') }}</label>
          <textarea v-model="memberRefundForm.reason" rows="3" class="input mt-1 w-full"></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="refundMember = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading || !memberRefundForm.reason.trim() || memberRefundForm.amount <= 0" @click="confirmMemberRefund">{{ t('payment.admin.confirmRefund') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  adminPaymentAPI,
  type AdminCarpoolGroup,
  type AdminCarpoolMember,
  type AdminCarpoolPlan,
  type CarpoolPlanInput,
} from '@/api/admin/payment'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { CarpoolStatus } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import { currencySymbol } from '@/components/payment/currency'

const { t } = useI18n()
const appStore = useAppStore()
const activeTab = ref<'current' | 'history' | 'plans'>('current')
const loading = ref(false)
const actionLoading = ref(false)
const groups = ref<AdminCarpoolGroup[]>([])
const plans = ref<AdminCarpoolPlan[]>([])
const planDialogOpen = ref(false)
const editingPlan = ref<AdminCarpoolPlan | null>(null)
const deletePlanTarget = ref<AdminCarpoolPlan | null>(null)
const planForm = reactive<CarpoolPlanInput>({ total_amount: 0, target_members: 1, note: '' })
const deliveryTarget = ref<AdminCarpoolGroup | null>(null)
const refundPendingTarget = ref<AdminCarpoolGroup | null>(null)
const refundPendingReason = ref('')
const refundMember = ref<AdminCarpoolMember | null>(null)
const memberRefundForm = reactive({ amount: 0, reason: '' })

// Keeps the three administrative datasets in one compact segmented control.
const tabs = computed(() => [
  { value: 'current' as const, label: t('payment.admin.carpool.current') },
  { value: 'history' as const, label: t('payment.admin.carpool.history') },
  { value: 'plans' as const, label: t('payment.admin.carpool.planConfig') },
])

// Mirrors the backend's per-member rounding rule for immediate form feedback.
const planPerMemberAmount = computed(() => {
  if (!Number.isFinite(planForm.total_amount) || planForm.total_amount <= 0 || !Number.isInteger(planForm.target_members) || planForm.target_members <= 0) return null
  const rawCents = planForm.total_amount * 100
  const totalCents = Math.round(rawCents)
  if (Math.abs(rawCents - totalCents) > 0.000001) return null
  return Math.round(totalCents / planForm.target_members) / 100
})

// The form is submittable only when all three business fields satisfy their contract.
const planFormValid = computed(() => planPerMemberAmount.value !== null && planForm.note.trim() !== '')

// Loads either the actionable queue or completed carpool history.
async function loadGroups() {
  loading.value = true
  try {
    const response = await adminPaymentAPI.getCarpoolGroups(activeTab.value === 'history')
    groups.value = response.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

// Loads the configurable carpool packages for the administrator.
async function loadPlans() {
  loading.value = true
  try {
    const response = await adminPaymentAPI.getCarpoolPlans()
    plans.value = response.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

// Refreshes the data owned by the selected administrative tab.
function loadActiveTab() {
  return activeTab.value === 'plans' ? loadPlans() : loadGroups()
}

// Switches administrative datasets without retaining stale rows.
function switchTab(value: 'current' | 'history' | 'plans') {
  if (activeTab.value === value) return
  activeTab.value = value
  groups.value = []
  plans.value = []
  loadActiveTab()
}

// Opens an empty form for a new carpool package.
function openCreatePlanDialog() {
  editingPlan.value = null
  planForm.total_amount = 0
  planForm.target_members = 1
  planForm.note = ''
  planDialogOpen.value = true
}

// Opens the same three-field form with an existing package's values.
function openEditPlanDialog(plan: AdminCarpoolPlan) {
  editingPlan.value = plan
  planForm.total_amount = plan.total_amount
  planForm.target_members = plan.target_members
  planForm.note = plan.note
  planDialogOpen.value = true
}

// Closes the plan editor and clears its edit target.
function closePlanDialog() {
  planDialogOpen.value = false
  editingPlan.value = null
}

// Persists the validated business fields through the create or update endpoint.
async function savePlan() {
  if (!planFormValid.value) return
  actionLoading.value = true
  const payload: CarpoolPlanInput = {
    total_amount: planForm.total_amount,
    target_members: planForm.target_members,
    note: planForm.note.trim(),
  }
  try {
    if (editingPlan.value) await adminPaymentAPI.updateCarpoolPlan(editingPlan.value.id, payload)
    else await adminPaymentAPI.createCarpoolPlan(payload)
    closePlanDialog()
    appStore.showSuccess(t('common.success'))
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

// Deletes the selected package after the backend verifies that it is unused.
async function confirmDeletePlan() {
  if (!deletePlanTarget.value) return
  actionLoading.value = true
  try {
    await adminPaymentAPI.deleteCarpoolPlan(deletePlanTarget.value.id)
    deletePlanTarget.value = null
    appStore.showSuccess(t('common.success'))
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

// Opens the confirmation for a service period that starts when the admin confirms.
function openDeliveryDialog(group: AdminCarpoolGroup) {
  deliveryTarget.value = group
}

// Records delivery; the backend owns both the opening timestamp and 30-day expiry.
async function confirmDelivery() {
  if (!deliveryTarget.value) return
  actionLoading.value = true
  try {
    await adminPaymentAPI.openCarpoolGroup(deliveryTarget.value.id)
    deliveryTarget.value = null
    appStore.showSuccess(t('common.success'))
    await loadGroups()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

// Opens the group-level refund transition dialog.
function openRefundPendingDialog(group: AdminCarpoolGroup) {
  refundPendingTarget.value = group
  refundPendingReason.value = ''
}

// Closes a group before administrators refund its member orders one by one.
async function confirmRefundPending() {
  if (!refundPendingTarget.value || !refundPendingReason.value.trim()) return
  actionLoading.value = true
  try {
    await adminPaymentAPI.markCarpoolRefundPending(refundPendingTarget.value.id, refundPendingReason.value.trim())
    refundPendingTarget.value = null
    appStore.showSuccess(t('common.success'))
    await loadGroups()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

// Opens a member refund with the full amount for failed formation or a 30-day prorated amount for active service.
function openMemberRefundDialog(group: AdminCarpoolGroup, member: AdminCarpoolMember) {
  refundMember.value = member
  if (group.status_reason === 'not_formed' || !group.expires_at) {
    memberRefundForm.amount = member.amount
  } else {
    const remainingDays = Math.max(0, Math.ceil((new Date(group.expires_at).getTime() - Date.now()) / 86_400_000))
    memberRefundForm.amount = Math.round((member.amount * remainingDays / 30) * 100) / 100
  }
  memberRefundForm.reason = group.status_reason === 'not_formed' ? t('payment.admin.carpool.notFormedRefundReason') : group.status_reason || ''
}

// Sends a manual original-method refund without any balance deduction.
async function confirmMemberRefund() {
  if (!refundMember.value || !memberRefundForm.reason.trim()) return
  actionLoading.value = true
  try {
    const response = await adminPaymentAPI.refundOrder(refundMember.value.order_id, {
      amount: memberRefundForm.amount,
      reason: memberRefundForm.reason.trim(),
      deduct_balance: false,
      force: false,
    })
    if (response.data.success) {
      appStore.showSuccess(t('payment.admin.refundSuccess'))
    } else {
      appStore.showSuccess(response.data.warning || t('payment.admin.refundPending'))
    }
    refundMember.value = null
    await loadGroups()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

// Queries a gateway refund that has not reached a terminal status.
async function queryRefund(orderID: number) {
  actionLoading.value = true
  try {
    const response = await adminPaymentAPI.queryRefund(orderID)
    if (response.data.success) appStore.showSuccess(t('payment.admin.refundSuccess'))
    else appStore.showSuccess(response.data.warning || t('payment.admin.refundPending'))
    await loadGroups()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

// Only completed and failed-refund orders can enter the existing admin refund flow.
function canRefund(status: string): boolean {
  return status === 'COMPLETED' || status === 'REFUND_FAILED'
}

// Formats an API timestamp in the administrator's local timezone.
function formatDateTime(value: string): string {
  return new Date(value).toLocaleString()
}

// Maps the persisted group status to an administrator-facing label.
function statusText(status: CarpoolStatus): string {
  switch (status) {
  case 'waiting': return t('payment.carpool.status.waiting')
  case 'purchasing': return t('payment.carpool.status.purchasing')
  case 'active': return t('payment.carpool.status.active')
  case 'refund_pending': return t('payment.carpool.status.refundPending')
  case 'refunded': return t('payment.carpool.status.refunded')
  case 'expired': return t('payment.carpool.status.expired')
  }
}

// Uses stable status colors across current and history records.
function statusClass(status: CarpoolStatus): string {
  switch (status) {
  case 'waiting': return 'bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300'
  case 'purchasing': return 'bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300'
  case 'active': return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300'
  case 'refund_pending': return 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300'
  case 'refunded': return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  case 'expired': return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
  }
}

// Translates the system reason used for automatic formation failure.
function statusReasonText(reason: string): string {
  return reason === 'not_formed' ? t('payment.admin.carpool.notFormed') : reason
}

onMounted(loadActiveTab)
</script>
