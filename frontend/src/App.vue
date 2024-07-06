<script setup>
import AppContent from './AppContent.vue'
import { onMounted, ref, watch } from 'vue'
import usePreferencesStore from './stores/preferences.js'
import useTailScaleStore from './stores/tailscale.js'
import { useI18n } from 'vue-i18n'
import { darkTheme } from 'naive-ui'
import { WindowSetDarkTheme, WindowSetLightTheme } from 'wailsjs/runtime/runtime.js'
import { EventsOn } from 'wailsjs/runtime'
import { darkThemeOverrides, themeOverrides } from '@/utils/theme.js'
import { Info } from 'wailsjs/go/services/systemService.js'

const prefStore = usePreferencesStore()
const tailScaleStore = useTailScaleStore()
const i18n = useI18n()
const initializing = ref(true)
onMounted(async () => {
    try {
        initializing.value = true
        if (prefStore.autoCheckUpdate) {
            prefStore.checkForUpdate()
        }
        Info().then(({ data }) => {
            // const {os, arch, version} = data
        })

        await tailScaleStore.load();

        tailScaleStore.timer = setInterval(() => {
            // Force update (In Vue 3, it might not be necessary if reactive state is used correctly)
        }, 30_000);

        EventsOn('update_all', async () => await tailScaleStore.load());
        EventsOn('update_files', async () => {
            tailScaleStore.files = await Files();
        });
        EventsOn('app_running', () => {
            tailScaleStore.appRunning = true;
        });
        EventsOn('app_not_running', () => {
            tailScaleStore.selectedPeer = null;
            tailScaleStore.appRunning = false;
        });
    } finally {
        initializing.value = false
    }
})

// watch theme and dynamically switch
watch(
    () => prefStore.isDark,
    (isDark) => (isDark ? WindowSetDarkTheme() : WindowSetLightTheme()),
)

// watch language and dynamically switch
watch(
    () => prefStore.general.language,
    (lang) => (i18n.locale.value = prefStore.currentLanguage),
)
</script>

<template>
    <n-config-provider
        :inline-theme-disabled="true"
        :locale="prefStore.themeLocale"
        :theme="prefStore.isDark ? darkTheme : undefined"
        :theme-overrides="prefStore.isDark ? darkThemeOverrides : themeOverrides"
        class="fill-height">
        <n-dialog-provider>
            <app-content :loading="initializing" />
        </n-dialog-provider>
    </n-config-provider>
</template>

<style lang="scss"></style>
