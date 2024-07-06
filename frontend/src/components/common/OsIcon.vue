<script setup>
import { computed } from 'vue'
import { 
    Linux, 
    Windows, 
    Apple, 
    Android, 
    Ubuntu, 
    Fedora, 
    Suse,
    Freebsd,
    QuestionCircleRegular

} from '@vicons/fa'

const props = defineProps({
    name: {
        type: [String],
        default: "",
    },
    size: {
        type: [Number, String],
        default: 16,
    },
})

const iconComponent = computed(() => {
    switch (props.name.toLowerCase()) {
        case 'linux':
            return Linux
        case 'ubuntu':
            return Ubuntu
        case 'fedora':
            return Fedora        
        case 'suse':
        case 'opensuse':
            return Suse
        case 'windows':
            return Windows
        case 'apple':
        case 'mac':
        case 'ios':
            return Apple
        case 'android':
            return Android
        case 'freebsd':
            return Freebsd
        default:
            return QuestionCircleRegular
    }
})

</script>

<template>
    <n-tooltip trigger="hover" placement="bottom" :show-arrow="false">
        <template #trigger>
            <n-icon :component="iconComponent" size="16" :depth="1" />
        </template>
        {{ props.name }}
    </n-tooltip>
  
</template>