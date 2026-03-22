<template>
  <VirtualScroller
    :data="renderConnections"
    :size="size"
  >
    <template v-slot:before>
      <ConnectionCtrl />
    </template>
    <template v-slot="{ item }: { item: Connection }">
      <ConnectionCard :conn="item" />
    </template>
  </VirtualScroller>
</template>

<script setup lang="ts">
import { renderConnections } from '@/store/connections'
import { connectionCardLines } from '@/store/settings'
import type { Connection } from '@/types'
import { computed } from 'vue'
import VirtualScroller from '../common/VirtualScroller.vue'
import ConnectionCtrl from '../sidebar/ConnectionCtrl.tsx'
import ConnectionCard from './ConnectionCard'
const size = computed(() => {
  return connectionCardLines.value.length * 28 + 4
})
</script>
