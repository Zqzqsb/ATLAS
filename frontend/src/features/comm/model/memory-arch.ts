import type { StageArch } from './comm'

export const memoryArch: StageArch = {
  id: 'memory',
  abstract: '"使用即沉淀（成功 → 知识 / 失败 → 任务）+ 解释（lineage / dry-plan）+ 权限（RLAC/CLAC）+ 漂移监控"——四件套缺一不可。',
  principles: [
    { name: '记忆是显式资产', desc: '不是黑盒向量；是 wiki Markdown / SL YAML / query_history 行——可审计、可 diff、可版本回滚。' },
    { name: '可解释 = 生成轨迹', desc: 'dry-plan / lineage 把 LLM 的选择展开为"用了哪些模型 / join / 列"；让数据团队能 review。' },
    { name: '权限是契约的一部分', desc: 'RLAC / CLAC 写在语义层，规划阶段强制执行——而不是在应用层补防火墙。' },
    { name: '漂移监控比 ML 重要', desc: 'schema 变了、表名换了、列删了——这些"软失败"远比模型退化常见，必须在数据层检测并主动失效相关 RC。' },
  ],
  subQuestions: [
    {
      id: 'learn',
      question: '"使用即沉淀"沉淀什么？',
      why: '沉淀对象决定下次召回的精度天花板。',
      variants: [
        { name: '确认的 NL-SQL 对', desc: 'few-shot 召回的源', vendors: ['WrenAI memory store', 'ATLAS', 'ktx'], accent: 'emerald' },
        { name: '失败任务卡', desc: '错的对话进任务队列等待数据团队处理', vendors: ['进阶治理'], accent: 'rose' },
        { name: 'wiki / instructions 增量', desc: '澄清答案、新规则、矛盾 → wiki 更新', vendors: ['ktx memory agent'], accent: 'violet' },
        { name: '黑盒 RLHF / 微调', desc: 'fine-tune LLM；难审计', vendors: ['闭源大厂'], accent: 'slate' },
      ],
      commonSense:
        '**沉淀进可读资产 (wiki / YAML / query_history)，不要沉淀进模型权重**——前者可评审、可回滚，后者一旦混入垃圾就洗不干净。',
    },
    {
      id: 'lineage',
      question: '可解释性怎么暴露？',
      why: '业务团队 / 监管 / 数据团队都需要知道"这个数字怎么来的"。',
      variants: [
        { name: 'dry-plan 展开轨迹', desc: '不连库展开 modeled SQL — 看 join / CTE / 计算列', vendors: ['WrenAI'], accent: 'amber' },
        { name: '生成 SQL + 引用模型列表', desc: '直接展示 SQL + 命中模型 / 表 / 列', vendors: ['多数 SL'], accent: 'emerald' },
        { name: '执行计划 (EXPLAIN)', desc: '物理计划；偏 DBA', vendors: ['通用'], accent: 'blue' },
        { name: '纯输出', desc: '只给 SQL 和结果；不解释', vendors: ['基线'], accent: 'rose' },
      ],
      commonSense:
        '**dry-plan + 命中模型清单 是面向业务最有用的解释**；EXPLAIN 太底层、SQL 太长——把"用了 customers · orders 1-N · lifetime_value 计算列"这种摘要做出来，业务才看得懂。',
    },
    {
      id: 'access',
      question: '权限模型在哪一层？',
      why: '权限放错层会"看似有但实际能绕过"——这是上下文层的高发漏洞。',
      variants: [
        {
          name: '语义层 RLAC / CLAC',
          desc: '规划阶段引擎注入；session property → WHERE / 列可见性',
          vendors: ['WrenAI', 'Cube row policies'],
          accent: 'emerald',
        },
        {
          name: '仓库 RLS (DB-level)',
          desc: 'PG / Snowflake / BigQuery 自带行级安全',
          vendors: ['Snowflake', 'BigQuery', 'Databricks UC'],
          accent: 'blue',
        },
        { name: '应用层过滤', desc: '在 BI / Agent 端补 WHERE', vendors: ['薄弱团队'], accent: 'rose' },
        { name: '无权限模型', desc: '完全开放；测试 / Demo', vendors: ['基线'], accent: 'slate' },
      ],
      commonSense:
        '**语义层 + 仓库 RLS 双重保险**：语义层 RLAC 避免"绕过 SL 直接连库"的漏洞；仓库 RLS 是最后兜底。应用层补 WHERE 是反模式。',
    },
    {
      id: 'observe',
      question: '可观测 / 漂移监控做什么？',
      why: '生产环境的"软失败"——schema 变了、列删了、enum 多了一个值——比模型退化常见 10 倍。',
      variants: [
        {
          name: 'Schema diff 监控',
          desc: '定期 introspect 比对；列变化 → 失效相关 RC',
          vendors: ['ATLAS self-maintenance', 'ktx scan diff'],
          accent: 'amber',
        },
        { name: 'NL2SQL 准确率看板', desc: '黄金集 + 抽样的滚动准确率', vendors: ['进阶团队'], accent: 'violet' },
        { name: '延迟 / 成本看板', desc: 'p50/p99 延迟 · token 成本 · LLM 调用数', vendors: ['通用 telemetry'], accent: 'blue' },
        { name: '审计日志', desc: '所有 SQL / 工具调用 / 修改落日志（含 user / session）', vendors: ['企业必备'], accent: 'rose' },
      ],
      commonSense:
        '**漂移监控应自动失效 RC，而不是只发告警**——schema 变了，相关的列描述 / 关系定义 / measure 应被打"is_expired"，下次召回时跳过、并触发自愈重生。光发邮件没人看。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-history',
      title: '记忆是契约的演化',
      body: '上下文层的"自学习"不是 ML——是把每次成功 / 失败的事实显式写回 wiki / SL / query_history，让契约越用越准。这种学习是可逆的、可审计的、不会污染。',
    },
    {
      icon: 'i-lucide-route',
      title: '解释力 = 团队信任',
      body: 'NL2SQL 系统能否"被业务团队信任使用"，几乎完全取决于解释力。dry-plan / 命中模型清单 / lineage 这些可视化是上下文层的"前门"——比准确率本身更重要。',
    },
    {
      icon: 'i-lucide-activity',
      title: '漂移监控 ≈ 自维护',
      body: 'schema diff → 失效 RC → 自愈重生（重 introspect / 重 embed）这条链路应被作为底层服务，让上下文层"在静止状态下也保持新鲜"——否则上下文会变成历史快照。',
    },
  ],
  matrix: {
    cols: ['沉淀', '解释', '权限', '漂移监控'],
    rows: [
      { vendor: 'ATLAS', school: 'agentic', cells: ['rc_change_log + RC', 'verify_sql 链', 'session policy', 'self-maintain heal-loop'] },
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['memory store', 'dry-plan', 'RLAC / CLAC', '由消费方'] },
      { vendor: 'ktx', school: 'open-context', cells: ['memory agent → wiki/SL', '生成 SQL + 模型列表', '语义层 + 仓库 RLS', 'scan diff (增量)'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['由消费方', 'SL 编译 SQL', 'dbt grants', '由消费方'] },
      { vendor: 'Cube', school: 'semantic-layer', cells: ['Cube cache', 'pre-agg / SQL', 'row policies', 'cube monitoring'] },
      { vendor: 'Cortex Analyst', school: 'managed-cloud', cells: ['👍/👎', 'Semantic Views 解释', 'Snowflake RBAC', '云内监控'] },
      { vendor: 'Databricks UC', school: 'managed-cloud', cells: ['UC lineage', 'metric views', 'UC ACL', 'UC monitoring'] },
    ],
  },
}
