<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  icon: string
  active: boolean
  stage?: string
  color?: 'blue' | 'cyan' | 'purple'
  duration?: number // Duration in ms
  completed?: boolean
}>()

const colorClasses = computed(() => {
  const colors = {
    blue: {
      gradient: 'from-blue-50 to-blue-100/50',
      border: 'border-blue-200',
      glow: 'shadow-blue-500/10',
      icon: 'text-blue-600',
      iconBg: 'bg-blue-100',
      pulse: 'bg-blue-500'
    },
    cyan: {
      gradient: 'from-cyan-50 to-cyan-100/50',
      border: 'border-cyan-200',
      glow: 'shadow-cyan-500/10',
      icon: 'text-cyan-600',
      iconBg: 'bg-cyan-100',
      pulse: 'bg-cyan-500'
    },
    purple: {
      gradient: 'from-purple-50 to-purple-100/50',
      border: 'border-purple-200',
      glow: 'shadow-purple-500/10',
      icon: 'text-purple-600',
      iconBg: 'bg-purple-100',
      pulse: 'bg-purple-500'
    }
  }
  return colors[props.color || 'blue']
})
</script>

<template>
  <div 
    class="realtime-card rounded-xl overflow-hidden transition-all duration-300"
    :class="[
      active 
        ? `bg-gradient-to-br ${colorClasses.gradient} border ${colorClasses.border} ${colorClasses.glow} shadow-lg scale-[1.02]`
        : 'bg-white border border-gray-200 shadow-sm scale-100'
    ]"
  >
    <!-- Header -->
    <div class="card-header p-4 border-b" :class="active ? colorClasses.border : 'border-gray-100'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div 
            class="w-10 h-10 rounded-xl flex items-center justify-center transition-all"
            :class="active ? `${colorClasses.iconBg} shadow-sm` : 'bg-gray-100'"
          >
            <div :class="[icon, 'text-xl', active ? colorClasses.icon : 'text-gray-400']" />
          </div>
          <div>
            <h3 class="font-bold text-sm" :class="active ? 'text-gray-900' : 'text-gray-500'">
              {{ title }}
            </h3>
            <p v-if="stage" class="text-xs text-gray-500 mt-0.5 font-medium">
              {{ stage }}
            </p>
          </div>
        </div>

        <!-- Status Indicator -->
        <div class="flex items-center gap-2">
          <template v-if="active && !completed">
            <div 
              class="w-2 h-2 rounded-full animate-pulse"
              :class="colorClasses.pulse"
            />
            <span class="text-xs font-medium text-gray-500">Processing...</span>
          </template>
          <template v-else-if="completed">
            <div class="w-5 h-5 rounded-full bg-green-100 flex items-center justify-center">
              <div class="i-carbon-checkmark text-xs text-green-600" />
            </div>
            <span v-if="duration" class="text-xs font-bold text-gray-500">{{ (duration / 1000).toFixed(2) }}s</span>
          </template>
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
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #d1d5db;
}
</style>
