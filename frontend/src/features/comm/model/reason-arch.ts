import type { StageArch } from './comm'

export const reasonArch: StageArch = {
  id: 'reason',
  abstract: '不论"端到端 LLM"还是"多步 Agent"，都能被还原成 Plan → Ground → Generate 三段。各家差异 = 把 LLM 放在哪几段、怎么循环、怎么校验。',
  principles: [
    { name: '把"问题→SQL"切开', desc: '不要让一次 LLM 调用同时做意图分解、schema 落地、SQL 生成；切开后每段都可监控、可重试、可缓存。' },
    { name: '校验闭环胜过精度', desc: '"先生成 SQL 再 dry-run / 修复"循环，比"训出更聪明的生成模型"更可靠。' },
    { name: 'Schema linking 是核心难题', desc: '不是 LLM 能力问题，而是上下文层质量问题：列描述 / enum / 同义词不到位，再聪明的 LLM 都会绑错。' },
  ],
  subQuestions: [
    {
      id: 'shape',
      question: '推理流水线长什么样？',
      why: '系统的"骨架"决定了能做多复杂的查询、错了多容易修。',
      variants: [
        {
          name: '端到端单次 LLM',
          desc: '一次 prompt → 一段 SQL；最简单也最脆',
          vendors: ['早期 NL2SQL', '部分云内置'],
          accent: 'rose',
        },
        {
          name: '多步固定流水线',
          desc: 'Plan → Recall → Generate → Validate 固定阶段',
          vendors: ['Snowflake Cortex Analyst', 'Fabric'],
          accent: 'blue',
        },
        {
          name: 'Agent ReAct 循环',
          desc: 'Reason → Act (tool) → Observe 自治循环；步骤数自定',
          vendors: ['ATLAS', 'DMS Data Agent'],
          accent: 'violet',
        },
        {
          name: '原语 + 外置 Agent 编排',
          desc: '系统暴露 fetch / dry-plan / dry-run / store；Agent 编',
          vendors: ['ktx', 'WrenAI'],
          accent: 'emerald',
        },
      ],
      commonSense:
        '**固定流水线适合云厂商封闭场景；Agent 循环适合复杂 / 长尾问题；原语 + 外置 Agent 适合可定制场景**。最忌讳的是"端到端单 LLM 调用"——任何错都是黑盒，没法定位、没法局部重试。',
    },
    {
      id: 'plan',
      question: '"意图分解"放不放？怎么放？',
      why: '业务问题往往含多个子查询（"上月营收 + 同比增长 + Top5 客户"），不分解就让 LLM 一次写出，容易丢条件。',
      variants: [
        { name: '不分解', desc: '直接生成；适合简单单语句', vendors: ['许多默认实现'], accent: 'rose' },
        { name: '改写 / 规范化', desc: '把口语化问题改写成"标准化"形式', vendors: ['ATLAS rewrite step'], accent: 'slate' },
        { name: '任务分类 + 路由', desc: '判断是 NL2SQL / 描述查询 / 元数据 / 走不同 prompt', vendors: ['Cortex Analyst', 'Fabric'], accent: 'blue' },
        { name: '多步分解 + 子查询合并', desc: '拆成子问题，分别生成 SQL 再 UNION/JOIN', vendors: ['进阶 Agent 工作流'], accent: 'violet' },
      ],
      commonSense:
        '**至少要有"问题改写"——把"上月"具体化成日期、把"客户"消歧到模型名、把单位 / 时区注入**。哪怕只做这一步，也能挡住大部分的低级错误。',
    },
    {
      id: 'ground',
      question: 'Schema linking 怎么做？',
      why: '"客户" → 哪张表？哪一列？这步错了，后面全错。',
      variants: [
        { name: 'LLM 自由选', desc: '把 schema 全塞 prompt 让 LLM 选', vendors: ['小规模默认'], accent: 'slate' },
        { name: '向量召回 top-k 模型', desc: 'fetch 相似 schema_items 限定候选', vendors: ['WrenAI', 'ATLAS', 'ktx'], accent: 'violet' },
        { name: '取值剖析消歧', desc: '看列里实际值（distinct sample）判断', vendors: ['ATLAS', 'WrenAI value profiling'], accent: 'amber' },
        { name: '强约束（语义层强制）', desc: 'strict mode 禁止裸物理表 / 未建模列', vendors: ['WrenAI strict_mode', 'Cube'], accent: 'emerald' },
      ],
      commonSense:
        '**召回缩小候选 + 取值剖析消歧 + strict mode 兜底**——三层组合是 schema linking 在生产里能跑稳的最低配置。少任何一层，长尾错误率都会上升。',
    },
    {
      id: 'gen',
      question: 'SQL 生成的"主体"在哪？',
      why: '把 join、计算列、关系展开放在 LLM 还是引擎，决定了正确性的上限。',
      variants: [
        { name: 'LLM 直接写物理 SQL', desc: 'LLM 写完整 SQL，包含 join / 计算', vendors: ['许多 NL2SQL'], accent: 'rose' },
        {
          name: 'LLM 写"模型 SQL" + 引擎展开',
          desc: 'LLM 引用模型名 / 计算列；引擎展开 CTE / join / RLAC',
          vendors: ['WrenAI (wren-core)', 'dbt SL', 'Cube'],
          accent: 'amber',
        },
        {
          name: 'LLM 选 measures / dimensions',
          desc: 'LLM 不写 SQL，只声明要哪些指标 / 维度；引擎全权生成',
          vendors: ['Snowflake Cortex Analyst', 'ktx sl_query'],
          accent: 'emerald',
        },
        {
          name: 'Agent 工具拼装',
          desc: 'Agent 用 verify_sql / set_rich_context 等工具增量推进',
          vendors: ['ATLAS'],
          accent: 'violet',
        },
      ],
      commonSense:
        '**让 LLM 做"声明"，让引擎做"展开"**——LLM 的价值是理解业务问题、把意图映射到模型；写 join 路径、注入 RLAC、防 fan-out 这些应该交给确定式引擎。LLM 直接写物理 SQL 是技术债。',
    },
    {
      id: 'loop',
      question: '错了之后怎么循环？',
      why: '"一次写对"是奢望；如何把错误变成结构化反馈让系统自己修复，决定了稳定性。',
      variants: [
        { name: '不循环', desc: '错就错；用户复述', vendors: ['基线'], accent: 'rose' },
        { name: '结构化错误 + Agent retry', desc: 'WrenError(phase, code) → Agent 据 phase 修复', vendors: ['WrenAI'], accent: 'amber' },
        { name: '内置 self-correction pipeline', desc: '系统内固定的"再尝试"链', vendors: ['Cortex Analyst', 'Fabric'], accent: 'blue' },
        { name: 'Agent 工具自动修复', desc: 'verify_sql 失败 → Agent 自然在 ReAct 里修', vendors: ['ATLAS'], accent: 'violet' },
      ],
      commonSense:
        '**结构化错误是基础设施**——返回 phase / code / 列名 / 期望类型，Agent 才有办法定位。返回字符串错误信息（"syntax error near LIMIT"）等于没给反馈。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-cpu',
      title: '语义引擎是真相源',
      body: '"LLM + 数据" 不等于 "上下文层"——决定正确性的是"谁来执行 join / 计算列 / 权限注入"。把这件事放在确定式引擎（wren-core / SL planner）里，不论 LLM 多差都不会出错。',
    },
    {
      icon: 'i-lucide-route',
      title: '可解释性 ≈ 生成轨迹',
      body: 'dry-plan / 计划解释把 LLM 的"黑盒选择"展开成"用了哪些模型 / join / 计算列"——这就是 NL2SQL 系统该有的可解释性，比"看 LLM 在想什么"靠谱得多。',
    },
    {
      icon: 'i-lucide-sliders',
      title: 'Schema linking 准 ≠ LLM 强',
      body: 'schema linking 失败的根因 90% 是上下文层缺信息（列描述 / enum / 同义词），10% 才是 LLM 不懂。把精力放在 wiki 而非 prompt 工程上。',
    },
  ],
  matrix: {
    cols: ['Plan', 'Ground', 'Generate', 'Repair'],
    rows: [
      { vendor: 'ATLAS', school: 'agentic', cells: ['rewrite', 'agent ReAct', 'agent + verify_sql', 'tool 内修'] },
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['agent 自分解', 'fetch + recall', 'agent → wren-core 展开', 'WrenError + retry'] },
      { vendor: 'ktx', school: 'open-context', cells: ['agent skill', 'hybrid 召回', 'sl_query 引擎', '工具粒度修复'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['由消费方', '由消费方', 'SL 编译', '由消费方'] },
      { vendor: 'Cortex Analyst', school: 'managed-cloud', cells: ['任务路由', '云内 schema', '基于 Semantic Views', 'self-correction'] },
      { vendor: 'Fabric DA', school: 'managed-cloud', cells: ['任务路由', 'Business Sem.', '内置生成', '内置 retry'] },
    ],
  },
}
