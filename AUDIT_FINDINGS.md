# what2watch — Audit Findings

> Audit-only pass. Nothing in this report has been changed in code. Every claim is backed by a `file:line` reference so you can jump to it.
> Date: 2026-06-09 · Branch: `main` · Auditor: Claude Code

---

## 1. Executive summary

This is a Go (Gin) + SQLite backend serving a fully client-rendered React/Vite SPA, doing "vibe" search over 463 movies/shows via OpenAI embeddings + an LLM reranker. The core pipeline is actually well-structured and the SQL layer is clean (parameterized queries throughout — no injection). But the app is **wide open and not production-safe**, and it has at least one bug that will crash the server under normal concurrent use.

The three scariest issues: **(1)** there is *no authentication of any kind* and `user_id` is a client-supplied random string, so anyone can read or delete anyone else's watch history (IDOR) — handlers.go:94/116. **(2)** Every expensive OpenAI-billed endpoint (`/recommend`, `/media`, `/media/:id/refresh`, `/admin/scrape`) is public and unthrottled, so a stranger with `curl` can run up your OpenAI bill at will (cost-DoS) — main.go:104-136. **(3)** The in-memory vector store has no locking, but ingest/refresh write to it while searches read from it, which is a concurrent map read/write in Go — a hard panic that takes down the whole process — embeddings.go:111-181.

The single highest-leverage fix is to **put the whole API behind authentication and derive `user_id` server-side from the session** instead of trusting the client. That one change closes the IDOR *and* removes the anonymous-abuse path to your OpenAI spend.

Secondary themes: SEO is effectively zero (crawlers get an empty `<div id="root">`), a 12MB compiled binary and a DB file are committed to git, `npm audit` shows 5 high / 3 moderate, `npm run lint` fails with 12 errors, and there's no `.env.example`. None of the security headers / TLS / reverse-proxy the README claims live in this repo (they're attributed to Caddy, which isn't here), so they're unverified.

---

## 2. Stack inventory

| Dimension | Finding |
|---|---|
| Backend framework | Go 1.25.6, Gin v1.11.0 (`go.mod:1-7`) |
| Frontend framework | React 19.2 + Vite 7.2 SPA (`frontend/package.json`) |
| Language | Backend: Go. Frontend: **plain JS/JSX, no TypeScript** (`@types/*` present but no `.ts`/`tsconfig.json`) |
| Styling | Tailwind CSS v4 via `@tailwindcss/vite` (`frontend/vite.config.js:3`) |
| Animation | framer-motion 12.29 (used in nearly every component) |
| Package mgr / Node | npm (package-lock.json); Dockerfile pins Node 20-alpine, Go 1.25-alpine (`Dockerfile`) |
| Data sources | OpenAI (embeddings `text-embedding-3-small` + chat `gpt-4o-mini`), TMDB (import tool), Reddit JSON API (scraper) |
| Database | SQLite via `mattn/go-sqlite3`, in-process; embeddings stored as JSON BLOBs (`database.go:320-333`) |
| Vector search | In-memory `map[string][]float32`, linear cosine scan (`embeddings.go:110-181`) |
| Rendering model | **CSR only.** `index.html` ships an empty root div; all content fetched client-side (`frontend/index.html:15`) |
| Hosting / deploy | Multi-stage Dockerfile + `entrypoint.sh` (seed then run). README claims Caddy reverse proxy on `chidaucf.win` — **not in repo** |
| Live URL | Likely `https://chidaucf.win` per `README.md:3` (unverified — not deployed/checked) |

### Build / lint / audit results

| Command | Result |
|---|---|
| `go build ./...` | ✅ Clean (exit 0) |
| `go vet ./...` | ✅ Clean (exit 0) — note: vet does **not** catch the data race below |
| `npm ci` | ✅ Clean, 180 packages |
| `npm run build` | ✅ Builds in ~0.8s. Output: `index.js` 356 KB (112 KB gzip, **single chunk, no code-splitting**), `index.css` 41 KB |
| `npm run lint` | ❌ **Fails — 12 errors** (see M-3) |
| `npm audit` | ❌ **8 vulns: 5 high, 3 moderate** (see H-2) |

