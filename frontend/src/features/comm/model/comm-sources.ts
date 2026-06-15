import type { SourceCatalog } from '../../arch/components/module/diagram/evidence-types'

/**
 * Shared SourceCatalog for the comm framework deck.
 *
 * For closed-source / managed vendors we can't deep-link into a codebase; we
 * link to official documentation, release notes, or whitepapers instead.
 * Each entry's `id` is referenced from VendorTake.refs.
 *
 * Naming: `<short-vendor>-<topic>` e.g. `sf-cortex-overview`, `dbx-mv-ref`.
 */
export const COMM_SOURCES: SourceCatalog = {
  /* ─── Snowflake Cortex Analyst / Semantic Views ─── */
  'sf-cortex-overview': {
    id: 'sf-cortex-overview',
    title: 'Cortex Analyst — overview',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst',
    type: 'official_doc',
  },
  'sf-semantic-views': {
    id: 'sf-semantic-views',
    title: 'Semantic Views — define & query',
    url: 'https://docs.snowflake.com/en/user-guide/views-semantic/overview',
    type: 'official_doc',
  },
  'sf-semantic-yaml': {
    id: 'sf-semantic-yaml',
    title: 'Semantic model YAML reference',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/semantic-model-spec',
    type: 'official_doc',
  },
  'sf-cortex-search': {
    id: 'sf-cortex-search',
    title: 'Cortex Search — managed retrieval',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-search/cortex-search-overview',
    type: 'official_doc',
  },
  'sf-vqr': {
    id: 'sf-vqr',
    title: 'Verified Query Repository (VQR)',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/verified-queries',
    type: 'official_doc',
  },

  /* ─── Databricks Unity Catalog / Metric Views / Genie ─── */
  'dbx-mv-ref': {
    id: 'dbx-mv-ref',
    title: 'Unity Catalog Metric Views',
    url: 'https://docs.databricks.com/aws/en/metric-views/',
    type: 'official_doc',
  },
  'dbx-genie': {
    id: 'dbx-genie',
    title: 'AI/BI Genie — overview',
    url: 'https://docs.databricks.com/aws/en/genie/',
    type: 'official_doc',
  },
  'dbx-uc-ai': {
    id: 'dbx-uc-ai',
    title: 'AI-generated table & column comments',
    url: 'https://docs.databricks.com/aws/en/comments/ai-comments',
    type: 'official_doc',
  },

  /* ─── Microsoft Fabric Data Agent ─── */
  'fabric-data-agent': {
    id: 'fabric-data-agent',
    title: 'Fabric Data Agent — overview',
    url: 'https://learn.microsoft.com/en-us/fabric/data-science/concept-data-agent',
    type: 'official_doc',
  },

  /* ─── Oracle Select AI ─── */
  'oracle-select-ai': {
    id: 'oracle-select-ai',
    title: 'Oracle Select AI',
    url: 'https://docs.oracle.com/en/cloud/paas/autonomous-database/serverless/adbsb/sql-generation-ai-autonomous.html',
    type: 'official_doc',
  },
  'oracle-ai-enrich': {
    id: 'oracle-ai-enrich',
    title: 'Oracle AI Catalog enrichment',
    url: 'https://docs.oracle.com/en-us/iaas/data-catalog/using/data-asset-enrichment.htm',
    type: 'official_doc',
  },

  /* ─── dbt Semantic Layer (公司开源, 但下钻到代码用 CodeRef; 文档放这里) ─── */
  'dbt-sl-overview': {
    id: 'dbt-sl-overview',
    title: 'dbt Semantic Layer — overview',
    url: 'https://docs.getdbt.com/docs/use-dbt-semantic-layer/dbt-sl',
    type: 'official_doc',
  },
}
