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
    class="min-h-[calc(100vh-4rem)] bg-gray-50"
    :class="{ 'p-8': padding }"
  >
    <div 
      :class="[maxWidthClass, { 'mx-auto': center }]"
      class="w-full transition-all duration-300 ease-in-out"
    >
      <!-- Page header -->
      <header v-if="title" class="mb-8 border-b border-gray-200 pb-6">
        <h1 class="text-3xl font-bold text-gray-900 tracking-tight">
          {{ title }}
        </h1>
        <p v-if="subtitle" class="mt-2 text-lg text-gray-500 font-medium max-w-3xl">
          {{ subtitle }}
        </p>
      </header>

      <!-- Slot header -->
      <slot name="header" />

      <!-- Main content -->
      <main class="space-y-6">
        <slot />
      </main>

      <!-- Footer -->
      <div class="mt-12 pt-6 border-t border-gray-200" v-if="$slots.footer">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>
