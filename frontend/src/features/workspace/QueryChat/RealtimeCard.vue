<script setup lang="ts">
import { computed, ref, watch, Transition } from 'vue'

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
  collapsible?: boolean // when true, completed cards can be collapsed (default: true)
}>()

// Whether this card is in idle state (not active, not completed, not pending)
const isIdle = computed(() => !props.active && !props.completed && !props.pending)

// Collapsible behavior — completed cards collapse by default
const isCollapsible = computed(() => props.collapsible !== false)
const isExpanded = ref(true) // starts expanded; auto-collapses on completion

// Auto-collapse when stage completes
watch(() => props.completed, (done) => {
  if (done && isCollapsible.value) {
    isExpanded.value = false
  }
})
// Auto-expand when stage becomes active
watch(() => props.active, (act) => {
  if (act) {
    isExpanded.value = true
  }
})

function toggleExpand() {
  if (props.completed && isCollapsible.value) {
    isExpanded.value = !isExpanded.value
  }
}

// Show content when active/completed, or show skeleton in idle
const showContent = computed(() => (props.active || props.completed) && isExpanded.value)
const showSkeleton = computed(() => isIdle.value)

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
    class="realtime-card rounded-2xl overflow-hidden transition-all duration-300"
    :class="[
      active 
        ? `bg-white border-2 ${colorClasses.border} shadow-lg ${colorClasses.glow}`
        : completed
          ? 'bg-white border border-emerald-200 shadow-sm'
          : pending
            ? `bg-white border border-dashed ${colorClasses.border}`
            : 'bg-white border border-gray-150 shadow-sm'
    ]"
  >
    <!-- Header -->
    <div 
      class="card-header px-5 transition-colors duration-200" 
      :class="[
        active 
          ? `${colorClasses.border} bg-gradient-to-r ${colorClasses.gradient} py-3.5 border-b` 
          : completed 
            ? `border-emerald-100 bg-gradient-to-r from-emerald-50/60 to-white py-3 border-b ${isCollapsible ? 'cursor-pointer select-none hover:from-emerald-50 hover:to-emerald-50/20' : ''}` 
            : pending 
              ? 'bg-gray-50/30 py-3' 
              : 'py-3.5'
      ]"
      @click="toggleExpand"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3 min-w-0">
          <!-- Step number badge for idle/pending; icon for active/completed -->
          <div 
            v-if="isIdle && stepNumber"
            class="w-7 h-7 rounded-full flex items-center justify-center border text-xs font-bold shadow-sm shrink-0"
            :class="colorClasses.step"
          >
            {{ stepNumber }}
          </div>
          <div 
            v-else
            class="w-8 h-8 rounded-lg flex items-center justify-center transition-all duration-200 shadow-sm shrink-0"
            :class="active ? `${colorClasses.iconBg} ring-1 ring-${props.color || 'blue'}-200` : completed ? 'bg-emerald-50 ring-1 ring-emerald-200' : pending ? `${colorClasses.iconBg} opacity-60` : 'bg-gray-50 border border-gray-100'"
          >
            <div 
              :class="[
                icon, 
                'text-base transition-colors duration-200', 
                active ? colorClasses.icon : completed ? 'text-emerald-600' : pending ? `${colorClasses.icon} opacity-60` : 'text-gray-400'
              ]" 
            />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2 flex-wrap">
              <h3 
                class="font-medium transition-colors duration-200" 
                :class="[
                  active || completed ? 'text-gray-900 text-[13px]' : pending ? 'text-gray-500 text-[13px]' : 'text-gray-400 text-xs'
                ]"
              >
                {{ title }}
              </h3>
              <!-- Inline summary (shown when collapsed or completed) -->
              <slot v-if="completed" name="summary" />
            </div>
            <p v-if="subtitle && completed && isExpanded" class="text-xs text-gray-400 mt-0.5">
              {{ subtitle }}
            </p>
          </div>
        </div>

        <!-- Status Indicator + Chevron -->
        <div class="flex items-center gap-2 shrink-0">
          <template v-if="active && !completed">
            <div class="flex items-center gap-1">
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" />
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" style="animation-delay: 0.2s" />
              <span class="processing-dot w-1.5 h-1.5 rounded-full" :class="colorClasses.pulse" style="animation-delay: 0.4s" />
            </div>
            <span class="text-xs text-gray-400 ml-1">Processing</span>
          </template>
          <template v-else-if="completed">
            <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200/60">
              <div class="i-lucide-check text-xs" />
              <span v-if="duration" class="text-[11px] font-semibold">{{ (duration / 1000).toFixed(2) }}s</span>
              <span v-else class="text-[11px] font-semibold">Done</span>
            </div>
            <!-- Chevron for collapse toggle -->
            <div 
              v-if="isCollapsible"
              class="i-lucide-chevron-down text-sm text-gray-400 transition-transform duration-200"
              :class="{ 'rotate-180': isExpanded }"
            />
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

    <!-- Content: only shown when active or completed AND expanded -->
    <Transition name="card-content">
    <div v-if="showContent" class="card-content px-5 py-4">
      <slot name="content" />
    </div>
    </Transition>

    <!-- Skeleton placeholder: shown in idle state with shimmer animation -->
    <div v-if="showSkeleton" class="skeleton-shimmer px-4 pb-3.5 pt-1.5">
      <slot name="skeleton">
        <!-- Default skeleton: 3 shimmer bars -->
        <div class="space-y-2">
          <div class="h-2 rounded-full skeleton-bar w-4/5" />
          <div class="h-2 rounded-full skeleton-bar w-3/5" />
          <div class="h-2 rounded-full skeleton-bar w-2/5" />
        </div>
      </slot>
    </div>
  </div>
</template>

<style scoped>
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

/* Skeleton shimmer animation — sweeping highlight from left to right */
.skeleton-shimmer {
  position: relative;
  overflow: hidden;
}

.skeleton-bar {
  background: linear-gradient(90deg, #f0f0f0 0%, #f0f0f0 100%);
  position: relative;
  overflow: hidden;
}

.skeleton-bar::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.6) 50%,
    transparent 100%
  );
  animation: shimmerSweep 2s ease-in-out infinite;
}

/* Shimmer for skeleton items that are NOT .skeleton-bar (tags, blocks etc) */
.skeleton-shimmer :deep(.skeleton-item) {
  position: relative;
  overflow: hidden;
}

.skeleton-shimmer :deep(.skeleton-item)::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.5) 50%,
    transparent 100%
  );
  animation: shimmerSweep 2s ease-in-out infinite;
}

@keyframes shimmerSweep {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

/* Card content reveal/collapse transition */
.card-content-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.card-content-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}
.card-content-enter-from {
  opacity: 0;
  max-height: 0;
  transform: translateY(-4px);
}
.card-content-leave-to {
  opacity: 0;
  max-height: 0;
  transform: translateY(-4px);
}
</style>
