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
        <div class="app min-h-screen bg-gray-50 text-gray-700">
          <AppHeader />
          <RouterView />

          <!-- Global Background Task Indicator -->
          <Transition name="slide-up">
            <div
              v-if="ctxGenStore.isMinimized && (ctxGenStore.isRunning || ctxGenStore.isComplete)"
              class="fixed bottom-6 right-6 z-[9999]"
            >
              <button
                class="group flex items-center gap-3 px-4 py-3 rounded-2xl shadow-2xl backdrop-blur-md border transition-all duration-300 hover:scale-105 cursor-pointer"
                :class="ctxGenStore.isRunning
                  ? 'bg-blue-600/90 border-blue-400/40 hover:bg-blue-500/95 shadow-blue-500/30'
                  : 'bg-green-600/90 border-green-400/40 hover:bg-green-500/95 shadow-green-500/30'"
                @click="ctxGenStore.restore()"
              >
                <!-- Animated spinner or checkmark -->
                <div v-if="ctxGenStore.isRunning" class="relative w-8 h-8 flex-shrink-0">
                  <svg class="w-8 h-8 animate-spin" viewBox="0 0 32 32" fill="none">
                    <circle cx="16" cy="16" r="13" stroke="rgba(255,255,255,0.2)" stroke-width="3" />
                    <path d="M16 3a13 13 0 0 1 13 13" stroke="white" stroke-width="3" stroke-linecap="round" />
                  </svg>
                  <span class="absolute inset-0 flex items-center justify-center text-white text-[10px] font-bold">
                    {{ ctxGenStore.overallProgress }}%
                  </span>
                </div>
                <div v-else class="w-8 h-8 flex items-center justify-center flex-shrink-0">
                  <div class="i-carbon-checkmark-filled text-xl text-white" />
                </div>

                <!-- Info -->
                <div class="flex flex-col items-start min-w-0">
                  <span class="text-white text-sm font-semibold truncate">
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
                  class="ml-2 w-6 h-6 rounded-full bg-white/20 hover:bg-white/40 flex items-center justify-center transition-colors"
                  @click.stop="ctxGenStore.reset()"
                >
                  <div class="i-carbon-close text-xs text-white" />
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
