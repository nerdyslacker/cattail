import { defineStore } from 'pinia'
// import { endsWith, get, isEmpty, map, now, size } from 'lodash'
import { EventsEmit, EventsOnce } from 'wailsjs/runtime'
import {
    Accounts,
    CopyClipboard,
    RemoveFile,
    AcceptFile,
    CurrentAccount,
    Namespaces,
    Files,
    Self,
    SetExitNode,
    SwitchTo,
    UploadFile,
    AdvertiseExitNode,
    Start,
    Stop,
    GetStatus,
    UpdateStatus,
} from 'wailsjs/go/services/tailScaleService.js'

const useTailScaleStore = defineStore('tailScaleStore', {
    state: () => ({
        account: '',
        otherAccounts: [],
        files: [],
        namespaces: null,
        self: {},
        selectedPeer: null,
        appRunning: true,
        timer: null,
    }),
    actions: {
        async load() {
            if (!this.appRunning) {
                return
            }

            this.account = await CurrentAccount()
            this.otherAccounts = await Accounts()
            this.files = await Files()
            this.namespaces = await Namespaces()
            this.self = await Self()
            if (this.selectedPeer === null) {
                this.selectedPeer = this.self
            } else {
                this.namespaces.forEach((namespace) => {
                    namespace.peers.forEach((peer) => {
                        if (peer.dns_name === this.selectedPeer.dns_name) {
                            this.selectedPeer = peer
                        }
                    })
                })
            }
        },
        async start() {
            await Start()
        },
        async stop() {
            await Stop()
        },     
        async getStatus() {
            return await GetStatus()
        },
        async updateStatus(prevOnline) {
            return await UpdateStatus(prevOnline)
        },
        async switchAccount(name) {
            // const name = event.target.text
            await SwitchTo(name)
        },
        async setExitNode(event) {
            console.log('setting exit node')
            event.target.disabled = true
            EventsOnce('exit_node_connect', () => {
                event.target.disabled = false
            })
            await SetExitNode(this.selectedPeer.dns_name)
        },
        async advertiseExitNode(event) {
            console.log('advertising exit node')
            event.target.disabled = true
            EventsOnce('advertise_exit_node_done', async () => {
                event.target.disabled = false
                this.self = await Self()
            })
            await AdvertiseExitNode(this.selectedPeer.dns_name)
        },
        async acceptFile(name) {
            await AcceptFile(name)
        },
        async rejectFile(name) {
            await RemoveFile(name)
        },
        async copyClipboard(name) {
            await CopyClipboard(name)
        },
        async sendFile(name) {
            await UploadFile(name)
        },
        dateDiff(ref) {
            const date = new Date(ref)
            const now = new Date()
            const res = (now - date) / 1000
            if (res < 3600) {
                return Math.round(res / 60) + ' minutes ago'
            } else if (res < 86400) {
                return Math.round(res / 3600) + ' hours ago'
            } else {
                return Math.round(res / 86400) + ' days ago'
            }
        },
        humanFileSize(bytes, si = false, dp = 1) {
            const thresh = si ? 1000 : 1024

            if (Math.abs(bytes) < thresh) {
                return bytes + ' B'
            }

            const units = si
                ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
                : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB']
            let u = -1
            const r = 10 ** dp

            do {
                bytes /= thresh
                ++u
            } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1)

            return bytes.toFixed(dp) + ' ' + units[u]
        },
    },
})

export default useTailScaleStore