DB state: root `vibe.db` has 463 media + 463 embeddings + 1 user. The committed `data/vibe.db` is empty (0 media).

---

## 3. Findings (by severity)

### 🔴 Critical

| # | What | Where | Why it matters | Effort |
|---|---|---|---|---|
| C-1 | **No authentication + IDOR.** `user_id` is a client-supplied free-text field; the server auto-creates users on first use. Any client can pass any `user_id` to read, mark, or delete another user's watch history. | `handlers.go:94` (GetSeen), `handlers.go:35` (PostSeen auto-creates, :48-59), `handlers.go:116` (DeleteSeen); identity generated as `'user_' + Math.random()` in `frontend/src/hooks/useLocalStorage.js:27` | Anyone can enumerate/overwrite/wipe any user's data. There is no session, token, or ownership check anywhere. | M |
| C-2 | **Public, unthrottled, OpenAI-billed endpoints.** `/recommend` embeds every query (OpenAI cost per call), `/media` + `/media/:id/refresh` run LLM generation, `/admin/scrape` runs LLM classify+extract over scraped Reddit. All anonymous, no rate limiting anywhere. | `main.go:104-136` (routes), `vibesearch.go:127` (embed per search), `handlers.go:316` (`/admin/scrape` has zero auth) | A stranger with `curl` can drive unbounded OpenAI spend and Reddit traffic. Classic cost-based DoS. | M |
| C-3 | **Data race → server crash.** `VectorStore` (`map[string][]float32`) has **no mutex**, but `Search` ranges the map while `Add`/`Remove` write to it. `/media` and `/media/:id/refresh` call `Add` concurrently with `/recommend` calls to `Search`. | `embeddings.go:111-181` (no locking), writes via `vibesearch.go:90` & `:365`, reads via `vibesearch.go:139` | Concurrent map read+write is a fatal `panic` in Go — it kills the whole process, not just one request. Guaranteed under real traffic. `go vet` won't flag it. | S |

### 🟠 High

