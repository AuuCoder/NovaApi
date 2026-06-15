<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page (Swiss editorial style) -->
  <div v-else class="swiss-root" :class="{ 'is-dark': isDark }">
    <!-- Header -->
    <header class="swiss-top">
      <div class="swiss-top__inner">
        <div class="swiss-mark">
          <img v-if="siteLogo" :src="siteLogo" alt="Logo" class="swiss-mark__logo" />
          <span v-else class="swiss-mark__dot" aria-hidden="true"></span>
          <span class="swiss-mark__name">{{ siteName }}</span>
        </div>
        <nav class="swiss-actions">
          <span class="swiss-meta">{{ editionLabel }}</span>
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="swiss-cta swiss-cta--ghost"
          >
            {{ t('home.docs') }}
          </a>
          <button
            @click="toggleTheme"
            class="swiss-link"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link v-if="isAuthenticated" :to="dashboardPath" class="swiss-cta">
            <span class="swiss-cta__avatar">{{ userInitial }}</span>
            {{ t('home.dashboard') }}
          </router-link>
          <router-link v-else to="/login" class="swiss-cta swiss-cta--ghost">
            {{ t('home.login') }}
          </router-link>
        </nav>
      </div>
    </header>

    <main class="swiss-main">
      <!-- Hero -->
      <section class="swiss-section swiss-hero">
        <div class="swiss-rule">
          <span class="swiss-rule__num">01</span>
          <span class="swiss-rule__line" aria-hidden="true"></span>
          <span class="swiss-rule__label">{{ t('home.sections.index') }}</span>
        </div>
        <h1 class="swiss-display">
          <span class="swiss-display__row swiss-display__row--lead">
            <span class="swiss-display__word">ONE</span>
            <span class="swiss-display__rule" aria-hidden="true"></span>
            <span class="swiss-display__word">ONE</span>
          </span>
          <span class="swiss-display__row swiss-display__row--trail">
            <span class="swiss-display__word swiss-display__word--accent">KEY.</span>
            <span class="swiss-display__word swiss-display__word--accent">WAY IN.</span>
          </span>
        </h1>
        <div class="swiss-hero__bottom">
          <p class="swiss-lede">{{ t('home.heroDescription') }}</p>
          <div class="swiss-hero__cta">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="swiss-button"
            >
              <span class="swiss-button__index">→</span>
              <span class="swiss-button__label">{{
                isAuthenticated ? t('home.goToDashboard') : t('home.getStarted')
              }}</span>
            </router-link>
            <span class="swiss-meta swiss-meta--inline">{{ t('home.startHint') }}</span>
          </div>
        </div>
      </section>

      <!-- Pillars -->
      <section class="swiss-section swiss-pillars">
        <div class="swiss-rule">
          <span class="swiss-rule__num">02</span>
          <span class="swiss-rule__line" aria-hidden="true"></span>
          <span class="swiss-rule__label">{{ t('home.sections.rules') }}</span>
        </div>
        <div class="swiss-pillars__grid">
          <article class="swiss-pillar">
            <div class="swiss-pillar__index">01</div>
            <h3 class="swiss-pillar__title">{{ t('home.features.unifiedGateway') }}</h3>
            <p class="swiss-pillar__body">{{ t('home.features.unifiedGatewayDesc') }}</p>
          </article>
          <article class="swiss-pillar">
            <div class="swiss-pillar__index">02</div>
            <h3 class="swiss-pillar__title">{{ t('home.features.multiAccount') }}</h3>
            <p class="swiss-pillar__body">{{ t('home.features.multiAccountDesc') }}</p>
          </article>
          <article class="swiss-pillar">
            <div class="swiss-pillar__index">03</div>
            <h3 class="swiss-pillar__title">{{ t('home.features.balanceQuota') }}</h3>
            <p class="swiss-pillar__body">{{ t('home.features.balanceQuotaDesc') }}</p>
          </article>
        </div>
      </section>

      <!-- Coverage -->
      <section class="swiss-section swiss-coverage">
        <div class="swiss-rule">
          <span class="swiss-rule__num">03</span>
          <span class="swiss-rule__line" aria-hidden="true"></span>
          <span class="swiss-rule__label">{{ t('home.sections.coverage') }}</span>
        </div>
        <ol class="swiss-coverage__list">
          <li
            v-for="(item, i) in coverage"
            :key="item.name"
            class="swiss-coverage__row"
          >
            <span class="swiss-coverage__num">{{ String(i + 1).padStart(2, '0') }}</span>
            <span class="swiss-coverage__name">{{ item.name }}</span>
            <span class="swiss-coverage__rule" aria-hidden="true"></span>
            <span class="swiss-coverage__hint">{{ item.hint }}</span>
            <span class="swiss-coverage__status" :data-state="item.state">{{
              item.state === 'live' ? t('home.providers.supported') : t('home.providers.soon')
            }}</span>
          </li>
        </ol>
      </section>

      <!-- Close -->
      <section class="swiss-section swiss-close">
        <div class="swiss-rule">
          <span class="swiss-rule__num">04</span>
          <span class="swiss-rule__line" aria-hidden="true"></span>
          <span class="swiss-rule__label">{{ t('home.sections.start') }}</span>
        </div>
        <p class="swiss-close__pull">{{ t('home.closePull') }}</p>
        <div class="swiss-close__cta">
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="swiss-button swiss-button--solid"
          >
            <span class="swiss-button__label">{{
              isAuthenticated ? t('home.goToDashboard') : t('home.getStarted')
            }}</span>
            <span class="swiss-button__index">→</span>
          </router-link>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="swiss-foot">
      <div class="swiss-foot__inner">
        <span class="swiss-meta">&copy; {{ currentYear }} {{ siteName }}</span>
        <span class="swiss-meta swiss-meta--mute">{{ t('home.footer.allRightsReserved') }}</span>
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="swiss-foot__link"
        >
          {{ t('home.docs') }}
        </a>
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

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'NovaAPI')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
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
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year + edition label (magazine masthead)
const currentYear = computed(() => new Date().getFullYear())
const editionLabel = computed(() => {
  const now = new Date()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  return `Vol. ${mm} / ${now.getFullYear()}`
})

