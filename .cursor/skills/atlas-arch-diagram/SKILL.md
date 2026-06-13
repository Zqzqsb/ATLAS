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
- **L1 Module detail** — click a drillable node → zoom-focus into **one internal
  architecture diagram with clear module boundaries**. L1 is the deepest level;
  it is **not** a stack of separate section-graphs and has **no further
  drill-down**. Everything (prompt rules, produced data shapes, storage tables)
  is a *detail/annotation on the diagram*, not its own section.

## File Map

```
features/arch/
├── index.vue                     # stage: overview<->module state machine + zoom transition
├── model/
│   ├── architecture.ts           # L0 source of truth: ACCENTS, ARCH_LAYERS[]
│   ├── flows.ts                  # module identity: FlowDef + flows[] + getFlow()
│   └── modules.ts                # L1 internal-architecture data: MODULES + getModule()
└── components/
    ├── overview/ ArchOverview · ArchLayer · ArchNode   # render ARCH_LAYERS
    └── module/
        ├── ModuleDetail.vue      # header + notes toggle + REGISTRY dispatch by flow.id
        ├── diagram/              # reusable diagram primitives
        │   ArchBox · Connector · PeekPanel · ChunkTreemap · LinkingDemo
        │   · HealLoopDemo · InsightNotes
        └── modules/              # per-module diagram composition: <Xxx>Detail.vue
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
   - **Icons/classes that live only in `model/*.ts`** (e.g. `node.icon`,
     `insights[].icon`) require UnoCSS to scan `.ts`. `uno.config.ts` adds `.ts`/
     `.js` to `content.pipeline.include` for this; without it those icons render
     as empty boxes. Use only `i-lucide-*` names (the one collection wired up).
   - presetIcons under pnpm can't auto-discover `@iconify-json/*`; `uno.config.ts`
     registers the `lucide` collection explicitly via `collections`. If icons
     vanish app-wide, that loader is the first place to check.

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

## Recipe B — Add a Module's Internal Architecture Diagram (the template)

Copy how `onboarding` is built. The L1 is ONE top-to-bottom architecture diagram:
boxes are modules with clear boundaries; connectors are the dataflow; long lists
are peek-on-demand details inside the boxes.

```
- [ ] 1. Research the real backend flow (handlers + scenarios + tools + models)
- [ ] 2. Add module identity   → model/flows.ts  (FlowDef in flows[])
- [ ] 3. Add architecture data → model/modules.ts (entry in MODULES)
- [ ] 4. Compose the diagram    → components/module/modules/<Xxx>Detail.vue
- [ ] 5. Register               → REGISTRY in ModuleDetail.vue
- [ ] 6. Make the node drillable → set node.flow in architecture.ts
- [ ] 7. Verify (typecheck + dev server)
```

**Step 2 — identity (`flows.ts`).** Add a `FlowDef { id, label, title, subtitle,
icon, accent }`. Used by the overview drill target + the `ModuleDetail` header only.

**Step 3 — architecture data (`modules.ts`).** Add a `MODULES[id]` entry whose `id`
equals the `FlowDef.id`. Model the diagram as boxes + their contents. Onboarding's
shape: `{ input, coordinator, worker{ prompt{blocks,rules}, tools, loop, output{types} },
storage{items} }`. Keep long lists here: prompt **rules**, produced data **types**,
**storage tables**. Different modules can have different shapes — model what the
real architecture is, not a fixed template.

**Step 4 — compose the diagram.** Create `modules/<Xxx>Detail.vue` taking
`defineProps<{ flow: FlowDef; showNotes?: boolean }>()`, look up `getModule(flow.id)`,
and lay out the three-zone grid (see "Three-Zone Layout" below) with the primitives:

```
<ArchBox icon title role accent [badge] [muted]> ...box body... </ArchBox>
<Connector [label="dispatch × N"] />          <!-- arrow + flowing dot between boxes -->
<ArchBox ...>
  ...sub-boxes (plain divs with ACCENTS.<c>.surface) for Prompt / Tools...
  <PeekPanel label count accent> ...collapsed list (rules / types)... </PeekPanel>
</ArchBox>
```

Rules for module boundaries & details:
- One box = one module (Input · Coordinator · Worker · Storage…). Nest sub-boxes
  (Prompt, Tools) as plain divs inside a box.
- Put dense lists (prompt rules, output type definitions) inside a `PeekPanel`
  (collapsed by default, click to expand inline — never a new zoom level).
- Annotate variations on the diagram (e.g. small-DB shortcut) as a dashed inline
  note, not a separate branch/section.

**Step 5 — register.** In `ModuleDetail.vue` add `{ <id>: <Xxx>Detail }` to
`REGISTRY`. Modules without an entry show a "建设中" placeholder.

## Diagram Primitives (reuse, don't reinvent)

- `ArchBox` — a module box: accent top-bar, icon + title + role pill + optional
  badge (`× N`), `muted` for dashed boundary/input boxes; body via default slot.
- `Connector` — vertical line + down chevron + optional `label`, with a subtle
  downward "data flow" dot. Place between boxes.
- `PeekPanel` — click-to-open **popover** (naive-ui `NPopover`, teleports to
  body so it floats above the diagram with **no layout reflow**); header = label +
  count, slot = the dense list (rules / type defs). Use for any dense detail.
  (Inline expansion was dropped because it squeezed neighboring columns.)
- `ChunkTreemap` — self-contained accelerated demo (squarified treemap with a
  looping pending→running→done fill) of the Coordinator's chunks being processed.
- `LinkingDemo` — self-contained accelerated demo of inference grounding: a recall
  funnel (all→HNSW→LLM) stacked over a **concurrency Gantt** (Agent ∥ Retrieval,
  schema-slot handoff, overlap savings). Use to make timing/parallelism legible.
- `HealLoopDemo` — self-contained accelerated demo of self-maintenance: per-row
  Rich Context lifecycle (fresh → `is_expired` → healing → healed / deleted) with
  the three soft-flag chips (`is_expired` / `is_stale` / `is_deleted`) flipping as
  the Coordinator marks and the Executor heals, then a re-embed sweep clears stale.
- `InsightNotes` — a stack of insight cards (+ optional `intro` line) for the
  hideable left notes column; takes `accent` + `items: Insight[]`.

## Layout Design SOP (the distilled principles)

The look every L1 module must converge on — apply these before writing markup:

1. **One internal architecture diagram, not a section stack.** A module is ONE
   top→bottom diagram. Prompt rules / output shapes / storage tables are
   *annotations inside boxes* (peek/dashed notes), never their own section-graphs.
2. **Three zones, one stage per row:** `notes ｜ spine ｜ demo` (see below). Every
   stage is a grid row; columns stay aligned across rows.
3. **Spine = boxes + connectors.** `ArchBox` per module (clear boundary), nested
   sub-boxes (Prompt/Tools) as plain `ACCENTS.<c>.surface` divs, `Connector`
   (with `label` = the data handed downstream) between them. Give the spine the
   largest column fraction.
4. **Merge tightly-coupled agents into one stage** (Coordinator+Worker /
   Coordinator+Executor) so a single tall right-column demo aligns beside the pair.
5. **Dense detail is peek-on-demand.** Long lists (rules, modes, signals, RC
   types) go in a `PeekPanel` popover — never inline-expanded (it reflows columns).
6. **Right column is a looping "accelerated demo"** of the stage's hardest idea
   (`ChunkTreemap`, `LinkingDemo`, `HealLoopDemo`…), aligned to the stage it
   explains; empty `<div class="hidden lg:block"/>` elsewhere to keep rows aligned.
7. **Engineering highlights as inline dashed notes** (concurrency timing,
   small-DB shortcut, verify gate). The "why" lives in the left notes column.
8. **Accent per stage, semantically:** violet=agent/reasoning, blue=retrieval/tools,
   amber=prompt/expire, emerald=output/heal/execute, indigo=storage, slate=input.
9. **Vertically center rows** (`items-start lg:items-center`) — no manual `pt-*`.
10. **All copy lives in `modules.ts`.** Components stay dumb. Ground every value in
    real backend code.

## Three-Zone Layout: notes ｜ spine ｜ demo

A bare vertical flow reads as "linear" and hides the engineering. Lay each module
out as up to three aligned columns, one stage per grid row (`items-start`):

- **Left — presenter notes (hideable).** The "why we designed it this way"
  insight cards, like PPT speaker notes. Off by default; `ModuleDetail.vue` owns a
  `showNotes` ref + "讲解备注" toggle and passes `:show-notes` down (it also widens
  its container `max-w-5xl → max-w-7xl` when on). Render via `InsightNotes`.
- **Center — the architecture spine.** Boxes (`ArchBox`) + `Connector`s. This is
  the focal column; give it the largest fraction.
- **Right — live demo lane.** A looping demo where it helps (`ChunkTreemap`,
  `LinkingDemo`), aligned to the stage it explains; empty cells elsewhere.

Implementation: bind the grid template to `showNotes` —
`showNotes ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
: 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]'`. Per stage emit cells in column
order: a left notes cell **`v-if="showNotes"`**, the spine cell, then the demo
cell. **Always render every center+right cell** (use an empty `<div class="hidden
lg:block" />` when a stage has no demo) so rows stay column-aligned. Use
`items-start lg:items-center` so notes/demo sit **vertically centered** against
the (taller) spine box per row — no manual `pt-*` offsets. Merge tightly-coupled
boxes (Coordinator+Worker) into one stage so a tall demo aligns beside them. Keep
all copy in `modules.ts` (`insights.{input,<stage>,…}`).

The `<Xxx>Detail.vue` takes `defineProps<{ flow: FlowDef; showNotes?: boolean }>()`.

If a module needs a genuinely new shape, extend the data in `modules.ts` and
compose existing primitives; only add a new primitive under `diagram/` when a
layout truly can't be expressed with the existing ones.

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

Then open `/arch`, drill into the node, and check: the diagram renders with
clear boxes + connectors, PeekPanels expand inline, accents/icons resolve (no
missing UnoCSS classes), and Esc/breadcrumb return to the panorama.
