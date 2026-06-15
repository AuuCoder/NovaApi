<template>
  <div :class="['auth-editorial', { 'auth-editorial--light': !isDark }]">
    <!-- Top bar -->
    <header class="auth-editorial__top">
      <div class="auth-editorial__top-inner">
        <router-link to="/home" class="auth-editorial__mark">
          <template v-if="settingsLoaded && siteLogo">
            <img :src="siteLogo" alt="Logo" style="height: 22px; width: auto" />
          </template>
          <span v-else class="auth-editorial__mark-dot" aria-hidden="true"></span>
          <span class="auth-editorial__mark-name">{{ siteName }}</span>
        </router-link>

        <nav class="auth-editorial__actions">
          <span class="auth-editorial__meta">
            {{ t('auth.editorial.volume', { num: volumeNum, year: currentYear }) }}
          </span>
          <button
            type="button"
            class="auth-editorial__icon-btn"
            :title="isDark ? t('auth.editorial.lightMode') : t('auth.editorial.darkMode')"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
          </button>
        </nav>
      </div>
    </header>

    <!-- Main grid -->
    <main class="auth-editorial__main">
      <!-- Left: brand / editorial column -->
      <section class="auth-editorial__brand">
        <div class="auth-editorial__marker">
          <span class="auth-editorial__marker-id">A1</span>
          <span class="auth-editorial__marker-line"></span>
          <span>{{ eyebrow }}</span>
        </div>

        <h1 class="auth-editorial__display">
          <span class="auth-editorial__display-row">{{ headTop }}</span>
          <span class="auth-editorial__display-row auth-editorial__display-row--accent">
            {{ headAccent }}
          </span>
        </h1>

        <p class="auth-editorial__lede">{{ lede }}</p>

        <ul class="auth-editorial__bullets">
          <li>
            <span class="auth-editorial__bullet-num">01</span>
            <span class="auth-editorial__bullet-label">{{ t('auth.editorial.feature1') }}</span>
          </li>
          <li>
            <span class="auth-editorial__bullet-num">02</span>
            <span class="auth-editorial__bullet-label">{{ t('auth.editorial.feature2') }}</span>
          </li>
          <li>
            <span class="auth-editorial__bullet-num">03</span>
            <span class="auth-editorial__bullet-label">{{ t('auth.editorial.feature3') }}</span>
          </li>
        </ul>
      </section>

      <!-- Right: form panel -->
      <section class="auth-editorial__panel">
        <div class="auth-editorial__marker">
          <span class="auth-editorial__marker-id">A2</span>
          <span class="auth-editorial__marker-line"></span>
          <span>{{ sectionLabel }}</span>
        </div>

        <slot />

        <div v-if="$slots.footer" class="auth-editorial__panel-foot">
          <slot name="footer" />
        </div>
      </section>
    </main>

    <!-- Bottom footer -->
    <footer class="auth-editorial__foot">
      <div class="auth-editorial__foot-inner">
        <span>&copy; {{ currentYear }} {{ siteName }}</span>
        <span>{{ t('auth.editorial.rights') }}</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(
  defineProps<{
    variant?: 'login' | 'register'
  }>(),
  {
    variant: 'login'
  }
)

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'NovaAPI')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true })
)
const siteSubtitle = computed(
  () =>
    appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform'
)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())
const volumeNum = computed(() => String(new Date().getMonth() + 1).padStart(2, '0'))

// ---- editorial copy varies by variant ----
const eyebrow = computed(() =>
  props.variant === 'register'
    ? t('auth.editorial.registerEyebrow')
    : t('auth.editorial.loginEyebrow')
)
const sectionLabel = computed(() =>
  props.variant === 'register'
    ? t('auth.editorial.registerSection')
    : t('auth.editorial.loginSection')
)
const headTop = computed(() =>
  props.variant === 'register'
    ? t('auth.editorial.registerHeadTop')
    : t('auth.editorial.loginHeadTop')
)
const headAccent = computed(() =>
  props.variant === 'register'
    ? t('auth.editorial.registerHeadAccent')
    : t('auth.editorial.loginHeadAccent')
)
const lede = computed(() => siteSubtitle.value)

// ---- theme toggle (mirrors AppSidebar logic) ----
const isDark = ref(document.documentElement.classList.contains('dark'))

function toggleTheme(): void {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  appStore.fetchPublicSettings()
  isDark.value = document.documentElement.classList.contains('dark')
})
</script>
