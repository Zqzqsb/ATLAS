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
  <section class="rounded-2xl border overflow-hidden shadow-sm" :class="a.surface">
    <div class="h-1.5 bg-gradient-to-r" :class="a.gradient" />
    <div class="px-4 pt-3 pb-3">
      <!-- Layer header -->
      <div class="flex items-center gap-2 mb-3">
        <div class="w-7 h-7 rounded-lg flex-center bg-gradient-to-br text-white shadow-sm" :class="a.gradient">
          <div :class="[layer.icon, 'text-sm text-white']" />
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
