<script setup lang="ts">
/**
 * InlineCode — tiny prose helper that renders text with `...` segments
 * highlighted as monospace pills (sky-blue for keys/idents).
 *
 *   Input:  "用 `is_calculated: true` + SQL 表达式"
 *   Output: 用 [is_calculated: true] + SQL 表达式
 *           where [..] is mono + bg-gray-100 + text-sky-700.
 */
import { computed } from 'vue'

const props = defineProps<{ text: string }>()

interface Seg { kind: 'text' | 'code'; value: string }

const segs = computed<Seg[]>(() => {
  const out: Seg[] = []
  const re = /`([^`\n]+)`/g
  let last = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(props.text)) !== null) {
    if (m.index > last) out.push({ kind: 'text', value: props.text.slice(last, m.index) })
    out.push({ kind: 'code', value: m[1]! })
    last = m.index + m[0].length
  }
  if (last < props.text.length) out.push({ kind: 'text', value: props.text.slice(last) })
  return out
})
</script>

<template><span><template v-for="(s, i) in segs" :key="i"><code
  v-if="s.kind === 'code'"
  class="px-1 py-0.5 mx-0.5 rounded text-[91%] font-mono bg-sky-50 text-sky-700 border border-sky-200/70 leading-none"
>{{ s.value }}</code><template v-else>{{ s.value }}</template></template></span></template>
