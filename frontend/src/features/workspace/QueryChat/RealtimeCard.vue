<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  icon: string
  active: boolean
  stage?: string
  color?: 'blue' | 'cyan' | 'purple'
}>()

const colorClasses = computed(() => {
  const colors = {
    blue: {
      gradient: 'from-blue-600/20 to-cyan-600/20',
      border: 'border-blue-500/30',
      glow: 'shadow-blue-500/20',
      icon: 'text-blue-400',
      pulse: 'bg-blue-400'
    },
    cyan: {
      gradient: 'from-cyan-600/20 to-teal-600/20',
      border: 'border-cyan-500/30',
      glow: 'shadow-cyan-500/20',
      icon: 'text-cyan-400',
      pulse: 'bg-cyan-400'
    },
    purple: {
      gradient: 'from-purple-600/20 to-pink-600/20',
      border: 'border-purple-500/30',
      glow: 'shadow-purple-500/20',
      icon: 'text-purple-400',
      pulse: 'bg-purple-400'
    }
  }
  return colors[props.color || 'blue']
})
</script>

<template>
  <div 
    class="realtime-card rounded-xl overflow-hidden transition-all duration-500 backdrop-blur-md"
    :class="[
      active 
        ? `bg-gradient-to-br ${colorClasses.gradient} border-2 ${colorClasses.border} ${colorClasses.glow} shadow-2xl scale-[1.02]`
        : 'bg-gradient-to-br from-gray-800/30 to-gray-900/30 border border-white/10 scale-100'
    ]"
  >
    <!-- Header -->
    <div class="card-header p-4 border-b" :class="active ? 'border-white/20' : 'border-white/10'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div 
            class="w-10 h-10 rounded-xl flex items-center justify-center backdrop-blur-md transition-all"
            :class="active ? `bg-white/10 ${colorClasses.glow} shadow-lg` : 'bg-white/5'"
          >
            <div :class="[icon, 'text-xl', active ? colorClasses.icon : 'text-gray-500']" />
          </div>
          <div>
            <h3 class="font-semibold" :class="active ? 'text-white' : 'text-gray-400'">
              {{ title }}
            </h3>
            <p v-if="stage" class="text-xs text-gray-500 mt-0.5">
              {{ stage }}
            </p>
          </div>
        </div>

        <!-- Status Indicator -->
        <div v-if="active" class="flex items-center gap-2">
          <div 
            class="w-2 h-2 rounded-full animate-pulse"
            :class="colorClasses.pulse"
          />
          <span class="text-xs text-gray-400">Processing...</span>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="card-content p-4 max-h-[400px] overflow-y-auto custom-scrollbar">
      <slot name="content" />
    </div>
  </div>
</template>

<style scoped>
.realtime-card {
  min-height: 200px;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

.card-header {
  background: linear-gradient(to bottom, rgba(255, 255, 255, 0.03), transparent);
}
</style>
