import { defineStore } from 'pinia'
import { lang } from '@/langs/index.js'
import { cloneDeep, findIndex, get, isEmpty, join, map, merge, pick, set, some, split } from 'lodash'
import {
    GetPreferences,
    RestorePreferences,
    SetPreferences,
} from 'wailsjs/go/services/preferencesService.js'
import { BrowserOpenURL } from 'wailsjs/runtime/runtime.js'
import { i18nGlobal } from '@/utils/i18n.js'
import { enUS, NButton, NSpace, useOsTheme, zhCN } from 'naive-ui'
import { h, nextTick } from 'vue'
import { compareVersion } from '@/utils/version.js'

const osTheme = useOsTheme()
const usePreferencesStore = defineStore('preferences', {
    /**
     * @typedef {Object} Preferences
     * @property {Object} general
     * @property {Object} editor
     * @property {FontItem[]} fontList
     */
    /**
     *
     * @returns {Preferences}
     */
    state: () => ({
        behavior: {
            asideWidth: 300,
            windowWidth: 0,
            windowHeight: 0,
            windowMaximised: false,
        },
        general: {
            theme: 'auto',
            language: 'auto',
            font: '',
            fontFamily: [],
            fontSize: 14,
            scanSize: 3000,
            keyIconStyle: 0,
            useSysProxy: false,
            useSysProxyHttp: false,
            skipVersion: '',
        },
        editor: {
            font: '',
            fontFamily: [],
            fontSize: 14,
            showLineNum: true,
            showFolding: true,
            dropText: true,
            links: true,
        },
        cli: {
            fontFamily: [],
            fontSize: 14,
            cursorStyle: 'block',
        },
        buildInDecoder: [],
        decoder: [],
        lastPref: {},
        fontList: [],
    }),
    getters: {
        getSeparator() {
            return ':'
        },

        themeOption() {
            return [
                {
                    value: 'light',
                    label: 'preferences.general.theme_light',
                },
                {
                    value: 'dark',
                    label: 'preferences.general.theme_dark',
                },
                {
                    value: 'auto',
                    label: 'preferences.general.theme_auto',
                },
            ]
        },

        /**
         * all available language
         * @returns {{label: string, value: string}[]}
         */
        langOption() {
            const options = Object.entries(lang).map(([key, value]) => ({
                value: key,
                label: value['name'],
            }))
            options.splice(0, 0, {
                value: 'auto',
                label: 'preferences.general.system_lang',
            })
            return options
        },
        /**
         * get current language setting
         * @return {string}
         */
        currentLanguage() {
            let lang = get(this.general, 'language', 'auto')
            if (lang === 'auto') {
                const systemLang = navigator.language || navigator.userLanguage
                lang = split(systemLang, '-')[0]
            }
            return lang || 'en'
        },

        isDark() {
            const th = get(this.general, 'theme', 'auto')
            if (th !== 'auto') {
                return th === 'dark'
            } else {
                return osTheme.value === 'dark'
            }
        },

        themeLocale() {
            const lang = this.currentLanguage
            switch (lang) {
                case 'zh':
                    return zhCN
                default:
                    return enUS
            }
        },

        showLineNum() {
            return get(this.editor, 'showLineNum', true)
        },

        showFolding() {
            return get(this.editor, 'showFolding', true)
        },

        dropText() {
            return get(this.editor, 'dropText', true)
        },

        editorLinks() {
            return get(this.editor, 'links', true)
        },
    },
    actions: {
        _applyPreferences(data) {
            for (const key in data) {
                set(this, key, data[key])
            }
        },

        /**
         * load preferences from local
         * @returns {Promise<void>}
         */
        async loadPreferences() {
            const { success, data } = await GetPreferences()
            if (success) {
                this.lastPref = cloneDeep(data)
                this._applyPreferences(data)
                // default value
                const showLineNum = get(data, 'editor.showLineNum')
                if (showLineNum === undefined) {
                    set(data, 'editor.showLineNum', true)
                }
                const showFolding = get(data, 'editor.showFolding')
                if (showFolding === undefined) {
                    set(data, 'editor.showFolding', true)
                }
                const dropText = get(data, 'editor.dropText')
                if (dropText === undefined) {
                    set(data, 'editor.dropText', true)
                }
                const links = get(data, 'editor.links')
                if (links === undefined) {
                    set(data, 'editor.links', true)
                }
                i18nGlobal.locale.value = this.currentLanguage
            }
        },

        /**
         * save preferences to local
         * @returns {Promise<boolean>}
         */
        async savePreferences() {
            const pf = pick(this, ['behavior', 'general', 'editor', 'cli', 'decoder'])
            const { success, msg } = await SetPreferences(pf)
            return success === true
        },

        /**
         * reset to last-loaded preferences
         * @returns {Promise<void>}
         */
        async resetToLastPreferences() {
            if (!isEmpty(this.lastPref)) {
                this._applyPreferences(this.lastPref)
            }
        },

        /**
         * restore preferences to default
         * @returns {Promise<boolean>}
         */
        async restorePreferences() {
            const { success, data } = await RestorePreferences()
            if (success === true) {
                const { pref } = data
                this._applyPreferences(pref)
                return true
            }
            return false
        },
    },
})

export default usePreferencesStore
