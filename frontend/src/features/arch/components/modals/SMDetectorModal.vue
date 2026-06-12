<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Change Detector</h3>
    <p class="text-sm text-gray-500 mb-3">DDLDetector compares rc_tables/rc_columns against INFORMATION_SCHEMA on the business database. Detects: ADD/DROP TABLE, ADD/DROP/RENAME/MODIFY COLUMN.</p>
    <div class="text-sm text-gray-500 font-semibold mb-1">Mechanism</div>
    <p class="text-sm text-gray-500">Full diff: SELECT from rc_tables → compare to INFORMATION_SCHEMA.TABLES. Any mismatch → delta detected → marks is_expired=true on affected rows.</p>
  </div>
</template>
