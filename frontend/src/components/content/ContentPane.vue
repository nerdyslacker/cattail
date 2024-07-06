<script setup>
import { h, computed, ref } from 'vue'
import { useThemeVars, NButton } from 'naive-ui'
const themeVars = useThemeVars()
import useTailScaleStore from '../../stores/tailscale.js'
import { CopyRegular } from '@vicons/fa'
const tailScaleStore = useTailScaleStore()

const accountOptions = computed(() => tailScaleStore.otherAccounts.map(acc => {
  return {
    label: acc,
    key: acc
  }
}));

const handleAccountSelect = (key) => {
  tailScaleStore.switchAccount(key)
}

const selected_peer = computed(() => tailScaleStore.selectedPeer);

const files = computed(() => tailScaleStore.files.map(file => {
  return {
    filename: file.name,
    size: tailScaleStore.humanFileSize(file.size)
  }
}));

const columns = [
  {
    title: 'Filename',
    key: 'filename',
    sorter: 'default'
  },
  {
    title: 'Size',
    key: 'size'
  },
  {
    title: 'Actions',
    key: 'actions',
    render(row) {
      return [h(
        NButton,
        {
          strong: true,
          tertiary: true,
          size: 'small',
          type: 'success',
          style: {
            marginRight: '6px'
          },
          onClick: () => tailScaleStore.acceptFile(row.filename)
        },
        { default: () => 'Accept' }
      ), h(
        NButton,
        {
          strong: true,
          tertiary: true,
          size: 'small',
          type: 'error',
          onClick: () => tailScaleStore.rejectFile(row.filename)
        },
        { default: () => 'Reject' }
      )]
    }
  }
]

const data = ref(files)

const dataTableInstRef = ref(null)

</script>

