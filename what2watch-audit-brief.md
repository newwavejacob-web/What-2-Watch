# what2watch — Site Audit Brief

> **For:** Claude Code, run from the root of the what2watch repo.
> **Goal:** Produce a complete, honest picture of the current state of this site and a prioritized plan to fix it.
> **Mode:** AUDIT ONLY. Do **not** modify, refactor, or "fix" anything in this pass. Investigate, then write a findings report. Wait for my approval before touching code.

---

## 0. Before you start — fill these in

If any of these are unknown, detect them from the repo and note your assumptions:

- **Live URL:** _(fill in if deployed — e.g. https://what2watch...)_
- **Deploy target:** _(Vercel / Netlify / Cloudflare Pages / other — check config files)_
- **Things I already know are broken / annoy me:** _(optional — list anything here so you prioritize it)_

---

## 1. Stack inventory (do this first)

Detect and document, in a short table:

- Framework + version (Next.js, Vite/React, SvelteKit, plain HTML, etc.)
- Language (JS / TS) and, if TS, the `strict` setting in `tsconfig.json`
- Styling approach (Tailwind, CSS modules, styled-components, raw CSS)
- Package manager + Node version (`package.json` engines, lockfile type)
- External APIs / data sources (TMDB, OMDb, JustWatch, a custom backend, etc.)
- Hosting / deploy config files present
- Rendering model (SSR, SSG, CSR, ISR) — this matters a lot for SEO below

Then run the build and capture what happens:

```
# adapt to the detected package manager
<pm> install
<pm> run build
<pm> run lint      # if a lint script exists
<pm> audit         # dependency vulnerabilities
```

Record: does it install cleanly? Does it build? Any warnings? How many lint errors/warnings? How many `npm audit` criticals/highs?

---

## 2. Audit dimensions

Go through each. For every finding, record: **what**, **where** (`file:line`), **why it matters**, **severity** (Critical / High / Medium / Low), and **rough effort** (S / M / L).

### 2.1 Security  ⚠️ highest priority for this kind of app
- **Exposed secrets.** Is any API key (TMDB, etc.) shipped to the client bundle? Search for keys in client-side code, `NEXT_PUBLIC_*` vars, hardcoded strings, and the built output. A movie-data API key visible in the browser network tab = critical.
- Is `.env` (or any secrets file) committed to git? Check `.gitignore` and git history.
- Dependency vulnerabilities from `npm audit` — list criticals/highs with the affected package.
- Any user input that hits an API/DB without sanitization (search box, URL params)?
- Security headers / CSP configured at the host or in middleware?

### 2.2 SEO & crawlability
- Per-page `<title>` and `<meta name="description">` — present and unique, or copy-pasted/missing?
- Open Graph + Twitter card tags (matters for sharing a movie page).
- Structured data (JSON-LD) — is there `Movie` / `TVSeries` / `BreadcrumbList` schema? Flag pages that should have it but don't.
- `sitemap.xml` and `robots.txt` present and correct?
- Canonical URLs set? Any duplicate-content risks?
- **Crawlability vs rendering:** if the app is client-rendered (CSR), is the movie content actually in the initial HTML, or does it only appear after JS runs? Flag if crawlers would see an empty shell.
- Semantic HTML (`<main>`, `<nav>`, `<h1>` hierarchy) vs a soup of `<div>`s.

### 2.3 Performance
- Bundle size — total JS shipped, largest chunks, any obviously huge dependencies pulled in for a small feature.
- Images — are posters/backdrops optimized (sizing, format, lazy-loading)? Movie sites live or die on image weight. Flag full-res images rendered into thumbnails.
- API patterns — request waterfalls, duplicate fetches, missing caching, fetching more data than rendered.
- Render performance — unnecessary re-renders, missing memoization on heavy lists/grids, unkeyed lists.
- Font loading strategy (FOIT/FOUT, render-blocking).
- Anything that would tank Core Web Vitals (LCP image, layout shift from late-loading content).

### 2.4 Accessibility (a11y)
- `alt` text on poster/thumbnail images.
- Keyboard navigation — can you tab through the search, results, and any modals? Visible focus states?
- Color contrast on text over poster backgrounds (common failure on movie UIs).
- Form labels on the search input.
- ARIA roles/landmarks where appropriate; flag misuse too.

### 2.5 Responsiveness & UX
- Mobile layout — does the grid/list hold up at small widths? Touch target sizes?
- Loading states, empty states ("no results"), and error states for API calls — present or missing?
- Is there a real 404 page?
- Any broken links or dead routes you can detect statically.

### 2.6 Code quality & architecture
- Project structure — is it organized or a flat dump?
- Component reuse vs copy-paste duplication.
- State management approach — reasonable, or prop-drilling / tangled?
- TypeScript usage — `any` everywhere? strict off? missing types on API responses?
- Error handling around fetches — try/catch, fallbacks, or unhandled rejections?
- Dead code, commented-out blocks, `TODO`/`FIXME`/`HACK` markers (search and count them).
- Magic numbers/strings, hardcoded URLs that should be env vars.

### 2.7 Tooling & maintainability
- README — does it explain setup, env vars, and how to run? Or is it the default template?
- `.env.example` present so the project is reproducible?
- Linting/formatting config (ESLint, Prettier) — present and actually passing?
- Tests — any at all? (Expect none; just note it.)
- CI/CD — any GitHub Actions or deploy automation?

---

## 3. Deliverable

Write your findings to a new file: **`AUDIT_FINDINGS.md`** in the repo root, structured exactly like this:

1. **Executive summary** — 5–8 sentences: overall health, the 3 scariest issues, and the single highest-leverage fix.
2. **Stack inventory** — the table from section 1 plus build/lint/audit results.
3. **Findings** — grouped by severity (Critical → Low). Each row: description, `file:line`, why it matters, effort (S/M/L). Use a table.
4. **Remediation roadmap** — an ordered checklist of what to fix and in what order, grouped into:
   - **Now** (security + anything broken/blocking)
   - **Next** (SEO, performance, a11y wins)
   - **Later** (refactors, tooling, polish)
   Each item gets an effort estimate so I can decide what to batch.
5. **Open questions** — anything you couldn't determine without me (the live URL, intended behavior, design decisions).

Keep the report blunt and specific. I'd rather hear it's a mess than get reassured. Don't pad it. Use `file:line` references everywhere so I can jump straight to the problem.

**Reminder: do not fix anything yet.** Audit, report, stop.
