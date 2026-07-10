<template>
  <div class="uclaw-shell">
    <header class="uclaw-topbar">
      <router-link to="/home" class="uclaw-brand">
        <span class="uclaw-brand-mark">
          <img :src="siteLogo || '/logo.png'" :alt="siteName" />
        </span>
        <span>
          <strong>UClaw</strong>
          <small>Portable AI Workbench</small>
        </span>
      </router-link>

      <nav class="uclaw-nav">
        <router-link to="/home">首页</router-link>
        <router-link to="/docs">文档</router-link>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'">
          {{ isAuthenticated ? '控制台' : '登录' }}
        </router-link>
      </nav>
    </header>

    <main>
      <section class="uclaw-hero">
        <div class="uclaw-hero-copy">
          <p class="uclaw-eyebrow">装在 U 盘里的 AI 工作台</p>
          <h1>UClaw</h1>
          <p class="uclaw-hero-lead">
            插上电脑、双击启动、输入激活码即可使用。视频、办公、文档、设计、搜索等常用 AI 能力随身带走，用完拔掉不留痕迹。
          </p>
          <div class="uclaw-actions">
            <button type="button" class="uclaw-primary" @click="showContact = true">
              <span>购买联系客服</span>
              <Icon name="chat" size="md" />
            </button>
            <a href="#capabilities" class="uclaw-secondary">
              查看能力
            </a>
          </div>
        </div>

        <div class="uclaw-device" aria-label="UClaw 产品能力概览">
          <div class="uclaw-device-head">
            <span></span>
            <strong>UClaw</strong>
            <em>READY</em>
          </div>
          <div class="uclaw-usb">
            <div class="uclaw-usb-metal"></div>
            <div class="uclaw-usb-body">
              <span>AI</span>
              <strong>随身工作台</strong>
            </div>
          </div>
          <div class="uclaw-console">
            <div v-for="item in heroSignals" :key="item.name">
              <span>{{ item.name }}</span>
              <strong>{{ item.value }}</strong>
            </div>
          </div>
        </div>
      </section>

      <section class="uclaw-strip">
        <div v-for="item in highlights" :key="item.title">
          <strong>{{ item.title }}</strong>
          <span>{{ item.text }}</span>
        </div>
      </section>

      <section id="capabilities" class="uclaw-section">
        <div class="uclaw-section-head">
          <span>01</span>
          <div>
            <h2>它能帮你做什么</h2>
            <p>不是单点工具，而是一套可以直接出活儿的 AI 办公与创作流程。</p>
          </div>
        </div>

        <div class="uclaw-capability-grid">
          <article v-for="group in capabilityGroups" :key="group.title" class="uclaw-capability">
            <span>{{ group.tag }}</span>
            <h3>{{ group.title }}</h3>
            <ul>
              <li v-for="item in group.items" :key="item">{{ item }}</li>
            </ul>
          </article>
        </div>
      </section>

      <section class="uclaw-section uclaw-why">
        <div class="uclaw-section-head">
          <span>02</span>
          <div>
            <h2>为什么值得购买</h2>
            <p>UClaw 的价值不在于多一个聊天窗口，而是把高频工作压缩成可随身携带的完整工作台。</p>
          </div>
        </div>

        <div class="uclaw-reason-list">
          <article v-for="reason in reasons" :key="reason.title" class="uclaw-reason">
            <span>{{ reason.index }}</span>
            <div>
              <h3>{{ reason.title }}</h3>
              <p>{{ reason.text }}</p>
            </div>
          </article>
        </div>
      </section>

      <section class="uclaw-section">
        <div class="uclaw-section-head">
          <span>03</span>
          <div>
            <h2>适合这些场景</h2>
            <p>个人、团队、销售和 IT 都能用同一套便携形态，把工具带到现场。</p>
          </div>
        </div>

        <div class="uclaw-audience-grid">
          <article v-for="audience in audiences" :key="audience.title">
            <strong>{{ audience.title }}</strong>
            <p>{{ audience.text }}</p>
          </article>
        </div>
      </section>

      <section class="uclaw-cta">
        <div>
          <p>需要一个免安装、可离线、数据留在自己手里的 AI 工具箱？</p>
          <h2>购买 UClaw，联系客服获取激活与交付方式。</h2>
        </div>
        <button type="button" class="uclaw-primary" @click="showContact = true">
          <span>购买联系客服</span>
          <Icon name="chat" size="md" />
        </button>
      </section>
    </main>

    <Transition name="uclaw-modal">
      <div v-if="showContact" class="uclaw-modal" @click.self="showContact = false">
        <section class="uclaw-modal-panel" role="dialog" aria-modal="true" aria-labelledby="uclaw-contact-title">
          <button type="button" class="uclaw-modal-close" aria-label="关闭" @click="showContact = false">
            <Icon name="x" size="md" />
          </button>
          <p class="uclaw-eyebrow">客服微信</p>
          <h2 id="uclaw-contact-title">扫码添加客服</h2>
          <img src="/uclaw-wechat.jpg" alt="客服微信二维码" />
          <p>请使用微信扫一扫，添加客服咨询购买、激活码和交付方式。</p>
        </section>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore, useAuthStore } from '@/stores'

