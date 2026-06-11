<script setup lang="ts">
import { NMessageProvider, NDialogProvider, NConfigProvider } from 'naive-ui'
import AppHeader from '@/components/layout/AppHeader.vue'
import { useContextGenerationStore } from '@/stores/contextGeneration'

const ctxGenStore = useContextGenerationStore()
</script>

<template>
  <NConfigProvider>
    <NMessageProvider>
      <NDialogProvider>
        <div class="app min-h-screen bg-slate-50 text-gray-700">
          <AppHeader />
          <RouterView />

          <!-- Global Background Task Indicator -->
          <Transition name="slide-up">
            <div
              v-if="ctxGenStore.isMinimized && (ctxGenStore.isRunning || ctxGenStore.isComplete)"
              class="fixed bottom-5 right-5 z-[9999]"
            >
              <button
                class="group flex items-center gap-3 px-4 py-2.5 rounded-lg shadow-lg border transition-all cursor-pointer"
                :class="ctxGenStore.isRunning
                  ? 'bg-primary-600 border-primary-500 hover:bg-primary-500'
                  : 'bg-emerald-600 border-emerald-500 hover:bg-emerald-500'"
                @click="ctxGenStore.restore()"
              >
                <!-- Animated spinner or checkmark -->
                <div v-if="ctxGenStore.isRunning" class="relative w-7 h-7 flex-shrink-0">
                  <svg class="w-7 h-7 animate-spin" viewBox="0 0 32 32" fill="none">
                    <circle cx="16" cy="16" r="13" stroke="rgba(255,255,255,0.2)" stroke-width="3" />
                    <path d="M16 3a13 13 0 0 1 13 13" stroke="white" stroke-width="3" stroke-linecap="round" />
                  </svg>
                  <span class="absolute inset-0 flex items-center justify-center text-white text-[10px] font-semibold">
                    {{ ctxGenStore.overallProgress }}%
                  </span>
                </div>
                <div v-else class="w-7 h-7 flex items-center justify-center flex-shrink-0">
                  <div class="i-lucide-check text-lg text-white" />
                </div>

                <!-- Info -->
                <div class="flex flex-col items-start min-w-0">
                  <span class="text-white text-sm font-medium truncate">
                    {{ ctxGenStore.isRunning ? 'Generating Context...' : 'Generation Complete' }}
                  </span>
                  <span class="text-white/60 text-xs truncate">
                    <template v-if="ctxGenStore.isRunning">
                      {{ ctxGenStore.formattedElapsed }} · {{ ctxGenStore.storageStats.tablesUpdated }}/{{ ctxGenStore.storageStats.tablesTotal }} tables
                    </template>
                    <template v-else>
                      Click to view results
                    </template>
                  </span>
                </div>

                <!-- Close button (dismiss for completed) -->
                <button
                  v-if="!ctxGenStore.isRunning"
                  class="ml-2 w-5 h-5 rounded bg-white/20 hover:bg-white/40 flex items-center justify-center transition-colors"
                  @click.stop="ctxGenStore.reset()"
                >
                  <div class="i-lucide-x text-xs text-white" />
                </button>
              </button>
            </div>
          </Transition>
        </div>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

<style>
.app {
  font-family: Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* Slide-up transition for floating indicator */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
