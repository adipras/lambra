# Frontend Analysis — JavaScript Application Graph

> REA static analysis (`analyze_javascript_application`) of `frontend/` source tree.
> Analyzed as inert syntax — no code was executed. Artifact (directory) digest:
> `fa4d9079b94c33b6a03de17122b06907fd8f6781a02708e1a8d5be7188dd4dd5`,
> Evidence `ev_16d763c671e1adab6359b112adc43812968016c76133f0ff8d562d7a099fab43`.

---

## 1. Parse Coverage

| Metric | Value |
|---|---|
| Relevant files | 56 |
| Parsed JavaScript files | 53 (0 parse failures) |
| Text bytes read | 564,739 |
| AST nodes visited | 37,414 |
| Findings | 890 |

100% of source files parsed cleanly — consistent with a small, unminified,
unbundled Vite development tree.

## 2. Semantic Graph Statistics (static relations)

| Family | Retained relations |
|---|---|
| data-flow | 9,927 |
| object-flow | 1,545 |
| closure | 1,163 |
| call-flow | 302 |
| boundary (JSON/parse/validate) | 21 |
| promise-ownership | 33 |
| timer | 8 |
| event / child-process / request / resource-lifecycle / configuration | 0 |

Interpretation:
- The dominant families (data-flow, object-flow, closure) match a typical React SPA:
  heavy prop/state plumbing and hook closures, few imperative side effects.
- Zero child-process / resource-lifecycle / timer families: no Node- or Electron-style
  code paths — the frontend is purely browser-side (analysis found no Electron
  applicability markers in any of the 53 sources).
- The zero `request`-family count reflects REA's extraction scope (fetch/`node:http`
  constructors): all HTTP in this app flows through the bundled axios instance
  (`src/api/axios.js`), which is data-flow, not constructor-flow.

## 3. Route Candidates (javascript-routes view)

Static candidates recovered (low confidence — inferred from navigation targets and
route-string constants; dynamic segments are not enumerable statically):

| Path | Mechanisms | Occurrences | Sources |
|---|---|---|---|
| `/` | navigation-target | 1 | `Sidebar.jsx` |
| `/login` | navigation-target | 1 | `Sidebar.jsx` (auth redirect/`Navigate` usage) |
| `/services` | navigation-target, route-string-constant | 6 | `Sidebar.jsx`, `ServiceNew.jsx`, +1 |
| `/settings` | navigation-target, route-string-constant | 2 | `Sidebar.jsx` |

Parameterized routes declared in `App.jsx` (`/services/:id`, `/endpoints/:id`,
`/services/new`) are constructed via JSX `<Route>` elements and dynamic strings, which
static path-string extraction does not promote to route candidates — the actual route
set is defined in `src/App.jsx` (7 routes total, `/login` outside the auth guard,
remaining 6 wrapped in `RequireAuth`).

## 4. Structure Notes (source-derived, corroborated by graph)

- Entry chain: `main.jsx` → `QueryClientProvider` → `BrowserRouter` → `App.jsx`.
- State: React Query for server state; zustand stores (`useAuthStore`, `useDiagramStore`)
  for session/UI state — the auth store is a new singleton holding token+user in
  localStorage with login/logout actions.
- API layer: single axios instance (`src/api/axios.js`) with request interceptor
  (Bearer token) and response interceptor (401 → session clear + redirect, exempting
  the login call).
- Diagram subsystem (`components/diagram/*`) is the largest UI cluster, consistent with
  the visual-relation editor being the platform's signature feature.

## 5. Security-Relevant Static Observations

- Token storage is `localStorage` (XSS-exfiltratable by design tradeoff); no token
  refresh mechanism (24 h expiry forces re-login).
- No `eval`/`new Function`/dynamic-injection patterns surfaced in the boundary family
  (21 JSON/parse relations, all parse/validate candidates).
- No outbound endpoints beyond the configured `VITE_API_BASE_URL` base (single axios
  instance).

## 6. Limitations

Static candidates do not prove runtime behavior; graph coverage is incomplete by design
(4,721 semantic unknowns retained in the parent Evidence). Route reachability,
permissions, and the complete application route set remain unverified by this analysis
alone; the route table above was cross-checked against `App.jsx` source.