<template>
  <div class="content-container flex-box-v">
    <n-tabs default-value="general" type="line" placement="top" size="large"
      tab-style="padding-left: 10px; padding-right: 10px;" animated>
      <n-tab-pane name="general" tab="General">
        <n-scrollbar style="max-height: 650px;">
          <n-space v-if="selected_peer != null" vertical style="padding-left: 10px; padding-right: 10px;">
            <n-card title="Information" size="small">
              <n-list hoverable clickable>
                <n-list-item>
                  <template #prefix>
                    <n-tag>Hostname </n-tag>
                  </template>
                  <template #suffix>
                    <n-button @click="tailScaleStore.copyClipboard(selected_peer.name)">
                      <n-icon :component="CopyRegular" size="16" :depth="1" />
                    </n-button>
                  </template>
                  <n-thing style="text-align: right;">
                    {{ selected_peer.name }}
                  </n-thing>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-tag>DNS name </n-tag>
                  </template>
                  <template #suffix>
                    <n-button @click="tailScaleStore.copyClipboard(selected_peer.dns_name)">
                      <n-icon :component="CopyRegular" size="16" :depth="1" />
                    </n-button>
                  </template>
                  <n-thing style="text-align: right;">
                    {{ selected_peer.dns_name }}
                  </n-thing>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-tag>Tailscale ID </n-tag>
                  </template>
                  <template #suffix>
                    <n-button @click="tailScaleStore.copyClipboard(selected_peer.id)">
                      <n-icon :component="CopyRegular" size="16" :depth="1" />
                    </n-button>
                  </template>
                  <n-thing style="text-align: right;">
                    {{ selected_peer.id }}
                  </n-thing>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-tag>Created </n-tag>
                  </template>
                  <template #suffix>
                    <n-button style="visibility: hidden;" @click="tailScaleStore.copyClipboard(selected_peer.created_at)">
                      <n-icon :component="CopyRegular" size="16" :depth="1" />
                    </n-button>
                  </template>
                  <n-thing style="text-align: right;">
                    {{ tailScaleStore.dateDiff(selected_peer.created_at) }}
                  </n-thing>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-tag>Last seen </n-tag>
                  </template>
                  <template #suffix>
                    <n-button style="visibility: hidden;" @click="tailScaleStore.copyClipboard(selected_peer.last_seen)">
                      <n-icon :component="CopyRegular" size="16" :depth="1" />
                    </n-button>
                  </template>
                  <n-thing style="text-align: right;">
                    {{ selected_peer.dns_name === tailScaleStore.self.dns_name || selected_peer.online ? "now" :
            tailScaleStore.dateDiff(selected_peer.last_seen) }}
                  </n-thing>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-tag>OS </n-tag>
                  </template>
                  <template #suffix>
                    <n-button style="visibility: hidden;" @click="tailScaleStore.copyClipboard(selected_peer.os)">
                      <n-icon :component="CopyRegular" size="16" :depth="1" />
                    </n-button>
                  </template>
                  <n-thing style="text-align: right;">
                    {{ selected_peer.os }}
                  </n-thing>
                </n-list-item>
                <n-list-item>
                  <template #prefix>
                    <n-tag>Addresses </n-tag>
                  </template>
                  <n-thing>
                    <n-flex>
                      <div v-for="ip in selected_peer.ips">
                        <n-tooltip trigger="hover" placement="bottom" :show-arrow="false">
                          <template #trigger>
                            <n-button @click="tailScaleStore.copyClipboard(ip)">
                              {{ ip }}
                            </n-button>
                          </template>
                          Click to copy
                        </n-tooltip>
                      </div>
                    </n-flex>
                  </n-thing>
                </n-list-item>
              </n-list>
              <n-space vertical v-if="selected_peer.dns_name !== tailScaleStore.self.dns_name">
                <n-divider></n-divider>
                <n-flex justify="center">
                  <input type="file" multiple name="fields[assetsFieldHandle][]" id="assetsFieldHandle"
                    style="width: 1px; height: 1px; opacity: 0; overflow: hidden; position: absolute" @change="onChange"
                    ref="file" />
                  <n-button size="large" @click="tailScaleStore.sendFile(selected_peer.dns_name)">
                    Send file
                  </n-button>
                </n-flex>
              </n-space>
            </n-card>
            <n-card title="Exit node" size="small">
              <n-list hoverable clickable>
                <n-list-item v-if="selected_peer.dns_name === tailScaleStore.self.dns_name">
                  <template #prefix>
                    <n-tag>Advertise exit node</n-tag>
                  </template>
                  <n-thing>
                    <n-switch size="large" @click="tailScaleStore.advertiseExitNode"
                      :value="selected_peer.exit_node_option" />
                  </n-thing>
                </n-list-item>
                <n-list-item v-if="selected_peer.dns_name !== tailScaleStore.self.dns_name">
                  <template #prefix>
                    <n-tag>Use exit node</n-tag>
                  </template>
                  <n-thing>
                    <n-switch v-if="!selected_peer.online || !selected_peer.exit_node_option" size="large"
                      @click="tailScaleStore.setExitNode" disabled />
                    <n-switch v-else size="large" @click="tailScaleStore.setExitNode"
                      :value="selected_peer.exit_node" />
                  </n-thing>
                </n-list-item>
              </n-list>
            </n-card>
          </n-space>
          <n-space v-else vertical>
            <n-skeleton height="40px" :repeat="5" style="margin-top: 8px;" :sharp="false" />
            <n-skeleton height="40px" width="66%" :sharp="false" />
          </n-space>
        </n-scrollbar>
      </n-tab-pane>
      <n-tab-pane name="files" :tab="`Files (${files.length})`">
        <n-scrollbar style="max-height: 670px;">
          <n-space vertical style="padding-left: 10px; padding-right: 10px;">
            <n-data-table ref="dataTableInstRef" :columns="columns" :data="data" />
          </n-space>
        </n-scrollbar>
      </n-tab-pane>
      <!-- <n-tab-pane name="preferences" tab="Preferences">
        TODO
      </n-tab-pane> -->
      <template #suffix>
        <div style="padding-right: 10px;">
          <n-dropdown trigger="click" :options="accountOptions" @select="handleAccountSelect">
            <n-tooltip trigger="hover" placement="bottom" :show-arrow="false">
              <template #trigger>
                <n-button>{{ tailScaleStore.account }}</n-button>
              </template>
              Click to switch accounts
            </n-tooltip>
          </n-dropdown>
        </div>
      </template>
    </n-tabs>
  </div>
</template>

<style lang="scss" scoped>
@import '@/styles/content';

.content-container {
  //padding: 5px 5px 0;
  //padding-top: 0;
  box-sizing: border-box;
  background-color: v-bind('themeVars.tabColor');
}
</style>

<style lang="scss">
.content-sub-tab {
  background-color: v-bind('themeVars.tabColor');
  height: 100%;
}

.content-sub-tab-pane {
  padding: 0 !important;
  height: 100%;
  background-color: v-bind('themeVars.bodyColor');
  overflow: hidden;
}

.n-tabs .n-tabs-bar {
  transition: none !important;
}

.n-upload-trigger+.n-upload-file-list {
  margin-top: 0 !important;
  margin-left: 5px;
}

.n-upload-file-list .n-upload-file .n-upload-file-info {
  padding-top: 3px !important;
  padding-bottom: 3px !important;
}
</style>