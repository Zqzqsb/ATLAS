<script setup lang="ts">
/**
 * CodeBlock — lightweight zero-dep syntax highlighter.
 *
 *   Goal: ditch the all-white <pre> look and give every supported lang a
 *   tasteful color scheme without pulling in shiki / prismjs / highlight.js.
 *
 *   Strategy: token-replace on a per-language regex set, escape HTML once,
 *   wrap matches in <span class="tk-*"> with a fixed palette. The output
 *   is a single innerHTML payload (safe because we escape first, then only
 *   inject our own span tags).
 *
 *   Supported langs: yaml, sql, python, bash/shell, json, ts/typescript,
 *   text (no highlight). Unknown lang falls back to text.
 */
import { computed } from 'vue'

const props = defineProps<{
  code: string
  lang?: string
}>()

const lang = computed(() => (props.lang || 'text').toLowerCase())

function escape(s: string): string {
  return s.replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c] as string))
}

/* ─── per-language tokenizers (regex applied IN ORDER, first match wins) ─── */
type Rule = { name: string; re: RegExp }

const RULES: Record<string, Rule[]> = {
  yaml: [
    { name: 'comment', re: /(^|\s)#[^\n]*/ },
    { name: 'string',  re: /"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/ },
    { name: 'key',     re: /^(\s*-\s*)?([A-Za-z_][\w-]*)(\s*:)/m },
    { name: 'number',  re: /\b\d+(?:\.\d+)?\b/ },
    { name: 'bool',    re: /\b(?:true|false|null|yes|no)\b/ },
    { name: 'punct',   re: /[\[\]{},|>]/ },
  ],
  sql: [
    { name: 'comment', re: /--[^\n]*|\/\*[\s\S]*?\*\// },
    { name: 'string',  re: /'(?:''|[^'])*'/ },
    { name: 'kw',      re: /\b(?:SELECT|FROM|WHERE|GROUP\s+BY|ORDER\s+BY|HAVING|LIMIT|OFFSET|JOIN|INNER|LEFT|RIGHT|FULL|OUTER|ON|AS|AND|OR|NOT|IN|IS|NULL|LIKE|BETWEEN|CASE|WHEN|THEN|ELSE|END|DISTINCT|UNION|ALL|WITH|CREATE|TABLE|VIEW|METRIC|INSERT|UPDATE|DELETE|INTO|VALUES|SET|TRUE|FALSE|MEASURE|DIMENSION|TIME_DIMENSION|FACT)\b/i },
    { name: 'fn',      re: /\b(?:COUNT|SUM|AVG|MIN|MAX|DATE_TRUNC|EXTRACT|CAST|COALESCE|NULLIF|ROUND|FLOOR|CEIL|ABS|NOW|CURRENT_DATE|CURRENT_TIMESTAMP)\b(?=\s*\()/i },
    { name: 'number',  re: /\b\d+(?:\.\d+)?\b/ },
    { name: 'punct',   re: /[(),;.]/ },
  ],
  python: [
    { name: 'comment', re: /#[^\n]*/ },
    { name: 'string',  re: /"""[\s\S]*?"""|'''[\s\S]*?'''|f?"(?:\\.|[^"\\])*"|f?'(?:\\.|[^'\\])*'/ },
    { name: 'kw',      re: /\b(?:def|class|return|if|elif|else|for|while|in|not|and|or|is|None|True|False|import|from|as|with|try|except|finally|raise|yield|lambda|pass|break|continue|global|nonlocal|async|await)\b/ },
    { name: 'number',  re: /\b\d+(?:\.\d+)?\b/ },
    { name: 'fn',      re: /\b[A-Za-z_]\w*(?=\s*\()/ },
    { name: 'punct',   re: /[(){}\[\],:.]/ },
  ],
  bash: [
    { name: 'comment', re: /#[^\n]*/ },
    { name: 'string',  re: /"(?:\\.|[^"\\])*"|'[^']*'/ },
    { name: 'prompt',  re: /^\$\s/m },
    { name: 'flag',    re: /(?:^|\s)(--?[A-Za-z][\w-]*)/ },
    { name: 'kw',      re: /\b(?:if|then|else|fi|for|do|done|while|case|esac|in|function|export|return|cd|ls|cat|grep|echo|pwd|mkdir|rm|cp|mv)\b/ },
    { name: 'number',  re: /\b\d+\b/ },
  ],
  json: [
    { name: 'string',  re: /"(?:\\.|[^"\\])*"(?=\s*:)/ },
    { name: 'value',   re: /"(?:\\.|[^"\\])*"/ },
    { name: 'number',  re: /-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b/ },
    { name: 'bool',    re: /\b(?:true|false|null)\b/ },
    { name: 'punct',   re: /[{}\[\],:]/ },
  ],
  ts: [
    { name: 'comment', re: /\/\/[^\n]*|\/\*[\s\S]*?\*\// },
    { name: 'string',  re: /`(?:\\.|[^`\\])*`|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/ },
    { name: 'kw',      re: /\b(?:const|let|var|function|return|if|else|for|while|class|interface|type|extends|implements|import|export|from|as|new|this|super|true|false|null|undefined|async|await|try|catch|finally|throw|in|of|typeof|instanceof|void|enum|public|private|protected|readonly|static)\b/ },
    { name: 'number',  re: /\b\d+(?:\.\d+)?\b/ },
    { name: 'fn',      re: /\b[A-Za-z_]\w*(?=\s*\()/ },
    { name: 'punct',   re: /[(){}\[\],:;.]/ },
  ],
  text: [],
}
RULES.shell = RULES.bash!
RULES.typescript = RULES.ts!
RULES.javascript = RULES.ts!
RULES.js = RULES.ts!

/** Color palette (mapped via class names → unocss inline rgb). */
const COLORS: Record<string, string> = {
  comment: 'text-gray-500 italic',
  string:  'text-emerald-300',
  value:   'text-emerald-300',
  key:     'text-sky-300',
  kw:      'text-violet-300 font-semibold',
  fn:      'text-amber-300',
  number:  'text-orange-300',
  bool:    'text-rose-300',
  punct:   'text-gray-400',
  prompt:  'text-emerald-400 font-bold',
  flag:    'text-cyan-300',
}

interface Token { type: string; value: string }

/** Tokenize using a "first regex that matches at position 0 wins" loop —
 *  o(n²) but for code snippets ≤ a few KB this is fine and avoids deps. */
function tokenize(src: string, rules: Rule[]): Token[] {
  if (!rules.length) return [{ type: 'text', value: src }]

  const compiled = rules.map((r) => ({
    name: r.name,
    re: new RegExp(r.re.source, r.re.flags.includes('m') ? 'my' : 'y'),
  }))

  const out: Token[] = []
  let i = 0
  let buf = ''
  while (i < src.length) {
    let matched: { name: string; text: string } | null = null
    for (const { name, re } of compiled) {
      re.lastIndex = i
      const m = re.exec(src)
      if (m && m.index === i) {
        // For yaml `key`, we have a leading-space capture group — keep
        // the prefix as text, the key as `key`, the trailing colon as punct.
        if (name === 'key' && m.length >= 4) {
          if (buf) { out.push({ type: 'text', value: buf }); buf = '' }
          if (m[1]) out.push({ type: 'text', value: m[1]! })
          out.push({ type: 'key', value: m[2]! })
          out.push({ type: 'punct', value: m[3]! })
          i += m[0].length
          matched = { name, text: m[0] }
          break
        }
        if (name === 'comment' && m[0].startsWith(' ')) {
          if (buf) { out.push({ type: 'text', value: buf }); buf = '' }
          const lead = m[0].match(/^\s+/)?.[0] ?? ''
          if (lead) out.push({ type: 'text', value: lead })
          out.push({ type: 'comment', value: m[0].slice(lead.length) })
          i += m[0].length
          matched = { name, text: m[0] }
          break
        }
        if (name === 'flag' && m[1]) {
          if (buf) { out.push({ type: 'text', value: buf }); buf = '' }
          const lead = m[0].slice(0, m[0].length - m[1].length)
          if (lead) out.push({ type: 'text', value: lead })
          out.push({ type: 'flag', value: m[1] })
          i += m[0].length
          matched = { name, text: m[0] }
          break
        }
        if (buf) { out.push({ type: 'text', value: buf }); buf = '' }
        out.push({ type: name, value: m[0] })
        i += m[0].length
        matched = { name, text: m[0] }
        break
      }
    }
    if (!matched) {
      buf += src[i]!
      i += 1
    }
  }
  if (buf) out.push({ type: 'text', value: buf })
  return out
}

const tokens = computed(() => {
  const rules: Rule[] = RULES[lang.value] ?? RULES.text ?? []
  return tokenize(props.code, rules)
})

const html = computed(() =>
  tokens.value
    .map((t) => {
      const cls = COLORS[t.type]
      const safe = escape(t.value)
      return cls ? `<span class="${cls}">${safe}</span>` : safe
    })
    .join(''),
)
</script>

<template>
  <pre
    class="px-3 py-3 text-[11.5px] font-mono text-gray-100 overflow-x-auto leading-relaxed border-t border-gray-700 whitespace-pre"
  ><code v-html="html" /></pre>
</template>