// Coverage matrix
const coverage = computed(() => [
  { name: t('home.providers.claude'), hint: t('home.coverage.claudeHint'), state: 'live' },
  { name: 'GPT', hint: t('home.coverage.gptHint'), state: 'live' },
  { name: t('home.providers.gemini'), hint: t('home.coverage.geminiHint'), state: 'live' },
  { name: t('home.providers.antigravity'), hint: t('home.coverage.antigravityHint'), state: 'live' },
  { name: t('home.providers.more'), hint: t('home.coverage.moreHint'), state: 'soon' },
])

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
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* =========================================================================
   Swiss editorial landing — international typographic style
   light: #f4f3ee paper / #111 ink / #c1392b accent
   dark:  #0e0e0d / #f1efe7 / #ff5a3d
   ========================================================================= */
.swiss-root {
  --bg: #f4f3ee;
  --ink: #111111;
  --mute: #6b6b6b;
  --line: rgba(17, 17, 17, 0.18);
  --accent: #c1392b;
  --display-font: 'Times New Roman', 'Songti SC', 'STSong', 'Source Han Serif SC', serif;
  --mono-font: ui-monospace, 'SF Mono', 'Menlo', 'Consolas', monospace;
  --body-font: system-ui, -apple-system, 'PingFang SC', 'Hiragino Sans GB', sans-serif;
  background: var(--bg);
  color: var(--ink);
  font-family: var(--body-font);
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  letter-spacing: 0.005em;
}
.swiss-root.is-dark {
  --bg: #0e0e0d;
  --ink: #f1efe7;
  --mute: #8b8b85;
  --line: rgba(241, 239, 231, 0.18);
  --accent: #ff5a3d;
}
.swiss-root::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: linear-gradient(to right, var(--line) 1px, transparent 1px);
  background-size: calc(100% / 12) 100%;
  opacity: 0.35;
  mix-blend-mode: multiply;
  z-index: 0;
}
.swiss-root.is-dark::before {
  mix-blend-mode: screen;
}

/* Header */
.swiss-top {
  position: relative;
  z-index: 2;
  border-bottom: 1px solid var(--line);
}
.swiss-top__inner {
  max-width: 1280px;
  margin: 0 auto;
  padding: 18px 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}
.swiss-mark {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-family: var(--mono-font);
  font-size: 12px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}
