/** Shared types for evidence-chip linking on black-box vendor decks. */
export interface SourceRef {
  id: string
  title: string
  url: string
  type: string
}

export type SourceCatalog = Record<string, SourceRef>
