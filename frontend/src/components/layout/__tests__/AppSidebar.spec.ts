import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const routerPath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../router/index.ts')
const routerSource = readFileSync(routerPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
  })
})

describe('AppSidebar models and pricing entry', () => {
  it('is visible without a feature flag or simple-mode restriction', () => {
    const entry = componentSource.match(/\{ path: '\/available-channels'[^\n]+\}/)?.[0]

    expect(entry).toBeDefined()
    expect(entry).not.toContain('featureFlag')
    expect(entry).not.toContain('hideInSimpleMode')
  })
})

describe('AppSidebar activity entries', () => {
  it('registers the blind-box menu and Token Arena label keys', () => {
    expect(componentSource).toContain("{ path: '/lottery', label: t('nav.rechargeLottery')")
    expect(componentSource).toContain("{ path: '/speed-rank', label: t('nav.speedRank')")
  })

  it('registers a standalone carpool menu and route', () => {
    expect(componentSource).toContain("{ path: '/carpool', label: t('nav.carpool')")
    const carpoolRoute = routerSource.match(/\{\n    path: '\/carpool',[\s\S]*?\n  \},/)?.[0]

    expect(carpoolRoute).toContain("name: 'CarpoolSubscription'")
    expect(carpoolRoute).toContain("descriptionKey: 'payment.carpool.title'")
  })

  it('gives the blind-box menu its restrained light-red treatment', () => {
    expect(componentSource).toContain("'lottery-menu-link': item.path === '/lottery'")
    expect(componentSource).toContain('.lottery-menu-link,')
    expect(componentSource).toContain('background: transparent;')
    expect(componentSource).toContain('animation: lottery-menu-pulse 3.6s ease-in-out infinite;')
  })
})

describe('AppSidebar contact entry', () => {
  it('registers Contact Us in the Other group', () => {
    expect(componentSource).toContain("{ path: '/contact', label: t('nav.contactUs'), icon: ContactIcon }")
  })
})

describe('AppSidebar carpool management entry', () => {
  it('registers carpool management under payment orders', () => {
    expect(componentSource).toContain("{ path: '/admin/orders/carpools', label: t('nav.carpoolManagement')")
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})