const appStore = useAppStore()
const authStore = useAuthStore()

const showContact = ref(false)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')

const heroSignals = [
  { name: '安装', value: '0' },
  { name: '数据', value: '本地' },
  { name: '流程', value: '20+' },
]

const highlights = [
  { title: '插上就用', text: '不用装软件，不改系统，用完拔掉。' },
  { title: '视频很强', text: '自动剪辑、字幕、调色、脚本到成片。' },
  { title: '办公全套', text: 'PPT、PDF、Excel、Word 一套带走。' },
  { title: '模型自由', text: 'OpenAI 或国内大模型都可接入。' },
]

const capabilityGroups = [
  {
    tag: 'VIDEO',
    title: '视频创作',
    items: ['自动剪掉口水词', '自动字幕与调色', '文案到配音再到成片', '口播、教程、产品演示快速产出'],
  },
  {
    tag: 'OFFICE',
    title: '办公文档',
    items: ['网页翻页 PPT', 'PDF 转换、OCR、合并拆分', 'Excel 清洗、合并、报表', 'Word 写作、排版、修订'],
  },
  {
    tag: 'CREATE',
    title: '内容设计',
    items: ['文章、网页、PDF 生成长图文', '页面、看板、原型和动画', 'AI 生图并保存本地', '越用越懂你的自我记忆'],
  },
  {
    tag: 'SEARCH',
    title: '信息检索',
    items: ['17 个搜索引擎一框搜', '国内国外信息源覆盖', '免 Key 搜索体验', '中国城市天气与生活指数'],
  },
]

const reasons = [
  {
    index: '01',
    title: '免安装，适合任何临时电脑',
    text: '公司电脑、客户现场、借来的笔记本都能用；不写注册表，不往系统里塞东西。',
  },
  {
    index: '02',
    title: '把复杂视频流程做成一键工作流',
    text: '剪辑、字幕、调色、脚本、配音、成片都能串起来，减少反复切换工具的时间。',
  },
  {
    index: '03',
    title: '隐私和数据留在自己手里',
    text: '文档、视频、配置和记录都在 U 盘里，断网、内网和敏感办公环境也更稳妥。',
  },
  {
    index: '04',
    title: '不绑定单一模型和订阅',
    text: '填入模型网址、密钥和模型名即可接入，价格和供应商由你自己控制。',
  },
]

const audiences = [
  { title: '自媒体 / 创作者', text: '随身完成视频、文案、PPT 和分享物料。' },
  { title: '办公族 / 白领', text: '把 PDF、Excel、Word 的重复处理交给 AI。' },
  { title: '企业 / 团队', text: '免安装、可离线、数据不外泄，适合内网和涉密环境。' },
  { title: '销售 / 渠道', text: '作为可交付成品分发给客户，带授权管理。' },
  { title: 'IT / 运维', text: '给非技术同事一个低门槛的 AI 工作入口。' },
]
</script>