.swiss-mark__dot {
  width: 8px;
  height: 8px;
  background: var(--accent);
  display: inline-block;
}
.swiss-mark__logo {
  height: 20px;
  width: auto;
  display: inline-block;
}
.swiss-mark__name {
  color: var(--ink);
  font-weight: 600;
}
.swiss-actions {
  display: flex;
  align-items: center;
  gap: 14px;
}
.swiss-meta {
  font-family: var(--mono-font);
  font-size: 11px;
  color: var(--mute);
  letter-spacing: 0.16em;
  text-transform: uppercase;
}
.swiss-meta--inline {
  display: inline-block;
  margin-left: 14px;
  padding-left: 14px;
  border-left: 1px solid var(--line);
}
.swiss-meta--mute {
  color: var(--mute);
}
.swiss-link {
  width: 34px;
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--line);
  color: var(--ink);
  background: transparent;
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}
.swiss-link:hover {
  background: var(--ink);
  color: var(--bg);
}
.swiss-cta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border: 1px solid var(--ink);
  font-family: var(--mono-font);
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--ink);
  background: transparent;
  text-decoration: none;
  transition: background 0.18s ease, color 0.18s ease;
}
.swiss-cta:hover {
  background: var(--ink);
  color: var(--bg);
}
.swiss-cta__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  background: var(--accent);
  color: #fff;
  font-family: var(--mono-font);
  font-size: 10px;
  font-weight: 700;
}

/* Main */
.swiss-main {
  position: relative;
  z-index: 1;
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 32px;
}
.swiss-section {
  padding: 96px 0;
  border-bottom: 1px solid var(--line);
}
.swiss-section:last-of-type {
  border-bottom: none;
}

/* Rule (section header bar) */
.swiss-rule {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 56px;
}
.swiss-rule__num {
  font-family: var(--mono-font);
  font-size: 11px;
  font-weight: 600;
  color: var(--ink);
  letter-spacing: 0.2em;
}
.swiss-rule__line {
  flex: 1;
  height: 1px;
  background: var(--line);
}
.swiss-rule__label {
  font-family: var(--mono-font);
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--mute);
}

/* Hero display */
.swiss-display {
  font-family: var(--display-font);
  font-weight: 400;
  font-style: italic;
  line-height: 0.92;
  letter-spacing: -0.02em;
  margin: 0;
  color: var(--ink);
}
.swiss-display__row {
  display: flex;
  align-items: baseline;
  gap: clamp(24px, 4vw, 64px);
  flex-wrap: wrap;
}
.swiss-display__row--lead {
  font-size: clamp(72px, 14vw, 200px);
}
.swiss-display__row--trail {
  font-size: clamp(72px, 14vw, 200px);
  font-style: normal;
  font-weight: 600;
  letter-spacing: -0.04em;
}
.swiss-display__word--accent {
  position: relative;
}
.swiss-display__word--accent::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0.08em;
  height: 0.08em;
  background: var(--accent);
}
.swiss-display__rule {
  display: inline-block;
  width: clamp(60px, 8vw, 140px);
  height: 1px;
  background: var(--ink);
  align-self: center;
  transform: translateY(-0.18em);
}
.swiss-hero__bottom {
  margin-top: 80px;
  display: grid;
  grid-template-columns: 5fr 7fr;
  gap: 64px;
  align-items: end;
}
.swiss-lede {
  font-family: var(--display-font);
  font-size: clamp(20px, 1.6vw, 24px);
  line-height: 1.45;
  color: var(--ink);
  margin: 0;
  max-width: 520px;
}
.swiss-hero__cta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-self: end;
}
.swiss-button {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  padding: 14px 22px;
  border: 1px solid var(--ink);
  background: transparent;
  color: var(--ink);
  font-family: var(--mono-font);
  font-size: 12px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  text-decoration: none;
  transition: background 0.2s ease, color 0.2s ease, transform 0.2s ease;
}
.swiss-button:hover {
  background: var(--ink);
  color: var(--bg);
  transform: translate(2px);
}
.swiss-button--solid {
  background: var(--ink);
  color: var(--bg);
}
.swiss-button--solid:hover {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}
.swiss-button__index {
  font-size: 14px;
}

/* Pillars */
.swiss-pillars__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0;
  border-top: 1px solid var(--line);
}
.swiss-pillar {
  padding: 40px 32px 40px 0;
  border-right: 1px solid var(--line);
  position: relative;
}
.swiss-pillar:last-child {
  border-right: none;
  padding-right: 0;
}
.swiss-pillar:not(:first-child) {
  padding-left: 32px;
}
.swiss-pillar__index {
  font-family: var(--mono-font);
  font-size: 11px;
  color: var(--accent);
  letter-spacing: 0.2em;
  margin-bottom: 28px;
}
.swiss-pillar__title {
  font-family: var(--display-font);
  font-weight: 500;
  font-size: clamp(28px, 2.6vw, 40px);
  line-height: 1.1;
  letter-spacing: -0.01em;
  margin: 0 0 16px;
  color: var(--ink);
}
.swiss-pillar__body {
  font-size: 14px;
  line-height: 1.65;
  color: var(--mute);
  margin: 0;
  max-width: 36ch;
}

