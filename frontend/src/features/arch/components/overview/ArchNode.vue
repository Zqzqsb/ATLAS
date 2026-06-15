<script setup lang="ts">
import { computed, inject } from 'vue'
import { ACCENTS, type ArchNode } from '../../model/architecture'
import EvidenceChip from '../module/diagram/EvidenceChip.vue'
import { SOURCE_CATALOG_KEY } from './source-catalog'

const props = defineProps<{ node: ArchNode }>()
const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()

const a = computed(() => ACCENTS[props.node.accent])
const drillable = computed(() => !!props.node.flow)
const catalog = inject(SOURCE_CATALOG_KEY, null)

function onClick(ev: MouseEvent) {
  if (!drillable.value) return
  emit('select', props.node, ev)
}
</script>

<template>
  <component
    :is="drillable ? 'button' : 'div'"
    class="group relative w-full h-full text-left rounded-xl border bg-white pl-4 pr-3.5 py-3 transition-all duration-200 overflow-hidden"
    :class="[
      drillable
        ? `cursor-pointer border-gray-200 shadow-sm hover:-translate-y-0.5 hover:shadow-md ${a.hover}`
        : 'border-gray-200/70 cursor-default',
    ]"
    @click="onClick"
  >
    <div class="absolute left-0 top-0 bottom-0 w-1 bg-gradient-to-b" :class="a.gradient" />
    <div class="flex items-start gap-2.5">
      <div class="w-8 h-8 rounded-lg flex-center flex-shrink-0 bg-gradient-to-br text-white shadow-sm" :class="a.gradient">
        <div :class="[node.icon, 'text-base text-white']" />
      </div>
      <div class="flex-1 min-w-0">
        <div class="text-sm font-bold text-gray-800 leading-tight truncate">{{ node.label }}</div>
        <div v-if="node.sublabel" class="text-xs text-gray-400 mt-0.5 leading-snug">{{ node.sublabel }}</div>
        <EvidenceChip
          v-if="node.refs && node.refs.length && catalog"
          class="mt-1.5"
          :refs="node.refs"
          :catalog="catalog"
          size="xs"
        />
      </div>
    </div>

    <!-- drill affordance -->
    <div
      v-if="drillable"
      class="absolute top-2.5 right-2.5 flex items-center gap-1 text-xs font-semibold opacity-0 -translate-x-1 transition-all duration-200 group-hover:opacity-100 group-hover:translate-x-0"
      :class="a.text"
    >
      展开
      <div class="i-lucide-maximize-2 text-[11px]" />
    </div>
  </component>
</template>
