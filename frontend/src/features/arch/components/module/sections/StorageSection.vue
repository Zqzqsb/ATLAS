<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../model/architecture'
import { STORAGE_KIND_META, type StorageSection, type StorageKind } from '../../../model/modules'
import SectionHeading from './SectionHeading.vue'

const props = defineProps<{ data: StorageSection }>()

const groups = computed(() => {
  const order: StorageKind[] = ['schema', 'context', 'catalog', 'log']
  return order
    .map((kind) => ({ kind, meta: STORAGE_KIND_META[kind], items: props.data.items.filter((i) => i.kind === kind) }))
    .filter((g) => g.items.length > 0)
})
</script>

<template>
  <section>
    <SectionHeading icon="i-lucide-hard-drive" :title="data.title" :subtitle="data.subtitle" accent="indigo" />

    <div class="space-y-3">
      <div v-for="g in groups" :key="g.kind" class="rounded-2xl border bg-white overflow-hidden" :class="ACCENTS[g.meta.accent].surface">
        <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-100">
          <div :class="[g.meta.icon, ACCENTS[g.meta.accent].text, 'text-sm']" />
          <span class="text-xs font-bold text-gray-700">{{ g.meta.label }}</span>
        </div>
        <div class="divide-y divide-gray-100">
          <div v-for="item in g.items" :key="item.table" class="flex items-center gap-3 px-4 py-2.5">
            <code class="px-2 py-0.5 rounded-md bg-gray-900 text-gray-100 font-mono text-[11px] flex-shrink-0">{{ item.table }}</code>
            <div class="flex-1 min-w-0">
              <div class="text-xs font-semibold text-gray-800">{{ item.label }}</div>
              <div class="text-xs text-gray-400 truncate">{{ item.note }}</div>
            </div>
            <code v-if="item.spec" class="hidden md:block text-[10px] font-mono text-gray-400 flex-shrink-0">{{ item.spec }}</code>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
