<script setup lang="ts">
const steps = [
  "SELECT TABLE_NAME, TABLE_ROWS FROM information_schema.TABLES WHERE TABLE_TYPE = 'BASE TABLE'",
  'For each table: UPSERT INTO rc_tables (datasource_id, table_name, row_count)',
  'SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY FROM information_schema.COLUMNS',
  'For each column: UPSERT INTO rc_columns (type, nullable, is_pk, is_fk)',
  'SELECT FK relations FROM information_schema.KEY_COLUMN_USAGE',
  'For each FK: UPSERT INTO rc_relations (from_table, from_column, to_table, to_column)',
  'UPDATE rc_datasources.last_sync_at = NOW()',
]
</script>

<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Schema Sync</h3>
    <p class="text-sm text-gray-500 mb-4">SchemaSyncWorker — extracts physical schema from business DB via INFORMATION_SCHEMA</p>
    <div class="space-y-2">
      <div v-for="(s, i) in steps" :key="i" class="flex items-start gap-2 text-sm text-gray-600">
        <span class="text-emerald-500 font-mono flex-shrink-0">{{ i + 1 }}.</span>
        <span>{{ s }}</span>
      </div>
    </div>
  </div>
</template>
