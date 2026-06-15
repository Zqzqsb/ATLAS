import type { StageArch } from './comm'

export const uxArch: StageArch = {
  id: 'ux',
  abstract: '问题入口决定了上下文层暴露什么形态：工具腰带 / 单条 SQL / 仓库 PR / 数据集。同一份契约要服务四种入口才不会让团队各做各的。',
  principles: [
    { name: '一份契约，多个入口', desc: 'wiki + 语义层是单一来源；NL chat / Agent SDK / IDE / BI 都从同一处取上下文，避免知识孤岛。' },
    { name: '澄清优先于猜测', desc: '问题欠定义时，让入口主动反问而不是赌一个 SQL；澄清问题应被记忆下来作为下一次的 few-shot。' },
    { name: '入口决定可控性', desc: 'NL chat 给业务速度但难审计；IDE / PR 给开发严格度。给用户选，而不是只押一种。' },
  ],
  subQuestions: [
    {
      id: 'agent-locus',
      question: 'Agent / LLM 放在哪里？',
      why: '"系统内置 Agent" 还是 "BYO 外置 Agent" 决定了系统的边界、安全模型、可替换性。',
      variants: [
        {
          name: '内置 Agent',
          desc: '产品自己跑 LLM 服务、内置对话与生成内核',
          vendors: ['ATLAS', 'Fabric Data Agent', 'DMS Data Agent', 'Snowflake Cortex Analyst'],
          accent: 'violet',
        },
        {
          name: '外置 Agent (BYO)',
          desc: '只暴露 MCP/SDK 工具，由用户 Agent (Claude / Cursor / Codex) 编排',
          vendors: ['ktx', 'WrenAI (post-2026-05)'],
          accent: 'emerald',
        },
        {
          name: '语义层 + 多前端',
          desc: '语义层作为 BI / 任意 LLM 的统一接入点；自身不绑 LLM',
          vendors: ['dbt Semantic Layer', 'Cube'],
          accent: 'amber',
        },
      ],
      commonSense:
        '内置 Agent 上手快、跑得稳；外置 Agent 给团队替换 LLM 模型与工具的自由度，但需要团队懂 MCP / function-calling。**有"模型/工具/工作流应当我可控"的诉求 → 选外置；要"业务部门点开就能用" → 选内置。**',
    },
    {
      id: 'channel-form',
      question: '主要交互通道是哪种？',
      why: '主通道决定了反馈环、错误展示、记忆触发的设计基线。',
      variants: [
        { name: 'Chat 多轮', desc: '业务问 → 系统答；澄清回路天然', vendors: ['ATLAS', 'Cortex Analyst', 'Fabric'], accent: 'slate' },
        { name: 'Agent 工具调用', desc: 'function-calling · 一次次原子调用', vendors: ['ktx (MCP)', 'WrenAI SDK'], accent: 'violet' },
        { name: 'IDE / 仓库工作流', desc: '改 YAML、走 PR 评审；CLI 跑校验', vendors: ['ktx', 'dbt SL', 'WrenAI CLI'], accent: 'amber' },
        { name: 'BI 嵌入', desc: '已有 Looker / Metabase 拉数', vendors: ['Cube', 'dbt SL', 'Databricks UC'], accent: 'blue' },
      ],
      commonSense:
        '不要押单一通道——一个上下文层至少要能同时服务"业务 chat" + "Agent 工具" + "BI 拉数"三种。**通道是消费层，契约才是真相源。**',
    },
    {
      id: 'clarify',
      question: '怎么处理"问题欠定义"？',
      why: '"上月活跃客户"——上月是哪个月？活跃指什么？这种歧义如果不澄清就直接生成 SQL，错的概率几乎是 100%。',
      variants: [
        { name: '直接猜', desc: 'LLM 自行选定义；错了让用户复述', vendors: ['多数早期 NL2SQL'], accent: 'rose' },
        { name: '系统澄清', desc: '系统主动问回；选项化（drop-down）', vendors: ['ATLAS', 'Cortex Analyst'], accent: 'violet' },
        { name: 'Skill 引导 Agent 澄清', desc: 'Markdown skill 教 Agent 在不确定时反问', vendors: ['WrenAI (query.md)'], accent: 'emerald' },
      ],
      commonSense:
        '澄清是质量阀门，不是 UX 麻烦——"主动澄清 + 把答案写回 instructions / wiki" 比 "聪明的 LLM" 更能降低长尾错误率。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-layers',
      title: '入口可换、契约不可换',
      body: '如果一个系统让你"换 LLM = 重新建模"，它就不是上下文层。上下文层的标志是：把 LLM、Agent、UI 降级为消费者，契约（语义层 + wiki）才是核心资产。',
    },
    {
      icon: 'i-lucide-message-circle-question',
      title: '澄清不是 UX 问题，是数据问题',
      body: '业务问题的"歧义"绝大部分来自 schema 含义不清；与其训 LLM 学会问，不如把 enum / 时间窗 / 单位 / 同义词写进 wiki 让 schema 自己讲清楚。',
    },
    {
      icon: 'i-lucide-puzzle',
      title: '入口决定反馈路径',
      body: 'Chat 通道天然适合 👍/👎；IDE 通道天然适合 git diff 评审；BI 通道天然适合"准确率/延迟"看板——不要混。',
    },
  ],
  matrix: {
    cols: ['Chat', 'Agent SDK', 'IDE', 'BI'],
    rows: [
      { vendor: 'ATLAS', school: 'agentic', cells: ['✓', '·', '·', '·'] },
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['·', '✓', '✓', '·'] },
      { vendor: 'ktx', school: 'open-context', cells: ['·', '✓', '✓', '·'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['·', '✓', '✓', '✓'] },
      { vendor: 'Cube', school: 'semantic-layer', cells: ['·', '✓', '·', '✓'] },
      { vendor: 'Snowflake Cortex', school: 'managed-cloud', cells: ['✓', '✓', '·', '✓'] },
      { vendor: 'Fabric Data Agent', school: 'managed-cloud', cells: ['✓', '·', '·', '✓'] },
    ],
  },
}
