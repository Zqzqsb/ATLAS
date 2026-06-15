<script setup lang="ts">
/**
 * EvidenceChip — tiny `[Sn]` reference chip that links to an external source
 * (official doc / blog / source code). Click opens the source URL in a new tab;
 * hover shows a popover with `Sn · title · type`. Multiple refs render as a
 * compact strip.
 *
 * For black-box vendors (Databricks / Snowflake) where we can't ground in a
 * codebase, every ArchBox / PeekPanel header pins these chips so the reader can
 * jump to the official evidence behind any claim.
 */
import { computed } from 'vue'
import { NPopover } from 'naive-ui'
import type { SourceRef } from './evidence-types'

const props = defineProps<{
  refs: string[]
  /** id → SourceRef map (each deck owns its own sources catalog) */
  catalog: Record<string, SourceRef>
  /** smaller / inline variant */
  size?: 'xs' | 'sm'
}>()

const TYPE_LABEL: Record<string, string> = {
  official_doc: '官方文档',
  official_blog: '官方博客',
  release_note: '发布说明',
  paper: '论文',
  source_code: '源码',
  demo: 'Demo',
  third_party_article: '第三方',
  community_discussion: '社区',
  benchmark: 'Benchmark',
  unknown: '未知',
}

const items = computed(() =>
  props.refs.map((id) => props.catalog[id]).filter(Boolean) as SourceRef[],
)

function open(url: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <span v-if="items.length" class="inline-flex items-center gap-0.5 flex-shrink-0">
    <NPopover
      v-for="s in items"
      :key="s.id"
      placement="top"
      trigger="hover"
      :delay="200"
      :width="260"
    >
      <template #trigger>
        <button
          type="button"
          class="inline-flex items-center font-mono font-bold rounded-[4px] border border-blue-200 bg-blue-50 text-blue-700 hover:bg-blue-100 hover:border-blue-300 transition-colors"
          :class="size === 'xs' ? 'text-[9px] leading-none px-1 py-0.5' : 'text-[10px] leading-none px-1.5 py-0.5'"
          @click.stop="open(s.url)"
        >
          {{ s.id }}
        </button>
      </template>
      <div class="space-y-1">
        <div class="flex items-center gap-1.5">
          <code class="text-[10px] font-mono font-bold px-1 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200">{{ s.id }}</code>
          <span class="text-[10px] font-semibold text-gray-400">{{ TYPE_LABEL[s.type] ?? s.type }}</span>
        </div>
        <div class="text-xs font-semibold text-gray-800 leading-snug">{{ s.title }}</div>
        <div class="text-[10px] text-blue-600 break-all leading-snug">{{ s.url }}</div>
        <div class="text-[10px] text-gray-400">点击在新标签页打开</div>
      </div>
    </NPopover>
  </span>
</template>
