/**
 * Snowflake deck — public evidence catalog.
 *
 * IDs / URLs mirror `WiseCat/.claude/skills/research/results/snowflake_cortex_analyst_semantic_views.yaml`.
 */
import type { SourceRef } from '../../arch/components/module/diagram/evidence-types'

export const SOURCES: Record<string, SourceRef> = {
  S1: {
    id: 'S1',
    title: 'Cortex Analyst',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst',
    type: 'official_doc',
  },
  S2: {
    id: 'S2',
    title: 'Overview of semantic views',
    url: 'https://docs.snowflake.com/en/user-guide/views-semantic/overview',
    type: 'official_doc',
  },
  S3: {
    id: 'S3',
    title: 'YAML specification for semantic views',
    url: 'https://docs.snowflake.com/en/user-guide/views-semantic/semantic-view-yaml-spec',
    type: 'official_doc',
  },
  S4: {
    id: 'S4',
    title: 'CREATE SEMANTIC VIEW',
    url: 'https://docs.snowflake.com/en/sql-reference/sql/create-semantic-view',
    type: 'official_doc',
  },
  S5: {
    id: 'S5',
    title: 'Cortex Analyst Verified Query Repository',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/verified-query-repository',
    type: 'official_doc',
  },
  S6: {
    id: 'S6',
    title: 'Improve literal search to enhance Cortex Analyst (Cortex Search integration)',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/cortex-analyst-search-integration',
    type: 'official_doc',
  },
  S7: {
    id: 'S7',
    title: 'Cortex Search Overview',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-search/cortex-search-overview',
    type: 'official_doc',
  },
  S8: {
    id: 'S8',
    title: 'Cortex Analyst REST API',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/rest-api',
    type: 'official_doc',
  },
  S9: {
    id: 'S9',
    title: 'Snowflake Cortex Analyst: Behind the Scenes',
    url: 'https://www.snowflake.com/en/blog/engineering/snowflake-cortex-analyst-behind-the-scenes/',
    type: 'official_blog',
  },
  S10: {
    id: 'S10',
    title: 'Best practices for semantic views',
    url: 'https://docs.snowflake.com/en/user-guide/views-semantic/best-practices-dev',
    type: 'official_doc',
  },
  S11: {
    id: 'S11',
    title: 'Semantic View Autopilot',
    url: 'https://docs.snowflake.com/en/user-guide/views-semantic/autopilot',
    type: 'official_doc',
  },
  S12: {
    id: 'S12',
    title: 'Suggestions for semantic models and views',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/verified-query-suggestions',
    type: 'official_doc',
  },
  S13: {
    id: 'S13',
    title: 'Snowflake-Labs/semantic-model-generator',
    url: 'https://github.com/Snowflake-Labs/semantic-model-generator',
    type: 'source_code',
  },
  S14: {
    id: 'S14',
    title: 'Custom instructions for Cortex Analyst',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/custom-instructions',
    type: 'official_doc',
  },
  S15: {
    id: 'S15',
    title: 'Querying a semantic view',
    url: 'https://docs.snowflake.com/en/user-guide/views-semantic/querying',
    type: 'official_doc',
  },
  S16: {
    id: 'S16',
    title: 'Cortex Analyst administrator monitoring',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/admin-observability',
    type: 'official_doc',
  },
  S17: {
    id: 'S17',
    title: 'Cortex Analyst semantic model specification',
    url: 'https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst/semantic-model-spec',
    type: 'official_doc',
  },
}
