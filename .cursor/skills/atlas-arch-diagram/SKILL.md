---
name: atlas-arch-diagram
description: >-
  Build and extend the ATLAS `/arch` route — a presentation-style panoramic
  architecture (L0) that zoom-focuses into per-module dataflow + internal detail
  (L1). Use when adding/editing the architecture overview, wiring a module's
  drill-down, authoring dataflow steps, strategy/prompt/storage/insight sections,
  or when the user mentions arch route, 全景架构, 模块内部细节, dataflow, or
  features/arch components.
---

# ATLAS Architecture Diagram (`/arch`)

A componentized, data-driven architecture presentation under
`frontend/src/features/arch/`. Two zoom levels:

- **L0 Panorama** — the whole system at a glance (layers → nodes).
- **L1 Module detail** — click a drillable node → zoom-focus into its
  end-to-end **dataflow** plus a dense, **non-expandable** detailed architecture
  (dispatch strategy, prompt engineering, storage, insights). L1 is the deepest
  level; do not add further drill-downs inside it.

## File Map

```
features/arch/
├── index.vue                     # stage: overview<->module state machine + zoom transition
├── model/
│   ├── architecture.ts           # L0 source of truth: ACCENTS, ARCH_LAYERS[]
│   ├── flows.ts                  # L1 dataflow steps: FlowDef + flows[] + getFlow()
│   └── modules.ts                # L1 detail sections: ModuleData + MODULES + getModule()
└── components/
    ├── overview/ ArchOverview · ArchLayer · ArchNode   # render ARCH_LAYERS
    └── module/
        ├── ModuleDetail.vue      # generic header + ribbon + REGISTRY dispatch by flow.id
        ├── DataflowStepper.vue   # animated step-by-step dataflow (reusable)
        ├── sections/             # reusable L1 section primitives
        │   SectionHeading · StrategySection · PromptSection · StorageSection · InsightSection
        └── modules/              # per-module composition: <Xxx>Detail.vue
```

## Core Principles

1. **Data-driven.** Layers/nodes/steps/sections live in `model/*.ts`. Components
   are dumb renderers. Reshape the diagram by editing data, not markup.
2. **Ground every detail in real backend code.** Read the actual Go source
   (`backend/internal/...`, `backend/server/handlers/...`) before writing copy.
   Numbers, table names, thresholds, prompt rules must match the code.
3. **Insight over decoration.** Each module ends with a "why we designed it this
   way" insight section — that is the point of the page.
4. **UnoCSS needs literal class strings.** Never interpolate color classes
   (`` `bg-${c}-100` ``); UnoCSS won't generate them. Use the `ACCENTS` map in
   `architecture.ts`, which holds full literal class strings per accent.

## Accent System

`ACCENTS[key]` (key: `slate|emerald|blue|amber|violet|indigo`) exposes literal
class bundles: `bar, dot, surface, iconBg, iconText, hover, text, chip, gradient`.
Add a new color by adding a full entry (all literal strings) — never build class
names dynamically.

## Recipe A — Edit the L0 Panorama

Edit `model/architecture.ts` → `ARCH_LAYERS`. Each layer = `{ id, title, subtitle,
icon, accent, cols, nodes[] }`; each node = `{ id, label, sublabel, icon, accent,
span?, flow?, codeRefs? }`.

- A node becomes **drillable** only when `flow` is set (matches a `FlowDef.id`).
- `codeRefs` documents the backend files the node maps to.
- No component edits needed — `ArchOverview` renders the array.

## Recipe B — Add a Module's Internal Detail (the template)

Copy how `onboarding` is built. Steps:

```
- [ ] 1. Research the real backend flow (handlers + scenarios + tools + models)
- [ ] 2. Add dataflow steps   → model/flows.ts  (FlowDef in flows[])
- [ ] 3. Add detail sections  → model/modules.ts (ModuleData in MODULES)
- [ ] 4. Compose the page      → components/module/modules/<Xxx>Detail.vue
- [ ] 5. Register              → REGISTRY in ModuleDetail.vue
- [ ] 6. Make the node drillable → set node.flow in architecture.ts
- [ ] 7. Verify (typecheck + dev server)
```

**Step 2 — dataflow (`flows.ts`).** Add a `FlowDef { id, label, title, subtitle,
icon, accent, steps[] }`. Each step = `{ id, title, subtitle, icon, accent,
summary, detail, artifact }` where `artifact = { input, output, store, code, lang }`.
This drives the animated `DataflowStepper` (auto-play + clickable steps + code panel).

**Step 3 — detail sections (`modules.ts`).** Add a `ModuleData { id, strategy,
prompt, storage, insights }`. `id` must equal the `FlowDef.id`. Section shapes:

| Section | Purpose | Shape |
|---|---|---|
| `strategy` | task registration / dispatch, branch handling (e.g. small vs large) | `{ title, subtitle, decision, options[] }` |
| `prompt` | prompt-engineering recipe | `{ title, subtitle, engine, tools[], blocks[], rules[] }` |
| `storage` | what gets persisted; grouped by `StorageKind` (`schema/context/catalog/log`) | `{ title, subtitle, items[] }` |
| `insights` | key design decisions + rationale | `{ title, subtitle, items[] }` |

**Step 4 — composition.** Create `modules/<Xxx>Detail.vue` taking `defineProps<{
flow: FlowDef }>()`, look up `getModule(flow.id)`, and stack sections separated by
`<hr class="border-gray-100" />`:

```
<SectionHeading icon=... title="端到端 Dataflow" .../>  <DataflowStepper :flow="flow" />
<StrategySection :data="mod.strategy" />
<PromptSection :data="mod.prompt" />
<StorageSection :data="mod.storage" />
<InsightSection :data="mod.insights" />
```

**Step 5 — register.** In `ModuleDetail.vue` add to `REGISTRY`:
`{ <id>: <Xxx>Detail }`. Modules without an entry fall back to the bare
`DataflowStepper`.

## Section Primitives (reuse, don't reinvent)

- `SectionHeading` — eyebrow icon + title + subtitle (every section starts with one).
- `StrategySection` — decision pill + N strategy cards with bullet points.
- `PromptSection` — left: numbered prompt building blocks; right: rules/constraints.
- `StorageSection` — tables grouped by kind (basic schema → context → vector catalog → log).
- `InsightSection` — 2-col "why" cards.
- `DataflowStepper` — the animated pipeline; reused as the first section.

If a module needs a genuinely new layout, add a new primitive under
`sections/` (data type in `modules.ts`) rather than hard-coding markup in the
composition file.

## Conventions

- Prose can be Simplified Chinese; keep code/paths/table names in English.
- Keep `index.vue` logic-free beyond the state machine + zoom transition.
- Zoom transition origin is computed from the clicked node's position — keep it
  subtle (~420ms scale+fade), not flashy.
- Don't put backend-inaccurate placeholder content; if unknown, read the code first.

## Verify

```bash
cd frontend && ./node_modules/.bin/vue-tsc -b --force        # typecheck (expect EXIT=0)
# with dev server running, confirm new files transform (HTTP 200):
curl -s -o /dev/null -w "%{http_code}\n" \
  http://localhost:5173/src/features/arch/components/module/modules/<Xxx>Detail.vue
```

Then open `/arch`, drill into the node, and check: dataflow plays, sections
render, accents/icons resolve (no missing UnoCSS classes), Esc/breadcrumb return.
