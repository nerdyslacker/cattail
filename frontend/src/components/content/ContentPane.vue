<script setup>
import { h, computed, ref, reactive } from 'vue'
import { useThemeVars, NButton } from 'naive-ui'
const themeVars = useThemeVars()
import useTailScaleStore from '../../stores/tailscale.js'
import { CopyRegular, Plus, TrashAlt } from '@vicons/fa'
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
 
const selected_peer = computed(() => tailScaleStore.selectedPeer)

const onToggleSetExitNode = async (event) => {
  await tailScaleStore.setExitNode(event);
  selected_peer.value.exit_node = !selected_peer.value.exit_node;
}

const onToggleAdvertiseExitNode = async (event) => {
  await tailScaleStore.advertiseExitNode(event);
  selected_peer.value.exit_node_option = !selected_peer.value.exit_node_option;
}

const onToggleAllowLANAccess = async (checked) => {
  await tailScaleStore.allowLANAccess(checked);
  selected_peer.value.allow_lan_access = checked;
}

const onToggleAcceptRoutes = async (checked) => {
  await tailScaleStore.acceptRoutes(checked);
  selected_peer.value.accept_routes = checked;
}

const onToggleRunSSH = async (checked) => {
  await tailScaleStore.runSSH(checked);
  selected_peer.value.run_ssh = checked;
}

const showModalRef = ref(false);
const showModal = showModalRef;
const routes = ref([...new Set(selected_peer?.value?.advertised_routes || [])]);
const routeValue = ref(null);

const onClickAdvertiseRoutes = async () => {
  if (routeValue.value) {
    const newRoutes = routeValue.value
      .split(',')
      .map((route) => route.trim())
      .filter((route) => route !== '');

    newRoutes.forEach((route) => {
      if (!routes.value.includes(route)) {
        routes.value.push(route);
      }
    });
  }
  
  await tailScaleStore.advertiseRoutes(routes.value.join(','));
  selected_peer.value.advertised_routes = [...new Set(routes.value)];

  showModalRef.value = false;
  routeValue.value = null;
}

const onClickRemoveRoute = async (route) => {
  routes.value = routes.value.filter((r) => r !== route);
  await tailScaleStore.advertiseRoutes(routes.value.join(','));
  selected_peer.value.advertised_routes = [...new Set(routes.value)];
}

const onModalCancel = () => {
  showModalRef.value = false
};

</script>

<template>
  <div class="content-container flex-box-v">
    <n-tabs default-value="general" type="line" placement="top" size="large"
      tab-style="padding-left: 10px; padding-right: 10px;" animated>
      <n-tab-pane name="general" tab="General">
        <n-scrollbar style="max-height: 650px;">
          <n-space v-if="selected_peer != null" vertical style="padding-left: 10px; padding-right: 10px;">
            <n-card title="Information" size="small">
              <n-list hoverable>
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
                    <n-tag>Last update </n-tag>
                  </template>
                  <template #suffix>
                    <n-button style="visibility: hidden;" @click="tailScaleStore.copyClipboard(selected_peer.last_write)">
                      <n-icon :component="CopyRegular" size="16" :depth="1" />
                    </n-button>
                  </template>
                  <n-thing style="text-align: right;">
                    {{ new Date(selected_peer.last_write).getUTCFullYear() != 1 ? tailScaleStore.dateDiff(selected_peer.last_write) : '-'  }}
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
            <n-card title="Options" size="small">
              <n-list hoverable>
                <n-list-item v-if="selected_peer.dns_name !== tailScaleStore.self.dns_name">
                  <template #prefix>
                    <n-tag>Use exit node</n-tag>
                  </template>
                  <n-thing style="text-align: right;">
                    <n-switch v-if="!selected_peer.online || !selected_peer.exit_node_option" size="large"
                    @click="onToggleSetExitNode" disabled />
                    <n-switch v-else size="large" @click="onToggleSetExitNode" :value="selected_peer.exit_node" />
                  </n-thing>
                </n-list-item>
                <n-list-item v-if="selected_peer.dns_name === tailScaleStore.self.dns_name">
                  <template #prefix>
                    <n-tag>Advertise exit node</n-tag>
                  </template>
                  <n-thing  style="text-align: right;">
                    <n-switch size="large" @click="onToggleAdvertiseExitNode"  :value="selected_peer.exit_node_option" />
                  </n-thing>
                </n-list-item>
                <n-list-item v-if="selected_peer.dns_name === tailScaleStore.self.dns_name">
                  <template #prefix>
                    <n-tag>Allow LAN access</n-tag>
                  </template>
                  <n-thing style="text-align: right;">
                    <n-switch size="large" @update:value="onToggleAllowLANAccess"  :value="selected_peer.allow_lan_access" />
                  </n-thing>
                </n-list-item>
                <n-list-item v-if="selected_peer.dns_name === tailScaleStore.self.dns_name">
                  <template #prefix>
                    <n-tag>Accept routes</n-tag>
                  </template>
                  <n-thing style="text-align: right;">
                    <n-switch size="large" @update:value="onToggleAcceptRoutes" :value="selected_peer.accept_routes" />
                  </n-thing>
                </n-list-item>
                <n-list-item v-if="selected_peer.dns_name === tailScaleStore.self.dns_name">
                  <template #prefix>
                    <n-tag>Run SSH</n-tag>
                  </template>
                  <n-thing style="text-align: right;">
                    <n-switch size="large" @update:value="onToggleRunSSH" :value="selected_peer.run_ssh" />
                  </n-thing>
                </n-list-item>
              </n-list>
            </n-card>
            <n-card title="Advertised Routes" size="small">
              <template #header-extra v-if="selected_peer.dns_name === tailScaleStore.self.dns_name">
                <n-button @click="showModal = true">
                   <n-icon :component="Plus" size="16" :depth="1" />
                </n-button>
                <n-modal
                  v-model:show="showModal"
                  :mask-closable="false"
                  preset="dialog"
                  :show-icon="false"
                  title="Add IP"
                  positive-text="Add"
                  negative-text="Cancel"
                  @positive-click="onClickAdvertiseRoutes"
                  @negative-click="onModalCancel"
                >
                  <n-input v-model:value="routeValue" type="text" placeholder="IP prefix to advertise" />
                </n-modal>
              </template>
              <n-list hoverable>
                <n-list-item v-for="route in selected_peer.advertised_routes">
                  {{ route }}
                  <template #suffix v-if="selected_peer.dns_name === tailScaleStore.self.dns_name">
                    <n-button @click="onClickRemoveRoute(route)">
                      <n-icon :component="TrashAlt" size="16" :depth="1" />
                    </n-button>
                  </template>
                </n-list-item>
                <n-list-item v-if="selected_peer.advertised_routes == null || selected_peer.advertised_routes?.length === 0">
                  <p style="text-align: center;">No advertised routes</p>
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