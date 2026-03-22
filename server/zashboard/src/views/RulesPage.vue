<template>
  <div class="relative size-full overflow-x-hidden">
    <template v-if="!isVirtualScroller">
      <RulesCtrl />
      <div
        class="flex flex-col gap-1 p-2"
        :style="padding"
      >
        <template v-if="rulesTabShow === RULE_TAB_TYPE.PROVIDER">
          <RuleProvider
            v-for="(ruleProvider, index) in renderRulesProvider"
            :key="ruleProvider.name"
            :ruleProvider="ruleProvider"
            :index="index + 1"
          />
        </template>
        <template v-else>
          <RuleCard
            v-for="rule in renderRules"
            :key="rule.payload"
            :rule="rule"
            :index="rules.indexOf(rule) + 1"
          />
        </template>
      </div>
    </template>
    <VirtualScroller
      v-else
      :data="renderRules"
      :size="64"
    >
      <template v-slot:before>
        <RulesCtrl />
      </template>
      <template v-slot="{ item: rule }: { item: Rule }">
        <RuleCard
          :key="rule.payload"
          :rule="rule"
          :index="rules.indexOf(rule) + 1"
        />
      </template>
    </VirtualScroller>
  </div>
</template>

<script setup lang="ts">
import VirtualScroller from '@/components/common/VirtualScroller.vue'
import RuleCard from '@/components/rules/RuleCard.vue'
import RuleProvider from '@/components/rules/RuleProvider.vue'
import RulesCtrl from '@/components/sidebar/RulesCtrl.tsx'
import { usePaddingForViews } from '@/composables/paddingViews'
import { RULE_TAB_TYPE } from '@/constant'
import { fetchRules, renderRules, renderRulesProvider, rules, rulesTabShow } from '@/store/rules'
import type { Rule } from '@/types'
import { computed } from 'vue'

fetchRules()

const { padding } = usePaddingForViews({
  offsetTop: 8,
  offsetBottom: 8,
})
const isVirtualScroller = computed(() => {
  return rulesTabShow.value === RULE_TAB_TYPE.RULES && renderRules.value.length > 200
})
</script>
