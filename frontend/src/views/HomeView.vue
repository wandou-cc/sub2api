<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div v-else class="home-shell">
    <div class="motion-grid" aria-hidden="true"></div>
    <header class="home-header">
      <nav class="home-nav">
        <router-link to="/home" class="brand-link" :aria-label="t('home.homeAriaLabel')">
          <span class="brand-mark">
            <img :src="siteLogo || '/logo.svg'" :alt="t('home.logoAlt')" />
          </span>
          <span class="brand-copy">
            <span>{{ siteName }}</span>
            <small>{{ t('home.brandTagline') }}</small>
          </span>
        </router-link>

        <div class="nav-actions">
          <router-link to="/uclaw" class="nav-link nav-link-uclaw">
            UClaw
          </router-link>

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="nav-icon"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            v-else
            to="/docs"
            class="nav-icon"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </router-link>

          <LocaleSwitcher />

          <button
            type="button"
            @click="toggleTheme"
            class="nav-icon"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="nav-cta"
          >
            <span class="user-dot">{{ userInitial }}</span>
            <span>{{ t('home.dashboard') }}</span>
            <Icon name="arrowRight" size="sm" />
          </router-link>
          <router-link v-else to="/login" class="nav-cta">
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="home-main">
      <section class="hero-section">
        <div class="hero-copy-block">
          <p class="hero-eyebrow">{{ t('home.heroSubtitle') }}</p>
          <h1 class="hero-title">{{ siteName }}</h1>
          <p class="hero-subtitle">{{ siteSubtitle }}</p>
          <p class="hero-description">{{ t('home.heroDescription') }}</p>

          <div class="hero-actions">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="primary-action"
            >
              <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
              <Icon name="arrowRight" size="md" />
            </router-link>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="secondary-action"
            >
              {{ t('home.docs') }}
            </a>
            <router-link
              v-else
              to="/docs"
              class="secondary-action"
            >
              {{ t('home.docs') }}
            </router-link>
          </div>
        </div>

        <div class="gateway-panel" aria-hidden="true">
          <div class="panel-topline">
            <span>{{ t('home.gateway.routeLabel') }}</span>
            <span>{{ t('home.gateway.liveStatus') }}</span>
          </div>
          <div class="panel-command">
            <span>$</span>
            <code>{{ t('home.gateway.requestPath') }}</code>
          </div>
          <div class="route-list">
            <div class="route-row route-row-active">
              <span>01</span>
              <strong>{{ t('home.providers.claude') }}</strong>
              <em>{{ t('home.gateway.okStatus') }}</em>
            </div>
            <div class="route-row">
              <span>02</span>
              <strong>{{ t('home.providers.gpt') }}</strong>
              <em>{{ t('home.gateway.readyStatus') }}</em>
            </div>
            <div class="route-row">
              <span>03</span>
              <strong>{{ t('home.providers.gptImage2') }}</strong>
              <em>{{ t('home.gateway.syncStatus') }}</em>
            </div>
          </div>
          <div class="panel-meter">
            <span></span>
          </div>
          <div class="panel-foot">
            <span>{{ t('home.gateway.singleKey') }}</span>
            <span>{{ t('home.gateway.payAsYouGo') }}</span>
          </div>
        </div>
      </section>

      <section class="stats-strip" :aria-label="t('home.statsLabel')">
        <div>
          <strong>{{ t('home.tags.subscriptionToApi') }}</strong>
          <span>{{ t('home.tags.subscription') }}</span>
        </div>
        <div>
          <strong>{{ t('home.tags.stickySession') }}</strong>
          <span>{{ t('home.tags.affinity') }}</span>
        </div>
        <div>
          <strong>{{ t('home.tags.realtimeBilling') }}</strong>
          <span>{{ t('home.tags.billing') }}</span>
        </div>
      </section>

      <section class="section-block">
        <div class="section-head">
          <span>01</span>
          <div>
            <h2>{{ t('home.solutions.title') }}</h2>
            <p>{{ t('home.solutions.subtitle') }}</p>
          </div>
        </div>

        <div class="feature-ledger">
          <article class="feature-row">
            <span class="feature-index">01</span>
            <div>
              <h3>{{ t('home.features.unifiedGateway') }}</h3>
              <p>{{ t('home.features.unifiedGatewayDesc') }}</p>
            </div>
            <Icon name="server" size="lg" />
          </article>
          <article class="feature-row">
            <span class="feature-index">02</span>
            <div>
              <h3>{{ t('home.features.multiAccount') }}</h3>
              <p>{{ t('home.features.multiAccountDesc') }}</p>
            </div>
            <Icon name="shield" size="lg" />
          </article>
          <article class="feature-row">
            <span class="feature-index">03</span>
            <div>
              <h3>{{ t('home.features.balanceQuota') }}</h3>
              <p>{{ t('home.features.balanceQuotaDesc') }}</p>
            </div>
            <Icon name="chart" size="lg" />
          </article>
        </div>
      </section>

      <section class="section-block providers-block">
        <div class="section-head">
          <span>02</span>
          <div>
            <h2>{{ t('home.providers.title') }}</h2>
            <p>{{ t('home.providers.description') }}</p>
          </div>
        </div>

        <div class="provider-grid">
          <div class="provider-chip">
            <span>C</span>
            <strong>{{ t('home.providers.claude') }}</strong>
            <em>{{ t('home.providers.supported') }}</em>
          </div>
          <div class="provider-chip">
            <span>G</span>
            <strong>{{ t('home.providers.gpt') }}</strong>
            <em>{{ t('home.providers.supported') }}</em>
          </div>
          <div class="provider-chip">
            <span>I</span>
            <strong>{{ t('home.providers.gptImage2') }}</strong>
            <em>{{ t('home.providers.supported') }}</em>
          </div>
          <div class="provider-chip muted">
            <span>+</span>
            <strong>{{ t('home.providers.more') }}</strong>
            <em>{{ t('home.providers.soon') }}</em>
          </div>
        </div>
      </section>

      <section class="closer-section">
        <p>{{ t('home.cta.description') }}</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="primary-action"
        >
          <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.cta.button') }}</span>
          <Icon name="arrowRight" size="md" />
        </router-link>
      </section>
    </main>

    <footer class="home-footer">
      <div>
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <nav>
          <router-link to="/uclaw">
            UClaw
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
          >
            {{ t('home.docs') }}
          </a>
          <router-link v-else to="/docs">
            {{ t('home.docs') }}
          </router-link>
        </nav>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || t('home.defaultSubtitle'))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.home-shell {
  position: relative;
  display: flex;
  min-height: 100vh;
  flex-direction: column;
  overflow-x: hidden;
  background:
    linear-gradient(color-mix(in srgb, var(--app-ink) 3.5%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--app-ink) 3.5%, transparent) 1px, transparent 1px),
    var(--app-bg);
  background-size: 80px 80px;
  color: var(--app-ink);
  font-family: var(--app-font-sans);
}

