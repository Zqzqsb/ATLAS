import type { StageArch } from './comm'

export const contextArch: StageArch = {
  id: 'context',
  abstract: '"Context Layer" 是 4 个独立子决策的组合：以什么形态定义语义？怎么构建？存哪？怎么取回？各家差异基本都落在这 4 维。',
  principles: [
    { name: '语义 = 可评审的契约', desc: '不是隐藏在 LLM 提示词里的字符串——是 wiki / YAML / 知识图谱里可读、可 diff、可评审的资产。' },
    { name: '构建至少要"半自动"', desc: '纯手工建模在大 schema 下不可持续；纯自动会捕获错误概念。从 introspect 起步，让人评审 + LLM 增量补全。' },
    { name: '存储决定可移植性', desc: '存在产品 metadata DB → 锁死；存 git 化 YAML / Markdown → 可迁移。' },
    { name: '召回必须自适应', desc: '小 schema 全量塞进 prompt 最简单可靠；大 schema 才上向量；对短查询要兜底关键词。' },
  ],
  subQuestions: [
    {
      id: 'define-form',
      question: '"语义"以什么形态定义？',
      why: '形态决定了表达力上限、可评审性、被 LLM 消费的成本。',
      variants: [
        {
          name: '正式语义层 (YAML / DSL)',
          desc: 'models / measures / relationships / cubes · 可编译为 SQL',
          vendors: ['dbt Semantic Layer', 'Cube', 'WrenAI MDL', 'Databricks Metric Views'],
          accent: 'amber',
        },
        {
          name: 'Wiki + 列描述',
          desc: 'Markdown 业务知识 + 表/列 description；语义"靠描述讲清"',
          vendors: ['ktx', 'Oracle AI Enrichment'],
          accent: 'emerald',
        },
        {
          name: 'Rich Context 注释 (JSON/DB)',
          desc: '行/列粒度的上下文条目存仓库 DB；以表为单位维护',
          vendors: ['ATLAS (rc_business_context)'],
          accent: 'violet',
        },
        {
          name: '语义视图 + 集成 LLM',
          desc: '视图即语义；模型由数据平台内置统一管控',
          vendors: ['Snowflake Semantic Views', 'Fabric Business Semantics'],
          accent: 'blue',
        },
      ],
      commonSense:
        '正式语义层精度最高但"建模税"也最高；wiki + 列描述上手最快但精度依赖描述质量。**理想形态 = 语义层（指标 / 关系）+ wiki（业务规则、enum、单位）双载体**——结构化的留给引擎、自由文本的留给人。',
    },
    {
      id: 'build-path',
      question: '怎么"建"出来？',
      why: '建模成本是上下文层落地的最大阻力点；纯手工 / 纯自动两端都不可行。',
      variants: [
        { name: '纯手工 YAML', desc: '团队写、走 git 评审；语义最准', vendors: ['dbt SL · Cube'], accent: 'amber' },
        { name: 'Introspect + LLM 生成', desc: '从库 schema 推 + LLM 补描述 / 关系', vendors: ['WrenAI generate-mdl', 'ktx scan', 'ATLAS onboarding'], accent: 'violet' },
        { name: '从 dbt / LookML 导入', desc: '复用已有建模产出', vendors: ['dbt SL', 'WrenAI', 'ktx'], accent: 'emerald' },
        { name: 'dlt / SaaS 抽数', desc: 'HubSpot / Stripe / Salesforce → DuckDB → 建模', vendors: ['WrenAI dlt-connector'], accent: 'blue' },
        { name: '云内自动注释', desc: '云平台扫表自动写元数据 / AI 注释', vendors: ['Oracle AI Enrichment', 'Databricks UC'], accent: 'rose' },
      ],
      commonSense:
        '**先 introspect 起骨架，再用 LLM 补语义肉，再让人在 PR 上评审**——三段式落地阻力最小。任何一段省略都会出大问题。',
    },
    {
      id: 'storage',
      question: '上下文存在哪里？',
      why: '存储介质决定可迁移性、可审计性、备份恢复策略、以及语义谁说了算。',
      variants: [
        { name: 'Git 化 YAML / Markdown', desc: '语义层即代码；diff 评审；备份 = git', vendors: ['ktx', 'WrenAI', 'dbt SL'], accent: 'emerald' },
        { name: '应用 metadata DB', desc: '存平台自己的 DB；备份 / 迁移依赖产品', vendors: ['Cube (cloud)', 'Fabric'], accent: 'blue' },
        { name: '仓内向量表 + 关系表', desc: '语义存仓库自身（VECTOR / 表）', vendors: ['ATLAS Lakebase', 'Snowflake Semantic Views'], accent: 'violet' },
        { name: '混合', desc: '契约存 git，索引 / cache 存 sqlite', vendors: ['ktx (.ktx/db.sqlite)'], accent: 'amber' },
      ],
      commonSense:
        '**契约存 git，索引存本地 / 仓内**——前者保证可评审、可移植，后者保证检索快、增量更新。把契约存进产品 DB 是技术债。',
    },
    {
      id: 'recall',
      question: '怎么"取回"上下文？',
      why: '召回策略直接影响 SQL 生成质量和延迟；错的策略会让大 schema 完全不可用。',
      variants: [
        { name: '全量注入', desc: '小 schema (≤30k char) 整个塞 prompt', vendors: ['WrenAI (默认)'], accent: 'emerald' },
        { name: '向量召回', desc: '把 schema_items embed → cosine top-k', vendors: ['WrenAI (大 schema)', 'ATLAS', 'ktx'], accent: 'violet' },
        { name: 'FTS / BM25 关键词', desc: '名字命中精确，但泛化弱', vendors: ['ktx FTS5'], accent: 'amber' },
        { name: 'Hybrid + RRF 融合', desc: '关键词 + token + 向量三路 RRF', vendors: ['ktx HybridSearchCore'], accent: 'blue' },
        { name: '人工 / 工具白名单', desc: '只暴露指定模型 / 表给某 Agent', vendors: ['Snowflake Cortex', 'Cube row policies'], accent: 'rose' },
      ],
      commonSense:
        '**体量自适应 + 多路兜底**：小 schema 全量、大 schema 走 hybrid（FTS + 向量），对短查询尤其要让 FTS 兜底——单一向量召回在表名精确匹配时反而退化。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-git-pull-request',
      title: '语义即代码 = 上下文层成熟度的标志',
      body: '能否走 git diff 评审 / 复用 dbt 产出 / 在不同环境间迁移——这些"工程基本面"决定了上下文层的天花板，远大于"用了什么 embedding"。',
    },
    {
      icon: 'i-lucide-layers',
      title: '语义层 + Wiki 双载体',
      body: '指标 / 关系 / measure 形式化进语义层（让引擎执法）；enum / 单位 / 业务规则进 wiki（让 LLM 看懂）。强迫所有东西进同一种载体，会要么过度工程，要么写不出。',
    },
    {
      icon: 'i-lucide-recycle',
      title: '召回失败比 SQL 错误更危险',
      body: '召回错了，LLM 在错的上下文里写出"看起来对"的 SQL——这种错最难发现。检索召回率应被作为一级指标监控，而不是只看 SQL 准确率。',
    },
  ],
  matrix: {
    cols: ['契约形态', '构建主线', '存储', '召回'],
    rows: [
      { vendor: 'ATLAS', school: 'agentic', cells: ['Rich Context (JSON)', 'introspect+LLM', '仓内表 / 向量', '向量+精确'] },
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['MDL YAML', '手工/LLM/dbt/dlt', 'Git', '体量自适应+向量'] },
      { vendor: 'ktx', school: 'open-context', cells: ['SL YAML + Wiki', 'introspect+LLM+评审', 'Git + sqlite', 'Hybrid (RRF)'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['SL YAML', '手工', 'Git', '由消费方决定'] },
      { vendor: 'Cube', school: 'semantic-layer', cells: ['Cube DSL', '手工', 'Git / Cloud DB', 'AI API'] },
      { vendor: 'Snowflake Cortex', school: 'managed-cloud', cells: ['Semantic Views', '云内', '仓内', '内置 LLM'] },
      { vendor: 'Databricks UC', school: 'managed-cloud', cells: ['Metric Views', '云内', '仓内', '内置'] },
      { vendor: 'Oracle AI Enrich', school: 'managed-cloud', cells: ['列注释 / RC', '自动 AI', '仓内', '内置'] },
    ],
  },
}
