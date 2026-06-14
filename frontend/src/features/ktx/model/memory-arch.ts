import type { MemoryArch } from './ktx'

export const memoryArch: MemoryArch = {
  id: 'memory',
  input: {
    label: '对话尾 · userMessage + assistantMessage',
    note: 'MCP memory_ingest 或对话结束后异步触发 · sourceType: research | external_ingest | backfill',
    sources: ['Claude Code session', 'Codex session', 'MCP tool call transcript'],
  },
  signals: [
    { name: 'detectCaptureSignals', desc: '启发式判断对话是否含可沉淀知识（metric 定义 / 业务规则 / 列含义）' },
    { name: 'prefilterSkipReason', desc: '闲聊 / 纯查询无新事实 → 跳过，省 LLM 调用' },
    { name: 'promptNameFor', desc: '按 sourceType 选不同 system prompt' },
    { name: 'stepBudgetFor', desc: '按信号强度分配 agent loop 步数预算' },
  ],
  worktreeFlow: [
    { name: 'withLock(config:repo)', desc: '抓 main HEAD，创建 session worktree（独立分支）', icon: 'i-lucide-lock' },
    { name: 'worktree-scoped services', desc: 'wiki/SL 写入隔离在 worktree 内', icon: 'i-lucide-git-fork' },
    { name: 'agentRunner.runLoop', desc: 'LLM + load_skill tool + wiki/SL emit tools', icon: 'i-lucide-bot' },
    { name: 'slValidator', desc: 'warehouse probe rows 校验所有触动 SL source', icon: 'i-lucide-flask-conical' },
    { name: 'squash merge', desc: '通过则 squash 回 main；失败 revertSourceToPreHead', icon: 'i-lucide-git-merge' },
  ],
  toolBelt: [
    { name: 'load_skill', desc: '按需加载 wiki_capture / sl_capture / dbt_ingest 等 skill prompt' },
    { name: 'wiki emit-tools', desc: '创建/更新/删除 wiki 页' },
    { name: 'sl emit-tools', desc: '创建/更新/删除 SL source YAML' },
    { name: 'ToolSession.actions', desc: '工具结果队列 → telemetry + wiki↔SL ref 自动同步' },
  ],
  validation: {
    title: 'SL 校验闸门',
    items: [
      { name: 'slValidator', desc: '对触动 SL source 跑 warehouse probe（LIMIT 行）' },
      { name: 'revertSourceToPreHead', desc: 'probe 不通过 → 回滚该 source 到 pre-head' },
      { name: 'syncFromWiki', desc: 'MemoryKnowledgeSlRefsPort 自动同步 wiki 里的 SL 引用' },
    ],
    note: '写入永远经过 worktree → squash：主分支始终一致，并发请求互不污染',
  },
  resultRecord: [
    { field: 'signalDetected', desc: '是否检测到可沉淀信号' },
    { field: 'actions[]', desc: 'wiki created/updated/removed + sl created/updated/removed' },
    { field: 'skillsLoaded[]', desc: '本次加载的 skill 名' },
    { field: 'commitHash', desc: 'squash 后的 main HEAD' },
    { field: 'status / stage', desc: 'MemoryRunStore 异步状态（running / done / error）' },
  ],
  insights: {
    input: 'Memory Agent 是 ktx "自改进"的核心——不是微调模型，而是在对话尾把有价值事实显式写回 wiki/SL，下次 ingest/search 就能召回。',
    isolation: [
      { icon: 'i-lucide-git-fork', title: 'worktree 隔离写入', body: '每个 memory run 在独立 git worktree 分支上写，通过 SL 校验后才 squash 合并；并发 memory_ingest 与 ingest WU 互不污染主分支。' },
      { icon: 'i-lucide-shield-check', title: 'SL probe 闸门', body: '写 SL 不是"模型说了算"——必须过 warehouse probe 验证 SQL 能跑；失败自动 revert，保证语义层始终可执行。' },
    ],
    learn: [
      { icon: 'i-lucide-recycle', title: '显式沉淀、可评审', body: '学到的东西变成 wiki Markdown / SL YAML，走 git diff 评审——不是黑盒记忆向量，而是团队可审计的知识资产。' },
    ],
  },
}
