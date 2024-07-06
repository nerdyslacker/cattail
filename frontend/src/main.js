import { createPinia } from 'pinia'
import { createApp, nextTick } from 'vue'
import App from './App.vue'
import './styles/style.scss'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import relativeTime from 'dayjs/plugin/relativeTime'
import { i18n } from '@/utils/i18n.js'
import usePreferencesStore from 'stores/preferences.js'
import { loadEnvironment } from '@/utils/platform.js'

dayjs.extend(duration)
dayjs.extend(relativeTime)

async function setupApp() {
    const app = createApp(App)
    app.use(i18n)
    app.use(createPinia())

    await loadEnvironment()
    const prefStore = usePreferencesStore()
    await prefStore.loadPreferences()
    // app.config.errorHandler = (err, instance, info) => {
    //     nextTick().then(() => {
    //         try {
    //             const content = err.toString()
    //             $notification.error(content, {
    //                 title: i18n.global.t('common.error'),
    //                 meta: 'Please see console output for more detail',
    //             })
    //             console.error(err)
    //         } catch (e) {}
    //     })
    // }
    app.mount('#app')
}

setupApp()
