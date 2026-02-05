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
    class="realtime-card rounded-xl overflow-hidden transition-all duration-500 ease-out"
    :class="[
      active 
        ? `bg-gradient-to-br ${colorClasses.gradient} border-2 ${colorClasses.border} ${colorClasses.glow} shadow-xl scale-[1.02]`
        : completed
          ? 'bg-white border-2 border-green-200 shadow-md shadow-green-500/5 scale-100'
          : 'bg-white border border-gray-200 shadow-sm scale-100 opacity-80'
    ]"
  >
    <!-- Header -->
    <div 
      class="card-header p-4 border-b transition-colors duration-300" 
      :class="active ? colorClasses.border : completed ? 'border-green-100' : 'border-gray-100'"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div 
            class="w-10 h-10 rounded-xl flex items-center justify-center transition-all duration-300"
            :class="active ? `${colorClasses.iconBg} shadow-sm` : completed ? 'bg-green-100 shadow-sm' : 'bg-gray-100'"
          >
            <div 
              :class="[
                icon, 
                'text-xl transition-colors duration-300', 
                active ? colorClasses.icon : completed ? 'text-green-600' : 'text-gray-400'
              ]" 
            />
          </div>
          <div>
            <h3 
              class="font-bold text-sm transition-colors duration-300" 
              :class="active || completed ? 'text-gray-900' : 'text-gray-500'"
            >
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
            <div class="flex items-center gap-1.5">
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" />
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" style="animation-delay: 0.2s" />
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" style="animation-delay: 0.4s" />
            </div>
            <span class="text-xs font-medium text-gray-500 ml-1">Processing</span>
          </template>
          <template v-else-if="completed">
            <div class="completed-badge flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-green-100">
              <div class="i-carbon-checkmark text-xs text-green-600" />
              <span v-if="duration" class="text-xs font-bold text-green-700">{{ (duration / 1000).toFixed(2) }}s</span>
              <span v-else class="text-xs font-bold text-green-700">Done</span>
            </div>
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

/* Processing dots animation */
.processing-dot {
  animation: processingPulse 1.2s ease-in-out infinite;
}

@keyframes processingPulse {
  0%, 100% {
    opacity: 0.3;
    transform: scale(0.8);
  }
  50% {
    opacity: 1;
    transform: scale(1);
  }
}

/* Completed badge animation */
.completed-badge {
  animation: badgeIn 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes badgeIn {
  from {
    opacity: 0;
    transform: scale(0.8) translateX(8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateX(0);
  }
}
</style>
