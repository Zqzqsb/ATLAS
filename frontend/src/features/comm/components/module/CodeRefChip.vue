<script setup lang="ts">
/**
 * CodeRefChip — small, file-level link into a public codebase (e.g. WrenAI repo).
 *
 * Used by the comm framework's per-vendor "takes" to ground claims in actual
 * source files. We assume file-path accuracy at file granularity (the path
 * exists on `main`); line numbers are best-effort.
 *
 * For closed-source vendors (Snowflake / Databricks / Fabric …) use
 * EvidenceChip + SourceCatalog instead — those link to docs/blogs rather
 * than repos.
 */
import { computed } from 'vue'
import { NPopover } from 'naive-ui'
import { codeRefUrl, codeRefLabel, REPO_REGISTRY, type CodeRef } from '../../model/comm'

const props = defineProps<{
  refs: CodeRef[]
  /** shrink for inline-with-text usage */
  size?: 'xs' | 'sm'
}>()

interface Resolved {
  ref: CodeRef
  url: string | null
  text: string
  repoLabel: string
}

const items = computed<Resolved[]>(() =>
  props.refs.map((r) => ({
    ref: r,
    url: codeRefUrl(r),
    text: codeRefLabel(r),
    repoLabel: REPO_REGISTRY[r.repo]?.label ?? r.repo,
  })),
)

function open(url: string | null) {
  if (!url) return
  window.open(url, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <span v-if="items.length" class="inline-flex items-center flex-wrap gap-1">
    <NPopover
      v-for="(it, i) in items"
      :key="i"
      placement="top"
      trigger="hover"
      :delay="200"
      :width="320"
    >
      <template #trigger>
        <button
          type="button"
          class="inline-flex items-center gap-1 font-mono rounded border transition-colors max-w-full"
          :class="[
            it.url
              ? 'border-violet-200 bg-violet-50 text-violet-700 hover:bg-violet-100 hover:border-violet-300 cursor-pointer'
              : 'border-gray-200 bg-gray-50 text-gray-500 cursor-not-allowed',
            size === 'xs' ? 'text-[9.5px] leading-none px-1 py-0.5' : 'text-[10.5px] leading-none px-1.5 py-0.5',
          ]"
          :disabled="!it.url"
          @click.stop="open(it.url)"
        >
          <span class="i-lucide-file-code text-[10px] flex-shrink-0" />
          <span class="truncate font-semibold">{{ it.text }}</span>
          <span
            v-if="it.ref.lines"
            class="text-[9px] text-violet-500 font-normal"
          >L{{ it.ref.lines[0] }}</span>
        </button>
      </template>

      <div class="space-y-1.5">
        <div class="flex items-center gap-1.5">
          <code class="text-[10px] font-mono font-bold px-1 py-0.5 rounded bg-violet-50 text-violet-700 border border-violet-200">
            {{ it.repoLabel }}
          </code>
          <span v-if="!it.url" class="text-[10px] text-gray-400">私有仓库</span>
        </div>
        <div class="text-[11px] font-mono text-gray-700 break-all leading-snug">{{ it.ref.path }}</div>
        <div v-if="it.ref.lines" class="text-[10px] text-gray-500">
          行 {{ it.ref.lines[0] }} – {{ it.ref.lines[1] }}
        </div>
        <div v-if="it.url" class="text-[10px] text-violet-600 break-all leading-snug">{{ it.url }}</div>
        <div class="text-[10px] text-gray-400">
          {{ it.url ? '点击在 GitHub 打开' : '内部仓库 · 不可外链' }}
        </div>
      </div>
    </NPopover>
  </span>
</template>
