<template>
  <div class="space-y-7">
    <!-- Quick Amount Buttons -->
    <div>
      <label class="mb-3 block text-sm font-semibold text-gray-900 dark:text-white">
        {{ t('payment.quickAmounts') }}
      </label>
      <div class="grid grid-cols-5 gap-2">
        <button
          v-for="amt in filteredAmounts"
          :key="amt"
          type="button"
          :class="[
            'min-h-14 rounded-xl border px-2 py-3 text-center text-base font-semibold tabular-nums transition-colors',
            modelValue === amt
              ? 'border-gray-900 bg-gray-900 text-white shadow-sm dark:border-white dark:bg-white dark:text-gray-950'
              : 'border-gray-200 bg-white text-gray-700 hover:border-gray-400 hover:text-gray-950 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-dark-400 dark:hover:text-white',
          ]"
          @click="selectAmount(amt)"
        >
          {{ amt }}
        </button>
      </div>
    </div>

    <!-- Custom Amount Input -->
    <div>
      <label class="mb-3 block text-sm font-semibold text-gray-900 dark:text-white">
        {{ t('payment.customAmount') }}
      </label>
      <div class="relative">
        <span class="absolute left-5 top-1/2 -translate-y-1/2 font-serif text-xl font-semibold text-gray-500 dark:text-gray-400">
          {{ inputCurrencySymbol }}
        </span>
        <input
          type="text"
          inputmode="decimal"
          :value="customText"
          :placeholder="placeholderText"
          class="input h-14 w-full rounded-xl pl-12 pr-5 text-base"
          @input="handleInput"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { currencySymbol } from './currency'

const props = withDefaults(defineProps<{
  amounts?: number[]
  modelValue: number | null
  min?: number
  max?: number
  currency?: string
}>(), {
  amounts: () => [1, 10, 30, 50, 100],
  min: 0,
  max: 0,
  currency: 'CNY',
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const customText = ref('')
const inputCurrencySymbol = computed(() => currencySymbol(props.currency))

// 0 = no limit
const filteredAmounts = computed(() =>
  props.amounts.filter((a) => (props.min <= 0 || a >= props.min) && (props.max <= 0 || a <= props.max))
)

const placeholderText = computed(() => {
  if (props.min > 0 && props.max > 0) return `${props.min} - ${props.max}`
  if (props.min > 0) return `≥ ${props.min}`
  if (props.max > 0) return `≤ ${props.max}`
  return t('payment.enterAmount')
})

const AMOUNT_PATTERN = /^\d*(\.\d{0,2})?$/

function selectAmount(amt: number) {
  customText.value = String(amt)
  emit('update:modelValue', amt)
}

function handleInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  if (!AMOUNT_PATTERN.test(val)) return
  customText.value = val
  if (val === '') {
    emit('update:modelValue', null)
    return
  }
  const num = parseFloat(val)
  if (!isNaN(num) && num > 0) {
    emit('update:modelValue', num)
  } else {
    emit('update:modelValue', null)
  }
}

watch(() => props.modelValue, (v) => {
  if (v !== null && String(v) !== customText.value) {
    customText.value = String(v)
  }
}, { immediate: true })
</script>
