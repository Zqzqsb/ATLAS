<script setup lang="ts">
import { computed } from 'vue'
import type { FlowDef } from '../../../model/flows'
import { getModule } from '../../../model/modules'
import DataflowStepper from '../DataflowStepper.vue'
import SectionHeading from '../sections/SectionHeading.vue'
import StrategySection from '../sections/StrategySection.vue'
import PromptSection from '../sections/PromptSection.vue'
import StorageSection from '../sections/StorageSection.vue'
import InsightSection from '../sections/InsightSection.vue'

const props = defineProps<{ flow: FlowDef }>()
const mod = computed(() => getModule(props.flow.id))
</script>

<template>
  <div class="space-y-10">
    <!-- 1. Dataflow (animated) -->
    <section>
      <SectionHeading icon="i-lucide-route" title="端到端 Dataflow" subtitle="数据如何一步步流过 onboarding 管线" accent="emerald" />
      <DataflowStepper :flow="flow" />
    </section>

    <template v-if="mod">
      <hr class="border-gray-100" />
      <!-- 2. Task registration & dispatch (small vs large DB) -->
      <StrategySection :data="mod.strategy" />

      <hr class="border-gray-100" />
      <!-- 3. Prompt engineering -->
      <PromptSection :data="mod.prompt" />

      <hr class="border-gray-100" />
      <!-- 4. Storage layout -->
      <StorageSection :data="mod.storage" />

      <hr class="border-gray-100" />
      <!-- 5. Design insights -->
      <InsightSection :data="mod.insights" />
    </template>
  </div>
</template>
