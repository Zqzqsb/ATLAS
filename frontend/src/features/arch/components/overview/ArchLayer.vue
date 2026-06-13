<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS, type ArchLayer, type ArchNode } from '../../model/architecture'
import ArchNode_ from './ArchNode.vue'

const props = defineProps<{ layer: ArchLayer }>()
const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()

const a = computed(() => ACCENTS[props.layer.accent])
const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${props.layer.cols}, minmax(0, 1fr))`,
}))
</script>

<template>
  <section class="rounded-2xl border overflow-hidden" :class="a.surface">
    <div class="h-1" :class="a.bar" />
    <div class="px-4 pt-3 pb-3">
      <!-- Layer header -->
      <div class="flex items-center gap-2 mb-3">
        <div class="w-6 h-6 rounded-md flex-center" :class="a.iconBg">
          <div :class="[layer.icon, a.iconText, 'text-sm']" />
        </div>
        <span class="text-sm font-bold text-gray-700">{{ layer.title }}</span>
        <span v-if="layer.subtitle" class="text-xs text-gray-400">· {{ layer.subtitle }}</span>
      </div>

      <!-- Nodes grid -->
      <div class="grid gap-3" :style="gridStyle">
        <div
          v-for="node in layer.nodes"
          :key="node.id"
          :style="{ gridColumn: `span ${node.span ?? 1}` }"
        >
          <ArchNode_ :node="node" @select="(n, ev) => emit('select', n, ev)" />
        </div>
      </div>
    </div>
  </section>
</template>
