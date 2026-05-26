<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { EventsOn, EventsOff } from 'wailsjs/runtime'
import { WriteSSHInput, ResizeSSHTerminal } from 'wailsjs/go/services/tailScaleService'
import '@xterm/xterm/css/xterm.css'

const props = defineProps({
  sessionId: { type: String, required: true },
  peerName: { type: String, required: true },
  active: { type: Boolean, default: true },
})

const emit = defineEmits(['close'])

const terminalContainer = ref(null)
let terminal = null
let fitAddon = null
let observer = null

const outputEvent = `ssh:output:${props.sessionId}`
const errorEvent = `ssh:error:${props.sessionId}`
const closedEvent = `ssh:closed:${props.sessionId}`

function doFit() {
  if (!fitAddon || !terminalContainer.value || !terminal) return
  const w = terminalContainer.value.clientWidth
  const h = terminalContainer.value.clientHeight
  if (w === 0 || h === 0) return
  fitAddon.fit()
  terminal.refresh(0, terminal.rows)
  const dims = fitAddon.proposeDimensions()
  if (dims && dims.cols > 0 && dims.rows > 0) {
    ResizeSSHTerminal(props.sessionId, dims.cols, dims.rows)
  }
}

onMounted(async () => {
  await nextTick()

  terminal = new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: { background: '#1e1e1e', foreground: '#d4d4d4' },
    allowTransparency: false,
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.open(terminalContainer.value)

  EventsOn(outputEvent, (encoded) => {
    try { terminal.write(atob(encoded)) } catch (e) { /* ignore */ }
  })
  EventsOn(errorEvent, (msg) => {
    terminal.writeln(`\r\n\x1b[31mError: ${msg}\x1b[0m`)
  })
  EventsOn(closedEvent, () => {
    terminal.writeln(`\r\n\x1b[33m--- Connection closed ---\x1b[0m`)
    terminal.options.disableStdin = true
    emit('close')
  })

  terminal.onData((data) => {
    WriteSSHInput(props.sessionId, data)
  })

  observer = new ResizeObserver(() => doFit())
  observer.observe(terminalContainer.value)

  await nextTick()
  await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)))
  doFit()
})

onBeforeUnmount(() => {
  if (observer) observer.disconnect()
  EventsOff(outputEvent)
  EventsOff(errorEvent)
  EventsOff(closedEvent)
  if (terminal) terminal.dispose()
})

watch(() => props.active, (isActive) => {
  if (isActive) {
    nextTick(() => {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => doFit())
      })
    })
  }
})
</script>

<template>
  <div ref="terminalContainer" class="terminal-container"></div>
</template>

<style scoped>
.terminal-container {
  width: 100%;
  height: 100vh;
  min-height: 500px;
  overflow: hidden;
}
</style>
