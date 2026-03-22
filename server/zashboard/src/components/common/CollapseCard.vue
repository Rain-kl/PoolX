<template>
  <div :class="`collapse ${showCollapse ? 'collapse-open' : 'collapse-close'}`">
    <div
      class="collapse-title cursor-pointer pr-4"
      @click="showCollapse = !showCollapse"
    >
      <slot name="title" />
      <slot
        v-if="!showCollapse"
        name="preview"
      />
    </div>
    <div
      class="collapse-content p-0"
      @transitionend="handlerTransitionEnd"
    >
      <div
        v-if="showContent"
        class="max-h-108 overflow-y-auto p-4 pt-0 max-md:p-2"
        :class="[SCROLLABLE_PARENT_CLASS, !showCollapse && 'opacity-0']"
      >
        <slot name="content" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { SCROLLABLE_PARENT_CLASS } from '@/helper/utils'
import { collapseGroupMap } from '@/store/settings'
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  name: string
}>()

const showCollapse = computed({
  get() {
    return collapseGroupMap.value[props.name]
  },
  set(value) {
    collapseGroupMap.value[props.name] = value
  },
})

watch(showCollapse, (value) => {
  if (value) {
    showContent.value = true
  }
})

const showContent = ref(showCollapse.value)

const handlerTransitionEnd = () => {
  if (!showCollapse.value) {
    showContent.value = false
  }
}
</script>
