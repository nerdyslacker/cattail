<script setup>
import { ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import useTailScaleStore from '../../stores/tailscale.js'
import { BrowsePrivateKey } from 'wailsjs/go/services/tailScaleService'
import { GetSSHCredential, SaveSSHCredential } from 'wailsjs/go/services/preferencesService'

const props = defineProps({
  show: { type: Boolean, default: false },
  peerName: { type: String, required: true },
  peerDNSName: { type: String, default: '' },
})

const emit = defineEmits(['update:show', 'connected'])

const tailScaleStore = useTailScaleStore()
const message = useMessage()

const deviceKey = () => props.peerDNSName || props.peerName

const method = ref('tailscale')
const host = ref('')
const port = ref(22)
const username = ref('root')
const password = ref('')
const privateKeyPath = ref('')
const connecting = ref(false)

watch(() => props.show, async (isOpen) => {
  if (isOpen) {
    const dk = deviceKey()
    if (!dk) return
    try {
      const resp = await GetSSHCredential(dk)
      if (resp.success && resp.data) {
        host.value = resp.data.host || ''
        port.value = resp.data.port || 22
        username.value = resp.data.username || 'root'
        privateKeyPath.value = resp.data.keyPath || ''
      }
    } catch (e) {
      // no saved credentials
    }
  }
})

const handleShow = (val) => {
  if (!val) {
    emit('update:show', false)
  }
}

const browseKeyFile = async () => {
  try {
    const path = await BrowsePrivateKey()
    if (path) privateKeyPath.value = path
  } catch (e) {
    // user cancelled
  }
}

const handleConnect = async () => {
  connecting.value = true
  try {
    let sessionId
    if (method.value === 'tailscale') {
      sessionId = await tailScaleStore.connectTailscaleSSH(props.peerName)
    } else {
      const key = deviceKey()
      await SaveSSHCredential(key, {
        host: host.value,
        port: port.value,
        username: username.value,
        keyPath: privateKeyPath.value,
      })
      sessionId = await tailScaleStore.connectDirectSSH(
        host.value || props.peerDNSName || props.peerName,
        port.value,
        username.value,
        password.value,
        privateKeyPath.value,
      )
    }
    emit('connected', sessionId)
    emit('update:show', false)
  } catch (e) {
    message.error('Connection failed: ' + e)
  } finally {
    connecting.value = false
  }
}
</script>

<template>
  <n-modal
    :show="show"
    :on-update:show="handleShow"
    :mask-closable="false"
    preset="card"
    title="SSH Connection"
    style="width: 480px"
    role="dialog"
    :bordered="true"
    closable
  >
    <n-radio-group v-model:value="method">
      <n-space vertical>
        <n-radio value="tailscale">
          Tailscale SSH (no credentials needed)
          <div style="font-size: 12px; color: #888; margin-top: 2px; padding-left: 2px;">
            Make sure SSH is advertised on the target host (<code>tailscale up --ssh</code>)
          </div>
        </n-radio>
        <n-radio value="direct">Direct SSH (credentials)</n-radio>
      </n-space>
    </n-radio-group>

    <template v-if="method === 'direct'">
      <n-divider />
      <n-form label-placement="top">
        <n-form-item label="Host">
          <n-input v-model:value="host" :placeholder="peerDNSName || peerName" />
        </n-form-item>
        <n-form-item label="Port">
          <n-input-number v-model:value="port" :min="1" :max="65535" />
        </n-form-item>
        <n-form-item label="Username">
          <n-input v-model:value="username" placeholder="root" />
        </n-form-item>
        <n-form-item label="Password">
          <n-input v-model:value="password" type="password" placeholder="(optional if using key)" />
        </n-form-item>
        <n-form-item label="SSH Private Key">
          <n-input-group>
            <n-input v-model:value="privateKeyPath" placeholder="Path to private key" />
            <n-button @click="browseKeyFile" secondary>Browse</n-button>
          </n-input-group>
        </n-form-item>
      </n-form>
    </template>

    <template #footer>
      <n-space justify="end">
        <n-button @click="emit('update:show', false)">Cancel</n-button>
        <n-button type="primary" @click="handleConnect" :loading="connecting">
          Connect
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>
