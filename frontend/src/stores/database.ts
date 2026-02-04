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

  async function addDatabase(config: DatabaseConfig): Promise<Database | null> {
    loading.value = true
    error.value = null
    try {
      const newDb = await databaseApi.create(config)
      databases.value.push(newDb)
      return newDb
    } catch (e: any) {
      error.value = e.message || 'Failed to add database'
      return null
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
    testConnection,
    getDatabaseById,
    updateDatabaseStatus
  }
})
