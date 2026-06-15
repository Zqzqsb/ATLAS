import type { StageArch } from './comm'

export const verifyArch: StageArch = {
  id: 'verify',
  abstract:
    '"看起来对" ≠ "可执行 ∧ 语义正确"。把校验拆成静态 / 策略 / 预执行 / 语义合理性四层闸门，能在执行前堵住 95% 的错。',
  principles: [
    {
      name: '失败应给出结构化诊断',
      desc: '不是 "syntax error near X"，而是 phase / code / 列名 / 期望类型——让 Agent 能据此精确修复。',
    },
    { name: '默认只读', desc: 'NL2SQL 系统永远只跑 SELECT，从源头切断 destructive 风险。' },
    {
      name: '在 live DB 之前先在沙箱 / dry-run',
      desc: '错的 SQL 不应该直接到生产；LIMIT 0 / EXPLAIN 几乎零成本能挡掉一类错。',
    },
    {
      name: '语义合理性比语法更难也更重要',
      desc: 'fan-out / 时间窗 / 单位 / NULL 语义都是"语法对了但答案错"的高发地带。',
    },
  ],
  subQuestions: [
    /* ─────── Q1: 静态校验 ─────── */
    {
      id: 'static',
      question: '静态校验做哪些？',
      why: '低成本就能拦掉一类错——且能给 Agent 最早期的反馈。',
      steps: [
        {
          id: 'verify-static-1',
          name: 'parse + qualify：语法 + 限定到 schema',
          desc: '语法解析 + 把表 / 列限定（qualify）到具体 schema——最基础也最早期的闸门。',
          icon: 'i-lucide-file-search',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'sqlglot 解析 + wren-core 把引用 qualify 到 MDL model / column；未建模引用直接报错。',
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (parse/qualify)' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '生成的 SQL 由 Snowflake 内置编译器解析 + 限定；越界引用在云内被拒。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'verify_sql 第一关就是 parse + 列限定到 catalog；失败结构化报错。',
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'ktx-daemon analyze：解析 SQL 并限定到 semantic-layer 已知实体。',
              code: [{ repo: 'ktx', path: 'python/ktx-daemon', label: 'ktx-daemon analyze' }],
            },
          ],
        },
        {
          id: 'verify-static-2',
          name: 'type-check + 引用存在',
          desc: '函数 / 表达式参数类型 + 引用列 / 主键是否存在于 catalog。',
          icon: 'i-lucide-spell-check-2',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'wren-core 做类型检查 + 从 MDL / catalog 校验列存在；类型不匹配在编译期暴露。',
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (type-check)' }],
            },
            {
              vendor: 'Cube',
              school: 'semantic-layer',
              desc: 'Cube schema-compiler 编译期就校验 measure / dimension 的类型与引用。',
              code: [{ repo: 'cube', path: 'packages/cubejs-schema-compiler', label: 'schema-compiler' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'verify_sql 校验引用列存在于 catalog；不存在则报"列不存在 + 候选列"让 Agent 重绑。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '云内编译器做类型 / 引用校验；细节不可见，但越界即拒。',
              refs: ['sf-cortex-overview'],
            },
          ],
        },
      ],
      commonSense:
        'parse + qualify + 列存在校验是必须的——sqlglot / DataFusion 这种引擎已经能免费做掉，不用就是浪费。"无静态校验、直接进 dry-run"是基线。',
    },

    /* ─────── Q2: 策略校验 ─────── */
    {
      id: 'policy',
      question: '策略校验包含什么？',
      why: '这一层决定"安全"——读写权限、行/列权限、被禁的危险函数。',
      steps: [
        {
          id: 'verify-policy-1',
          name: 'read-only 黑名单',
          desc: '禁 INSERT / UPDATE / DELETE / ALTER / CREATE / DROP / TRUNCATE / GRANT / REVOKE——NL2SQL 的底线。',
          icon: 'i-lucide-lock',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'wren-core 只生成 SELECT；非读语句在解析期就被拒，永不到达数据库。',
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '生成范围限定为只读查询；叠加 Snowflake 账户级 RBAC 双重保险。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'ktx-daemon validate-read-only：硬性拒绝一切非 SELECT 语句，连接本身也是只读。',
              code: [{ repo: 'ktx', path: 'python/ktx-daemon', label: 'validate-read-only' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'dryrun gate：执行前黑名单拦截 destructive 语句，只放行 SELECT。',
            },
          ],
        },
        {
          id: 'verify-policy-2',
          name: 'RLAC / CLAC：行列级权限',
          desc: 'session property 注入 WHERE（行级）/ 控制列可见性（列级）——建模即治理。',
          icon: 'i-lucide-shield-half',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '在语义层定义 RLAC / CLAC，wren-core 在 SQL 展开时注入 WHERE / 隐藏列——LLM 无法绕过。',
              detail: {
                summary:
                  'WrenAI 把权限当作"语义层的一等公民"——RLAC（行）和 CLAC（列）写在 MDL 里，由 wren-core 在展开 SQL 时强制注入，而不是在应用层补 WHERE（那一定会被绕过）。',
                bullets: [
                  {
                    label: 'RLAC = 自动注入 WHERE',
                    icon: 'i-lucide-rows-3',
                    accent: 'emerald',
                    body: 'session 里携带 user.region，引擎在每条查询自动追加 `WHERE region = @user.region`——用户写的 SQL 看不到这层。',
                  },
                  {
                    label: 'CLAC = 列可见性',
                    icon: 'i-lucide-columns-3',
                    accent: 'amber',
                    body: '敏感列对无权限用户在 model 层就不暴露——和"选择性暴露"同机制，权限不足者根本看不到该列。',
                  },
                  {
                    label: '为什么必须在引擎层',
                    icon: 'i-lucide-shield-check',
                    accent: 'violet',
                    body: 'LLM 写的是逻辑 SQL、碰不到物理表——权限在展开时注入，从架构上杜绝"绕过 SL 直连库"的漏洞。',
                  },
                ],
                closing: '一句话：权限写进契约、引擎执法——应用层补防火墙是反模式。',
              },
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (RLAC/CLAC)' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '直接复用 Snowflake 的行访问策略 / 列遮蔽 / RBAC——权限在仓库层，天然不可绕过。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'Cube',
              school: 'semantic-layer',
              desc: 'Cube row-level security：在 cube 定义里写 queryRewrite / securityContext 注入过滤。',
              code: [{ repo: 'cube', path: 'packages/cubejs-server-core', label: 'security context' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'session policy 注入：按用户上下文在生成的 SQL 上追加行过滤。',
            },
          ],
        },
        {
          id: 'verify-policy-3',
          name: '危险函数 + 行数上限',
          desc: '黑名单时间敏感 / 危险 UDF；默认 LIMIT 100 · 硬顶防"全表扫"。',
          icon: 'i-lucide-gauge',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'policy 层：denied funcs 黑名单 + 默认 LIMIT / 硬顶，防止昂贵或不确定查询。',
              selfContained: true,
            },
            {
              vendor: 'Cube',
              school: 'semantic-layer',
              desc: 'Cube 限制查询行数 / 时间范围，配合 pre-aggregation 防全表扫。',
              selfContained: true,
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '结果行数 / 仓库资源由 Snowflake warehouse 配额与超时托管。',
              refs: ['sf-cortex-overview'],
            },
          ],
        },
      ],
      commonSense:
        'read-only 是 NL2SQL 不可商量的底线；RLAC / CLAC 应该写在语义层（建模即治理），而不是在应用层补——后者一定会被绕过。',
    },

    /* ─────── Q3: 预执行 ─────── */
    {
      id: 'dryrun',
      question: '怎么"预执行"？',
      why: '"在 live DB 跑一下"是最有效的验证，但要做得便宜（不返回行）。',
      steps: [
        {
          id: 'verify-dryrun-1',
          name: 'dry-plan：不连库展开逻辑',
          desc: '把 modeled SQL 展开成物理 SQL——不连库就能看 join / 计算列对不对。',
          icon: 'i-lucide-route',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'wren-core dry-plan：纯本地把逻辑 SQL 展开成物理 SQL + 命中模型清单——零数据库成本验证逻辑。',
              detail: {
                summary:
                  'dry-plan 是 WrenAI 独有的"便宜验证"——它不连数据库，仅靠 MDL 就能把 LLM 写的逻辑 SQL 展开成物理 SQL，让你在花任何 DB 成本前就看清 join 路径和计算列是否正确。',
                bullets: [
                  {
                    label: '不连库',
                    icon: 'i-lucide-plug-zap',
                    accent: 'emerald',
                    body: '展开完全靠 MDL 元数据 + wren-core——不发起任何数据库连接，因此可以在生成循环里反复跑，几乎零成本。',
                  },
                  {
                    label: '看 join / CTE 展开',
                    icon: 'i-lucide-git-merge',
                    accent: 'amber',
                    body: '输出展开后的物理 SQL + 用到的 model / relationship / 计算列清单——逻辑错（join 错表）在这一步就暴露。',
                  },
                  {
                    label: '兼做可解释性',
                    icon: 'i-lucide-eye',
                    accent: 'violet',
                    body: '同一份 dry-plan 输出可直接作为"这个数字怎么来的"解释给业务看——校验和 lineage 一举两得。',
                  },
                ],
                closing: '一句话：dry-plan 把"验证逻辑"和"解释逻辑"合二为一，且不花一分钱 DB 成本。',
              },
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (dry-plan)' }],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'MetricFlow 编译期把 metric 展开成 SQL——本质也是"不执行先展开"，可在 CI 校验。',
              code: [{ repo: 'metricflow', path: 'metricflow/engine', label: 'metricflow/engine' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '无对用户暴露的"不连库 dry-plan"；可解释性靠返回生成的 SQL + 命中语义对象。',
              notSupported: 'Cortex 不开放独立的离线展开步骤——生成与执行都在云内，中间态对用户黑盒。',
              refs: ['sf-cortex-overview'],
            },
          ],
        },
        {
          id: 'verify-dryrun-2',
          name: 'dry-run：到 live DB 但不返回行',
          desc: 'LIMIT 0 / EXPLAIN——提交 SQL 验证语法 / 列 / 权限，但不真正取数。',
          icon: 'i-lucide-flask-conical',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'dry-run：把展开后的物理 SQL 以不返回行的方式提交，验证语法 / 列 / 权限在真实库成立。',
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (dry-run)' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'verify_sql 的 dry-run 关：LIMIT 0 提交到目标库，捕获真实的语法 / 权限错误。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '执行在云内托管；可设 LIMIT，但"预执行验证"对用户不单独暴露。',
              refs: ['sf-cortex-overview'],
            },
          ],
        },
      ],
      commonSense:
        '**dry-plan + dry-run 双闸**：dry-plan 不连库就能验证逻辑（join 对不对、计算列展开对不对）；dry-run 再到 live DB 验证语法 / 列 / 权限。这两步一个没有都是赌博。一次 dry-run（LIMIT 0）几乎免费——应放在所有 LLM retry 之前。',
    },

    /* ─────── Q4: 语义合理性 ─────── */
    {
      id: 'semantic',
      question: '语义合理性怎么校验？',
      why: '语法 OK、能跑出数据，结果可能依然是错的——这是上下文层的硬骨头。',
      steps: [
        {
          id: 'verify-semantic-1',
          name: 'fan-out / chasm trap 检测',
          desc: '多对多关系下 SUM 重复计数——planner 应按 grain 自动拆 CTE，而非靠人记。',
          icon: 'i-lucide-alert-triangle',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'wren-core planner 按 measure 的 grain 检测 fan-out，自动拆 CTE 聚合后再 join，避免重复计数。',
              detail: {
                summary:
                  'fan-out 是 NL2SQL 最隐蔽的坑——1-N join 后做 SUM 会把"一"的那行数据按"多"的条数重复累加。WrenAI 让 planner 按 measure 的粒度自动处理，而不是指望 LLM 或人记得每个 measure 在每张表的 grain。',
                bullets: [
                  {
                    label: '问题',
                    icon: 'i-lucide-bug',
                    accent: 'rose',
                    body: 'customers 1-N orders，若先 join 再 `SUM(customers.credit)`，每个客户的额度会被按订单数重复计——结果偷偷放大。',
                  },
                  {
                    label: '引擎解法',
                    icon: 'i-lucide-layers',
                    accent: 'emerald',
                    body: 'planner 识别 measure 的 grain，把不同粒度的聚合拆进独立 CTE 各自 GROUP BY，再按 key join——数学上正确。',
                  },
                  {
                    label: '为什么不能靠人',
                    icon: 'i-lucide-brain',
                    accent: 'amber',
                    body: '一个 measure 在 5 张表里有 5 种 grain——人和 LLM 都记不全；这必须是引擎的确定式职责。',
                  },
                ],
                closing: '一句话：fan-out 必须由 planner 解决——这是语义引擎相对"LLM 直接写 SQL"最硬的优势之一。',
              },
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (planner)' }],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'ktx-sl 在构建 join 图时就检测并解决 chasm / fan trap——是它"自动建语义层"的核心卖点。',
              code: [{ repo: 'ktx', path: 'python/ktx-sl', label: 'ktx-sl (join graph)' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Semantic View 的 relationship + measure 定义让引擎按定义聚合，规避 fan-out；细节黑盒。',
              refs: ['sf-semantic-views'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '当前靠召回的关系 RC + 校验闸门提示 Agent 注意 grain；尚无引擎级 fan-out 自动拆解。',
              notSupported: 'ATLAS 暂无 planner 级 fan-out 自动检测；依赖 RC 关系提示 + 人工 review 兜底。',
            },
          ],
        },
        {
          id: 'verify-semantic-2',
          name: '单位 / 时间窗 / NULL 语义',
          desc: '"上月"是哪个时区起止？金额是分还是元？COUNT(*) vs COUNT(col)——靠 wiki / 列描述写清。',
          icon: 'i-lucide-ruler',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '单位 / 时间窗 / NULL 语义写进 MDL column description + instructions，召回时注入 prompt。',
              selfContained: true,
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Rich Context 的"单位 / 业务规则"类目专门承载这些——linking 时随列一起召回。',
              selfContained: true,
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: 'Semantic Model 的 description / instructions 字段承载单位与口径约定。',
              refs: ['sf-semantic-yaml'],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'metric 的 description + time spine 定义时间窗口；单位约定靠 description 文本。',
              code: [{ repo: 'dbt-sl', path: 'dbt_semantic_interfaces/type_enums', label: 'time spine' }],
            },
          ],
        },
        {
          id: 'verify-semantic-3',
          name: 'LLM-as-judge / 人工抽样',
          desc: '另一 LLM 检查"答案是否回答了原问题" + 高频问题人工核——兜底语义正确性。',
          icon: 'i-lucide-scale',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              desc: 'eval pipeline 可接 LLM-as-judge 跑黄金集回归；人工抽样由团队流程承载。',
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '生产前 / 高频问题人工核 + 错样本回流；LLM-judge 用于规模筛选。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: 'Verified Queries 让人工审过的 SQL 成为"标准答案"，是 Cortex 版的语义兜底。',
              refs: ['sf-vqr'],
            },
          ],
        },
      ],
      commonSense:
        '**fan-out / chasm 必须由 planner 解决**（人不可能记得每个 measure 在每张表的 grain）。**单位 / 时间窗 / NULL 必须由 wiki / 列描述写清**，不能依赖 LLM 现猜。LLM-as-judge 适合规模筛选，但不能替代人工抽样。',
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
