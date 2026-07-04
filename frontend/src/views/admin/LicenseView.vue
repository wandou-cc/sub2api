<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-wrap items-center gap-3">
        <div class="min-w-0 flex-1">
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.license.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('admin.license.description') }}
          </p>
        </div>
        <button class="btn btn-secondary" :disabled="loading" @click="loadCodes">
          <Icon name="refresh" size="md" :class="{ 'animate-spin': loading }" />
        </button>
        <button class="btn btn-primary" @click="showCreateDialog = true">
          <Icon name="plus" size="md" class="mr-2" />
          {{ t('admin.license.createCodes') }}
        </button>
      </div>

      <div class="grid gap-4 md:grid-cols-4">
        <div v-for="item in stats" :key="item.key" class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <div class="text-sm text-gray-500 dark:text-dark-400">{{ item.label }}</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ item.value }}</div>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <input
          v-model="search"
          type="text"
          class="input max-w-sm"
          :placeholder="t('admin.license.searchPlaceholder')"
        />
        <select v-model="statusFilter" class="input w-40">
          <option value="">{{ t('admin.license.allStatuses') }}</option>
          <option v-for="status in statuses" :key="status" :value="status">
            {{ statusLabel(status) }}
          </option>
        </select>
      </div>

      <div class="card overflow-hidden">
        <DataTable :columns="columns" :data="filteredCodes" :loading="loading">
          <template #cell-code="{ value }">
            <div class="flex items-center gap-2">
              <code class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ value }}</code>
              <button
                class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                :title="t('common.copy')"
                @click="copyText(value)"
              >
                <Icon name="copy" size="sm" />
              </button>
            </div>
          </template>

          <template #cell-licenseId="{ value }">
            <code v-if="value" class="font-mono text-xs text-gray-700 dark:text-gray-300">{{ value }}</code>
            <span v-else class="text-gray-400">-</span>
          </template>

          <template #cell-features="{ value }">
            <div class="flex flex-wrap gap-1">
              <span
                v-for="feature in value"
                :key="feature"
                class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-700 dark:bg-dark-700 dark:text-gray-200"
              >
                {{ feature }}
              </span>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', statusClass(value)]">{{ statusLabel(value) }}</span>
          </template>

          <template #cell-usbFingerprint="{ value }">
            <code v-if="value" class="font-mono text-xs text-gray-700 dark:text-gray-300">{{ value }}</code>
            <span v-else class="text-gray-400">-</span>
          </template>

          <template #cell-createdAt="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-expiresAt="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : t('admin.license.neverExpires') }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                class="rounded-lg p-1.5 text-gray-500 hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400"
                :title="t('admin.license.editFeatures')"
                @click="openEditFeatures(row)"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                v-if="row.status === 'disabled' || row.status === 'refunded' || row.status === 'revoked'"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20 dark:hover:text-green-400"
                :title="t('admin.license.enable')"
                @click="enable(row)"
              >
                <Icon name="check" size="sm" />
              </button>
              <button
                v-if="row.status === 'unused' || row.status === 'active'"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-yellow-50 hover:text-yellow-600 dark:hover:bg-yellow-900/20 dark:hover:text-yellow-400"
                :title="t('admin.license.disable')"
                @click="disable(row)"
              >
                <Icon name="ban" size="sm" />
              </button>
              <button
                v-if="row.status === 'unused' || row.status === 'active'"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('admin.license.refund')"
                @click="refund(row)"
              >
                <Icon name="xCircle" size="sm" />
              </button>
              <button
                v-if="row.licenseId && row.status === 'active'"
                class="rounded-lg p-1.5 text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('admin.license.revokeLicense')"
                @click="revoke(row)"
              >
                <Icon name="lock" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
      </div>
    </div>

    <BaseDialog
      :show="showCreateDialog"
      :title="t('admin.license.createCodes')"
      width="normal"
      @close="showCreateDialog = false"
    >
      <form id="license-create-form" class="space-y-4" @submit.prevent="create">
        <div>
          <label class="input-label">{{ t('admin.license.count') }}</label>
          <input v-model.number="createForm.count" type="number" min="1" max="1000" required class="input" />
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.license.product') }}</label>
            <input v-model="createForm.product" type="text" class="input" placeholder="uclaw-usb" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.license.productBatch') }}</label>
            <input v-model="createForm.productBatch" type="text" class="input" placeholder="dev-2026-06" />
          </div>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.license.prefix') }}</label>
            <input v-model="createForm.prefix" type="text" class="input" placeholder="UCLAW" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.license.expiresAt') }}</label>
            <input v-model="createForm.expiresAt" type="datetime-local" class="input" />
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.license.features') }}</label>
          <input v-model="createForm.features" type="text" class="input" placeholder="openmontage" />
        </div>
      </form>

      <template #footer>
        <button type="button" class="btn btn-secondary" @click="showCreateDialog = false">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" form="license-create-form" class="btn btn-primary" :disabled="creating">
          {{ creating ? t('common.processing') : t('common.create') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showEditFeaturesDialog"
      :title="t('admin.license.editFeatures')"
      width="normal"
      @close="showEditFeaturesDialog = false"
    >
      <form id="license-edit-features-form" class="space-y-4" @submit.prevent="saveFeatures">
        <div>
          <label class="input-label">{{ t('admin.license.code') }}</label>
          <code class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ editingCode?.code }}</code>
        </div>
        <div>
          <label class="input-label">{{ t('admin.license.features') }}</label>
          <input v-model="editFeaturesForm.features" type="text" class="input" placeholder="openmontage, video-use" />
          <p class="mt-1 text-xs text-gray-400">{{ t('admin.license.featuresHint') }}</p>
        </div>
      </form>

      <template #footer>
        <button type="button" class="btn btn-secondary" @click="showEditFeaturesDialog = false">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" form="license-edit-features-form" class="btn btn-primary" :disabled="savingFeatures">
          {{ savingFeatures ? t('common.processing') : t('common.save') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import licenseAPI, { type LicenseCode, type LicenseCodeStatus } from '@/api/admin/license'

const { t } = useI18n()
const appStore = useAppStore()

const codes = ref<LicenseCode[]>([])
const loading = ref(false)
const creating = ref(false)
const showCreateDialog = ref(false)
const search = ref('')
const statusFilter = ref<LicenseCodeStatus | ''>('')

const showEditFeaturesDialog = ref(false)
const savingFeatures = ref(false)
const editingCode = ref<LicenseCode | null>(null)
const editFeaturesForm = reactive({ features: '' })

const statuses: LicenseCodeStatus[] = ['unused', 'active', 'disabled', 'expired', 'revoked', 'refunded']

const createForm = reactive({
  count: 1,
  product: '',
  productBatch: '',
  features: 'openmontage',
  prefix: 'UCLAW',
  expiresAt: '',
})

const columns = [
  { key: 'code', label: t('admin.license.code') },
  { key: 'licenseId', label: t('admin.license.licenseId') },
  { key: 'product', label: t('admin.license.product') },
  { key: 'productBatch', label: t('admin.license.productBatch') },
  { key: 'features', label: t('admin.license.features') },
  { key: 'status', label: t('admin.license.status') },
  { key: 'usbFingerprint', label: t('admin.license.usbFingerprint') },
  { key: 'createdAt', label: t('admin.license.createdAt') },
  { key: 'expiresAt', label: t('admin.license.expiresAt') },
  { key: 'actions', label: t('common.actions') },
]

const filteredCodes = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  return codes.value.filter((item) => {
    if (statusFilter.value && item.status !== statusFilter.value) {
      return false
    }
    if (!keyword) {
      return true
    }
    return [item.code, item.codeId, item.licenseId, item.product, item.productBatch, item.usbFingerprint]
      .filter((value) => value)
      .some((value) => value.toLowerCase().includes(keyword))
  })
})

