import type { StageArch } from './comm'

export const uxArch: StageArch = {
  id: 'ux',
  abstract:
    '问题入口决定了上下文层暴露什么形态：工具腰带 / 单条 SQL / 仓库 PR / 数据集。同一份契约要服务四种入口才不会让团队各做各的。',
  principles: [
    {
      name: '一份契约，多个入口',
      desc: 'wiki + 语义层是单一来源；NL chat / Agent SDK / IDE / BI 都从同一处取上下文，避免知识孤岛。',
    },
    {
      name: '澄清优先于猜测',
      desc: '问题欠定义时，让入口主动反问而不是赌一个 SQL；澄清问题应被记忆下来作为下一次的 few-shot。',
    },
    {
      name: '入口决定可控性',
      desc: 'NL chat 给业务速度但难审计；IDE / PR 给开发严格度。给用户选，而不是只押一种。',
    },
  ],
  subQuestions: [
    /* ─────── Q1: Agent / LLM 放哪里 ─────── */
    {
      id: 'agent-locus',
      question: 'Agent / LLM 放在哪里？',
      why: '"系统内置 Agent" 还是 "BYO 外置 Agent" 决定了系统的边界、安全模型、可替换性。',
      steps: [
        {
          id: 'ux-locus-1',
          name: 'LLM 的归属：内置 vs 外置 vs 不绑',
          desc: '产品内置 LLM 内核 / 只暴露工具让用户 Agent 编排 / 语义层不绑 LLM 服务多前端——三种边界。',
          icon: 'i-lucide-bot',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '外置 Agent（BYO）：暴露 MCP / SDK 原语，由用户的 Claude / Cursor / Codex 编排——LLM 可替换。',
              detail: {
                summary:
                  'WrenAI 把自己定位成"被 Agent 调用的语义引擎"——它不强绑一个内置 LLM，而是暴露一组 MCP / SDK 原语，让用户拿自己的 Agent（Claude Code / Cursor / Codex）来编排推理。',
                bullets: [
                  {
                    label: 'LLM 可替换',
                    icon: 'i-lucide-replace',
                    accent: 'emerald',
                    body: '换模型 = 换你 Agent 侧的配置，不需要动 WrenAI 本身——避免"换 LLM = 重新建模"的锁定。',
                  },
                  {
                    label: '原语而非黑盒',
                    icon: 'i-lucide-puzzle',
                    accent: 'violet',
                    body: 'fetch / dry-plan / dry-run / generate 都是独立工具——Agent 自由组合，能力边界清晰。',
                  },
                  {
                    label: '代价：需懂 MCP',
                    icon: 'i-lucide-graduation-cap',
                    accent: 'amber',
                    body: '团队得会 MCP / function-calling 才能用好——不像内置 Agent"点开就能用"。',
                  },
                ],
                closing: '一句话：要"模型 / 工具 / 工作流我可控" → 选外置；要"业务部门点开就用" → 选内置。',
              },
              code: [{ repo: 'wrenai', path: 'wren-ai-service/src/pipelines', label: 'pipelines (原语)' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '内置 Agent：Snowflake 跑自己的 LLM 服务 + 对话内核，业务用户开箱即用，不可替换模型。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '外置 Agent（BYO）：只起本地 MCP server，由你的 Claude Code / Codex / Cursor 编排——和 WrenAI 同形态。',
              code: [{ repo: 'ktx', path: 'README.md#L133-L145', label: 'agent integration' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '内置 Agent：自跑 LLM 服务 + ReAct 对话内核，业务用户直接对话。',
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: '不绑 LLM：语义层作为 BI / 任意 LLM 的统一接入点，自身不提供对话 Agent。',
              notSupported: 'dbt SL 自身不内置也不编排 LLM——它是被多前端消费的"指标 API"。',
            },
            {
              vendor: 'Fabric DA',
              school: 'managed-cloud',
              desc: '内置 Agent：Fabric Data Agent 托管 LLM + 对话，绑定 Fabric 生态。',
              refs: ['fabric-data-agent'],
            },
          ],
        },
      ],
      commonSense:
        '内置 Agent 上手快、跑得稳；外置 Agent 给团队替换 LLM 模型与工具的自由度，但需要团队懂 MCP / function-calling。**有"模型/工具/工作流应当我可控"的诉求 → 选外置；要"业务部门点开就能用" → 选内置。**',
    },

    /* ─────── Q2: 交互通道 ─────── */
    {
      id: 'channel-form',
      question: '主要交互通道是哪种？',
      why: '主通道决定了反馈环、错误展示、记忆触发的设计基线。',
      steps: [
        {
          id: 'ux-channel-1',
          name: '四类通道：Chat / Agent / IDE / BI',
          desc: '业务 chat 多轮 / Agent function-calling / IDE 仓库 PR / BI 嵌入拉数——一个上下文层应同时服务多种。',
          icon: 'i-lucide-share-2',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'Agent 工具调用为主（SDK / MCP）+ IDE / CLI 仓库工作流；契约统一，通道是消费层。',
              code: [{ repo: 'wrenai', path: 'wren-ai-service', label: 'wren-ai-service' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Chat（业务多轮）+ REST API + BI 嵌入——云内多通道，统一走 Semantic Model。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Chat 多轮为主，澄清回路天然——面向业务人员的对话入口。',
              selfContained: true,
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'Agent 工具（MCP）+ IDE / CLI 仓库工作流——开发者 / Agent 双入口。',
              code: [{ repo: 'ktx', path: 'README.md#L154-L166', label: 'CLI / MCP' }],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'IDE 仓库工作流 + BI 嵌入（Hex / Tableau 通过 SL API 拉指标）；无原生 chat。',
              code: [{ repo: 'metricflow', path: 'metricflow', label: 'metricflow API' }],
            },
            {
              vendor: 'Cube',
              school: 'semantic-layer',
              desc: 'BI 嵌入为主（REST / GraphQL / SQL API）+ AI API；语义层多前端接入点。',
              code: [{ repo: 'cube', path: 'packages/cubejs-api-gateway', label: 'api-gateway' }],
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Genie chat + BI（Metric View 给 Notebook / Dashboard 拉数）。',
              refs: ['dbx-genie'],
            },
          ],
        },
      ],
      commonSense:
        '不要押单一通道——一个上下文层至少要能同时服务"业务 chat" + "Agent 工具" + "BI 拉数"三种。**通道是消费层，契约才是真相源。**',
    },

    /* ─────── Q3: 澄清 ─────── */
    {
      id: 'clarify',
      question: '怎么处理"问题欠定义"？',
      why: '"上月活跃客户"——上月是哪个月？活跃指什么？这种歧义如果不澄清就直接生成 SQL，错的概率几乎是 100%。',
      steps: [
        {
          id: 'ux-clarify-1',
          name: '欠定义时：猜 vs 反问',
          desc: '直接猜（错了让用户复述）/ 系统主动澄清（选项化）/ skill 教 Agent 在不确定时反问。',
          icon: 'i-lucide-message-circle-question',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'skill / pipeline 教 Agent 在歧义时反问，澄清答案可写回 instructions 成为下次上下文。',
              code: [{ repo: 'wrenai', path: 'wren-ai-service/src/pipelines/generation', label: 'clarification' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '系统主动澄清：歧义时反问 / 给出建议问题，引导用户收敛。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '系统主动澄清 + 选项化（drop-down）；澄清答案被记录为后续 few-shot。',
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'dbt SL 不处理自然语言澄清——歧义消解在上游消费方。',
              notSupported: 'dbt SL 无 NL 入口，"澄清"概念不适用；由消费方的对话层负责。',
            },
          ],
        },
      ],
      commonSense:
        '澄清是质量阀门，不是 UX 麻烦——"主动澄清 + 把答案写回 instructions / wiki" 比 "聪明的 LLM" 更能降低长尾错误率。"直接猜"是基线。',
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
