<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isVisible = ref(false)
const currentStep = ref(-1) // -1=idle, 0-4=steps
const containerRef = ref<HTMLElement>()

const steps = [
  {
    id: 'detect',
    title: 'DDL Change Detection',
    desc: 'Agent monitors schema changes in real-time',
    icon: 'i-atlas-scan-search',
    color: 'amber',
    gradientFrom: 'from-amber-500',
    gradientTo: 'to-orange-500',
    bgColor: 'bg-amber-50',
    borderColor: 'border-amber-300',
    ddl: 'ALTER TABLE orders ADD COLUMN refund_reason VARCHAR(255);',
    detail: 'DDL watcher detected new column "refund_reason" in table "orders"',
  },
  {
    id: 'mark',
    title: 'Mark Stale Context',
    desc: 'Flag affected contexts as outdated',
    icon: 'i-atlas-clock',
    color: 'red',
    gradientFrom: 'from-red-500',
    gradientTo: 'to-rose-500',
    bgColor: 'bg-red-50',
    borderColor: 'border-red-300',
    ddl: null,
    detail: 'Marked 3 related contexts as stale: orders.status mapping, orders table description, orders-payments relation',
  },
  {
    id: 'refresh',
    title: 'LLM Context Refresh',
    desc: 'Regenerate descriptions with new schema',
    icon: 'i-atlas-wand-2',
    color: 'violet',
    gradientFrom: 'from-violet-500',
    gradientTo: 'to-purple-500',
    bgColor: 'bg-violet-50',
    borderColor: 'border-violet-300',
    ddl: null,
    detail: 'LLM regenerated: "orders table now includes refund_reason for tracking refund justifications"',
  },
  {
    id: 'embed',
    title: 'Re-embed Vectors',
    desc: 'Update HNSW index with refreshed content',
    icon: 'i-atlas-radar',
    color: 'blue',
    gradientFrom: 'from-blue-500',
    gradientTo: 'to-cyan-500',
    bgColor: 'bg-blue-50',
    borderColor: 'border-blue-300',
    ddl: null,
    detail: 'Updated 3 embeddings in rc_embeddings (dim=1536), HNSW index rebuilt',
  },
  {
    id: 'verify',
    title: 'Closed-Loop Verification',
    desc: 'Validate updated context with test queries',
    icon: 'i-atlas-shield-check',
    color: 'emerald',
    gradientFrom: 'from-emerald-500',
    gradientTo: 'to-teal-500',
    bgColor: 'bg-emerald-50',
    borderColor: 'border-emerald-300',
    ddl: null,
    detail: 'Verification passed: "Find refund reasons for cancelled orders" → correct SQL generated',
  },
]

// Change log entries (appear as steps complete)
const changeLogs = ref<{ step: string; time: string; status: string }[]>([])

let observer: IntersectionObserver | null = null
let animTimer: number | null = null

function startAnimation() {
  changeLogs.value = []
  currentStep.value = 0
  const times = ['10:30:01', '10:30:02', '10:30:05', '10:30:07', '10:30:09']

  let step = 0
  animTimer = window.setInterval(() => {
    const s = steps[step]
    const t = times[step]
    if (!s || !t) {
      if (animTimer) clearInterval(animTimer)
      return
    }
    changeLogs.value.push({
      step: s.title,
      time: t,
      status: step < 4 ? '✓' : '✅',
    })
    step++
    if (step < steps.length) {
      currentStep.value = step
    } else {
      if (animTimer) clearInterval(animTimer)
    }
  }, 1800)
}

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      if (entry && entry.isIntersecting && !isVisible.value) {
        isVisible.value = true
        setTimeout(startAnimation, 600)
      }
    },
    { threshold: 0.2 }
  )
  if (containerRef.value) observer.observe(containerRef.value)
})

onUnmounted(() => {
  observer?.disconnect()
  if (animTimer) clearInterval(animTimer)
})
</script>

