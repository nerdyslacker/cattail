<script setup>
import ContentPane from './components/content/ContentPane.vue'
import BrowserPane from './components/sidebar/BrowserPane.vue'
import { computed, onMounted, reactive, ref, watchEffect } from 'vue'
import { debounce } from 'lodash'
import { useThemeVars } from 'naive-ui'
import usePreferencesStore from './stores/preferences.js'
import useTailScaleStore from './stores/tailscale.js'
import ToolbarControlWidget from '@/components/common/ToolbarControlWidget.vue'
import { EventsOn, WindowIsFullscreen, WindowIsMaximised, WindowToggleMaximise } from 'wailsjs/runtime/runtime.js'
import { isMacOS } from '@/utils/platform.js'
import iconUrl from '@/assets/images/icon.png'
import ResizeableWrapper from '@/components/common/ResizeableWrapper.vue'
import { extraTheme } from '@/utils/extra_theme.js'

const themeVars = useThemeVars()

const props = defineProps({
    loading: Boolean,
})

const data = reactive({
    navMenuWidth: 50,
    toolbarHeight: 38,
})

const prefStore = usePreferencesStore()
const exThemeVars = computed(() => {
    return extraTheme(prefStore.isDark)
})
// const preferences = ref({})
// provide('preferences', preferences)

const saveSidebarWidth = debounce(prefStore.savePreferences, 1000, { trailing: true })
const handleResize = () => {
    saveSidebarWidth()
}

const logoWrapperWidth = computed(() => {
    return `${data.navMenuWidth + prefStore.behavior.asideWidth - 4}px`
})

const logoPaddingLeft = ref(10)
const maximised = ref(false)
const hideRadius = ref(false)
const wrapperStyle = computed(() => {
    return hideRadius.value
        ? {}
        : {
              border: `1px solid ${themeVars.value.borderColor}`,
              borderRadius: '10px',
          }
})
const spinStyle = computed(() => {
    return hideRadius.value
        ? {
              backgroundColor: themeVars.value.bodyColor,
          }
        : {
              backgroundColor: themeVars.value.bodyColor,
              borderRadius: '10px',
          }
})

const onToggleFullscreen = (fullscreen) => {
    hideRadius.value = fullscreen
    if (fullscreen) {
        logoPaddingLeft.value = 10
    } else {
        logoPaddingLeft.value = isMacOS() ? 70 : 10
    }
}

const onToggleMaximize = (isMaximised) => {
    if (isMaximised) {
        maximised.value = true
        if (!isMacOS()) {
            hideRadius.value = true
        }
    } else {
        maximised.value = false
        if (!isMacOS()) {
            hideRadius.value = false
        }
    }
}

const tailScaleStore = useTailScaleStore()
const isOnline = ref(false)
const updateStatus = async () => {
  const prevOnline = isOnline.value
  isOnline.value = await tailScaleStore.updateStatus(prevOnline)
}

const onToggleTailScale = async (checked) => {
  if (checked) {
    await tailScaleStore.start()
  } else {
    await tailScaleStore.stop()
  }
  await updateStatus()
}

watchEffect(async () => {
  await updateStatus()
})

EventsOn('window_changed', (info) => {
    const { fullscreen, maximised } = info
    onToggleFullscreen(fullscreen === true)
    onToggleMaximize(maximised)
})

onMounted(async () => {
    const fullscreen = await WindowIsFullscreen()
    onToggleFullscreen(fullscreen === true)
    const maximised = await WindowIsMaximised()
    onToggleMaximize(maximised)
})
</script>

<template>
    <!-- app content-->
    <n-spin :show="props.loading" :style="spinStyle" :theme-overrides="{ opacitySpinning: 0 }">
        <div id="app-content-wrapper" :style="wrapperStyle" class="flex-box-v">
            <!-- title bar -->
            <div
                id="app-toolbar"
                :style="{ height: data.toolbarHeight + 'px' }"
                class="flex-box-h"
                style="--wails-draggable: drag"
                @dblclick="WindowToggleMaximise">
                <!-- title -->
                <div
                    id="app-toolbar-title"
                    :style="{
                        width: logoWrapperWidth,
                        minWidth: logoWrapperWidth,
                        paddingLeft: `${logoPaddingLeft}px`,
                    }">
                    <n-space :size="3" :wrap="false" :wrap-item="false" align="center">
                        <n-avatar :size="32" :src="iconUrl" color="#0000" style="min-width: 32px" />
                        <div style="min-width: 68px; font-weight: 800">Cattail</div>
                        <transition name="fade">
                            <n-switch size="medium" @update:value="onToggleTailScale" :value="isOnline" />
                        </transition>
                    </n-space>
                </div>
                <!-- browser tabs -->
                <div v-show="true" class="app-toolbar-tab flex-item-expand">
                    <content-value-tab />
                </div>
                <div class="flex-item-expand"></div>
                <!-- simulate window control buttons -->
                <toolbar-control-widget
                    v-if="!isMacOS()"
                    :maximised="maximised"
                    :size="data.toolbarHeight"
                    style="align-self: flex-start" />
            </div>

            <!-- content -->
            <div
                id="app-content"
                :style="prefStore.generalFont"
                class="flex-box-h flex-item-expand"
                style="--wails-draggable: none">
                <!-- browser page -->
                <div v-show="true" class="content-area flex-box-h flex-item-expand">
                    <resizeable-wrapper
                        v-model:size="prefStore.behavior.asideWidth"
                        :min-size="400"
                        :max-size="400"
                        :offset="data.navMenuWidth"
                        class="flex-item"
                        @update:size="handleResize">
                        <browser-pane class="app-side flex-item-expand" />
                    </resizeable-wrapper>
                    <content-pane class="flex-item-expand" />
                </div>
            </div>
        </div>
    </n-spin>
</template>

<style lang="scss" scoped>
#app-content-wrapper {
    width: 100vw;
    height: 100vh;
    overflow: hidden;
    box-sizing: border-box;
    background-color: v-bind('themeVars.bodyColor');
    color: v-bind('themeVars.textColorBase');

    #app-toolbar {
        background-color: v-bind('exThemeVars.titleColor');
        border-bottom: 1px solid v-bind('exThemeVars.splitColor');

        &-title {
            padding-left: 10px;
            padding-right: 10px;
            box-sizing: border-box;
            align-self: center;
            align-items: baseline;
        }
    }

    .app-toolbar-tab {
        align-self: flex-end;
        margin-bottom: -1px;
        margin-left: 3px;
    }

    #app-content {
        height: calc(100% - 60px);

        .content-area {
            overflow: hidden;
        }
    }

    .app-side {
        overflow: hidden;
        height: 100%;
        background-color: v-bind('exThemeVars.sidebarColor');
        border-right: 1px solid v-bind('exThemeVars.splitColor');
    }
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}

.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}
</style>