<style scoped>
.uclaw-shell {
  min-height: 100vh;
  overflow-x: hidden;
  background:
    linear-gradient(color-mix(in srgb, var(--app-ink) 3%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--app-ink) 3%, transparent) 1px, transparent 1px),
    var(--app-bg);
  background-size: 76px 76px;
  color: var(--app-ink);
}

.uclaw-topbar {
  position: sticky;
  top: 0;
  z-index: 30;
  display: flex;
  min-height: 72px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  border-bottom: 1px solid var(--app-line);
  background: color-mix(in srgb, var(--app-bg) 90%, transparent);
  padding: 0 max(24px, calc((100vw - 1180px) / 2));
  backdrop-filter: blur(16px);
}

.uclaw-brand,
.uclaw-nav,
.uclaw-actions,
.uclaw-primary,
.uclaw-secondary {
  display: inline-flex;
  align-items: center;
}

.uclaw-brand {
  min-width: 0;
  gap: 12px;
  text-decoration: none;
}

.uclaw-brand-mark {
  display: inline-flex;
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--app-ink);
  border-radius: 8px;
  background: var(--app-surface);
}

.uclaw-brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.uclaw-brand strong,
.uclaw-brand small {
  display: block;
}

.uclaw-brand strong {
  font-size: 16px;
  line-height: 1.1;
}

.uclaw-brand small {
  margin-top: 3px;
  color: var(--app-soft);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.uclaw-nav {
  gap: 10px;
}

.uclaw-nav a,
.uclaw-secondary {
  min-height: 38px;
  border: 1px solid var(--app-line);
  border-radius: 999px;
  background: var(--app-surface);
  padding: 0 14px;
  color: var(--app-ink);
  font-size: 13px;
  font-weight: 800;
  text-decoration: none;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease;
}

.uclaw-nav a:hover,
.uclaw-secondary:hover {
  border-color: var(--app-ink);
  transform: translateY(-1px);
}

main {
  width: min(1180px, calc(100% - 48px));
  margin: 0 auto;
}

.uclaw-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(360px, 0.86fr);
  gap: 70px;
  align-items: center;
  padding: 96px 0 78px;
  border-bottom: 1px solid var(--app-line);
}

.uclaw-eyebrow {
  margin: 0 0 16px;
  color: #0f766e;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0;
  text-transform: uppercase;
}

.uclaw-hero h1 {
  margin: 0;
  font-family: var(--app-font-display);
  font-size: clamp(68px, 12vw, 142px);
  font-weight: 900;
  line-height: 0.9;
  letter-spacing: 0;
}

.uclaw-hero-lead {
  max-width: 650px;
  margin: 28px 0 0;
  color: var(--app-muted);
  font-size: 20px;
  font-weight: 650;
  line-height: 1.8;
}

.uclaw-actions {
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 34px;
}

