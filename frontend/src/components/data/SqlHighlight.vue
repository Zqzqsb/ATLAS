<script setup lang="ts">
import { computed } from 'vue'
import { NCode, NScrollbar } from 'naive-ui'

const props = withDefaults(defineProps<{
  code: string
  language?: string
  maxHeight?: string
  showLineNumbers?: boolean
}>(), {
  language: 'sql',
  maxHeight: '300px',
  showLineNumbers: true
})

const formattedCode = computed(() => {
  // Basic SQL formatting
  if (props.language === 'sql') {
    return props.code
      .replace(/\bSELECT\b/gi, 'SELECT')
      .replace(/\bFROM\b/gi, '\nFROM')
      .replace(/\bWHERE\b/gi, '\nWHERE')
      .replace(/\bAND\b/gi, '\n  AND')
      .replace(/\bOR\b/gi, '\n  OR')
      .replace(/\bJOIN\b/gi, '\nJOIN')
      .replace(/\bLEFT JOIN\b/gi, '\nLEFT JOIN')
      .replace(/\bRIGHT JOIN\b/gi, '\nRIGHT JOIN')
      .replace(/\bINNER JOIN\b/gi, '\nINNER JOIN')
      .replace(/\bGROUP BY\b/gi, '\nGROUP BY')
      .replace(/\bORDER BY\b/gi, '\nORDER BY')
      .replace(/\bHAVING\b/gi, '\nHAVING')
      .replace(/\bLIMIT\b/gi, '\nLIMIT')
  }
  return props.code
})
</script>

<template>
  <div class="sql-highlight rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
    <NScrollbar :style="{ maxHeight }">
      <NCode
        :code="formattedCode"
        :language="language"
        :show-line-numbers="showLineNumbers"
        class="text-sm"
      />
    </NScrollbar>
  </div>
</template>

<style scoped>
.sql-highlight :deep(.n-code) {
  background: #1e1e1e;
  padding: 1rem;
}

.sql-highlight :deep(.n-code__line-numbers) {
  padding-right: 1rem;
  color: #6b7280;
}
</style>
