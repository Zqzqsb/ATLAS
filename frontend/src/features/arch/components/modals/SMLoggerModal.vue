<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Change Logger</h3>
    <p class="text-sm text-gray-500 mb-3">ChangeLogger writes audit entries to rc_change_log. Captures old_value → new_value, change_type, entity_type, reason, and changed_by.</p>
    <p class="text-xs text-gray-400 font-mono bg-gray-50 rounded p-2">Example: change_type: "ddl_add_column", entity_type: "column", entity_name: "orders.status", new_value: "VARCHAR(50)", reason: "DDL detected by agent"</p>
  </div>
</template>
