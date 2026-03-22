<template>
  <div class="flex flex-1 items-center gap-2 truncate pb-1">
    <template v-if="proxyGroup.now">
      <Component
        class="h-4 w-4 shrink-0 outline-none"
        :is="isFixed ? LockClosedIcon : ArrowRightCircleIcon"
        @mouseenter="tipForFixed"
      />
      <div
        v-if="isNowAGroup"
        class="hover:bg-base-300 -ml-1 flex cursor-pointer items-center gap-1 rounded-lg px-2 hover:shadow-sm"
        @click.stop="scrollToGroup(proxyGroup.now)"
      >
        <ProxyName
          :name="proxyGroup.now"
          class="text-base-content/80 text-xs md:text-sm"
        />
      </div>
      <ProxyName
        v-else
        :name="proxyGroup.now"
        class="text-base-content/80 text-xs md:text-sm"
      />
      <template v-if="finalOutbound && displayFinalOutbound">
        <ArrowRightCircleIcon class="h-4 w-4 shrink-0" />
        <ProxyName
          :name="finalOutbound"
          class="text-base-content/80 text-xs md:text-sm"
        />
      </template>
    </template>
    <template v-else-if="proxyGroup.type.toLowerCase() === PROXY_TYPE.LoadBalance">
      <CheckCircleIcon class="h-4 w-4 shrink-0" />
      <span class="text-base-content/80 text-xs md:text-sm">
        {{ $t('loadBalance') }}
      </span>
    </template>
  </div>
</template>

<script setup lang="ts">
import { PROXY_TYPE } from '@/constant'
import { useTooltip } from '@/helper/tooltip'
import { scrollToGroup } from '@/helper/utils'
import { getNowProxyNodeName, proxyGroupList, proxyMap } from '@/store/proxies'
import { displayFinalOutbound } from '@/store/settings'
import { ArrowRightCircleIcon, CheckCircleIcon, LockClosedIcon } from '@heroicons/vue/24/outline'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ProxyName from './ProxyName.vue'

const props = defineProps<{
  name: string
  mobile?: boolean
}>()
const proxyGroup = computed(() => proxyMap.value[props.name])
const { showTip } = useTooltip()
const { t } = useI18n()

const isFixed = computed(() => {
  return proxyGroup.value.fixed === proxyGroup.value.now
})

const tipForFixed = (e: Event) => {
  if (!isFixed.value) {
    return
  }

  showTip(e, t('tipForFixed', { type: proxyGroup.value.type }), {
    delay: [500, 0],
  })
}

const isNowAGroup = computed(() => {
  return proxyGroupList.value.includes(proxyGroup.value.now)
})

const finalOutbound = computed(() => {
  const now = getNowProxyNodeName(proxyGroup.value.now)

  if (now === proxyGroup.value.now) {
    return ''
  }

  return now
})
</script>
