<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Verify &amp; Execute</h3>
    <p class="text-sm text-gray-500 mb-3">Final SQL verification and execution. DryRun (EXPLAIN) validates syntax and execution plan. ExecuteQuery runs the final SQL on the business database.</p>
    <div class="text-sm text-gray-500 font-semibold mb-1">Fallback</div>
    <p class="text-sm text-gray-500">If execution fails, the ReAct agent receives the error as an observation and can revise the SQL up to the iteration budget.</p>
  </div>
</template>
