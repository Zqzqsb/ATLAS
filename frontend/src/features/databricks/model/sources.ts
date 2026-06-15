/**
 * Databricks deck — public evidence catalog.
 *
 * IDs and URLs mirror `WiseCat/.claude/skills/research/results/databricks_uc_metric_views.yaml`
 * so any chip in the deck is one click away from the same external source the
 * yaml evaluation cited. Add new sources here as the deck grows.
 */
import type { SourceRef } from '../../arch/components/module/diagram/evidence-types'

export const SOURCES: Record<string, SourceRef> = {
  S1: {
    id: 'S1',
    title: 'Unity Catalog business semantics',
    url: 'https://docs.databricks.com/aws/en/business-semantics/',
    type: 'official_doc',
  },
  S2: {
    id: 'S2',
    title: 'Unity Catalog metric views',
    url: 'https://docs.databricks.com/aws/en/business-semantics/metric-views/',
    type: 'official_doc',
  },
  S3: {
    id: 'S3',
    title: 'Agent metadata',
    url: 'https://docs.databricks.com/aws/en/business-semantics/agent-metadata',
    type: 'official_doc',
  },
  S4: {
    id: 'S4',
    title: 'Metric view YAML syntax reference',
    url: 'https://docs.databricks.com/aws/en/business-semantics/metric-views/yaml-reference',
    type: 'official_doc',
  },
  S5: {
    id: 'S5',
    title: 'Model metric views (basic modeling)',
    url: 'https://docs.databricks.com/aws/en/business-semantics/metric-views/basic-modeling',
    type: 'official_doc',
  },
  S6: {
    id: 'S6',
    title: 'Row filters and column masks (Unity Catalog)',
    url: 'https://docs.databricks.com/aws/en/data-governance/unity-catalog/filters-and-masks/',
    type: 'official_doc',
  },
  S7: {
    id: 'S7',
    title: 'Curate an effective Genie Space',
    url: 'https://docs.databricks.com/aws/en/genie/best-practices',
    type: 'official_doc',
  },
  S8: {
    id: 'S8',
    title: 'Tune Genie Space quality',
    url: 'https://docs.databricks.com/aws/en/genie/tune-quality',
    type: 'official_doc',
  },
  S9: {
    id: 'S9',
    title: 'Use a Genie Space to explore business data',
    url: 'https://docs.databricks.com/aws/en/genie/talk-to-genie',
    type: 'official_doc',
  },
  S10: {
    id: 'S10',
    title: 'Genie Spaces concepts',
    url: 'https://docs.databricks.com/aws/en/genie/concepts',
    type: 'official_doc',
  },
  S11: {
    id: 'S11',
    title: 'Lakehouse Federation',
    url: 'https://docs.databricks.com/aws/en/query-federation/',
    type: 'official_doc',
  },
  S12: {
    id: 'S12',
    title: 'Unity Catalog overview',
    url: 'https://docs.databricks.com/aws/en/data-governance/unity-catalog/',
    type: 'official_doc',
  },
  S13: {
    id: 'S13',
    title: 'Conversation API for Genie spaces',
    url: 'https://docs.databricks.com/api/workspace/genie',
    type: 'official_doc',
  },
  S14: {
    id: 'S14',
    title: 'Materialize a metric view',
    url: 'https://docs.databricks.com/aws/en/business-semantics/metric-views/materialize',
    type: 'official_doc',
  },
}