/* Coverage */
.swiss-coverage__list {
  list-style: none;
  margin: 0;
  padding: 0;
  border-top: 1px solid var(--line);
}
.swiss-coverage__row {
  display: grid;
  grid-template-columns: 60px 1.2fr auto 2fr 90px;
  align-items: center;
  gap: 24px;
  padding: 24px 0;
  border-bottom: 1px solid var(--line);
  transition: padding-left 0.2s ease;
}
.swiss-coverage__row:hover {
  padding-left: 12px;
}
.swiss-coverage__num {
  font-family: var(--mono-font);
  font-size: 11px;
  color: var(--mute);
  letter-spacing: 0.2em;
}
.swiss-coverage__name {
  font-family: var(--display-font);
  font-size: clamp(22px, 2vw, 30px);
  letter-spacing: -0.01em;
  color: var(--ink);
}
.swiss-coverage__rule {
  width: 36px;
  height: 1px;
  background: var(--line);
}
.swiss-coverage__hint {
  font-size: 13px;
  color: var(--mute);
  font-family: var(--mono-font);
  letter-spacing: 0.04em;
}
.swiss-coverage__status {
  font-family: var(--mono-font);
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  text-align: right;
}
.swiss-coverage__status[data-state='live'] {
  color: var(--accent);
}
.swiss-coverage__status[data-state='soon'] {
  color: var(--mute);
}

/* Close */
.swiss-close__pull {
  font-family: var(--display-font);
  font-style: italic;
  font-weight: 400;
  font-size: clamp(34px, 4vw, 64px);
  line-height: 1.18;
  letter-spacing: -0.015em;
  margin: 0;
  max-width: 22ch;
  color: var(--ink);
}
.swiss-close__cta {
  margin-top: 56px;
}

/* Footer */
.swiss-foot {
  position: relative;
  z-index: 2;
  border-top: 1px solid var(--line);
}
.swiss-foot__inner {
  max-width: 1280px;
  margin: 0 auto;
  padding: 28px 32px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 24px;
}
.swiss-foot__link {
  margin-left: auto;
  font-family: var(--mono-font);
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--mute);
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: color 0.18s ease, border-color 0.18s ease;
}
.swiss-foot__link:hover {
  color: var(--ink);
  border-bottom-color: var(--ink);
}

/* Reveal animation */
.swiss-section {
  opacity: 0;
  transform: translateY(12px);
  animation: swiss-reveal 0.8s cubic-bezier(0.2, 0.7, 0.2, 1) forwards;
}
.swiss-hero {
  animation-delay: 0.05s;
}
.swiss-pillars {
  animation-delay: 0.15s;
}
.swiss-coverage {
  animation-delay: 0.25s;
}
.swiss-close {
  animation-delay: 0.35s;
}
@keyframes swiss-reveal {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 960px) {
  .swiss-main {
    padding: 0 24px;
  }
  .swiss-top__inner {
    padding: 16px 24px;
  }
  .swiss-foot__inner {
    padding: 24px;
  }
  .swiss-section {
    padding: 64px 0;
  }
  .swiss-rule {
    margin-bottom: 36px;
  }
  .swiss-hero__bottom {
    grid-template-columns: 1fr;
    gap: 36px;
    margin-top: 56px;
  }
  .swiss-hero__cta {
    justify-self: start;
  }
  .swiss-pillars__grid {
    grid-template-columns: 1fr;
  }
  .swiss-pillar {
    padding: 32px 0;
    border-right: none;
    border-bottom: 1px solid var(--line);
  }
  .swiss-pillar:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }
  .swiss-pillar:not(:first-child) {
    padding-left: 0;
  }
  .swiss-coverage__row {
    grid-template-columns: 40px 1fr 70px;
    gap: 16px;
  }
  .swiss-coverage__rule,
  .swiss-coverage__hint {
    display: none;
  }
}
@media (max-width: 540px) {
  .swiss-actions {
    gap: 8px;
  }
  .swiss-meta {
    font-size: 10px;
  }
  .swiss-display__row {
    gap: 18px;
  }
  .swiss-display__rule {
    width: 36px;
  }
  .swiss-cta {
    padding: 6px 10px;
    font-size: 10px;
  }
}
</style>
