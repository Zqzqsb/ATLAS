import type { StageArch } from './comm'

export const memoryArch: StageArch = {
  id: 'memory',
  abstract:
    '"使用即沉淀（成功 → 知识 / 失败 → 任务）+ 解释（lineage / dry-plan）+ 权限（RLAC/CLAC）+ 漂移监控"——四件套缺一不可。',
  principles: [
    {
      name: '记忆是显式资产',
      desc: '不是黑盒向量；是 wiki Markdown / SL YAML / query_history 行——可审计、可 diff、可版本回滚。',
    },
    {
      name: '可解释 = 生成轨迹',
      desc: 'dry-plan / lineage 把 LLM 的选择展开为"用了哪些模型 / join / 列"；让数据团队能 review。',
    },
    {
      name: '权限是契约的一部分',
      desc: 'RLAC / CLAC 写在语义层，规划阶段强制执行——而不是在应用层补防火墙。',
    },
    {
      name: '漂移监控比 ML 重要',
      desc: 'schema 变了、表名换了、列删了——这些"软失败"远比模型退化常见，必须在数据层检测并主动失效相关 RC。',
    },
  ],
  subQuestions: [
    /* ─────── Q1: 使用即沉淀 ─────── */
    {
      id: 'learn',
      question: '"使用即沉淀"沉淀什么？',
      why: '沉淀对象决定下次召回的精度天花板。',
      steps: [
        {
          id: 'memory-learn-1',
          name: '沉淀进什么载体',
          desc: '确认的 NL-SQL 对 / 失败任务卡 / wiki 增量 / 黑盒 RLHF 微调——载体决定可审计性。',
          icon: 'i-lucide-archive',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '确认的 NL→SQL 对进 memory store 作 few-shot；澄清 / 新规则进 instructions——全是可读资产。',
              detail: {
                summary:
                  'WrenAI 的"学习"不是改模型权重，而是把每次成功的 (NL, SQL) 对、每次澄清答案显式写进 memory store / instructions——可读、可 diff、可回滚，越用越准而不会被污染。',
                bullets: [
                  {
                    label: '成功 → few-shot',
                    icon: 'i-lucide-check-circle-2',
                    accent: 'emerald',
                    body: '确认对的 (问题, SQL) 进 memory store，下次相似问题在 retrieval 阶段优先命中。',
                  },
                  {
                    label: '澄清 → instructions',
                    icon: 'i-lucide-pencil-line',
                    accent: 'amber',
                    body: '一次澄清（"上月=自然月、含税"）写回 instructions——下次同类问题不再追问。',
                  },
                  {
                    label: '可逆、不污染',
                    icon: 'i-lucide-undo-2',
                    accent: 'violet',
                    body: '沉淀进文件 / 行而非权重——混入垃圾可以 git revert，不像 fine-tune 一旦脏了洗不干净。',
                  },
                ],
                closing: '一句话：沉淀进可读资产（memory store / instructions），不要沉淀进模型权重。',
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/store.py', label: 'memory/store.py' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '👍/👎 + 把审过的对话提升为 Verified Query——VQR 是 Cortex 版的"成功沉淀"。',
              refs: ['sf-vqr'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'memory agent 把澄清 / 新规则 / 矛盾消解增量写回 wiki + semantic-layer——可读资产。',
              code: [{ repo: 'ktx', path: 'packages/cli/src/context', label: 'context engine' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '确认的 NL-SQL 进 query_history + RC 变更进 rc_change_log；失败对话进任务队列。',
            },
          ],
        },
      ],
      commonSense:
        '**沉淀进可读资产 (wiki / YAML / query_history)，不要沉淀进模型权重**——前者可评审、可回滚，后者一旦混入垃圾就洗不干净。失败任务卡（错对话进队列）适合进阶治理。',
    },

    /* ─────── Q2: 可解释性 ─────── */
    {
      id: 'lineage',
      question: '可解释性怎么暴露？',
      why: '业务团队 / 监管 / 数据团队都需要知道"这个数字怎么来的"。',
      steps: [
        {
          id: 'memory-lineage-1',
          name: '"这个数字怎么来的"怎么展示',
          desc: 'dry-plan 展开轨迹 / SQL + 命中模型清单 / EXPLAIN 物理计划 / 纯输出不解释。',
          icon: 'i-lucide-route',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'dry-plan：不连库展开 modeled SQL + 列出命中的 model / join / 计算列——面向业务最有用的解释。',
              detail: {
                summary:
                  'WrenAI 的可解释性复用了 dry-plan——同一份"不连库展开"既是校验手段，也是给业务看的 lineage：它把 LLM 的黑盒选择摊开成"用了 customers、orders（1-N）、lifetime_value 计算列"这种人话。',
                bullets: [
                  {
                    label: '展开 = 解释',
                    icon: 'i-lucide-eye',
                    accent: 'emerald',
                    body: 'dry-plan 输出物理 SQL + 命中对象清单——业务不用读 SQL，看"用了哪些模型 / 关系 / 计算列"就懂。',
                  },
                  {
                    label: '比 EXPLAIN 友好',
                    icon: 'i-lucide-smile',
                    accent: 'amber',
                    body: 'EXPLAIN 是物理计划、偏 DBA；命中模型清单是业务语言——受众不同，后者才是"前门"。',
                  },
                  {
                    label: '可审计',
                    icon: 'i-lucide-file-check',
                    accent: 'violet',
                    body: '每次查询都能留下"用了什么"的轨迹——监管 / 数据团队 review 有据可查。',
                  },
                ],
                closing: '一句话：dry-plan 把"验证逻辑"和"解释逻辑"合二为一——解释力就是团队信任。',
              },
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (dry-plan)' }],
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              primary: true,
              desc: 'Unity Catalog lineage：表 / 列级血缘 + Metric View 定义，平台原生可视化。',
              refs: ['dbx-genie', 'dbx-mv-ref'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'verify_sql 链 + 生成 SQL + 命中 RC 清单作为解释；偏 SQL 级而非引擎展开。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '返回生成的 SQL + 命中的 Semantic View 对象作为解释。',
              refs: ['sf-semantic-views'],
            },
          ],
        },
      ],
      commonSense:
        '**dry-plan + 命中模型清单 是面向业务最有用的解释**；EXPLAIN 太底层、SQL 太长——把"用了 customers · orders 1-N · lifetime_value 计算列"这种摘要做出来，业务才看得懂。"纯输出、只给 SQL 和结果"是基线。',
    },

    /* ─────── Q3: 权限 ─────── */
    {
      id: 'access',
      question: '权限模型在哪一层？',
      why: '权限放错层会"看似有但实际能绕过"——这是上下文层的高发漏洞。',
      steps: [
        {
          id: 'memory-access-1',
          name: '权限注入在哪一层',
          desc: '语义层 RLAC/CLAC（引擎注入）/ 仓库 RLS（DB 自带）/ 应用层补 WHERE（反模式）/ 无权限。',
          icon: 'i-lucide-shield-half',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '语义层 RLAC / CLAC：规划阶段 wren-core 注入；session property → WHERE / 列可见性，无法绕过。',
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (RLAC/CLAC)' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '仓库级：直接复用 Snowflake 行访问策略 / 列遮蔽 / RBAC——权限在数据层，天然兜底。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'Cube',
              school: 'semantic-layer',
              desc: 'Cube row policies：securityContext 在查询重写时注入过滤。',
              code: [{ repo: 'cube', path: 'packages/cubejs-server-core', label: 'security context' }],
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Unity Catalog ACL + 行列级策略——仓库层统一治理。',
              refs: ['dbx-mv-ref'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'session policy 注入行过滤；建议叠加仓库 RLS 双保险。',
            },
          ],
        },
      ],
      commonSense:
        '**语义层 + 仓库 RLS 双重保险**：语义层 RLAC 避免"绕过 SL 直接连库"的漏洞；仓库 RLS 是最后兜底。应用层补 WHERE 是反模式——一定会被绕过。',
    },

    /* ─────── Q4: 漂移监控 ─────── */
    {
      id: 'observe',
      question: '可观测 / 漂移监控做什么？',
      why: '生产环境的"软失败"——schema 变了、列删了、enum 多了一个值——比模型退化常见 10 倍。',
      steps: [
        {
          id: 'memory-observe-1',
          name: 'Schema 漂移：检测 + 自动失效',
          desc: '定期 introspect 比对，列变化 → 失效相关 RC → 触发自愈重生（而非只发告警）。',
          icon: 'i-lucide-activity',
          takes: [
            {
              vendor: 'ATLAS',
              school: 'agentic',
              primary: true,
              desc: 'self-maintenance heal-loop：schema diff → 标记过时 RC is_expired → 重 introspect + 重生 RC。',
              detail: {
                summary:
                  'ATLAS 把"漂移"当作一等公民——self-maintenance 周期性 introspect 比对 schema，发现列删 / 改名 / 新 enum 就把相关 Rich Context 打上 is_expired，并触发自愈重生，而不是发一封没人看的告警邮件。',
                bullets: [
                  {
                    label: '检测软失败',
                    icon: 'i-lucide-search',
                    accent: 'amber',
                    body: '定期 introspect 比对上次快照——列删 / 改名 / 类型变 / enum 多值，都是"语法没报错但语义已坏"的软失败。',
                  },
                  {
                    label: '主动失效 RC',
                    icon: 'i-lucide-ban',
                    accent: 'rose',
                    body: '相关列描述 / 关系 / measure 打 is_expired——下次召回直接跳过，避免用过时上下文写出错 SQL。',
                  },
                  {
                    label: '自愈重生',
                    icon: 'i-lucide-refresh-cw',
                    accent: 'emerald',
                    body: '失效后触发重 introspect + 重 embed——让上下文"静止时也保持新鲜"，不沦为历史快照。',
                  },
                ],
                closing: '一句话：漂移监控必须自动失效 + 自愈，光发邮件没人看。',
              },
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'MDL 与库 schema 比对可发现漂移；自动失效 / 重生当前更多由消费方流程承载。',
              notSupported: 'WrenAI 暂无内置的"漂移→自动失效 RC→自愈"闭环；schema 变更后需重跑 introspect / 索引。',
              code: [{ repo: 'wrenai', path: 'wren-ai-service/src/pipelines/indexing', label: 'indexing/' }],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'scan diff（增量）：重跑 ingest 时比对，更新变化的 model / wiki 条目。',
              code: [{ repo: 'ktx', path: 'packages/cli/src/connectors', label: 'scan connectors' }],
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'UC monitoring + lineage 感知 schema 变更；失效策略平台托管。',
              refs: ['dbx-genie'],
            },
          ],
        },
        {
          id: 'memory-observe-2',
          name: '准确率 / 延迟 / 成本 + 审计日志',
          desc: '黄金集滚动准确率 · p50/p99 延迟 · token 成本 · 全量 SQL / 工具调用落审计日志。',
          icon: 'i-lucide-gauge',
          takes: [
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '准确率看板 + 延迟 / 成本 telemetry + 全量审计日志（含 user / session）。',
              selfContained: true,
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '云内 usage / 准确率 metrics + Snowflake 账户级审计。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              desc: 'eval 黄金集滚动准确率；延迟 / 成本 telemetry 由部署侧承载。',
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '隐私友好的本地 telemetry（不记 SQL / schema / 路径），可 opt-out。',
              code: [{ repo: 'ktx', path: 'README.md#L260-L272', label: 'telemetry' }],
            },
          ],
        },
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