.uclaw-primary {
  min-height: 48px;
  justify-content: center;
  gap: 10px;
  border: 1px solid var(--app-ink);
  border-radius: 999px;
  background: var(--app-ink);
  box-shadow: 8px 8px 0 color-mix(in srgb, #0f766e 60%, var(--app-line));
  color: var(--app-bg);
  padding: 0 22px;
  font-size: 14px;
  font-weight: 900;
  transition:
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.uclaw-primary:hover {
  box-shadow: 6px 6px 0 color-mix(in srgb, #b8413a 55%, var(--app-line));
  transform: translateY(-2px);
}

.uclaw-secondary {
  min-height: 48px;
  justify-content: center;
  padding: 0 22px;
}

.uclaw-device {
  border: 1px solid var(--app-ink);
  border-radius: 8px;
  background: var(--app-surface);
  box-shadow: 14px 14px 0 var(--app-ink);
  overflow: hidden;
}

.uclaw-device-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid var(--app-line);
  padding: 16px 18px;
}

.uclaw-device-head span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #b8413a;
  box-shadow:
    20px 0 0 #bc8a35,
    40px 0 0 #0f766e;
}

.uclaw-device-head strong {
  font-size: 16px;
}

.uclaw-device-head em {
  border: 1px solid #0f766e;
  border-radius: 999px;
  color: #0f766e;
  padding: 4px 9px;
  font-size: 11px;
  font-style: normal;
  font-weight: 900;
}

.uclaw-usb {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  align-items: center;
  gap: 0;
  padding: 44px 30px 30px;
}

.uclaw-usb-metal {
  height: 72px;
  border: 1px solid var(--app-ink);
  border-right: 0;
  border-radius: 8px 0 0 8px;
  background:
    linear-gradient(90deg, transparent 18px, var(--app-line) 18px 20px, transparent 20px),
    color-mix(in srgb, var(--app-surface-muted) 80%, #fff);
}

.uclaw-usb-body {
  display: flex;
  min-height: 118px;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  border: 1px solid var(--app-ink);
  border-radius: 8px;
  background: #0f766e;
  color: #fffdf8;
  padding: 0 24px;
}

.uclaw-usb-body span {
  display: inline-flex;
  width: 54px;
  height: 54px;
  align-items: center;
  justify-content: center;
  border: 1px solid #fffdf8;
  border-radius: 8px;
  font-weight: 900;
}

.uclaw-usb-body strong {
  min-width: 0;
  font-size: 24px;
  text-align: right;
}

.uclaw-console {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 1px solid var(--app-line);
}

.uclaw-console div {
  min-width: 0;
  padding: 22px;
  border-right: 1px solid var(--app-line);
}

.uclaw-console div:last-child {
  border-right: 0;
}

.uclaw-console span,
.uclaw-console strong {
  display: block;
}

.uclaw-console span {
  color: var(--app-soft);
  font-size: 12px;
  font-weight: 800;
}

.uclaw-console strong {
  margin-top: 8px;
  font-size: 28px;
}

.uclaw-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-bottom: 1px solid var(--app-line);
}

.uclaw-strip div {
  min-width: 0;
  border-right: 1px solid var(--app-line);
  padding: 28px 22px;
}

.uclaw-strip div:last-child {
  border-right: 0;
}

.uclaw-strip strong,
.uclaw-strip span {
  display: block;
}

.uclaw-strip strong {
  font-size: 18px;
}

.uclaw-strip span {
  margin-top: 8px;
  color: var(--app-muted);
  font-size: 14px;
  line-height: 1.7;
}

.uclaw-section {
  padding: 82px 0;
  border-bottom: 1px solid var(--app-line);
}

.uclaw-section-head {
  display: grid;
  grid-template-columns: 70px minmax(0, 1fr);
  gap: 24px;
  margin-bottom: 34px;
}

.uclaw-section-head > span {
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

.uclaw-section h2 {
  margin: 0;
  font-size: clamp(32px, 5vw, 50px);
  line-height: 1.15;
  letter-spacing: 0;
}

.uclaw-section-head p {
  max-width: 680px;
  margin: 12px 0 0;
  color: var(--app-muted);
  font-size: 16px;
  line-height: 1.8;
}

.uclaw-capability-grid,
.uclaw-audience-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.uclaw-capability,
.uclaw-audience-grid article {
  min-width: 0;
  border: 1px solid var(--app-line);
  border-radius: 8px;
  background: var(--app-surface);
  padding: 20px;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.uclaw-capability:hover,
.uclaw-audience-grid article:hover {
  border-color: var(--app-ink);
  box-shadow: 8px 8px 0 var(--app-line);
  transform: translateY(-4px);
}

.uclaw-capability span {
  display: inline-flex;
  border: 1px solid #0f766e;
  border-radius: 999px;
  color: #0f766e;
  padding: 4px 9px;
  font-size: 11px;
  font-weight: 900;
}

.uclaw-capability h3 {
  margin: 18px 0 14px;
  font-size: 22px;
}

.uclaw-capability ul {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  color: var(--app-muted);
  font-size: 14px;
  line-height: 1.65;
  list-style: none;
}

.uclaw-capability li {
  position: relative;
  padding-left: 16px;
}

.uclaw-capability li::before {
  content: '';
  position: absolute;
  top: 0.75em;
  left: 0;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #b8413a;
}

.uclaw-reason-list {
  border-top: 1px solid var(--app-line);
}

.uclaw-reason {
  display: grid;
  grid-template-columns: 70px minmax(0, 1fr);
  gap: 24px;
  align-items: start;
  min-height: 138px;
  border-bottom: 1px solid var(--app-line);
  padding: 26px 0;
}

.uclaw-reason:last-child {
  border-bottom: 0;
}

.uclaw-reason > span {
  color: #b8413a;
  font-family: var(--app-font-mono);
  font-size: 13px;
  font-weight: 900;
}

.uclaw-reason h3 {
  margin: 0;
  font-size: 24px;
}

.uclaw-reason p,
.uclaw-audience-grid p {
  margin: 10px 0 0;
  color: var(--app-muted);
  font-size: 15px;
  line-height: 1.8;
}

.uclaw-audience-grid strong {
  display: block;
  font-size: 18px;
}

.uclaw-cta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 56px 0 72px;
}

.uclaw-cta p {
  margin: 0 0 12px;
  color: #0f766e;
  font-size: 13px;
  font-weight: 900;
}

.uclaw-cta h2 {
  max-width: 780px;
  margin: 0;
  font-size: clamp(28px, 4vw, 44px);
  line-height: 1.25;
}

.uclaw-modal {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, #000 58%, transparent);
  padding: 20px;
}

.uclaw-modal-panel {
  position: relative;
  width: min(100%, 430px);
  border: 1px solid var(--app-ink);
  border-radius: 8px;
  background: var(--app-surface);
  box-shadow: 12px 12px 0 var(--app-ink);
  padding: 26px;
  text-align: center;
}

.uclaw-modal-close {
  position: absolute;
  top: 14px;
  right: 14px;
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--app-line);
  border-radius: 8px;
  color: var(--app-muted);
}

.uclaw-modal-close:hover {
  border-color: var(--app-ink);
  color: var(--app-ink);
}

.uclaw-modal-panel h2 {
  margin: 0 0 18px;
  font-size: 28px;
}

.uclaw-modal-panel img {
  width: min(100%, 320px);
  border: 1px solid var(--app-line);
  border-radius: 8px;
  background: #fff;
}

.uclaw-modal-panel p:last-child {
  margin: 18px auto 0;
  max-width: 320px;
  color: var(--app-muted);
  font-size: 14px;
  line-height: 1.7;
}

.uclaw-modal-enter-active,
.uclaw-modal-leave-active {
  transition: opacity 0.2s ease;
}

.uclaw-modal-enter-from,
.uclaw-modal-leave-to {
  opacity: 0;
}

@media (max-width: 980px) {
  .uclaw-topbar {
    align-items: flex-start;
    flex-direction: column;
    padding: 16px 24px;
  }

  .uclaw-nav {
    width: 100%;
    flex-wrap: wrap;
  }

  main {
    width: min(100% - 32px, 720px);
  }

  .uclaw-hero {
    grid-template-columns: minmax(0, 1fr);
    gap: 44px;
    padding: 72px 0 64px;
  }

  .uclaw-strip,
  .uclaw-capability-grid,
  .uclaw-audience-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .uclaw-strip,
  .uclaw-capability-grid,
  .uclaw-audience-grid,
  .uclaw-console {
    grid-template-columns: minmax(0, 1fr);
  }

  .uclaw-strip div,
  .uclaw-console div {
    border-right: 0;
    border-bottom: 1px solid var(--app-line);
  }

  .uclaw-strip div:last-child,
  .uclaw-console div:last-child {
    border-bottom: 0;
  }

  .uclaw-section-head,
  .uclaw-reason {
    grid-template-columns: minmax(0, 1fr);
    gap: 16px;
  }

  .uclaw-usb {
    grid-template-columns: minmax(0, 1fr);
    padding: 30px 18px 22px;
  }

  .uclaw-usb-metal {
    display: none;
  }

  .uclaw-usb-body {
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
  }

  .uclaw-usb-body strong {
    text-align: left;
  }

  .uclaw-cta {
    align-items: flex-start;
    flex-direction: column;
  }

  .uclaw-primary,
  .uclaw-secondary {
    width: 100%;
  }
}
</style>