.motion-grid {
  pointer-events: none;
  position: fixed;
  inset: 0;
  z-index: 0;
  opacity: 0.55;
  background: linear-gradient(
    115deg,
    transparent 0%,
    transparent 42%,
    color-mix(in srgb, var(--app-accent) 10%, transparent) 50%,
    transparent 58%,
    transparent 100%
  );
  transform: translateX(-60%);
  animation: home-scan 8s ease-in-out infinite;
}

.home-header {
  position: sticky;
  top: 0;
  z-index: 20;
  border-bottom: 1px solid var(--app-line);
  background: color-mix(in srgb, var(--app-bg) 88%, transparent);
  backdrop-filter: blur(18px);
  animation: home-drop 0.55s ease-out both;
}

.home-nav {
  display: flex;
  max-width: 1180px;
  min-height: 72px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin: 0 auto;
  padding: 0 24px;
}

.brand-link {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
  color: var(--app-ink);
  text-decoration: none;
}

.brand-mark {
  display: flex;
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--app-ink);
  border-radius: 8px;
  background: var(--app-surface);
  transition:
    box-shadow 0.25s ease,
    transform 0.25s ease;
}

.brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.brand-link:hover .brand-mark {
  box-shadow: 5px 5px 0 var(--app-ink);
  transform: translate(-2px, -2px);
}

.brand-copy {
  display: grid;
  min-width: 0;
  line-height: 1.1;
}

