/**
 * Provide/inject key for the deck-level source-id → SourceRef catalog.
 *
 * Black-box decks (Databricks / Snowflake) provide their own catalog at the
 * overview root; ArchNode injects it (when present) to render `[Sn]` chips on
 * L0 cards. Lives in its own `.ts` so it can be imported anywhere without
 * pulling a Vue SFC, and so `<script setup>` doesn't need to re-export it
 * (which production builds reject).
 */
import type { InjectionKey } from 'vue'
import type { SourceCatalog } from '../module/diagram/evidence-types'

export const SOURCE_CATALOG_KEY: InjectionKey<SourceCatalog> = Symbol('source-catalog')
