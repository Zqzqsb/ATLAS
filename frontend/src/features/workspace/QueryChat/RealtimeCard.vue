<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  subtitle?: string // optional subtitle shown below title when completed
  icon: string
  active: boolean
  pending?: boolean // query started but this stage hasn't begun yet
  stage?: string
  color?: 'blue' | 'cyan' | 'purple'
  duration?: number // Duration in ms
  completed?: boolean
  stepNumber?: number // 1, 2, 3 — shown in idle/pending state
}>()

// Whether this card is in idle state (not active, not completed, not pending)
const isIdle = computed(() => !props.active && !props.completed && !props.pending)

// Show content only when there's something meaningful to display
const showContent = computed(() => props.active || props.completed)

const colorClasses = computed(() => {
  const colors = {
    blue: {
      gradient: 'from-blue-50/80 to-white',
      border: 'border-blue-200',
      glow: 'shadow-blue-500/10',
      icon: 'text-blue-600',
      iconBg: 'bg-blue-100',
      pulse: 'bg-blue-500',
      step: 'text-blue-400 bg-blue-50 border-blue-100',
    },
    cyan: {
      gradient: 'from-cyan-50/80 to-white',
      border: 'border-cyan-200',
      glow: 'shadow-cyan-500/10',
      icon: 'text-cyan-600',
      iconBg: 'bg-cyan-100',
      pulse: 'bg-cyan-500',
      step: 'text-cyan-400 bg-cyan-50 border-cyan-100',
    },
    purple: {
      gradient: 'from-purple-50/80 to-white',
      border: 'border-purple-200',
      glow: 'shadow-purple-500/10',
      icon: 'text-purple-600',
      iconBg: 'bg-purple-100',
      pulse: 'bg-purple-500',
      step: 'text-purple-400 bg-purple-50 border-purple-100',
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
        ? `bg-white border-2 ${colorClasses.border} shadow-md ${colorClasses.glow}`
        : completed
          ? 'bg-white border border-emerald-200 shadow-sm'
          : pending
            ? `bg-white border border-dashed ${colorClasses.border}`
            : 'bg-white border border-gray-200/60'
    ]"
  >
    <!-- Header -->
    <div 
      class="card-header px-4 transition-colors duration-200" 
      :class="[
        active 
          ? `${colorClasses.border} bg-gradient-to-r ${colorClasses.gradient} py-3 border-b` 
          : completed 
            ? 'border-emerald-100 bg-gradient-to-r from-emerald-50/60 to-white py-3 border-b' 
            : pending 
              ? 'bg-gray-50/30 py-2.5' 
              : 'py-2.5'
      ]"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <!-- Step number badge for idle/pending; icon for active/completed -->
          <div 
            v-if="isIdle && stepNumber"
            class="w-7 h-7 rounded-full flex items-center justify-center border text-xs font-bold"
            :class="colorClasses.step"
          >
            {{ stepNumber }}
          </div>
          <div 
            v-else
            class="w-8 h-8 rounded-lg flex items-center justify-center transition-all duration-200 shadow-sm"
            :class="active ? `${colorClasses.iconBg} ring-1 ring-${props.color || 'blue'}-200` : completed ? 'bg-emerald-50 ring-1 ring-emerald-200' : pending ? `${colorClasses.iconBg} opacity-60` : 'bg-gray-50 border border-gray-100'"
          >
            <div 
              :class="[
                icon, 
                'text-lg transition-colors duration-200', 
                active ? colorClasses.icon : completed ? 'text-emerald-600' : pending ? `${colorClasses.icon} opacity-60` : 'text-gray-400'
              ]" 
            />
          </div>
          <div>
            <h3 
              class="font-medium transition-colors duration-200" 
              :class="[
                active || completed ? 'text-gray-900 text-sm' : pending ? 'text-gray-500 text-sm' : 'text-gray-400 text-[13px]'
              ]"
            >
              {{ title }}
            </h3>
            <p v-if="subtitle && completed" class="text-xs text-gray-400 mt-0.5">
              {{ subtitle }}
            </p>
          </div>
        </div>

        <!-- Status Indicator -->
        <div class="flex items-center gap-2">
          <template v-if="active && !completed">
            <div class="flex items-center gap-1">
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" />
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" style="animation-delay: 0.2s" />
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" style="animation-delay: 0.4s" />
            </div>
            <span class="text-xs text-gray-400 ml-1">Processing</span>
          </template>
          <template v-else-if="completed">
            <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200/60 shadow-sm">
              <div class="i-lucide-check text-xs" />
              <span v-if="duration" class="text-xs font-semibold">{{ (duration / 1000).toFixed(2) }}s</span>
              <span v-else class="text-xs font-semibold">Done</span>
            </div>
          </template>
          <template v-else-if="pending">
            <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-gray-50 ring-1 ring-gray-200/60">
              <span class="pending-dot w-1 h-1 rounded-full bg-gray-300" />
              <span class="pending-dot w-1 h-1 rounded-full bg-gray-300" style="animation-delay: 0.3s" />
              <span class="pending-dot w-1 h-1 rounded-full bg-gray-300" style="animation-delay: 0.6s" />
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Content: only shown when active or completed -->
    <div v-if="showContent" class="card-content p-4 max-h-[560px] overflow-y-auto custom-scrollbar">
      <slot name="content" />
    </div>
  </div>
</template>

<style scoped>
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

/* Pending dots animation (slower, subtler) */
.pending-dot {
  animation: pendingPulse 1.8s ease-in-out infinite;
}

@keyframes pendingPulse {
  0%, 100% {
    opacity: 0.2;
    transform: scale(0.7);
  }
  50% {
    opacity: 0.7;
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
