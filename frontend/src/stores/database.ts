import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Database, DatabaseConfig, SchemaInfo } from '@/types'
import { databaseApi } from '@/api'

export const useDatabaseStore = defineStore('database', () => {
  // State
  const databases = ref<Database[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const databasesByType = computed(() => ({
    mariadb: databases.value.filter(db => db.type === 'mariadb'),
    sqlite: databases.value.filter(db => db.type === 'sqlite'),
    mysql: databases.value.filter(db => db.type === 'mysql'),
    postgresql: databases.value.filter(db => db.type === 'postgresql')
  }))

  const connectedDatabases = computed(() =>
    databases.value.filter(db => db.status === 'connected')
  )

  const databasesWithContext = computed(() =>
    databases.value.filter(db => db.hasRichContext)
  )

  const databaseCount = computed(() => databases.value.length)

  // Actions
  async function fetchDatabases() {
    loading.value = true
    error.value = null
    try {
      databases.value = await databaseApi.list()
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch databases'
      console.error('Failed to fetch databases:', e)
    } finally {
      loading.value = false
    }
  }

  async function addDatabase(config: DatabaseConfig): Promise<{ success: boolean; error?: string }> {
    loading.value = true
    error.value = null
    try {
      // Step 1: Add connection (pure connection management)
      await databaseApi.create(config)
      // Step 2: Sync schema (creates rc_datasources + discovers schema)
      try {
        await databaseApi.syncConnectionSchema(config.name)
      } catch (syncErr: any) {
        console.warn('Schema sync failed after connection:', syncErr)
        // Connection succeeded but sync failed — still consider success
      }
      // Step 3: Refresh list to show updated data
      await fetchDatabases()
      return { success: true }
    } catch (e: any) {
      // Extract error message from backend response
      const msg = e.response?.data?.error || e.message || 'Failed to add database'
      error.value = msg
      return { success: false, error: msg }
    } finally {
      loading.value = false
    }
  }

  async function removeDatabase(id: string): Promise<boolean> {
    try {
      await databaseApi.delete(id)
      const index = databases.value.findIndex(db => db.id === id)
      if (index >= 0) {
        databases.value.splice(index, 1)
      }
      return true
    } catch (e: any) {
      error.value = e.message || 'Failed to remove database'
      return false
    }
  }

  async function deleteDatasource(lakebaseId: number): Promise<boolean> {
    try {
      await databaseApi.deleteDatasource(lakebaseId)
      // Remove from local state by matching lakebaseId
      const index = databases.value.findIndex(db => db.metadata?.lakebaseId === lakebaseId)
      if (index >= 0) {
        databases.value.splice(index, 1)
      }
      return true
    } catch (e: any) {
      error.value = e.message || 'Failed to delete datasource'
      return false
    }
  }

  async function syncSchema(lakebaseId: number): Promise<{ success: boolean; tables?: number; columns?: number }> {
    try {
      const result = await databaseApi.syncSchema(lakebaseId)
      // Refresh list to update counts
      await fetchDatabases()
      return { success: true, tables: result.tables, columns: result.columns }
    } catch (e: any) {
      error.value = e.message || 'Failed to sync schema'
      return { success: false }
    }
  }

  async function testConnection(id: string): Promise<{ success: boolean; message?: string }> {
    const result = await databaseApi.testConnection(id)

    // Update database status
    const db = databases.value.find(d => d.id === id)
    if (db) {
      db.status = result.success ? 'connected' : 'error'
    }

    return result
  }

  function getDatabaseById(id: string): Database | undefined {
    return databases.value.find(db => db.id === id)
  }

  function updateDatabaseStatus(id: string, status: Database['status']) {
    const db = databases.value.find(d => d.id === id)
    if (db) {
      db.status = status
    }
  }

  return {
    // State
    databases,
    loading,
    error,

    // Computed
    databasesByType,
    connectedDatabases,
    databasesWithContext,
    databaseCount,

    // Actions
    fetchDatabases,
    addDatabase,
    removeDatabase,
    deleteDatasource,
    syncSchema,
    testConnection,
    getDatabaseById,
    updateDatabaseStatus
  }
})