const stats = computed(() => [
  { key: 'total', label: t('admin.license.total'), value: codes.value.length },
  { key: 'unused', label: statusLabel('unused'), value: codes.value.filter((item) => item.status === 'unused').length },
  { key: 'active', label: statusLabel('active'), value: codes.value.filter((item) => item.status === 'active').length },
  { key: 'revoked', label: statusLabel('revoked'), value: codes.value.filter((item) => item.status === 'revoked').length },
])

function statusLabel(status: LicenseCodeStatus): string {
  return t(`admin.license.statuses.${status}`)
}

function statusClass(status: LicenseCodeStatus): string {
  switch (status) {
    case 'active':
      return 'badge-success'
    case 'unused':
      return 'badge-primary'
    case 'disabled':
    case 'expired':
      return 'badge-warning'
    default:
      return 'badge-danger'
  }
}

async function loadCodes(): Promise<void> {
  loading.value = true
  try {
    codes.value = await licenseAPI.listCodes()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    loading.value = false
  }
}

async function create(): Promise<void> {
  creating.value = true
  try {
    await licenseAPI.createCodes({
      count: createForm.count,
      product: createForm.product.trim() || undefined,
      productBatch: createForm.productBatch.trim() || undefined,
      prefix: createForm.prefix.trim() || undefined,
      expiresAt: createForm.expiresAt ? new Date(createForm.expiresAt).toISOString() : undefined,
      features: createForm.features.split(',').map((item) => item.trim()).filter(Boolean),
    })
    showCreateDialog.value = false
    appStore.showSuccess(t('admin.license.created'))
    await loadCodes()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    creating.value = false
  }
}

async function disable(row: LicenseCode): Promise<void> {
  await runAction(() => licenseAPI.disableCode(row.codeId))
}

async function enable(row: LicenseCode): Promise<void> {
  await runAction(() => licenseAPI.enableCode(row.codeId))
}

async function refund(row: LicenseCode): Promise<void> {
  if (window.confirm(t('admin.license.confirmRefund'))) {
    await runAction(() => licenseAPI.refundCode(row.codeId))
  }
}

async function revoke(row: LicenseCode): Promise<void> {
  if (window.confirm(t('admin.license.confirmRevoke'))) {
    await runAction(() => licenseAPI.revokeLicense(row.licenseId))
  }
}

function openEditFeatures(row: LicenseCode): void {
  editingCode.value = row
  editFeaturesForm.features = row.features.join(', ')
  showEditFeaturesDialog.value = true
}

async function saveFeatures(): Promise<void> {
  if (!editingCode.value) {
    return
  }
  const features = editFeaturesForm.features
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
  savingFeatures.value = true
  try {
    await licenseAPI.updateCodeFeatures(editingCode.value.codeId, { features })
    showEditFeaturesDialog.value = false
    appStore.showSuccess(t('admin.license.featuresUpdated'))
    await loadCodes()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    savingFeatures.value = false
  }
}

async function runAction(action: () => Promise<LicenseCode>): Promise<void> {
  try {
    await action()
    appStore.showSuccess(t('common.success'))
    await loadCodes()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function copyText(value: string): Promise<void> {
  await navigator.clipboard.writeText(value)
  appStore.showSuccess(t('common.copied'))
}

onMounted(loadCodes)
</script>