<template>
  <div ref="containerRef" class="relative">
    <!-- Section header -->
    <div class="text-center mb-6">
      <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-amber-50 border border-amber-200 mb-3">
        <div class="i-atlas-bot text-amber-600" />
        <span class="text-sm font-semibold text-amber-700">Innovation #4</span>
      </div>
      <h3 class="text-3xl font-bold text-gray-900 mb-2">Agent Self-Maintaining</h3>
      <p class="text-lg text-gray-500 max-w-2xl mx-auto">
        Schema changes? The agent detects, repairs, and verifies — all autonomously, zero human intervention
      </p>
    </div>

    <div class="max-w-5xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Left: DDL trigger -->
      <div class="lg:col-span-1">
        <div class="rounded-2xl border-2 border-amber-200 bg-amber-50/50 p-4 mb-3">
          <div class="flex items-center gap-2 mb-3">
            <div class="i-atlas-terminal text-amber-600" />
            <span class="font-semibold text-gray-800 text-sm">DDL Change Event</span>
          </div>
          <div
            class="font-mono text-xs bg-gray-900 text-green-400 rounded-lg p-3 transition-all duration-500"
            :class="isVisible ? 'opacity-100' : 'opacity-0'"
          >
            <div class="text-gray-500 mb-1">mysql></div>
            <div class="leading-relaxed">
              <span class="text-yellow-400">ALTER TABLE</span> orders<br>
              <span class="text-yellow-400">ADD COLUMN</span> refund_reason<br>
              <span class="text-cyan-400">VARCHAR</span>(255);
            </div>
            <div class="text-gray-500 mt-2">Query OK, 0 rows affected</div>
          </div>
        </div>

        <!-- Change Log panel -->
        <div class="rounded-2xl border-2 border-gray-200 bg-white p-4">
          <div class="flex items-center gap-2 mb-3">
            <div class="i-atlas-scroll-text text-gray-500" />
            <span class="font-semibold text-gray-800 text-sm">Change Log</span>
          </div>
          <div class="space-y-2 max-h-[260px] overflow-y-auto">
            <TransitionGroup name="log-list">
              <div
                v-for="(log, idx) in changeLogs"
                :key="idx"
                class="flex items-center gap-2 px-3 py-2 rounded-lg bg-gray-50 border border-gray-100 text-xs"
              >
                <span class="text-gray-400 font-mono">{{ log.time }}</span>
                <span class="text-gray-700 flex-1">{{ log.step }}</span>
                <span>{{ log.status }}</span>
              </div>
            </TransitionGroup>
            <div v-if="changeLogs.length === 0" class="text-center text-gray-400 text-xs py-4">
              Waiting for events...
            </div>
          </div>
        </div>
      </div>

      <!-- Right: 5-step pipeline -->
      <div class="lg:col-span-2">
        <div class="space-y-2">
          <div
            v-for="(step, idx) in steps"
            :key="step.id"
            class="relative"
          >
            <!-- Step card -->
            <div
              class="flex items-start gap-3 px-4 py-2.5 rounded-2xl border-2 transition-all duration-500"
              :class="[
                currentStep === idx
                  ? `${step.borderColor} ${step.bgColor} shadow-lg scale-[1.02]`
                  : currentStep > idx
                    ? 'border-gray-200 bg-gray-50/50 opacity-70'
                    : 'border-gray-200 bg-white opacity-50',
              ]"
            >
              <!-- Step number + icon -->
              <div class="flex flex-col items-center gap-0.5 shrink-0">
                <div
                  class="w-9 h-9 rounded-xl bg-gradient-to-br flex-center text-white transition-all duration-300"
                  :class="[
                    currentStep >= idx
                      ? `${step.gradientFrom} ${step.gradientTo}`
                      : 'from-gray-300 to-gray-400',
                  ]"
                >
                  <div v-if="currentStep > idx" class="i-atlas-check text-lg" />
                  <div v-else-if="currentStep === idx" :class="step.icon" class="text-lg animate-pulse" />
                  <div v-else :class="step.icon" class="text-lg" />
                </div>
                <span class="text-xs font-mono text-gray-400">{{ idx + 1 }}/5</span>
              </div>

              <!-- Content -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <h5 class="font-bold text-gray-800">{{ step.title }}</h5>
                  <span
                    v-if="currentStep === idx"
                    class="px-2 py-0.5 rounded-full text-xs font-semibold animate-pulse"
                    :class="`${step.bgColor} ${step.borderColor} border`"
                  >
                    Processing...
                  </span>
                  <span
                    v-else-if="currentStep > idx"
                    class="px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-700 border border-emerald-200"
                  >
                    Done
                  </span>
                </div>
                <p class="text-sm text-gray-500 mb-1">{{ step.desc }}</p>

                <!-- Detail text (visible when active or completed) -->
                <div
                  v-if="currentStep >= idx"
                  class="text-xs text-gray-600 bg-white/60 rounded-lg px-3 py-2 border border-white/80 transition-all duration-500"
                  :class="currentStep === idx ? 'opacity-100' : 'opacity-60'"
                >
                  {{ step.detail }}
                </div>
              </div>
            </div>

            <!-- Connector line -->
            <div
              v-if="idx < steps.length - 1"
              class="absolute left-[26px] -bottom-2 w-0.5 h-2 transition-all duration-300"
              :class="currentStep > idx ? 'bg-emerald-300' : 'bg-gray-200'"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-list-enter-active {
  transition: all 0.4s ease;
}
.log-list-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
