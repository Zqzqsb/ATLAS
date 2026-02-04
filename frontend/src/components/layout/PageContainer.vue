<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  title?: string
  subtitle?: string
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full'
  padding?: boolean
  center?: boolean
}>(), {
  maxWidth: 'full',
  padding: true,
  center: false
})

const maxWidthClass = computed(() => {
  const map: Record<string, string> = {
    sm: 'max-w-screen-sm',
    md: 'max-w-screen-md',
    lg: 'max-w-screen-lg',
    xl: 'max-w-screen-xl',
    '2xl': 'max-w-screen-2xl',
    full: 'max-w-full'
  }
  return map[props.maxWidth] || 'max-w-full'
})
</script>

<template>
  <div 
    class="min-h-[calc(100vh-3.5rem)] bg-gray-50 dark:bg-gray-950"
    :class="{ 'p-6': padding }"
  >
    <div 
      :class="[maxWidthClass, { 'mx-auto': center }]"
      class="w-full"
    >
      <!-- Page header -->
      <header v-if="title" class="mb-6">
        <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100">
          {{ title }}
        </h1>
        <p v-if="subtitle" class="mt-1 text-gray-500 dark:text-gray-400">
          {{ subtitle }}
        </p>
      </header>

      <!-- Slot header -->
      <slot name="header" />

      <!-- Main content -->
      <main>
        <slot />
      </main>

      <!-- Footer -->
      <slot name="footer" />
    </div>
  </div>
</template>
