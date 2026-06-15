import type { StageArch } from './comm'

export const verifyArch: StageArch = {
  id: 'verify',
  abstract: '"看起来对" ≠ "可执行 ∧ 语义正确"。把校验拆成静态 / 策略 / 预执行 / 语义合理性四层闸门，能在执行前堵住 95% 的错。',
  principles: [
    { name: '失败应给出结构化诊断', desc: '不是 "syntax error near X"，而是 phase / code / 列名 / 期望类型——让 Agent 能据此精确修复。' },
    { name: '默认只读', desc: 'NL2SQL 系统永远只跑 SELECT，从源头切断 destructive 风险。' },
    { name: '在 live DB 之前先在沙箱 / dry-run', desc: '错的 SQL 不应该直接到生产；LIMIT 0 / EXPLAIN 几乎零成本能挡掉一类错。' },
    { name: '语义合理性比语法更难也更重要', desc: 'fan-out / 时间窗 / 单位 / NULL 语义都是"语法对了但答案错"的高发地带。' },
  ],
  subQuestions: [
    {
      id: 'static',
      question: '静态校验做哪些？',
      why: '低成本就能拦掉一类错——且能给 Agent 最早期的反馈。',
      variants: [
        { name: 'parse + qualify', desc: '语法解析 + 表 / 列限定到 schema', vendors: ['sqlglot', 'wren-core'], accent: 'amber' },
        { name: 'type-check', desc: '函数 / 表达式参数类型', vendors: ['wren-core', 'Cube'], accent: 'amber' },
        { name: '引用列存在 / 主键存在', desc: '从 catalog 校验列 / FK', vendors: ['多数 SL'], accent: 'amber' },
        { name: '无静态校验', desc: '直接进 dry-run', vendors: ['基线'], accent: 'rose' },
      ],
      commonSense:
        'parse + qualify + 列存在校验是必须的——sqlglot / DataFusion 这种引擎已经能免费做掉，不用就是浪费。',
    },
    {
      id: 'policy',
      question: '策略校验包含什么？',
      why: '这一层决定"安全"——读写权限、行/列权限、被禁的危险函数。',
      variants: [
        {
          name: 'read-only 黑名单',
          desc: '禁 INSERT / UPDATE / DELETE / ALTER / CREATE / DROP / TRUNCATE / GRANT / REVOKE…',
          vendors: ['ktx daemon validate-read-only', 'ATLAS dryrun gate'],
          accent: 'rose',
        },
        {
          name: 'RLAC / CLAC',
          desc: 'session property 注入 WHERE / 列可见性',
          vendors: ['WrenAI', 'Cube row policies', 'Snowflake'],
          accent: 'emerald',
        },
        { name: 'denied funcs', desc: '黑名单危险 UDF / 时间敏感函数', vendors: ['WrenAI policy'], accent: 'amber' },
        {
          name: 'row limit',
          desc: '默认 LIMIT 100 · 硬顶 1000 · 防"全表扫"',
          vendors: ['WrenAI', 'Cube', 'Cortex Analyst'],
          accent: 'blue',
        },
      ],
      commonSense:
        'read-only 是 NL2SQL 不可商量的底线；RLAC / CLAC 应该写在语义层（建模即治理），而不是在应用层补——后者一定会被绕过。',
    },
    {
      id: 'dryrun',
      question: '怎么"预执行"？',
      why: '"在 live DB 跑一下"是最有效的验证，但要做得便宜（不返回行）。',
      variants: [
        { name: 'EXPLAIN / 计划解释', desc: '不跑数据，看物理计划', vendors: ['多数 SL'], accent: 'amber' },
        {
          name: 'LIMIT 0 / dry-run',
          desc: '提交 SQL 但不返回行 — 校验语法 / 列 / 权限',
          vendors: ['WrenAI dry-run', 'ATLAS'],
          accent: 'emerald',
        },
        {
          name: 'dry-plan (不连库)',
          desc: 'wren-core 把 modeled SQL 展开 — 不连库即可看到 join / 计算列',
          vendors: ['WrenAI'],
          accent: 'violet',
        },
        { name: '沙箱执行', desc: '在副本 / 测试库执行', vendors: ['企业内常见'], accent: 'blue' },
      ],
      commonSense:
        '**dry-plan + dry-run 双闸**：dry-plan 不连库就能验证逻辑（join 对不对、计算列展开对不对）；dry-run 再到 live DB 验证语法 / 列 / 权限。这两步一个没有都是赌博。',
    },
    {
      id: 'semantic',
      question: '语义合理性怎么校验？',
      why: '语法 OK、能跑出数据，结果可能依然是错的——这是上下文层的硬骨头。',
      variants: [
        {
          name: 'fan-out / chasm trap',
          desc: '多对多关系下 SUM 重复计数；planner 应自动检测',
          vendors: ['WrenAI planner', 'ktx SL'],
          accent: 'amber',
        },
        {
          name: '单位 / 时间窗',
          desc: '"上月" → 哪个时区 / 起止？金额 → 单位是分还是元？',
          vendors: ['由 wiki / instructions 承载'],
          accent: 'emerald',
        },
        { name: 'NULL 语义', desc: 'COUNT(*) vs COUNT(col)；JOIN 漏行', vendors: ['列描述 / wiki'], accent: 'blue' },
        { name: 'LLM-as-judge', desc: '让另一 LLM 检查"答案是否回答了原问题"', vendors: ['进阶 eval'], accent: 'violet' },
        { name: '人工抽样', desc: '生产前 / 高频问题人工核', vendors: ['所有严肃团队'], accent: 'rose' },
      ],
      commonSense:
        '**fan-out / chasm 必须由 planner 解决**（人不可能记得每个 measure 在每张表的 grain）。**单位 / 时间窗 / NULL 必须由 wiki / 列描述写清**，不能依赖 LLM 现猜。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-shield-check',
      title: '校验是 4 层闸门，不是 1 个 LLM',
      body: '把"对错判断"压在 LLM 上是反模式——四层闸门里只有"语义合理性"才需要 LLM；其余三层都是确定式 + 仓库引擎可以做的事，应当尽量前置、尽量便宜。',
    },
    {
      icon: 'i-lucide-alert-triangle',
      title: 'fan-out 是隐形大坑',
      body: '"orders SUM(total)"在有 1-N 关系时如果 join 路径选错，结果会偷偷地放大。planner 必须检测 fan-out 并按 measure_group 拆 CTE 才能避免——这是任何 NL2SQL 团队都该早点踩到的坑。',
    },
    {
      icon: 'i-lucide-flask-conical',
      title: 'dry-run 比 retry 便宜',
      body: '一次 LLM retry = 一次 LLM 调用 + 一次 DB 查询的成本；一次 dry-run（LIMIT 0）= 几乎免费。生产场景应该把 dry-run 放在所有 LLM retry 之前。',
    },
  ],
  matrix: {
    cols: ['Static', 'Policy', 'Dry-run', 'Semantic'],
    rows: [
      { vendor: 'ATLAS', school: 'agentic', cells: ['parse', 'read-only', 'dryrun', 'agent verify_sql'] },
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['sqlglot+wren-core', 'strict/RLAC/CLAC/limit', 'dry-plan + dry-run', 'fan-out detect'] },
      { vendor: 'ktx', school: 'open-context', cells: ['daemon analyze', 'validate-read-only', 'sl_query plan', 'fan-out + chasm'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['SL 编译时', '消费方', '消费方', 'SL planner'] },
      { vendor: 'Cube', school: 'semantic-layer', cells: ['Cube 编译时', 'row policies', 'pre-aggregations', 'cube semantics'] },
      { vendor: 'Cortex Analyst', school: 'managed-cloud', cells: ['内置', '云权限', '内置 plan', 'self-correction'] },
    ],
  },
}
