<template>
  <div
    class="relative flex h-full flex-col overflow-y-auto"
    ref="scrollContainerRef"
    @scroll.passive="handleScroll"
  >
    <!-- 左侧菜单 -->
    <SettingsMenu
      ref="menuComponentRef"
      :menu-items="menuItems"
      :active-menu-key="activeMenuKey"
      @menu-click="handleMenuClick"
    />
    <!-- 右侧内容区域 -->
    <div
      class="grid grid-cols-1 gap-2 p-2"
      :style="padding"
    >
      <div class="flex flex-col gap-4">
        <div
          v-for="item in menuItems"
          :key="item.key"
          :id="`item-${item.key}`"
          :data-key="item.key"
          class="card"
        >
          <div
            class="settings-title mt-4 px-4"
            v-if="![SETTINGS_MENU_KEY.general, SETTINGS_MENU_KEY.backend].includes(item.key)"
          >
            {{ $t(item.label) }}
          </div>
          <component :is="item.component" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BackendSettings from '@/components/settings/BackendSettings.vue'
import ConnectionsSettings from '@/components/settings/ConnectionsSettings.vue'
import GeneralSettings from '@/components/settings/GeneralSettings.vue'
import OverviewSettings from '@/components/settings/OverviewSettings.vue'
import ProxiesSettings from '@/components/settings/ProxiesSettings.vue'
import SettingsMenu from '@/components/settings/SettingsMenu.vue'
import { usePaddingForViews } from '@/composables/paddingViews'
import { isSettingVisible } from '@/composables/settings'
import { SETTINGS_MENU_KEY } from '@/constant'
import { settingsMenuOrder } from '@/store/settings'
import {
  ArrowsRightLeftIcon,
  CubeTransparentIcon,
  GlobeAltIcon,
  HomeIcon,
  ServerIcon,
} from '@heroicons/vue/24/outline'
import { throttle } from 'lodash'
import type { Component } from 'vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

type MenuItem = {
  key: SETTINGS_MENU_KEY
  label: string
  icon: Component
  component: Component
}

const { padding } = usePaddingForViews()
const route = useRoute()

const menuComponentRef = ref<InstanceType<typeof SettingsMenu> | null>(null)
const scrollContainerRef = ref<HTMLDivElement>()
const menuItems = computed<MenuItem[]>(() => {
  const itemsMap = new Map<SETTINGS_MENU_KEY, MenuItem>([
    [
      SETTINGS_MENU_KEY.general,
      {
        key: SETTINGS_MENU_KEY.general,
        label: 'zashboardSettings',
        icon: HomeIcon,
        component: GeneralSettings,
      },
    ],
    [
      SETTINGS_MENU_KEY.overview,
      {
        key: SETTINGS_MENU_KEY.overview,
        label: 'overviewSettings',
        icon: CubeTransparentIcon,
        component: OverviewSettings,
      },
    ],
    [
      SETTINGS_MENU_KEY.backend,
      {
        key: SETTINGS_MENU_KEY.backend,
        label: 'backendSettings',
        icon: ServerIcon,
        component: BackendSettings,
      },
    ],
    [
      SETTINGS_MENU_KEY.proxies,
      {
        key: SETTINGS_MENU_KEY.proxies,
        label: 'proxySettings',
        icon: GlobeAltIcon,
        component: ProxiesSettings,
      },
    ],
    [
      SETTINGS_MENU_KEY.connections,
      {
        key: SETTINGS_MENU_KEY.connections,
        label: 'connectionSettings',
        icon: ArrowsRightLeftIcon,
        component: ConnectionsSettings,
      },
    ],
  ])

  // 根据 settingsMenuOrder 排序，并过滤隐藏的项
  return settingsMenuOrder.value
    .map((key) => itemsMap.get(key))
    .filter((item): item is MenuItem => item !== undefined && isSettingVisible(item.key))
})
const activeMenuKey = ref<SETTINGS_MENU_KEY>(menuItems.value[0]?.key || SETTINGS_MENU_KEY.general)

// 当 menuItems 变化时，如果当前激活的项被隐藏，则切换到第一个可见项
watch(
  menuItems,
  (newItems) => {
    if (newItems.length > 0) {
      if (!newItems.find((item) => item.key === activeMenuKey.value)) {
        activeMenuKey.value = newItems[0].key
      }
    } else {
      // 如果所有设置项都被隐藏，保持当前值（虽然不会显示）
      // 这种情况应该很少见，因为至少应该有一个设置项可见
    }
  },
  { immediate: true },
)
const getItemRef = (key: SETTINGS_MENU_KEY) => {
  return document.getElementById(`item-${key}`)
}

const isTriggerByClick = ref(false)
const timeoutId = ref<number>()

const handleMenuClick = (key: SETTINGS_MENU_KEY) => {
  activeMenuKey.value = key

  const index = menuItems.value.findIndex((item) => item.key === key)
  if (index !== -1) {
    isTriggerByClick.value = true
    clearTimeout(timeoutId.value)
    timeoutId.value = setTimeout(() => {
      isTriggerByClick.value = false
    }, 1000)
    const element = getItemRef(key)
    if (element && scrollContainerRef.value) {
      const containerRect = scrollContainerRef.value.getBoundingClientRect()
      const elementRect = element.getBoundingClientRect()
      const scrollTop = scrollContainerRef.value.scrollTop
      const targetScrollTop = scrollTop + elementRect.top - containerRect.top - 86

      scrollContainerRef.value.scrollTo({
        top: targetScrollTop,
        behavior: 'smooth',
      })
    }
  }
}

const scrollTop = ref(0)
const updateActiveMenuByScroll = () => {
  if (!scrollContainerRef.value || isTriggerByClick.value) return

  const containerRect = scrollContainerRef.value.getBoundingClientRect()
  const newScrollTop = scrollContainerRef.value.scrollTop
  const containerCenter =
    containerRect.top + containerRect.height * (newScrollTop > scrollTop.value ? 0.8 : 0.3)

  let minDistance = Infinity
  let closestKey: SETTINGS_MENU_KEY | null = null

  menuItems.value.forEach((item) => {
    const element = getItemRef(item.key)
    if (!element) return

    const elementRect = element.getBoundingClientRect()
    const elementCenter = elementRect.top + elementRect.height / 2
    const distance = Math.abs(elementCenter - containerCenter)

    if (distance < minDistance) {
      minDistance = distance
      closestKey = item.key
    }
  })

  // 如果找到了最近的元素，更新激活菜单
  if (closestKey && closestKey !== activeMenuKey.value) {
    activeMenuKey.value = closestKey
  }

  scrollTop.value = newScrollTop
}

const handleScroll = throttle(updateActiveMenuByScroll, 100)

onMounted(() => {
  requestAnimationFrame(async () => {
    const scrollTo = route.query.scrollTo as SETTINGS_MENU_KEY
    if (scrollTo) {
      handleMenuClick(scrollTo)
    }
  })
})
</script>
