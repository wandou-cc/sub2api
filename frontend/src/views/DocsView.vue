<template>
  <div class="docs-shell">
    <header class="docs-topbar">
      <router-link to="/home" class="docs-brand">
        <span class="docs-logo">
          <img :src="siteLogo || '/logo.png'" :alt="siteName" />
        </span>
        <span>{{ siteName }}</span>
      </router-link>

      <nav class="docs-topbar-actions">
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="docs-topbar-link">
          {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
        </router-link>
      </nav>
    </header>

    <div class="docs-layout">
      <aside class="docs-sidebar">
        <button
          type="button"
          class="docs-nav-item"
          :class="{ 'docs-nav-item-active': activeSection.id === 'overview' }"
          @click="showOverview"
        >
          <Icon name="book" size="sm" />
          <span>文档总览</span>
        </button>

        <div v-for="section in docsSections" :key="section.id" class="docs-nav-section">
          <button
            type="button"
            class="docs-nav-item"
            :class="{ 'docs-nav-item-active': activeSection.id === section.id && !activePage }"
            @click="openSection(section.id)"
          >
            <Icon :name="section.icon" size="sm" />
            <span>{{ section.title }}</span>
            <Icon
              name="chevronDown"
              size="xs"
              class="ml-auto transition-transform"
              :class="{ 'rotate-180': expandedSections.has(section.id) }"
            />
          </button>

          <div v-if="expandedSections.has(section.id)" class="docs-nav-pages">
            <button
              v-for="page in section.pages"
              :key="page.id"
              type="button"
              class="docs-nav-page"
              :class="{ 'docs-nav-page-active': activePage?.id === page.id }"
              @click="openPage(section.id, page.id)"
            >
              {{ page.title }}
            </button>
          </div>
        </div>
      </aside>

      <main class="docs-content">
        <div v-if="!activePage" class="docs-hero">
          <p class="docs-eyebrow">codeingforce Docs</p>
          <h1>{{ activeSection.title }}</h1>
          <p>{{ activeSection.description }}</p>
        </div>

        <div v-if="!activePage && activeSection.id === 'overview'" class="docs-cards">
          <button
            v-for="section in docsSections"
            :key="section.id"
            type="button"
            class="docs-card"
            @click="openSection(section.id)"
          >
            <span class="docs-card-icon">
              <Icon :name="section.icon" size="lg" />
            </span>
            <span class="docs-card-title">{{ section.title }}</span>
            <span class="docs-card-description">{{ section.description }}</span>
            <span class="docs-card-pages">
              <span v-for="page in section.pages.slice(0, 3)" :key="page.id">{{ page.title }}</span>
            </span>
          </button>
        </div>

        <div v-else-if="!activePage" class="docs-page-list">
          <button
            v-for="page in activeSection.pages"
            :key="page.id"
            type="button"
            class="docs-page-card"
            @click="openPage(activeSection.id, page.id)"
          >
            <span>
              <strong>{{ page.title }}</strong>
              <small>{{ page.summary }}</small>
            </span>
            <Icon name="chevronRight" size="sm" />
          </button>
        </div>

        <article v-else class="docs-article">
          <nav class="docs-breadcrumb">
            <button type="button" @click="showOverview">文档总览</button>
            <span>/</span>
            <button type="button" @click="openSection(activeSection.id)">{{ activeSection.title }}</button>
            <span>/</span>
            <span>{{ activePage.title }}</span>
          </nav>

          <header class="docs-article-header">
            <p class="docs-eyebrow">{{ activeSection.title }}</p>
            <h1>{{ activePage.title }}</h1>
            <p>{{ activePage.summary }}</p>
          </header>

          <div class="docs-prose">
            <p class="docs-lead">{{ activePage.lead }}</p>

            <section v-for="block in activePage.blocks" :key="block.heading">
              <h2>{{ block.heading }}</h2>
              <p v-for="paragraph in block.paragraphs" :key="paragraph">{{ paragraph }}</p>
              <ul v-if="block.items">
                <li v-for="item in block.items" :key="item">{{ item }}</li>
              </ul>
              <pre v-if="block.code"><code>{{ resolveCode(block.code) }}</code></pre>
              <p v-if="block.note" class="docs-tip">{{ block.note }}</p>
            </section>
          </div>
        </article>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

type DocsIcon = 'book' | 'terminal' | 'cube' | 'globe'

interface DocsBlock {
  heading: string
  paragraphs?: string[]
  items?: string[]
  code?: keyof typeof codeExamples
  note?: string
}

interface DocsPage {
  id: string
  title: string
  summary: string
  lead: string
  blocks: DocsBlock[]
}

interface DocsSection {
  id: string
  title: string
  description: string
  icon: DocsIcon
  pages: DocsPage[]
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const activeSectionId = ref('overview')
const activePageId = ref('')
const expandedSections = ref(new Set<string>(['about-us', 'deploy']))

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'codeingforce')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const baseUrl = computed(() => {
  const configured = appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl
  return configured || window.location.origin
})
const apiRootUrl = computed(() => baseUrl.value.replace(/\/v1\/?$/, '').replace(/\/+$/, ''))
const openAIBaseUrl = computed(() => `${apiRootUrl.value}/v1`)

const codeExamples = {
  claudeInstall: () => `npm install -g @anthropic-ai/claude-code
claude --version`,
  ccSwitchInstall: () => `npm install -g cc-switch
cc --version`,
  ccSwitchAdd: () => `cc add codeingforce \\
  --base-url ${baseUrl.value} \\
  --token YOUR_API_KEY

cc use codeingforce
cc list`,
  claudeEnv: () => `export ANTHROPIC_BASE_URL="${baseUrl.value}"
export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"
claude`,
  codexInstall: () => `npm install -g @openai/codex
codex --version`,
  codexConfig: () => `# ~/.codex/config.toml
model_provider = "codeingforce"
model = "gpt-5"

[model_providers.codeingforce]
name = "codeingforce"
base_url = "${openAIBaseUrl.value}"
wire_api = "responses"
env_key = "CODEINGFORCE_API_KEY"`,
  codexEnv: () => `export CODEINGFORCE_API_KEY="YOUR_API_KEY"
codex "检查当前项目的测试风险"`,
  geminiInstall: () => `npm install -g @google/gemini-cli
gemini --version`,
  geminiEnv: () => `export GOOGLE_GEMINI_BASE_URL="${baseUrl.value}"
export GEMINI_API_KEY="YOUR_API_KEY"
gemini "用一句话说明这个项目的作用"`,
  pythonAnthropic: () => `from anthropic import Anthropic

client = Anthropic(
    base_url="${baseUrl.value}",
    api_key="YOUR_API_KEY",
)

msg = client.messages.create(
    model="claude-sonnet-4-5",
    max_tokens=1024,
    messages=[{"role": "user", "content": "你好"}],
)
print(msg.content[0].text)`,
  pythonOpenAI: () => `from openai import OpenAI

client = OpenAI(
    base_url="${openAIBaseUrl.value}",
    api_key="YOUR_API_KEY",
)

resp = client.chat.completions.create(
    model="gpt-5",
    messages=[{"role": "user", "content": "你好"}],
)
print(resp.choices[0].message.content)`,
  nodeAnthropic: () => `import Anthropic from "@anthropic-ai/sdk"

const client = new Anthropic({
  baseURL: "${baseUrl.value}",
  apiKey: "YOUR_API_KEY",
})

const msg = await client.messages.create({
  model: "claude-sonnet-4-5",
  max_tokens: 1024,
  messages: [{ role: "user", content: "你好" }],
})
console.log(msg.content[0].text)`,
  curlOpenAI: () => `curl ${openAIBaseUrl.value}/chat/completions \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5",
    "messages": [
      { "role": "user", "content": "hi" }
    ]
  }'`,
  openclawInstall: () => `npm install -g openclaw@latest
openclaw --version`,
  openclawOnboard: () => `openclaw onboard \\
  --provider openai-compatible \\
  --base-url ${openAIBaseUrl.value} \\
  --api-key YOUR_API_KEY`,
  opencodeInstall: () => `curl -fsSL https://opencode.ai/install | bash
# 或
npm install -g opencode-ai`,
  opencodeConfig: () => `{
  "provider": {
    "codeingforce": {
      "name": "codeingforce",
      "type": "openai",
      "options": {
        "baseURL": "${openAIBaseUrl.value}",
        "apiKey": "YOUR_API_KEY"
      },
      "models": {
        "gpt-5": { "name": "GPT-5" },
        "claude-sonnet-4-5": { "name": "Claude Sonnet 4.5" }
      }
    }
  }
}`,
  clineConfig: () => `API Provider: OpenAI Compatible
Base URL: ${openAIBaseUrl.value}
API Key: YOUR_API_KEY
Model ID: gpt-5`,
  windowsVerify: () => `node -v
npm -v`,
  windowsRegistry: () => `npm config set registry https://registry.npmmirror.com`,
  macosNvm: () => `brew install nvm
mkdir -p ~/.nvm
echo 'export NVM_DIR="$HOME/.nvm"' >> ~/.zshrc
echo '[ -s "$(brew --prefix)/opt/nvm/nvm.sh" ] && . "$(brew --prefix)/opt/nvm/nvm.sh"' >> ~/.zshrc
source ~/.zshrc`,
  macosNode: () => `nvm install --lts
nvm use --lts
nvm alias default 'lts/*'
node -v
npm -v`,
  networkHealth: () => `curl -I ${baseUrl.value.replace(/\/$/, '')}/health`,
  unsetProxy: () => `# macOS / Linux
unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy

# Windows PowerShell
$env:HTTP_PROXY=$null
$env:HTTPS_PROXY=$null`,
}

const docsSections = computed<DocsSection[]>(() => [
  {
    id: 'about-us',
    title: '了解我们',
    description: '项目定位、隐私边界和真实可验证的网关行为说明。',
    icon: 'book',
    pages: [
      {
        id: 'about-us-overview',
        title: '为什么这是一个真实中转',
        summary: '用请求链路和后台能力说明 codeingforce 做了什么。',
        lead: 'codeingforce 的核心是把 OpenAI、Anthropic、Gemini 等协议统一到一个网关里，再通过后台账号、分组、模型映射、计费和监控能力转发到上游。',
        blocks: [
          {
            heading: '请求链路',
            paragraphs: [
              '客户端把请求发到当前站点的兼容接口，网关根据 API Key 找到用户、分组、可用渠道和上游账号，再把请求转给对应平台。',
              '调用完成后，系统根据上游返回的 usage、模型配置和分组倍率写入使用日志，并更新余额、订阅额度或 Key 限额。',
            ],
            items: [
              'OpenAI 兼容路径：/v1/chat/completions、/v1/responses 等。',
              'Anthropic 兼容路径：/v1/messages。',
              'Gemini 兼容路径：/v1beta/models/{model}:generateContent。',
            ],
          },
          {
            heading: '后台可验证能力',
            paragraphs: [
              '管理员可以配置账号、渠道、模型映射、分组倍率、渠道监控、用量聚合和错误监控。用户侧可以查看自己的 Key、用量、可用渠道和渠道状态。',
            ],
            items: [
              'Key 绑定分组，分组决定可用模型和计费规则。',
              '用量记录展示请求模型、Token、费用、耗时和渠道相关信息。',
              '渠道状态用于判断某个上游账号或模型是否健康。',
            ],
          },
          {
            heading: '不夸大承诺',
            paragraphs: [
              '文档里的示例只说明本项目的接入方式。上游模型是否可用、价格、速率和上下文能力，以你部署后台的账号、渠道和模型配置为准。',
            ],
          },
        ],
      },
      {
        id: 'about-us-privacy',
        title: '隐私边界：系统存什么',
        summary: '按当前代码实现说明使用日志、错误日志和 API Key 的保存边界。',
        lead: 'codeingforce 当前会保存必要的鉴权、计费和排障数据。这里按代码里的 schema 说明，不写超出实现的隐私承诺。',
        blocks: [
          {
            heading: '使用日志保存的字段',
            paragraphs: [
              '每次成功或进入计费流程的请求会写入使用日志。日志用于账单、限额、统计和问题排查。',
            ],
            items: [
              '用户、API Key、上游账号、分组、渠道和请求 ID。',
              '请求模型、实际上游模型、模型映射链和计费层级。',
              '输入/输出/缓存 Token、图片数量和图片尺寸元数据。',
              '费用、倍率、计费模式、流式标记、总耗时、首 token 耗时。',
              'User-Agent、IP 地址和创建时间。',
            ],
          },
          {
            heading: '错误日志保存的字段',
            paragraphs: [
              '如果部署启用了运维错误监控，系统会保存上游状态码、错误阶段、错误类型和经过脱敏、截断后的错误正文或错误详情。',
              '错误正文用于排查上游报错，不应把敏感业务内容主动放进报错信息里。',
            ],
          },
          {
            heading: '不会写入使用日志的内容',
            paragraphs: [
              '当前使用日志表没有 request body 或 response body 字段，常规用量记录不保存完整请求正文和完整响应正文。',
              'API Key 当前以 key 字段保存完整值用于鉴权唯一性，不是只保存哈希。生产部署应限制数据库访问权限，并避免把 Key 写进公开仓库、前端代码或聊天记录。',
            ],
          },
        ],
      },
    ],
  },
  {
    id: 'deploy',
    title: '接入部署',
    description: 'Claude Code、Codex、Gemini CLI 和 SDK 的接入步骤。',
    icon: 'terminal',
    pages: [
      {
        id: 'claude-code',
        title: 'Claude Code CLI 接入',
        summary: '安装 Claude Code，创建 Key，并通过 CC-Switch 或环境变量接入。',
        lead: 'Claude Code 走 Anthropic Messages 协议。你需要创建一个支持 Claude 模型的 API Key，然后把 Anthropic Base URL 指向本站。',
        blocks: [
          { heading: '安装 Claude Code', paragraphs: ['先确认本机已经安装 Node.js。'], code: 'claudeInstall' },
          { heading: '安装 CC-Switch', paragraphs: ['CC-Switch 用于管理 Claude Code 的多个 provider，适合在官方和本站之间切换。'], code: 'ccSwitchInstall' },
          {
            heading: '创建 API Key',
            paragraphs: ['登录控制台，进入 API 密钥页面，新建一个 Key。分组要选择支持 Claude / Anthropic 协议的分组。'],
            items: ['给 Key 起一个能识别用途的名称。', '设置额度或周期限制，避免异常消耗。', '保存后复制完整 Key。'],
          },
          { heading: '用 CC-Switch 添加本站', paragraphs: ['provider 名可以自定义。切换后建议新开终端再运行 Claude Code。'], code: 'ccSwitchAdd' },
          { heading: '环境变量方式', paragraphs: ['如果不想使用切换工具，可以直接在 shell 中设置环境变量。'], code: 'claudeEnv' },
        ],
      },
      {
        id: 'codex',
        title: 'OpenAI Codex CLI 接入',
        summary: '把 Codex 指向 OpenAI 兼容网关，使用支持 GPT 模型的分组。',
        lead: 'Codex CLI 使用 OpenAI 协议。创建 Key 时请选择支持 OpenAI / GPT 模型的分组，然后配置 Codex 的 model provider。',
        blocks: [
          { heading: '安装 Codex', paragraphs: ['Codex 可以通过 npm 全局安装。'], code: 'codexInstall' },
          {
            heading: '配置 model provider',
            paragraphs: ['Codex 读取 ~/.codex/config.toml。下面示例把 provider 指向当前站点的 OpenAI 兼容 /v1 地址。'],
            code: 'codexConfig',
          },
          { heading: '写入环境变量并验证', paragraphs: ['Key 放在环境变量里，不要直接提交到仓库。'], code: 'codexEnv' },
          {
            heading: '模型选择',
            paragraphs: ['model 字段应填写后台分组里真实可用的模型名。若请求报模型不存在，先检查 Key 绑定分组和后台模型映射。'],
          },
        ],
      },
      {
        id: 'gemini-cli',
        title: 'Gemini CLI 接入',
        summary: '安装 Gemini CLI，并把 base URL 与 Key 指向本站 Gemini 兼容接口。',
        lead: 'Gemini CLI 适合长上下文任务。创建 Key 时请选择支持 Gemini 模型的分组。',
        blocks: [
          { heading: '安装 Gemini CLI', paragraphs: ['Gemini CLI 通过 npm 安装。'], code: 'geminiInstall' },
          {
            heading: '配置环境变量',
            paragraphs: ['GOOGLE_GEMINI_BASE_URL 指向当前站点，GEMINI_API_KEY 使用你在控制台创建的 Key。'],
            code: 'geminiEnv',
          },
          {
            heading: '排查方式',
            paragraphs: ['如果 CLI 仍连到官方地址，先检查当前 shell 是否加载了旧配置，再用新终端验证环境变量。'],
          },
        ],
      },
      {
        id: 'sdk-quick',
        title: 'SDK 与 cURL 直调',
        summary: 'Python / Node SDK 与 cURL 的最小接入示例。',
        lead: '大多数官方 SDK 只需要改 base_url 或 baseURL。API Key 使用控制台生成的 Key。',
        blocks: [
          { heading: 'Python Anthropic SDK', code: 'pythonAnthropic' },
          { heading: 'Python OpenAI SDK', code: 'pythonOpenAI' },
          { heading: 'Node.js Anthropic SDK', code: 'nodeAnthropic' },
          { heading: 'cURL OpenAI 兼容接口', code: 'curlOpenAI' },
          {
            heading: '协议路径',
            items: ['Anthropic SDK 使用站点根地址。', 'OpenAI SDK 使用 /v1 地址。', 'Gemini SDK 或 CLI 使用 Gemini 兼容路径和对应环境变量。'],
          },
        ],
      },
    ],
  },
  {
    id: 'tools',
    title: '工具与客户端',
    description: '常见终端和 IDE 客户端的接入配置。',
    icon: 'cube',
    pages: [
      {
        id: 'cc-switch',
        title: 'CC-Switch：Claude Code 配置切换',
        summary: '用本地 provider 名管理多个 Claude Code 端点。',
        lead: 'CC-Switch 适合频繁在官方 Anthropic 和 codeingforce 之间切换的用户。',
        blocks: [
          { heading: '安装', code: 'ccSwitchInstall' },
          { heading: '添加本站 provider', code: 'ccSwitchAdd' },
          {
            heading: '注意事项',
            items: ['切换后新开终端窗口，确保环境变量生效。', 'Key 泄露后先禁用对应 Key，再重新创建。', '如果模型不可用，检查 Key 绑定分组是否支持 Claude 模型。'],
          },
        ],
      },
      {
        id: 'openclaw',
        title: 'OpenClaw：终端 AI 助手',
        summary: '使用 OpenAI Compatible endpoint 接入本站。',
        lead: 'OpenClaw 是终端 AI 助手，适合通过 OpenAI 兼容网关配置自定义 provider。',
        blocks: [
          { heading: '安装', paragraphs: ['OpenClaw 通常要求较新的 Node.js，建议使用 Node.js 22 或更新版本。'], code: 'openclawInstall' },
          { heading: '接入本站', paragraphs: ['如果版本支持 onboard，可以按下面方式写入 provider。'], code: 'openclawOnboard' },
          {
            heading: '模型',
            paragraphs: ['主模型和轻量模型都应选择后台分组真实开放的模型。请求失败时先在可用渠道页确认模型是否存在。'],
          },
        ],
      },
      {
        id: 'opencode',
        title: 'OpenCode：终端编程助手',
        summary: '在 OpenCode 配置 OpenAI 兼容 provider。',
        lead: 'OpenCode 支持自定义 OpenAI compatible provider，适合在项目根目录进行终端协作。',
        blocks: [
          { heading: '安装', code: 'opencodeInstall' },
          { heading: '配置 provider', paragraphs: ['编辑 ~/.config/opencode/config.json，没有文件就新建。'], code: 'opencodeConfig' },
          { heading: '运行', paragraphs: ['进入项目根目录后运行 opencode，再通过模型选择命令切到 codeingforce 下的模型。'] },
        ],
      },
      {
        id: 'cline',
        title: 'Cline：VSCode 编程助手扩展',
        summary: '在 VSCode Cline 中选择 OpenAI Compatible。',
        lead: 'Cline 在 VSCode 内运行，可以读取文件和执行命令。接入本站时使用 OpenAI Compatible provider。',
        blocks: [
          {
            heading: '安装扩展',
            paragraphs: ['打开 VSCode 扩展面板，搜索 Cline 并安装。'],
          },
          { heading: '配置 API Provider', code: 'clineConfig' },
          {
            heading: '使用建议',
            paragraphs: ['Cline 会请求确认文件修改和命令执行。给它较明确的任务边界，并为 Key 配置额度限制。'],
          },
        ],
      },
    ],
  },
  {
    id: 'environment',
    title: '环境准备',
    description: 'Node.js 安装、终端验证和网络排查。',
    icon: 'globe',
    pages: [
      {
        id: 'nodejs-windows',
        title: 'Node.js：Windows',
        summary: '使用官方 MSI 安装包，并用 PowerShell 验证。',
        lead: 'Claude Code、Codex、Gemini CLI 等工具通常依赖 Node.js 和 npm。Windows 用户推荐先安装官方 LTS 版本。',
        blocks: [
          {
            heading: '下载安装',
            paragraphs: ['访问 Node.js 官网下载 Windows Installer (.msi)，选择 64-bit LTS 版本。安装时保持 Add to PATH 选项开启。'],
          },
          { heading: '验证', paragraphs: ['安装完成后新开 PowerShell，旧终端不会刷新 PATH。'], code: 'windowsVerify' },
          {
            heading: '多版本和镜像',
            paragraphs: ['需要多版本切换时可以安装 nvm-windows。企业或校园网络 npm 下载慢时再切换镜像源。'],
            code: 'windowsRegistry',
          },
        ],
      },
      {
        id: 'nodejs-macos',
        title: 'Node.js：macOS',
        summary: '使用 Homebrew + nvm 管理 Node.js LTS。',
        lead: 'macOS 推荐用 nvm 管理 Node.js，便于在不同 CLI 需要的版本之间切换。',
        blocks: [
          {
            heading: '安装 Homebrew',
            paragraphs: ['如果还没有 Homebrew，先按 Homebrew 官网命令安装，并根据提示把 brew 加入 PATH。'],
          },
          { heading: '安装 nvm', code: 'macosNvm' },
          { heading: '安装 Node.js LTS', code: 'macosNode' },
          {
            heading: '常见问题',
            items: ['command not found: nvm：重新打开终端或 source ~/.zshrc。', 'which node 不在 ~/.nvm：说明当前 shell 没有使用 nvm 的 Node。', 'npm 下载慢：按需设置 npm registry。'],
          },
        ],
      },
      {
        id: 'network',
        title: '网络与代理',
        summary: '检查站点健康、代理环境和企业网络白名单。',
        lead: '客户端只需要能访问当前 codeingforce 站点。上游网络由服务器侧账号和网关负责。',
        blocks: [
          {
            heading: '检查健康接口',
            paragraphs: ['先从本机终端请求健康检查，确认 DNS、TLS 和站点入口可达。'],
            code: 'networkHealth',
          },
          {
            heading: '代理变量',
            paragraphs: ['如果终端设置了代理，SDK 和 cURL 可能绕路，导致延迟或失败。排查时可以临时清空代理变量。'],
            code: 'unsetProxy',
          },
          {
            heading: '企业或校园网络',
            paragraphs: ['如果网络有白名单策略，把当前站点域名加入允许列表。服务器到上游的连通性在后台渠道状态和日志中排查。'],
          },
        ],
      },
    ],
  },
])

const overviewSection = computed<DocsSection>(() => ({
  id: 'overview',
  title: '接入文档',
  description: `${siteName.value} 的文档菜单已按接入、工具和环境整理。当前 Base URL：${baseUrl.value}`,
  icon: 'book',
  pages: [],
}))

const activeSection = computed(() => docsSections.value.find((section) => section.id === activeSectionId.value) || overviewSection.value)
const activePage = computed(() => activeSection.value.pages.find((page) => page.id === activePageId.value))

watch(
  () => route.query,
  () => {
    const cat = typeof route.query.cat === 'string' ? route.query.cat : ''
    const page = typeof route.query.page === 'string' ? route.query.page : ''
    const section = docsSections.value.find((item) => item.id === cat)
    const selectedPage = section?.pages.find((item) => item.id === page)
    if (section && selectedPage) {
      activeSectionId.value = section.id
      activePageId.value = selectedPage.id
      expandedSections.value.add(section.id)
      return
    }
    if (section) {
      activeSectionId.value = section.id
      activePageId.value = ''
      expandedSections.value.add(section.id)
      return
    }
    activeSectionId.value = 'overview'
    activePageId.value = ''
  },
  { immediate: true },
)

// 返回文档总览并清空查询参数。
function showOverview() {
  activeSectionId.value = 'overview'
  activePageId.value = ''
  router.push({ path: '/docs' })
}

// 打开分组首页，并把 cat 写入 URL 以支持分享。
function openSection(sectionId: string) {
  activeSectionId.value = sectionId
  activePageId.value = ''
  expandedSections.value.add(sectionId)
  router.push({ path: '/docs', query: { cat: sectionId } })
}

// 打开具体文档页，并同步 cat/page 查询参数。
function openPage(sectionId: string, pageId: string) {
  activeSectionId.value = sectionId
  activePageId.value = pageId
  expandedSections.value.add(sectionId)
  router.push({ path: '/docs', query: { cat: sectionId, page: pageId } })
}

// 根据当前站点配置生成示例代码。
function resolveCode(name: keyof typeof codeExamples) {
  return codeExamples[name]()
}
</script>

<style scoped>
.docs-shell {
  min-height: 100vh;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--app-accent) 7%, transparent), transparent 18rem),
    var(--app-bg);
  color: var(--app-ink);
}