.brand-copy span {
  overflow: hidden;
  font-size: 16px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand-copy small {
  margin-top: 3px;
  color: var(--app-soft);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: uppercase;
}

.nav-actions {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.nav-link,
.nav-icon,
.nav-cta {
  border: 1px solid var(--app-line);
  background: var(--app-surface);
  color: var(--app-ink);
  text-decoration: none;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.nav-link,
.nav-cta {
  display: inline-flex;
  min-height: 38px;
  align-items: center;
  gap: 8px;
  border-radius: 999px;
  padding: 0 14px;
  font-size: 13px;
  font-weight: 700;
}

.nav-icon {
  display: inline-flex;
  width: 38px;
  height: 38px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
}

.nav-link:hover,
.nav-icon:hover {
  border-color: var(--app-ink);
  transform: translateY(-1px);
}

.nav-cta {
  border-color: var(--app-ink);
  background: var(--app-ink);
  color: var(--app-bg);
}

.nav-cta:hover,
.primary-action:hover,
.secondary-action:hover {
  transform: translateY(-2px);
}

.user-dot {
  display: inline-flex;
  width: 22px;
  height: 22px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--app-bg);
  color: var(--app-ink);
  font-size: 11px;
  font-weight: 800;
}

.home-main {
  flex: 1;
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
}

.hero-section {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(340px, 0.9fr);
  gap: 72px;
  align-items: center;
  padding: 104px 0 84px;
  border-bottom: 1px solid var(--app-line);
  animation: home-rise 0.7s ease-out 0.08s both;
}

.hero-copy-block {
  min-width: 0;
}

.hero-eyebrow {
  margin: 0 0 18px;
  color: var(--app-accent);
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.hero-title {
  max-width: 760px;
  margin: 0;
  overflow-wrap: anywhere;
  font-family: var(--app-font-display);
  font-size: 112px;
  font-weight: 700;
  line-height: 0.92;
  letter-spacing: 0;
}

.hero-subtitle {
  max-width: 620px;
  margin: 28px 0 0;
  color: var(--app-ink);
  font-size: 30px;
  font-weight: 700;
  line-height: 1.25;
}

.hero-description {
  max-width: 610px;
  margin: 18px 0 0;
  color: var(--app-muted);
  font-size: 16px;
  line-height: 1.9;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 34px;
}

.primary-action,
.secondary-action {
  display: inline-flex;
  min-height: 48px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-radius: 999px;
  padding: 0 22px;
  font-size: 14px;
  font-weight: 800;
  text-decoration: none;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.primary-action {
  border: 1px solid var(--app-ink);
  background: var(--app-ink);
  color: var(--app-bg);
  box-shadow: 8px 8px 0 var(--app-line);
}

.secondary-action {
  border: 1px solid var(--app-line);
  background: var(--app-surface);
  color: var(--app-ink);
}

.gateway-panel {
  min-width: 0;
  border: 1px solid var(--app-ink);
  border-radius: 8px;
  background: var(--app-surface);
  box-shadow: 14px 14px 0 var(--app-ink);
  transform-origin: center;
  animation: home-panel 0.75s ease-out 0.2s both, home-float 5s ease-in-out 1s infinite;
}

:global(.dark) .gateway-panel {
  box-shadow: 14px 14px 0 #000;
}

.panel-topline,
.panel-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: var(--app-soft);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.panel-topline {
  border-bottom: 1px solid var(--app-line);
  padding: 16px 18px;
}

.panel-command {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 28px 22px;
  border-bottom: 1px solid var(--app-line);
  font-family: var(--app-font-mono);
  font-size: 15px;
}

.panel-command span {
  color: var(--app-accent);
  font-weight: 900;
}

.panel-command code {
  white-space: normal;
  overflow-wrap: anywhere;
}

.route-list {
  padding: 8px 0;
}

.route-row {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr) auto;
  align-items: center;
  gap: 16px;
  min-height: 58px;
  padding: 0 22px;
  border-bottom: 1px solid var(--app-line);
  transition:
    background-color 0.25s ease,
    transform 0.25s ease;
}

.route-row:last-child {
  border-bottom: 0;
}

.route-row span {
  color: var(--app-soft);
  font-family: var(--app-font-mono);
  font-size: 12px;
}

.route-row strong {
  overflow: hidden;
  font-size: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.route-row em {
  border: 1px solid var(--app-line);
  border-radius: 999px;
  padding: 5px 9px;
  color: var(--app-accent);
  font-size: 11px;
  font-style: normal;
  font-weight: 800;
}

.route-row-active {
  background: color-mix(in srgb, var(--app-accent) 8%, transparent);
}

.route-row:nth-child(2) {
  animation: route-pulse 4s ease-in-out 0.8s infinite;
}

.route-row:nth-child(3) {
  animation: route-pulse 4s ease-in-out 1.6s infinite;
}

.route-row:hover {
  background: color-mix(in srgb, var(--app-accent) 10%, transparent);
  transform: translateX(4px);
}

.panel-meter {
  padding: 20px 22px 4px;
}

.panel-meter span {
  display: block;
  height: 8px;
  border: 1px solid var(--app-ink);
  border-radius: 999px;
  background:
    linear-gradient(90deg, var(--app-accent) 0 72%, transparent 72% 100%),
    var(--app-surface-muted);
  background-size: 140% 100%;
  animation: meter-flow 2.8s ease-in-out infinite;
}

.panel-foot {
  padding: 16px 22px 22px;
}

.stats-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-bottom: 1px solid var(--app-line);
  animation: home-rise 0.7s ease-out 0.22s both;
}

.stats-strip div {
  min-width: 0;
  padding: 28px 24px;
  border-right: 1px solid var(--app-line);
  transition:
    background-color 0.25s ease,
    transform 0.25s ease;
}

.stats-strip div:last-child {
  border-right: 0;
}

.stats-strip div:hover {
  background: color-mix(in srgb, var(--app-accent) 7%, transparent);
  transform: translateY(-4px);
}

.stats-strip strong,
.stats-strip span {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stats-strip strong {
  font-size: 18px;
}

.stats-strip span {
  margin-top: 8px;
  color: var(--app-soft);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.section-block {
  padding: 84px 0;
  border-bottom: 1px solid var(--app-line);
  animation: home-rise 0.7s ease-out 0.28s both;
}

.section-head {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: 24px;
  align-items: start;
  margin-bottom: 34px;
}

.section-head > span {
  display: inline-flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--app-ink);
  border-radius: 50%;
  font-family: var(--app-font-mono);
  font-size: 12px;
  font-weight: 900;
}

.section-head h2 {
  margin: 0;
  font-size: 42px;
  line-height: 1.12;
  letter-spacing: 0;
}

.section-head p {
  max-width: 560px;
  margin: 12px 0 0;
  color: var(--app-muted);
  font-size: 16px;
  line-height: 1.8;
}

.feature-ledger {
  border-top: 1px solid var(--app-line);
}

.feature-row {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr) 42px;
  gap: 24px;
  align-items: center;
  min-height: 158px;
  border-bottom: 1px solid var(--app-line);
  transition:
    background-color 0.25s ease,
    transform 0.25s ease;
}

.feature-row:hover {
  background: color-mix(in srgb, var(--app-accent) 6%, transparent);
  transform: translateX(8px);
}

.feature-index {
  color: var(--app-accent);
  font-family: var(--app-font-mono);
  font-size: 13px;
  font-weight: 900;
}

.feature-row h3 {
  margin: 0;
  font-size: 24px;
  line-height: 1.3;
}

.feature-row p {
  max-width: 720px;
  margin: 10px 0 0;
  color: var(--app-muted);
  font-size: 15px;
  line-height: 1.8;
}

.feature-row svg {
  color: var(--app-accent);
}

.provider-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
}

.provider-chip {
  display: grid;
  min-height: 154px;
  align-content: space-between;
  border: 1px solid var(--app-line);
  border-radius: 8px;
  background: var(--app-surface);
  padding: 18px;
  transition:
    border-color 0.25s ease,
    box-shadow 0.25s ease,
    transform 0.25s ease;
}

.provider-chip:hover {
  border-color: var(--app-ink);
  box-shadow: 8px 8px 0 var(--app-line);
  transform: translateY(-6px);
}

.provider-chip span {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--app-ink);
  border-radius: 50%;
  font-weight: 900;
}

.provider-chip strong {
  overflow: hidden;
  font-size: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-chip em {
  color: var(--app-accent);
  font-size: 12px;
  font-style: normal;
  font-weight: 800;
}

.provider-chip.muted {
  background: var(--app-surface-muted);
  color: var(--app-muted);
}

.closer-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 54px 0;
  border-bottom: 1px solid var(--app-line);
  animation: home-rise 0.7s ease-out 0.34s both;
}

.closer-section p {
  max-width: 620px;
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  line-height: 1.5;
}

.home-footer {
  border-top: 0;
}

.home-footer > div {
  display: flex;
  max-width: 1180px;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin: 0 auto;
  padding: 28px 24px;
  color: var(--app-muted);
  font-size: 13px;
}

.home-footer p {
  margin: 0;
}

.home-footer nav {
  display: flex;
  gap: 16px;
}

.home-footer a {
  color: var(--app-muted);
  font-weight: 700;
  text-decoration: none;
}

.home-footer a:hover {
  color: var(--app-ink);
}

@keyframes home-drop {
  from {
    opacity: 0;
    transform: translateY(-12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes home-rise {
  from {
    opacity: 0;
    transform: translateY(24px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes home-panel {
  from {
    opacity: 0;
    transform: translateY(18px) rotate(-1deg);
  }
  to {
    opacity: 1;
    transform: translateY(0) rotate(0);
  }
}

@keyframes home-float {
  0%,
  100% {
    translate: 0 0;
  }
  50% {
    translate: 0 -10px;
  }
}

@keyframes home-scan {
  0%,
  35% {
    transform: translateX(-60%);
  }
  70%,
  100% {
    transform: translateX(60%);
  }
}

@keyframes meter-flow {
  0%,
  100% {
    background-position: 0 0;
  }
  50% {
    background-position: 100% 0;
  }
}

@keyframes route-pulse {
  0%,
  100% {
    background: transparent;
  }
  50% {
    background: color-mix(in srgb, var(--app-accent) 7%, transparent);
  }
}

@media (max-width: 980px) {
  .home-nav {
    align-items: flex-start;
    flex-direction: column;
    padding: 16px 24px;
  }

  .nav-actions {
    width: 100%;
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .home-main {
    width: min(100% - 32px, 720px);
  }

  .hero-section {
    grid-template-columns: minmax(0, 1fr);
    gap: 44px;
    padding: 72px 0 64px;
  }

  .hero-title {
    font-size: 72px;
  }

  .hero-subtitle {
    font-size: 24px;
  }

  .stats-strip,
  .provider-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .stats-strip div,
  .stats-strip div:last-child {
    border-right: 0;
    border-bottom: 1px solid var(--app-line);
  }

  .stats-strip div:last-child {
    border-bottom: 0;
  }

  .feature-row {
    grid-template-columns: 48px minmax(0, 1fr);
  }

  .feature-row svg {
    display: none;
  }

  .closer-section {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 560px) {
  .home-nav,
  .home-footer > div {
    padding-right: 16px;
    padding-left: 16px;
  }

  .nav-link {
    display: none;
  }

  .nav-link-uclaw {
    display: inline-flex;
  }

  .hero-title {
    font-size: 56px;
  }

  .hero-subtitle {
    font-size: 21px;
  }

  .primary-action,
  .secondary-action,
  .nav-cta {
    width: 100%;
  }

  .gateway-panel {
    box-shadow: 8px 8px 0 var(--app-ink);
  }

  .route-row {
    grid-template-columns: 32px minmax(0, 1fr);
  }

  .route-row em {
    grid-column: 2;
    justify-self: start;
  }

  .section-block {
    padding: 64px 0;
  }

  .section-head {
    grid-template-columns: minmax(0, 1fr);
    gap: 16px;
  }

  .section-head h2 {
    font-size: 32px;
  }

  .feature-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 12px;
    padding: 28px 0;
  }

  .home-footer > div {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-header,
  .hero-section,
  .gateway-panel,
  .stats-strip,
  .section-block,
  .closer-section,
  .route-row,
  .panel-meter span,
  .motion-grid {
    animation-duration: 1ms;
    animation-iteration-count: 1;
    transition-duration: 1ms;
  }

  .gateway-panel {
    animation-name: home-panel;
  }

  .motion-grid {
    display: none;
  }
}
</style>
