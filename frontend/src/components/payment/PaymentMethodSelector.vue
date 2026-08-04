<template>
  <div>
    <label class="mb-3 block text-sm font-semibold text-gray-900 dark:text-white">
      {{ t('payment.paymentMethod') }}
    </label>
    <div
      data-testid="payment-method-grid"
      class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4"
    >
      <button
        v-for="method in sortedMethods"
        :key="method.type"
        type="button"
        :title="methodLabel(method)"
        :disabled="!method.available"
        :class="[
          'relative flex h-[68px] min-w-0 items-center justify-between rounded-xl border px-4 text-left transition-colors',
          !method.available
            ? 'cursor-not-allowed border-gray-200 bg-gray-50 opacity-50 dark:border-dark-700 dark:bg-dark-800/50'
            : selected === method.type
              ? 'border-gray-950 bg-white text-gray-950 shadow-sm dark:border-white dark:bg-dark-800 dark:text-white'
              : 'border-gray-300 bg-white text-gray-700 hover:border-gray-400 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-dark-500',
        ]"
        @click="method.available && emit('select', method.type)"
      >
        <span class="flex min-w-0 flex-1 items-center gap-3">
          <span :class="['flex h-9 w-9 shrink-0 items-center justify-center rounded-lg', methodIconBackgroundClass(method.type)]">
            <img :src="methodIcon(method.type)" :alt="methodLabel(method)" class="h-6 w-6 object-contain" />
          </span>
          <span class="flex min-w-0 flex-1 flex-col items-start gap-1 leading-none">
            <span data-testid="payment-method-label" class="block w-full truncate text-sm font-semibold">
              {{ methodLabel(method) }}
            </span>
            <span
              v-if="method.fee_rate > 0"
              class="text-xs text-gray-500 dark:text-gray-400"
            >
              {{ t('payment.fee') }} {{ method.fee_rate }}%
            </span>
          </span>
        </span>
        <span
          :class="[
            'ml-3 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border',
            selected === method.type
              ? 'border-gray-950 bg-gray-950 text-white dark:border-white dark:bg-white dark:text-gray-950'
              : 'border-gray-300 text-transparent dark:border-dark-500',
          ]"
        >
          <Icon name="check" size="xs" :stroke-width="2.5" />
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { METHOD_ORDER, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from './providerConfig'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'
import airwallexIcon from '@/assets/icons/airwallex.svg'
import paymentIcon from '@/assets/icons/payment.svg'

export interface PaymentMethodOption {
  type: string
  display_name?: string
  fee_rate: number
  available: boolean
}

const props = defineProps<{
  methods: PaymentMethodOption[]
  selected: string
}>()

const emit = defineEmits<{
  select: [type: string]
}>()

const { t } = useI18n()

const METHOD_ICONS: Record<string, string> = {
  alipay: alipayIcon,
  wxpay: wxpayIcon,
  stripe: stripeIcon,
  airwallex: airwallexIcon,
  credit_card: paymentIcon,
}

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

function methodIcon(type: string): string {
  if (isBuiltInAlipayMethod(type)) return METHOD_ICONS.alipay
  if (isBuiltInWxpayMethod(type)) return METHOD_ICONS.wxpay
  if (type === 'airwallex') return METHOD_ICONS.airwallex
  return METHOD_ICONS[type] || paymentIcon
}

function methodLabel(method: PaymentMethodOption): string {
  return method.display_name || t(`payment.methods.${method.type}`, method.type)
}

// Keep each provider icon legible without turning the entire payment option into a brand-color block.
function methodIconBackgroundClass(type: string): string {
  if (isBuiltInAlipayMethod(type)) return 'bg-[#E8F7FD] dark:bg-[#02A9F1]/15'
  if (isBuiltInWxpayMethod(type)) return 'bg-[#EAF8EA] dark:bg-[#09BB07]/15'
  if (type === 'stripe') return 'bg-[#EEEEFF] dark:bg-[#676BE5]/15'
  if (type === 'airwallex') return 'bg-[#FFF0EB] dark:bg-[#FF6B3D]/15'
  return 'bg-gray-100 dark:bg-dark-700'
}
</script>
