<script setup>
import { useThemeVars } from 'naive-ui'
import { computed, onMounted } from 'vue'
import useTailScaleStore from '../../stores/tailscale.js'
import OsIcon from '@/components/common/OsIcon.vue'

const props = defineProps({

})

const themeVars = useThemeVars()
const tailScaleStore = useTailScaleStore()

// Ensure the data is loaded
onMounted(async () => {
    if (!tailScaleStore.namespaces) {
        await tailScaleStore.load();
    }
});

// Access namespaces from the store
const namespaces = computed(() => {
    return tailScaleStore.namespaces && tailScaleStore.namespaces.map(namespace => {
        const sortedPeers = [...namespace.peers].sort((a, b) => {
            if (a.online === b.online) {
                // If both peers have the same online status, sort by last_seen
                return new Date(b.last_seen) - new Date(a.last_seen)
            }
            // Otherwise, sort by online status (online first)
            return b.online - a.online
        })
        return {
            ...namespace,
            peers: sortedPeers
        }
    })
});

</script>

<template>
    <div class="nav-pane-container flex-box-v">
        <div>
            <n-list v-if="tailScaleStore.self != null" bordered hoverable clickable>
                <n-list-item @click="tailScaleStore.selectedPeer = tailScaleStore.self">
                    <template #prefix>
                        <n-tag :bordered="false" type="info">
                            <os-icon :name="tailScaleStore.self.os" size="16" />
                        </n-tag>
                    </template>
                    <template #suffix>

                    </template>
                    <n-thing>
                        <n-flex>
                            {{ tailScaleStore.self.name }}
                            <n-tag size="small" round>
                                This machine
                            </n-tag>
                        </n-flex>
                    </n-thing>
                </n-list-item>
            </n-list>

            <n-scrollbar style="max-height: 660px;">
                <div v-for="namespace in namespaces">
                    <n-divider title-placement="left">
                        {{ namespace.name }}
                    </n-divider>
                    <n-list hoverable clickable>
                        <n-list-item v-for="peer in namespace.peers">
                            <template #prefix>
                                <n-tag :bordered="false" type="info">
                                    <os-icon :name="peer.os" size="16" />
                                </n-tag>
                            </template>
                            <template #suffix>
                                <n-tag v-if="peer.online" :bordered="false" type="success">
                                    Online
                                </n-tag>
                                <n-tag v-else :bordered="false">
                                    {{ tailScaleStore.dateDiff(peer.last_seen) }}
                                </n-tag>
                            </template>
                            <n-thing @click="tailScaleStore.selectedPeer = peer">
                                <n-flex>
                                    <n-ellipsis style="max-width: 200px">
                                        {{ peer.name }}
                                    </n-ellipsis>
                                    <n-tag v-if="peer.exit_node" size="small" round type="error">
                                        Exit node
                                    </n-tag>
                                    <n-tag v-else-if="peer.exit_node_option" size="small" round type="warning">
                                        Exit node
                                    </n-tag>
                                </n-flex>
                            </n-thing>
                        </n-list-item>
                    </n-list>
                </div>
            </n-scrollbar>
        </div>
        <!-- bottom function bar -->
        <!-- <div class="nav-pane-bottom flex-box-v">
            <transition mode="out-in" name="fade">
                <div class="flex-box-h nav-pane-func" style="text-align: center;">
                    © {{ new Date().getFullYear() }} <img src="" height="25" alt="">
                </div>
            </transition>
        </div> -->
    </div>
</template>

<style lang="scss" scoped>
@import '@/styles/style';

:deep(.toggle-btn) {
    border-style: solid;
    border-width: 1px;
    border-radius: 3px;
    padding: 4px;
}

:deep(.toggle-on) {
    border-color: v-bind('themeVars.iconColorDisabled');
    background-color: v-bind('themeVars.iconColorDisabled');
}

:deep(.toggle-off) {
    border-color: #0000;
}

.nav-pane-top {
    //@include bottom-shadow(0.1);
    color: v-bind('themeVars.iconColor');
    border-bottom: v-bind('themeVars.borderColor') 1px solid;
}

.nav-pane-bottom {
    @include top-shadow(0.1);
    color: v-bind('themeVars.iconColor');
    border-top: v-bind('themeVars.borderColor') 1px solid;
}
</style>