| # | What | Where | Why it matters | Effort |
|---|---|---|---|---|
| H-1 | **12 MB compiled binary committed to git**, plus a tracked DB file. `tmdb-import` (12,331,826 bytes, a platform-specific Mach-O) and `data/vibe.db` are both tracked. | `git ls-files` → `tmdb-import`, `data/vibe.db` | Bloats every clone forever (it's in history), platform-specific, and `.gitignore`'s `*.db` / `data/vibe.db` rules are a no-op because the file was committed before they were added. | S |
| H-2 | **Dependency vulns: 5 high, 3 moderate.** vite (path traversal + arbitrary file read via dev WS), rollup (path traversal arbitrary write), picomatch (ReDoS), postcss (XSS in stringify). | `npm audit` output | All fixable via `npm audit fix`. The vite/rollup ones are mostly dev-server risks, but postcss XSS touches the build output. Leaving them rotting is how a build pipeline gets popped. | S |
| H-3 | **SEO is effectively zero — CSR empty shell.** Initial HTML is `<div id="root"></div>`; all movie data is injected after JS + API round-trips. No SSR/SSG, no JSON-LD `Movie`/`TVSeries`, no `sitemap.xml`, no `robots.txt`, no canonical, no per-item OG/Twitter tags. | `frontend/index.html:14-16`, no `robots.txt`/`sitemap.xml` in `frontend/public/` (only `favicon.svg`) | For a discovery site, organic search is the whole growth channel. Crawlers see an empty page; individual titles are unshareable and unindexable. | L |
| H-4 | **Server-side security posture unverified / absent in-repo.** Gin runs in **debug mode** (no `gin.ReleaseMode`), no CSP/security headers, no CORS configured despite `gin-contrib/cors` being a dependency. README attributes all of this to Caddy, which is **not in this repo**. | `main.go:118` (`gin.Default()`, no `SetMode`), no header middleware anywhere, `go.mod:4` (cors dep unused) | Debug mode leaks routes/stack info and is slower. If the Caddy layer isn't actually there in prod, there are *no* security headers at all. Can't confirm from this repo. | M |

### 🟡 Medium

| # | What | Where | Why it matters | Effort |
|---|---|---|---|---|
| M-1 | **Search input has no label.** The primary `<input>` has placeholder + `type="text"` but no `<label>`/`aria-label`. | `frontend/src/components/SearchBar.jsx:117-133` | Screen readers announce an unlabeled text field. The whole app is one search box — this is the one control that must be labeled. | S |
| M-2 | **Low-contrast text everywhere.** Heavy use of `text-muted/20`, `/30`, `text-[10px]` on a near-black background (footer, keyboard hints, query meta). | `App.jsx:282` (`text-muted/30`), `App.jsx:331-339` (`text-muted/20`), `SearchBar.jsx` hints | Almost certainly fails WCAG AA contrast. "Cyberpunk dim" aesthetic, but functionally unreadable for many users. | S |
| M-3 | **`npm run lint` fails (12 errors).** One legitimate: `react-hooks/set-state-in-effect` calling `setPlaceholder` synchronously in an effect. The rest are "`motion` is defined but never used" across 5 files where `motion.*` *is* used — a parser/config mismatch making lint unusable as a gate. | `SearchBar.jsx:13` (real), `QuickActions.jsx:1`, `RecommendationCard.jsx:2`, `WatchHistory.jsx:1`, `SearchBar.jsx:2` (false positives), config `frontend/eslint.config.js` | Lint can't be a CI gate while it's red, and the false positives suggest the flat-config/parser is misconfigured for JSX member usage. | M |
| M-4 | **No `.env.example`.** Two required keys (`OPENAI_API_KEY`, `TMDB_API_KEY`) exist only in the gitignored `.env`. | repo root (absent); keys read at `main.go:36`, `tmdb/tmdb.go` | A fresh clone can't run without reverse-engineering env vars from source. Reproducibility gap. | S |
| M-5 | **In-memory vector store won't scale.** Linear O(n) cosine scan over every embedding, all loaded into RAM at boot; no ANN index, no persistence of the index itself. | `embeddings.go:150-181`, loaded at `vibesearch.go:39-46` | Fine at 463 rows. At 50k+ it's slow and memory-heavy, and every restart re-reads all BLOBs. (Tied to C-3 — same struct also needs locking.) | M |

### 🟢 Low

| # | What | Where | Why it matters | Effort |
|---|---|---|---|---|
| L-1 | **Every route registered twice** (with and without `/api` prefix) for "backwards compatibility." | `main.go:91-136` | Doubles the surface to keep in sync; the legacy un-prefixed set is dead weight if the SPA only calls `/api`. | S |
| L-2 | **Dead code.** `EuclideanDistance` (`embeddings.go:203`), `NewClientWithModel` (`llm.go:36`), `SortByVibeScore`/`ByVibeScore` (`vibesearch.go:384-393`) have no callers. 7 `TODO/FIXME`-style markers across the Go/JSX source. | as listed | Noise; misleads readers about what's actually used. | S |
| L-3 | **`generateID` collision risk.** IDs are the title with non-alphanumerics stripped, so "Wall-E" and "Wall E" collapse to the same ID; namespace is title-only. | `vibesearch.go:397-411` | Two distinct titles can collide and overwrite; comment even flags it as not-for-production. | M |
| L-4 | **Single 356 KB JS bundle, no code-splitting**, framer-motion loaded eagerly app-wide. | `npm run build` output; `frontend/vite.config.js` | 112 KB gzip is acceptable but the particle background + motion could be lazy-loaded for a faster first paint. | M |
| L-5 | **Docs/infra drift.** README describes Caddy, security headers, and `chidaucf.win` that don't exist in the repo; `.gitignore` lists `data/vibe.db` though it's already tracked (rule is a no-op). | `README.md:11-13`, `.gitignore:21` | Documentation overstates what's in the codebase; misleads anyone setting up from scratch. | S |

---

## 4. Remediation roadmap

### Now — security + crash bugs (do before any public exposure)
- [ ] **C-3 — Add a `sync.RWMutex` to `VectorStore`** (RLock in `Search`, Lock in `Add`/`Remove`/`LoadFromMap`). This is the cheapest fix with the biggest stability payoff. **(S)**
- [ ] **C-1 — Real auth + server-derived `user_id`.** Introduce sessions/tokens; stop trusting the client's `user_id`; add ownership checks on `/seen`. **(M)**
- [ ] **C-2 — Gate + throttle the expensive endpoints.** Auth on `/admin/*`, `/media`, `/media/:id/refresh`; rate-limit `/recommend` and `/vibe`. **(M)**
- [ ] **H-4 — `gin.SetMode(gin.ReleaseMode)`** and confirm the Caddy/security-header layer actually exists in prod (or add headers in-app). **(S–M)**
- [ ] **H-1 — Purge the 12 MB binary + DB from git.** `git rm --cached tmdb-import data/vibe.db`; consider history scrub if clone size matters. **(S)**
- [ ] **H-2 — `npm audit fix`** and re-run. **(S)**

### Next — SEO, a11y, correctness wins
- [ ] **H-3 — Decide on a rendering strategy for crawlability.** Add `robots.txt` + `sitemap.xml` now (S); longer-term, SSR/prerender title pages or emit JSON-LD so content is in the initial HTML. **(S now / L full)**
- [ ] **M-1 — Label the search input** (`aria-label` or visually-hidden `<label>`). **(S)**
- [ ] **M-2 — Raise low-contrast text** to meet WCAG AA. **(S)**
- [ ] **M-3 — Fix lint:** resolve the real `set-state-in-effect` and repair the eslint/JSX config so `motion` usage is recognized; make lint green. **(M)**
- [ ] **M-4 — Add `.env.example`** with both keys documented. **(S)**

### Later — refactors, scale, polish
- [ ] **M-5 — Plan the vector-search path** (persisted index / ANN library) before the catalog grows. **(M)**
- [ ] **L-1 — Drop the duplicate un-prefixed routes** once the SPA is confirmed `/api`-only. **(S)**
- [ ] **L-2 — Delete dead code**, sweep the 7 TODO markers. **(S)**
- [ ] **L-3 — Replace `generateID`** with a hash/UUID to remove collisions. **(M)**
- [ ] **L-4 — Code-split / lazy-load** the particle background + heavy motion. **(M)**
- [ ] **L-5 — Reconcile README** with what's actually in the repo. **(S)**
- [ ] Tests: there are **none**. Worth a minimal suite around the search pipeline and the `/seen` ownership checks once auth lands.

---

## 5. Open questions

1. **Is the app actually deployed, and is Caddy really in front of it?** The README claims TLS + security headers + reverse proxy on `chidaucf.win`, but none of that is in this repo. If Caddy isn't present in prod, H-4 escalates to Critical. — *Need the live URL / deploy config to confirm.*
2. **Is `user_id` meant to be anonymous-per-device by design, or is real login planned?** That decides whether C-1's fix is "add sessions" or "scope data to an opaque device token + still authorize." — *Your call on the intended product model.*
3. **Is the Reddit scraper meant to run in prod** (`ENABLE_SCRAPER=false` by default)? If yes, it's another unauthenticated OpenAI-cost path via `/admin/scrape` and needs the same gating. — *Confirm intended behavior.*
4. **Which `vibe.db` is the source of truth?** Root `vibe.db` has 463 titles; the committed `data/vibe.db` is empty, and `entrypoint.sh` seeds on boot. — *Confirm the seed path produces the 463-title set in prod.*
5. **Is no-TypeScript a deliberate choice?** API responses are untyped on the frontend; a `tsconfig` with `strict` would catch a class of bugs, but that's a bigger lift. — *Your preference.*

---

*End of audit. No code was modified. Awaiting your go-ahead before touching anything.*
