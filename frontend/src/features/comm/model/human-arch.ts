import type { StageArch } from './comm'

export const humanArch: StageArch = {
  id: 'human',
  abstract: '完全自动建模 / 自学 = 容易跑偏。把"人参与"细化成事前 / 事中 / 事后 / 知识精炼四种形态，让人成为精度的锚点而非瓶颈。',
  principles: [
    { name: '人参与必须低摩擦', desc: '👍/👎 一键反馈、PR 评审 / git diff——和团队已有工具集成；让"修一条规则"不需要工程师写代码。' },
    { name: '反馈必须落地为契约', desc: '一次澄清 = 一次 wiki 更新 = 下次召回；否则人的成本就被浪费。' },
    { name: '精度的瓶颈是知识，不是模型', desc: '团队对业务的隐性理解（同义词 / enum / 规则）才是关键资产，记忆它们才是上下文层的核心工作。' },
  ],
  subQuestions: [
    {
      id: 'before',
      question: '事前评审走什么形态？',
      why: '上线前的评审决定语义层质量上限。',
      variants: [
        { name: 'Git PR 评审', desc: 'YAML diff / Markdown diff 走 PR', vendors: ['ktx', 'WrenAI', 'dbt SL'], accent: 'emerald' },
        { name: '应用 UI 审批', desc: '在产品里点确认 / 拒绝', vendors: ['Cube Cloud', 'Fabric', 'Snowflake'], accent: 'blue' },
        { name: '自动通过', desc: 'LLM / Agent 直接写入', vendors: ['危险默认'], accent: 'rose' },
      ],
      commonSense:
        '**Git PR 是金标准**——历史可追溯、可回滚、与代码评审同流程。应用 UI 审批适合非技术团队但缺审计、缺并发处理。"自动通过"在生产场景永远不该出现。',
    },
    {
      id: 'during',
      question: '事中怎么收集反馈？',
      why: '生成出错时立刻能修，是上下文层"使用即沉淀"的入口。',
      variants: [
        { name: '👍/👎 + 文字', desc: '业务用户一键打分 + 自由评论', vendors: ['Cortex Analyst', 'Fabric'], accent: 'rose' },
        { name: '在线编辑 SQL', desc: '用户改 SQL；改后回流为 few-shot', vendors: ['ATLAS', 'WrenAI memory store'], accent: 'amber' },
        { name: '错样本标注', desc: '把错误归类（schema linking / fan-out / 时间窗）', vendors: ['进阶团队'], accent: 'violet' },
        { name: '主动澄清问题', desc: '系统问回；澄清答案被记录', vendors: ['ATLAS'], accent: 'emerald' },
      ],
      commonSense:
        '**"👍/👎" 几乎没有信息量；最有用的是"用户改后的 SQL"——把它和原 NL 一起回流到 query_history，是上下文层最廉价的精度来源**。',
    },
    {
      id: 'after',
      question: '事后回归 / Eval 怎么做？',
      why: '没有 eval 就没有改进；只看用户反馈是 selection bias 严重。',
      variants: [
        { name: '黄金 NL→SQL 集', desc: '团队维护一组"标准答案"，回归跑', vendors: ['进阶团队'], accent: 'amber' },
        { name: 'LLM-as-judge', desc: '另一 LLM 评分；规模大但精度有限', vendors: ['学术 / OSS'], accent: 'violet' },
        { name: '人工抽样', desc: '随机抽 N% 生产对话人工核', vendors: ['所有严肃团队'], accent: 'rose' },
        { name: '执行结果对比', desc: '比对当前 SQL 与 baseline SQL 的结果数字', vendors: ['进阶'], accent: 'blue' },
      ],
      commonSense:
        '**黄金集 + 抽样核对 = 最低配；LLM-as-judge 适合规模筛选但不能替代抽样**。Eval 应该跑在 PR 流程里——契约改了 → 自动跑 eval → 通过才合并。',
    },
    {
      id: 'curate',
      question: '知识精炼谁来做？',
      why: '业务知识冷热数据不停变化；不持续精炼，wiki 会变成"垃圾场"。',
      variants: [
        { name: '半自动 grill', desc: 'LLM 扫 wiki 找矛盾 / 重复 / 过时 → 人审定', vendors: ['ktx memory agent', 'WrenAI enrich-context'], accent: 'emerald' },
        { name: '完全人工', desc: '专人 / 数据团队定期清理', vendors: ['传统数据治理'], accent: 'amber' },
        { name: '自然淘汰', desc: '不主动管；冷数据自动失效', vendors: ['过于乐观'], accent: 'rose' },
      ],
      commonSense:
        '**"半自动 grill" 是最务实的形态**——LLM 把可疑条目挑出来（矛盾 / 重复 / 长期不被引用），人只需评审 yes/no。把人的精力放在"决策"上而不是"扫地"上。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-recycle',
      title: '反馈 = 契约更新，不是日志',
      body: '一次"用户改了 SQL"如果只进了日志而不更新 wiki / SL / query_history，下次还会犯一样的错。**反馈必须落地为契约——闭环的衡量标准就是这件事**。',
    },
    {
      icon: 'i-lucide-git-pull-request',
      title: 'PR 流程是质量的护城河',
      body: '语义层 / wiki 走 git PR + 自动 eval，把"语义改动"变成"代码改动"——这件事的价值远超用什么 embedding、什么 LLM。',
    },
    {
      icon: 'i-lucide-sparkles',
      title: 'LLM 应该做"建议"，人做"决策"',
      body: '矛盾标注、同义词归并、cube 抽取 这些任务由 LLM 提候选、人确认；让人去做"扫地"的工作就是浪费——也容易被跳过，最终 wiki 不维护。',
    },
  ],
  matrix: {
    cols: ['事前评审', '事中反馈', '事后 Eval', '知识精炼'],
    rows: [
      { vendor: 'ATLAS', school: 'agentic', cells: ['UI 审批', '改 SQL · 澄清', '抽样核', 'memory agent'] },
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['Git PR', 'memory store', '黄金集 / eval skill', 'enrich-context (grill)'] },
      { vendor: 'ktx', school: 'open-context', cells: ['Git PR', 'memory agent', '由消费方', 'memory agent grill'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['Git PR (dbt 流程)', '由消费方', '由消费方', '由消费方'] },
      { vendor: 'Cortex Analyst', school: 'managed-cloud', cells: ['UI 审批', '👍/👎', '云内 metrics', '云内'] },
      { vendor: 'Fabric DA', school: 'managed-cloud', cells: ['UI 审批', '👍/👎', '内置', '内置'] },
    ],
  },
}
