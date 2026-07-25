<template>
  <AppLayout>
    <section class="flex min-h-[calc(100vh-10rem)] flex-col items-center justify-center px-4 pb-12 text-center">
      <p class="text-sm font-medium text-primary-600 dark:text-primary-400">
        {{ t('contact.wechatLabel') }}
      </p>
      <h2 class="mt-2 text-xl font-semibold text-gray-900 dark:text-white">
        {{ t('contact.helpTitle') }}
      </h2>
      <p class="mt-4 max-w-lg text-sm leading-7 text-gray-600 dark:text-gray-300">
        {{ t('contact.helpDescription') }}
      </p>

      <img
        :src="wechatContactImage"
        :alt="t('contact.wechatImageAlt')"
        class="mt-6 w-full max-w-[280px] object-contain"
      />

      <p class="mt-5 max-w-md text-sm leading-6 text-gray-500 dark:text-gray-400">
        {{ t('contact.scanHint') }}
      </p>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import wechatContactImage from '@/assets/icons/IMG_1643.jpg'

const { t } = useI18n()
</script>
