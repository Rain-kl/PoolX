<template>
  <svg
    class="-mt-0.5 mr-0.5 inline-block align-middle"
    width="14"
    height="12"
    viewBox="0 0 14 12"
  >
    <rect
      x="0"
      y="8"
      width="3"
      height="4"
      rx="0.5"
      :fill="level >= 1 ? 'currentColor' : 'oklch(var(--bc) / 0.2)'"
    />
    <rect
      x="5"
      y="4"
      width="3"
      height="8"
      rx="0.5"
      :fill="level >= 2 ? 'currentColor' : 'oklch(var(--bc) / 0.2)'"
    />
    <rect
      x="10"
      y="0"
      width="3"
      height="12"
      rx="0.5"
      :fill="level >= 3 ? 'currentColor' : 'oklch(var(--bc) / 0.2)'"
    />
  </svg>
</template>

<script setup lang="ts">
import { NOT_CONNECTED } from '@/constant'
import { lowLatency, mediumLatency } from '@/store/settings'
import { computed } from 'vue'

const props = defineProps<{
  latency: number
}>()

const level = computed(() => {
  if (!props.latency || props.latency === NOT_CONNECTED) return 0
  if (props.latency < lowLatency.value) return 3
  if (props.latency < mediumLatency.value) return 2
  return 1
})
</script>