.docs-topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  min-height: 4rem;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--app-line);
  background: color-mix(in srgb, var(--app-bg) 94%, transparent);
  padding: 0.75rem 1.5rem;
  backdrop-filter: blur(12px);
}

.docs-brand {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
  color: var(--app-ink);
  font-weight: 750;
  text-decoration: none;
}

.docs-logo {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  flex-shrink: 0;
  overflow: hidden;
  border-radius: 0.5rem;
  background: var(--app-surface);
}

.docs-logo img {
  height: 100%;
  width: 100%;
  object-fit: contain;
}

.docs-topbar-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.docs-topbar-link {
  border-radius: 0.5rem;
  background: var(--app-ink);
  color: var(--app-bg);
  font-size: 0.875rem;
  font-weight: 650;
  padding: 0.55rem 0.9rem;
  text-decoration: none;
}

.docs-layout {
  display: grid;
  grid-template-columns: minmax(220px, 280px) minmax(0, 1fr);
  gap: 2rem;
  align-items: start;
  margin: 0 auto;
  max-width: 1180px;
  padding: 2rem 1.5rem 4rem;
}

.docs-sidebar {
  position: sticky;
  top: 5.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.docs-nav-section {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.docs-nav-item,
.docs-nav-page {
  border: 0;
  background: transparent;
  color: var(--app-muted);
  cursor: pointer;
  text-align: left;
}

.docs-nav-item {
  display: flex;
  min-height: 2.5rem;
  width: 100%;
  align-items: center;
  gap: 0.625rem;
  border-radius: 0.5rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition:
    background-color 0.16s ease,
    color 0.16s ease;
}

.docs-nav-item:hover,
.docs-nav-page:hover {
  background: var(--app-surface-muted);
  color: var(--app-ink);
}

.docs-nav-item-active {
  background: var(--app-ink);
  color: var(--app-bg);
}

.docs-nav-pages {
  margin-left: 1.25rem;
  border-left: 1px solid var(--app-line);
  padding: 0.25rem 0 0.5rem 0.75rem;
}

.docs-nav-page {
  display: block;
  width: 100%;
  border-radius: 0.375rem;
  padding: 0.45rem 0.625rem;
  font-size: 0.8125rem;
  transition:
    background-color 0.16s ease,
    color 0.16s ease;
}

.docs-nav-page-active {
  background: color-mix(in srgb, var(--app-accent) 12%, transparent);
  color: var(--app-accent);
  font-weight: 600;
}

.docs-content {
  min-width: 0;
  padding-bottom: 4rem;
}

.docs-hero {
  margin-bottom: 1.5rem;
  max-width: 48rem;
}

.docs-eyebrow {
  margin: 0 0 0.625rem;
  color: var(--app-accent);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: uppercase;
}

.docs-hero h1,
.docs-article-header h1 {
  margin: 0;
  color: var(--app-ink);
  font-size: 2rem;
  font-weight: 800;
  line-height: 1.15;
}

.docs-hero p:last-child,
.docs-article-header p {
  margin: 0.875rem 0 0;
  color: var(--app-muted);
  font-size: 0.9375rem;
  line-height: 1.7;
}

.docs-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));
  gap: 1rem;
}

