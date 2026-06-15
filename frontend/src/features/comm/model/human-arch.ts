import type { StageArch } from './comm'

export const humanArch: StageArch = {
  id: 'human',
  abstract:
    '完全自动建模 / 自学 = 容易跑偏。把"人参与"细化成事前 / 事中 / 事后 / 知识精炼四种形态，让人成为精度的锚点而非瓶颈。',
  principles: [
    {
      name: '人参与必须低摩擦',
      desc: '👍/👎 一键反馈、PR 评审 / git diff——和团队已有工具集成；让"修一条规则"不需要工程师写代码。',
    },
    {
      name: '反馈必须落地为契约',
      desc: '一次澄清 = 一次 wiki 更新 = 下次召回；否则人的成本就被浪费。',
    },
    {
      name: '精度的瓶颈是知识，不是模型',
      desc: '团队对业务的隐性理解（同义词 / enum / 规则）才是关键资产，记忆它们才是上下文层的核心工作。',
    },
  ],
  subQuestions: [
    /* ─────── Q1: 事前评审 ─────── */
    {
      id: 'before',
      question: '事前评审走什么形态？',
      why: '上线前的评审决定语义层质量上限。',
      steps: [
        {
          id: 'human-before-1',
          name: '语义改动怎么过审',
          desc: 'Git PR 评审（YAML / Markdown diff）vs 应用 UI 审批 vs 自动通过——决定了可追溯性。',
          icon: 'i-lucide-git-pull-request',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'MDL YAML 入 git，语义改动走标准 PR diff 评审——和代码评审同流程，可追溯、可回滚。',
              detail: {
                summary:
                  'WrenAI 把"改语义"等同于"改代码"——MDL 的每个 model / relationship / 计算列都是 git 里的 YAML，任何修改都通过 PR 提交，reviewer 看 diff 就知道改了什么。',
                bullets: [
                  {
                    label: '改动即 diff',
                    icon: 'i-lucide-file-diff',
                    accent: 'emerald',
                    body: '新增一个计算列 / 改一个 join 关系 = 一行 YAML diff——reviewer 一眼看清影响面，不用猜。',
                  },
                  {
                    label: '可回滚',
                    icon: 'i-lucide-undo-2',
                    accent: 'violet',
                    body: '改错了 git revert 即可——和"在产品 UI 里点了确认就生效、出错难追"形成鲜明对比。',
                  },
                  {
                    label: '可接 CI eval',
                    icon: 'i-lucide-check-check',
                    accent: 'amber',
                    body: 'PR 触发自动 eval（黄金集回归）——契约改了先跑评测、通过才合并，把质量门禁前置。',
                  },
                ],
                closing: '一句话：git PR 是语义层质量的金标准——历史可追溯、可回滚、与代码评审同流程。',
              },
              code: [{ repo: 'wrenai', path: 'docs/core/reference/mdl.md', label: 'MDL (git YAML)' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Semantic Model 在 Studio UI 里编辑 / 审批；也可导出 YAML 入 git，但主流程是云内 UI。',
              refs: ['sf-cortex-overview', 'sf-semantic-yaml'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'semantic-layer YAML + wiki Markdown 全入 git——和 WrenAI 同金标准，PR diff 评审。',
              code: [{ repo: 'ktx', path: 'README.md#L168-L184', label: 'project layout' }],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: '复用 dbt 项目的 git 流程：semantic_models YAML 跟着 dbt model 一起走 PR。',
              code: [{ repo: 'dbt-sl', path: 'dbt_semantic_interfaces', label: 'dbt-semantic-interfaces' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Rich Context 改动在 UI 审批 + rc_change_log 留痕；当前非 git PR 流程。',
              notSupported: 'RC 存仓内表而非 git，事前评审走应用 UI 审批；缺 PR diff 的并发 / 分支能力。',
            },
            {
              vendor: 'Fabric DA',
              school: 'managed-cloud',
              desc: 'Business semantics 在 Fabric 门户里配置 + 审批，云内 UI 流程。',
              refs: ['fabric-data-agent'],
            },
          ],
        },
      ],
      commonSense:
        '**Git PR 是金标准**——历史可追溯、可回滚、与代码评审同流程。应用 UI 审批适合非技术团队但缺审计、缺并发处理。"自动通过"在生产场景永远不该出现。',
    },

    /* ─────── Q2: 事中反馈 ─────── */
    {
      id: 'during',
      question: '事中怎么收集反馈？',
      why: '生成出错时立刻能修，是上下文层"使用即沉淀"的入口。',
      steps: [
        {
          id: 'human-during-1',
          name: '反馈的形态：从 👍/👎 到改 SQL',
          desc: '👍/👎 + 文字 / 用户在线改 SQL / 错样本标注 / 系统主动澄清——信息量递增。',
          icon: 'i-lucide-message-square-heart',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '用户修正后的 SQL 回流 memory store，与原 NL 一起作为 few-shot——比 👍/👎 信息量高得多。',
              detail: {
                summary:
                  'WrenAI 最有价值的事中反馈不是评分，而是"用户改后的 SQL"——它把 (原始问题, 修正后 SQL) 这一对存进 memory store，下次相似问题直接召回当 few-shot。',
                bullets: [
                  {
                    label: '改后 SQL = 黄金样本',
                    icon: 'i-lucide-sparkles',
                    accent: 'amber',
                    body: '用户把错的 SQL 改对，这一对 (NL, SQL) 是人工标注过的最高质量样本——远胜一个 👍。',
                  },
                  {
                    label: '回流即召回源',
                    icon: 'i-lucide-recycle',
                    accent: 'emerald',
                    body: '存进 memory store 后，相似问题在 retrieval 阶段优先命中——一次修正惠及未来所有同类问题。',
                  },
                  {
                    label: '低摩擦',
                    icon: 'i-lucide-feather',
                    accent: 'violet',
                    body: '用户本来就要把 SQL 改对才能用——顺手回流不增加额外负担，是最廉价的精度来源。',
                  },
                ],
                closing: '一句话："👍/👎"几乎没有信息量；"用户改后的 SQL"才是上下文层最便宜的精度来源。',
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/seed_queries.py', label: 'memory/seed_queries.py' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '👍/👎 + 反馈文字，且支持把审过的对话提升为 Verified Query——评分 + 标准答案双通道。',
              refs: ['sf-cortex-overview', 'sf-vqr'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '用户在线改 SQL 回流 query_history + 系统主动澄清，澄清答案被记录为下次的上下文。',
            },
            {
              vendor: 'Fabric DA',
              school: 'managed-cloud',
              desc: '👍/👎 + 文字反馈，云内收集。',
              refs: ['fabric-data-agent'],
            },
          ],
        },
      ],
      commonSense:
        '**"👍/👎" 几乎没有信息量；最有用的是"用户改后的 SQL"——把它和原 NL 一起回流到 query_history，是上下文层最廉价的精度来源**。错样本标注（按 schema linking / fan-out / 时间窗归类）适合进阶团队。',
    },

    /* ─────── Q3: 事后 Eval ─────── */
    {
      id: 'after',
      question: '事后回归 / Eval 怎么做？',
      why: '没有 eval 就没有改进；只看用户反馈是 selection bias 严重。',
      steps: [
        {
          id: 'human-after-1',
          name: '回归基线：黄金集 + 抽样',
          desc: '团队维护"标准答案"集回归跑 + 随机抽 N% 生产对话人工核——eval 的最低配。',
          icon: 'i-lucide-clipboard-check',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'eval skill / 黄金 NL→SQL 集回归；执行结果对比 baseline——可挂进 PR 流程做门禁。',
              code: [
                { repo: 'wrenai', path: 'wren-ai-service/eval', label: 'eval/' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '云内提供准确率 / 使用 metrics；Verified Queries 充当事实上的回归基线。',
              refs: ['sf-cortex-overview', 'sf-vqr'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '黄金集 + 人工抽样核 + 执行结果对比；错样本回流驱动改进。',
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'dbt SL 自身不做 NL2SQL eval——由消费方维护问题→SQL 的回归集。',
              notSupported: 'dbt SL 是指标编译器，无 NL2SQL eval 概念；评测在上游消费方。',
            },
          ],
        },
        {
          id: 'human-after-2',
          name: 'LLM-as-judge：规模筛选',
          desc: '另一 LLM 给答案评分——规模大但精度有限，只能作初筛不能替代人工。',
          icon: 'i-lucide-gavel',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              desc: 'eval pipeline 可接 LLM-judge 做大规模初筛，再人工复核可疑样本。',
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'LLM-judge 跑全量初筛 + 人工抽样精核——两层结合控成本。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '主要依赖 Verified Queries + 人工审；LLM-judge 非内置一等能力。',
              refs: ['sf-vqr'],
            },
          ],
        },
      ],
      commonSense:
        '**黄金集 + 抽样核对 = 最低配；LLM-as-judge 适合规模筛选但不能替代抽样**。Eval 应该跑在 PR 流程里——契约改了 → 自动跑 eval → 通过才合并。',
    },

    /* ─────── Q4: 知识精炼 ─────── */
    {
      id: 'curate',
      question: '知识精炼谁来做？',
      why: '业务知识冷热数据不停变化；不持续精炼，wiki 会变成"垃圾场"。',
      steps: [
        {
          id: 'human-curate-1',
          name: '矛盾 / 重复 / 过时怎么清',
          desc: '半自动 grill（LLM 挑可疑条目 → 人审定）vs 完全人工 vs 自然淘汰。',
          icon: 'i-lucide-sparkles',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'enrich-context / grill：LLM 扫 wiki + MDL 找矛盾 / 重复 / 缺失，提候选给人评审 yes/no。',
              detail: {
                summary:
                  'WrenAI 的知识精炼走"LLM 提建议、人做决策"——enrich-context skill 扫描已有上下文，主动挑出矛盾、重复、缺 enum / 单位的条目，把人的精力从"扫地"解放到"拍板"。',
                bullets: [
                  {
                    label: 'LLM 当扫描器',
                    icon: 'i-lucide-scan-line',
                    accent: 'amber',
                    body: 'LLM 通读 wiki / MDL，标出"两条规则互相矛盾""这个 enum 缺含义""这列没单位"——人很难全量扫到的。',
                  },
                  {
                    label: '人当决策者',
                    icon: 'i-lucide-user-check',
                    accent: 'emerald',
                    body: '每个候选条目人只需 yes/no——把判断留给人，把苦力留给 LLM。',
                  },
                  {
                    label: '增量补全',
                    icon: 'i-lucide-plus-circle',
                    accent: 'violet',
                    body: '在已有 MDL 上从 raw 文档补 enum 含义 / 同义词 / cubes——精炼即扩充，不是单纯删减。',
                  },
                ],
                closing: '一句话：把人的精力放在"决策"上而不是"扫地"上——这是知识精炼能持续的唯一形态。',
              },
              code: [
                { repo: 'wrenai', path: 'wren-ai-service/src/pipelines/indexing', label: 'enrich-context' },
              ],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'memory agent grill：摄入 wiki 时自动去重 + 标矛盾 → 人评审，是 ktx 的核心卖点。',
              code: [{ repo: 'ktx', path: 'packages/cli/src/context', label: 'context engine (grill)' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'self-maintenance memory agent：扫 RC 找过时 / 矛盾，结合 schema diff 触发重生。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '云内对 Semantic Model 的维护 / 校验；矛盾检测非公开一等能力。',
              refs: ['sf-cortex-overview'],
            },
          ],
        },
      ],
      commonSense:
        '**"半自动 grill" 是最务实的形态**——LLM 把可疑条目挑出来（矛盾 / 重复 / 长期不被引用），人只需评审 yes/no。把人的精力放在"决策"上而不是"扫地"上。完全人工不可持续，自然淘汰过于乐观。',
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