.docs-card {
  display: flex;
  min-height: 13rem;
  flex-direction: column;
  gap: 0.75rem;
  border: 1px solid var(--app-line);
  border-radius: 0.5rem;
  background: var(--app-surface);
  padding: 1.25rem;
  color: inherit;
  cursor: pointer;
  text-align: left;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.docs-card:hover {
  border-color: var(--app-accent);
  box-shadow: 0 16px 35px -28px rgba(15, 23, 42, 0.45);
  transform: translateY(-2px);
}

.docs-card-icon {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: var(--app-surface-muted);
  color: var(--app-accent);
}

.docs-card-title {
  color: var(--app-ink);
  font-size: 1.0625rem;
  font-weight: 700;
}

.docs-card-description {
  color: var(--app-muted);
  font-size: 0.875rem;
  line-height: 1.6;
}

.docs-card-pages {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  margin-top: auto;
}

.docs-card-pages span {
  border: 1px solid var(--app-line);
  border-radius: 999px;
  padding: 0.1875rem 0.5rem;
  color: var(--app-muted);
  font-size: 0.75rem;
}

.docs-page-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.docs-page-card {
  display: flex;
  min-height: 5rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid var(--app-line);
  border-radius: 0.5rem;
  background: var(--app-surface);
  color: var(--app-muted);
  cursor: pointer;
  padding: 1rem 1.125rem;
  text-align: left;
  transition:
    border-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.docs-page-card:hover {
  border-color: var(--app-accent);
  color: var(--app-ink);
  transform: translateY(-1px);
}

.docs-page-card strong,
.docs-page-card small {
  display: block;
}

.docs-page-card strong {
  color: var(--app-ink);
  font-size: 0.9375rem;
}

.docs-page-card small {
  margin-top: 0.25rem;
  font-size: 0.8125rem;
  line-height: 1.5;
}

.docs-article {
  border: 1px solid var(--app-line);
  border-radius: 0.5rem;
  background: var(--app-surface);
  padding: 2rem;
}

.docs-breadcrumb {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  color: var(--app-soft);
  font-size: 0.8125rem;
}

.docs-breadcrumb button {
  border: 0;
  background: transparent;
  color: var(--app-muted);
  cursor: pointer;
  padding: 0;
}

.docs-breadcrumb button:hover {
  color: var(--app-ink);
}

.docs-article-header {
  border-bottom: 1px solid var(--app-line);
  margin-bottom: 1.5rem;
  padding-bottom: 1.5rem;
}

.docs-prose {
  color: var(--app-muted);
  font-size: 0.9375rem;
  line-height: 1.75;
}

.docs-prose h2 {
  margin: 2rem 0 0.75rem;
  color: var(--app-ink);
  font-size: 1.125rem;
  font-weight: 750;
}

.docs-prose p {
  margin: 0 0 1rem;
}

.docs-prose ul {
  margin: 0 0 1rem;
  padding-left: 1.25rem;
}

.docs-prose li {
  margin-bottom: 0.375rem;
}

.docs-lead {
  border-left: 3px solid var(--app-accent);
  border-radius: 0 0.5rem 0.5rem 0;
  background: var(--app-surface-muted);
  padding: 0.875rem 1rem;
  color: var(--app-ink);
  font-weight: 600;
}

.docs-tip {
  border-left: 3px solid var(--app-warning);
  border-radius: 0 0.5rem 0.5rem 0;
  background: color-mix(in srgb, var(--app-warning) 12%, transparent);
  padding: 0.75rem 1rem;
  color: var(--app-ink);
}

.docs-prose pre {
  overflow-x: auto;
  border-radius: 0.5rem;
  background: #0f172a;
  color: #e5e7eb;
  font-size: 0.8125rem;
  line-height: 1.65;
  margin: 1rem 0;
  padding: 1rem;
}

.docs-prose code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

@media (max-width: 900px) {
  .docs-topbar {
    padding-inline: 1rem;
  }

  .docs-layout {
    grid-template-columns: 1fr;
    gap: 1.5rem;
    padding: 1.25rem 1rem 3rem;
  }

  .docs-sidebar {
    position: static;
  }

  .docs-article {
    padding: 1.25rem;
  }
}
</style>
