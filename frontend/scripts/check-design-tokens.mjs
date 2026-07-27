#!/usr/bin/env bun
// Guard against design-token regressions. Fails the build if source code under
// frontend/src uses forbidden raw Tailwind palette classes or opacity-suffixed
// `text-white/N`. See DESIGN.md and issue #1411.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = fileURLToPath(new URL('..', import.meta.url));
const SRC_DIR = join(ROOT, 'src');

const PALETTE =
  '(red|green|yellow|amber|blue|orange|purple|pink|emerald|sky|indigo|violet|fuchsia|rose|slate|zinc|gray|neutral|stone|cyan|teal|lime)';
const RULES = [
  {
    pattern: /\btext-white\/\d+\b/g,
    message: 'Use text-ds-text-primary or text-ds-text-muted instead of text-white/N (see DESIGN.md).',
  },
  {
    pattern: new RegExp(String.raw`\b(bg|text|border|ring|shadow|from|to|via|divide)-${PALETTE}-\d{2,3}(/\d+)?\b`, 'g'),
    message: 'Use design-system tokens (bg-ds-*, text-ds-*, border-ds-*, ring-ds-*) instead of raw Tailwind palette.',
  },
];

const INCLUDED_EXT = new Set(['.ts', '.tsx']);
const IGNORED = /\.(test|spec)\.(ts|tsx)$/;

async function walk(dir) {
  const entries = await readdir(dir, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await walk(full)));
    } else if (INCLUDED_EXT.has(entry.name.slice(entry.name.lastIndexOf('.')))) {
      if (!IGNORED.test(entry.name)) files.push(full);
    }
  }
  return files;
}

const files = await walk(SRC_DIR);
const violations = [];

for (const file of files) {
  const text = await readFile(file, 'utf8');
  const lines = text.split('\n');
  for (const { pattern, message } of RULES) {
    for (let i = 0; i < lines.length; i += 1) {
      for (const match of lines[i].matchAll(pattern)) {
        violations.push({
          file: relative(ROOT, file),
          line: i + 1,
          match: match[0],
          message,
        });
      }
    }
  }
}

if (violations.length > 0) {
  console.error('\nDesign-token policy violations:\n');
  for (const v of violations) {
    console.error(`  ${v.file}:${v.line}  [${v.match}]  ${v.message}`);
  }
  console.error(`\n${violations.length} violation(s). See DESIGN.md and issue #1411.`);
  process.exit(1);
}

// Reduced-motion SSoT guard (issue #4315): prefers-reduced-motion must be
// enforced by a single universal block, not a class-name allowlist that
// silently misses arbitrary animate-[…] utilities and future animations.
const indexCss = await readFile(join(SRC_DIR, 'index.css'), 'utf8');
const rmIdx = indexCss.indexOf('@media (prefers-reduced-motion: reduce)');
// Read to the at-rule's own closing brace — the only `}` in column 0 after it,
// since the nested `*` rule closes indented. This used to slice a fixed 400
// characters, which silently failed the moment a comment was added inside the
// block: the declarations it looks for slid past the window and the guard
// reported the SSoT block missing when it was right there.
const rmEnd = rmIdx === -1 ? -1 : indexCss.indexOf('\n}', rmIdx);
const rmBlock = rmIdx === -1 || rmEnd === -1 ? '' : indexCss.slice(rmIdx, rmEnd + 2);
if (!rmBlock.includes('*,') || !rmBlock.includes('animation-duration: 0.01ms')) {
  console.error(
    '\nreduced-motion: index.css must enforce prefers-reduced-motion with a universal\n' +
      '(`*, *::before, *::after`) block that near-zeroes animation/transition durations,\n' +
      'not a per-class allowlist. See DESIGN.md Motion section and issue #4315.',
  );
  process.exit(1);
}

console.log(`design-tokens: OK (${files.length} source files scanned, test files skipped).`);
console.log('reduced-motion: OK (index.css uses the universal prefers-reduced-motion block).');

// Tap-target guard (issue #4368): DESIGN.md's "Interactive Element Minimum
// Size" rule requires every interactive control to hit a 44x44 CSS px target
// (WCAG 2.5.5 AAA).
//
// This guard covers checkbox and radio inputs ONLY, and that narrow scope is
// the point. index.css already floors `select`, `input[type=number|text|search]`
// and bare `<input>` at min-height:44px globally, so those controls cannot be
// undersized no matter what a page writes -- and that CSS says in as many words
// that per-component min-h-[44px] classes are therefore unnecessary. Checking
// them here would demand redundant classes and contradict it.
//
// Checkbox and radio are genuinely different: index.css only gives them an
// accent-color, and a 16-20px box must *stay* that size to still read as a
// checkbox. The 44px target has to come from the wrapping <label>, so that the
// whole row is tappable. A sibling <label htmlFor> does not count -- it leaves
// the small box as its own undersized target with no row to hit. Nothing but a
// source check can catch that: it is invisible to the type checker, to the
// global CSS, and to a DOM test that only asserts behaviour.
const TAP_TARGET = 'min-h-[44px]';

/**
 * Opening tag of the element starting at `<` index `start`.
 *
 * Scans for the `>` that actually closes the tag rather than the first one:
 * JSX attributes routinely embed `>` inside braces (`onChange={(e) => ...}`),
 * and stopping at that arrow truncates the tag before its className, which made
 * an earlier version of this guard report compliant controls as violations.
 */
function openingTag(text, start) {
  let depth = 0;
  for (let i = start; i < text.length; i += 1) {
    const ch = text[i];
    if (ch === '{') depth += 1;
    else if (ch === '}') depth -= 1;
    else if (ch === '>' && depth === 0) return text.slice(start, i + 1);
  }
  return text.slice(start);
}

/** Opening tag of the nearest <label> still open at `index`, or null. */
function enclosingLabelTag(text, index) {
  let depth = 0;
  let i = index;
  while (i > 0) {
    const close = text.lastIndexOf('</label>', i);
    const open = text.lastIndexOf('<label', i);
    if (open === -1) return null;
    if (close > open) {
      depth += 1;
      i = close - 1;
      continue;
    }
    if (depth === 0) return openingTag(text, open);
    depth -= 1;
    i = open - 1;
  }
  return null;
}

const LABEL_SIZED = new Set(['checkbox', 'radio']);

/** Index of the `}` closing the object literal that starts at `i`. */
function objectEnd(text, i) {
  let d = 0;
  for (let j = i; j < text.length; j += 1) {
    if (text[j] === '{') d += 1;
    else if (text[j] === '}') {
      d -= 1;
      if (d === 0) return j;
    }
  }
  return -1;
}

/**
 * Blank out comment bodies, preserving length and newlines so byte offsets and
 * reported line numbers still line up. Several doc comments mention `<select>`
 * in prose, and matching those reported violations in files with no JSX at all.
 */
function stripComments(text) {
  return text
    .replace(/\/\*[\s\S]*?\*\//g, (m) => m.replace(/[^\n]/g, ' '))
    .replace(/\/\/[^\n]*/g, (m) => ' '.repeat(m.length));
}

const tapViolations = [];
for (const file of files) {
  const raw = await readFile(file, 'utf8');
  const text = stripComments(raw);
  const lineOf = (idx) => text.slice(0, idx).split('\n').length;

  // Find every <input> tag first, then read its type out of the resolved tag.
  // Matching `/<input\b[^>]*type="checkbox"/` instead would stop at the first
  // `>` in the attribute list, so an input whose earlier attribute embeds one
  // (`min={a > b ? a : b}`, an inline handler) would not match at all and would
  // skip the guard silently -- a false negative in the very mechanism meant to
  // stop drift. Nothing in the tree does that today; the point is that it
  // cannot start.
  for (const m of text.matchAll(/<input\b/g)) {
    const tag = openingTag(text, m.index);
    const type = /type="(\w+)"/.exec(tag)?.[1];
    if (!type || !LABEL_SIZED.has(type)) continue;
    if (tag.includes(TAP_TARGET)) continue;
    const label = enclosingLabelTag(text, m.index);
    if (!label?.includes(TAP_TARGET)) {
      tapViolations.push({
        file: relative(ROOT, file),
        line: lineOf(m.index),
        what: `<input type="${type}">`,
        why: label
          ? `wrapping <label> must carry ${TAP_TARGET}`
          : `must be wrapped in a <label> carrying ${TAP_TARGET} (a sibling <label htmlFor> leaves the box itself undersized)`,
      });
    }
  }
}

if (tapViolations.length > 0) {
  console.error('\nTap-target policy violations (DESIGN.md Interactive Element Minimum Size):\n');
  for (const v of tapViolations) {
    console.error(`  ${v.file}:${v.line}  ${v.what}  ${v.why}`);
  }
  console.error(`\n${tapViolations.length} violation(s). See DESIGN.md and issue #4368.`);
  process.exit(1);
}

console.log('tap-targets: OK (every checkbox/radio sits in a 44px label row; index.css floors the rest).');

// Badge-contrast guard (issue #4367): DESIGN.md's Opacity rule forbids mixing
// Tailwind opacity suffixes with design tokens for the text or background of
// *meaningful* information (state badges, alerts, forced-action banners),
// because the effective colour then depends on whatever sits beneath.
//
// This flags one specific, measurable shape: a translucent semantic background
// paired with the *matching* semantic foreground in the same className, e.g.
// `bg-ds-error/30 … text-ds-error`. That combination has no fixed contrast
// ratio at all -- measured across the four table felts it ranges 1.13:1 to
// 4.59:1, and 29 of 30 combinations fall below WCAG AA. badgeStyles.ts exists
// precisely to replace it, and its helpers hold 5.8:1-12.1:1 regardless of
// background.
//
// Deliberately NOT flagged: a translucent semantic background whose foreground
// is inherited (issue #4367's "A-2", 42 sites). Those are only a problem when
// the inherited colour is itself low-contrast, which this cannot see, and
// DESIGN.md still permits opacity for decorative tints. Widening this rule
// needs a per-site judgement call, not a regex.
const BADGE_TOKEN = /\bbg-ds-(warning|info|success|error)\/\d+\b/g;

/**
 * Every `className` value in `text`, as {start, end} offsets covering the whole
 * value — the quoted string, or the brace-balanced expression.
 *
 * Matching per *line* instead would miss a violation split across a multi-line
 * ternary (background on one line, foreground on the next), which is exactly how
 * several of the sites fixed for #4367 were written. A guard that the common
 * formatting of the thing it guards can slip past is not much of a guard.
 */
function classNameValues(text) {
  const out = [];
  for (const m of text.matchAll(/className=/g)) {
    const i = m.index + m[0].length;
    if (text[i] === '"' || text[i] === "'") {
      const end = text.indexOf(text[i], i + 1);
      if (end !== -1) out.push({ start: i, end: end + 1 });
    } else if (text[i] === '{') {
      let depth = 0;
      for (let j = i; j < text.length; j += 1) {
        if (text[j] === '{') depth += 1;
        else if (text[j] === '}') {
          depth -= 1;
          if (depth === 0) {
            out.push({ start: i, end: j + 1 });
            break;
          }
        }
      }
    }
  }
  return out;
}

const badgeViolations = [];
for (const file of files) {
  if (file.endsWith('badgeStyles.ts')) continue; // the sanctioned home for these tokens
  const text = stripComments(await readFile(file, 'utf8'));
  for (const { start, end } of classNameValues(text)) {
    const value = text.slice(start, end);
    for (const m of value.matchAll(BADGE_TOKEN)) {
      const kind = m[1];
      // Only when the same className also sets the matching semantic foreground.
      if (!new RegExp(String.raw`\btext-ds-${kind}\b`).test(value)) continue;
      badgeViolations.push({
        file: relative(ROOT, file),
        line: text.slice(0, start + m.index).split('\n').length,
        kind,
        match: m[0],
      });
    }
  }
}

if (badgeViolations.length > 0) {
  console.error('\nBadge-contrast policy violations (DESIGN.md Opacity rule):\n');
  for (const v of badgeViolations) {
    console.error(
      `  ${v.file}:${v.line}  [${v.match} + text-ds-${v.kind}]  ` +
        `use badge${v.kind[0].toUpperCase()}${v.kind.slice(1)}Colors from styles/badgeStyles.ts`,
    );
  }
  console.error(`\n${badgeViolations.length} violation(s). See DESIGN.md and issue #4367.`);
  process.exit(1);
}

console.log('badge-contrast: OK (no semantic foreground sits on a translucent semantic background).');

// Keyboard-label guard (issue #4369): every `label` on an ActionBinding names a
// shared entry under `kbd.action.*` in common.json, and ActionShortcutsPanel
// renders `t('kbd.action.' + label)`. A typo would therefore print the raw i18n
// key on a game page — the exact defect #4374 existed to clean up.
//
// ActionBinding.label is typed `string` rather than a union of the known names,
// because a union would force an explicit `useMemo<ActionBinding[]>` annotation
// on all 56 pages: those array literals have no contextual type and would widen.
// This check buys the same safety without the ceremony.
const commonJa = JSON.parse(await readFile(join(SRC_DIR, 'i18n/locales/ja/common.json'), 'utf8'));
const commonEn = JSON.parse(await readFile(join(SRC_DIR, 'i18n/locales/en/common.json'), 'utf8'));
const knownLabels = new Set(Object.keys(commonJa.kbd?.action ?? {}));
const knownLabelsEn = new Set(Object.keys(commonEn.kbd?.action ?? {}));

const labelViolations = [];
for (const file of files) {
  const text = stripComments(await readFile(file, 'utf8'));
  // Only `label:` inside a binding object — one that also carries key: and
  // action:. Plenty of unrelated option lists use a `label` property too
  // (`{ value: 0, label: 'easy' }`), and those are not i18n keys.
  for (const m of text.matchAll(/\{\s*key:\s*'[^']+'\s*,\s*action:/g)) {
    const end = objectEnd(text, m.index);
    if (end === -1) continue;
    const body = text.slice(m.index, end + 1);
    const label = /\blabel:\s*'([^']*)'/.exec(body);
    if (!label) continue;
    const missing = [];
    if (!knownLabels.has(label[1])) missing.push('ja');
    if (!knownLabelsEn.has(label[1])) missing.push('en');
    if (missing.length > 0) {
      labelViolations.push({
        file: relative(ROOT, file),
        line: text.slice(0, m.index).split('\n').length,
        label: label[1],
        missing: missing.join(' + '),
      });
    }
  }
}

if (labelViolations.length > 0) {
  console.error('\nKeyboard-label violations (ActionBinding.label with no kbd.action.* entry):\n');
  for (const v of labelViolations) {
    console.error(`  ${v.file}:${v.line}  label: '${v.label}'  missing kbd.action.${v.label} in ${v.missing}`);
  }
  console.error(`\n${labelViolations.length} violation(s). See DESIGN.md and issue #4369.`);
  process.exit(1);
}

console.log(`kbd-labels: OK (every ActionBinding label resolves; ${knownLabels.size} shared labels).`);

// --- Mobile shell-height contract (issue #4373) -----------------------------
//
// Two classes keep every game page inside a 375x667 viewport, and both look like
// redundant cruft to a reader or an autofixer:
//
//   * `min-h-0` on the flex column in App.tsx — without it that column's
//     automatic minimum size is its content height, so it grows the document
//     instead of shrinking to the viewport. Removing it made 179 of 219 game
//     pages scroll vertically.
//   * the `max-h-[45vh]` cap on GameFooter — the footer is `shrink-0`, so without
//     a cap it takes whatever its controls want and the play area is squeezed
//     instead (measured: tallest footer 558px, 26 pages left under 80px).
//
// This has regressed five times (#965, #993, #1058, #1367, #4373), which is why
// it is a guard and not only a line in DESIGN.md.
const shellViolations = [];

const appText = await readFile(join(SRC_DIR, 'App.tsx'), 'utf8');
// Identify the column by the classes it has, in any order: a single regex
// requiring `flex-1` to textually precede `min-w-0` would still fail the build if
// a class sorter reordered them, but with a misleading "markup changed" message
// instead of naming the real problem.
const appColumn = [...appText.matchAll(/className="([^"]*)"/g)]
  .map((m) => m[1])
  .find((cls) => /\bflex-1\b/.test(cls) && /\bmin-w-0\b/.test(cls));
if (appColumn === undefined) {
  shellViolations.push('src/App.tsx: could not find the main flex column (flex-1 + min-w-0) — did its markup change?');
} else if (!/\bmin-h-0\b/.test(appColumn)) {
  shellViolations.push(
    `src/App.tsx: the main flex column lost min-h-0 (found "${appColumn}") — the document will grow past the viewport on mobile`,
  );
}

// The shell column must clip: its children each scroll themselves, so any overflow
// reaching it would otherwise propagate to the viewport and grow the document.
const shellText = await readFile(join(SRC_DIR, 'components/GamePageShell.tsx'), 'utf8');
const shellColumn = [...shellText.matchAll(/className=\{`([^`]*)`\}/g)]
  .map((m) => m[1])
  .find((cls) => /\bflex-1\b/.test(cls) && /\bflex-col\b/.test(cls) && /\brelative\b/.test(cls));
if (shellColumn === undefined) {
  shellViolations.push('src/components/GamePageShell.tsx: could not find the shell column — did its markup change?');
} else {
  for (const need of ['min-h-0', 'overflow-hidden']) {
    if (!shellColumn.split(/\s+/).includes(need)) {
      shellViolations.push(
        `src/components/GamePageShell.tsx: the shell column lost ${need} (found "${shellColumn}") — a descendant's overflow will grow the document on mobile`,
      );
    }
  }
}

const footerText = await readFile(join(SRC_DIR, 'components/GameFooter.tsx'), 'utf8');
for (const need of ['max-h-[45vh]', 'overflow-y-auto', 'sm:max-h-none', 'sm:overflow-y-visible']) {
  if (!footerText.includes(need)) {
    shellViolations.push(
      `src/components/GameFooter.tsx: lost "${need}" — the footer will crowd out the play area on mobile`,
    );
  }
}

if (shellViolations.length > 0) {
  console.error('\nMobile shell-height violations:\n');
  for (const v of shellViolations) console.error(`  ${v}`);
  console.error('\nSee the "Mobile vertical budget" section of DESIGN.md and issue #4373.');
  process.exit(1);
}

console.log(
  'shell-height: OK (App.tsx + GamePageShell columns keep min-h-0, the shell clips, GameFooter keeps its 45vh mobile cap).',
);

// --- Unmount-safe state writes in hooks (issue #4447) -----------------------
//
// A hook that awaits and then writes its own state can have that write land after
// the component is gone. React makes it a silent no-op, so nothing complains until
// the surrounding environment is torn down, where React reaches for `window` and
// throws from `dispatchSetState` — which failed whole CI runs while reporting every
// test as passed (#4444). The fix is `useIsMounted()` and a check after the await.
//
// Checked PER SITE, not per file. The first version of this guard skipped any file
// that mentioned `useIsMounted` anywhere, which let `useAcesUpGame` pass with an
// unguarded catch branch — a guard that grants a file blanket absolution is worse
// than none, because it reads as coverage it does not provide.
const unmountViolations = [];
const hookDir = join(SRC_DIR, 'hooks');
const hookFiles = (await readdir(hookDir)).filter(
  (f) => (f.endsWith('.ts') || f.endsWith('.tsx')) && !f.includes('.test.'),
);
const GUARD_RE = /\bisMounted\(\)|\bmountedRef\.current\b/;
for (const file of hookFiles) {
  const text = stripComments(await readFile(join(hookDir, file), 'utf8'));
  const lines = text.split('\n');
  // Setters this file owns; passing someone else's setter is not this file's problem.
  const owned = new Set([...text.matchAll(/const \[[^,\]]+,\s*(set[A-Z]\w*)\]\s*=\s*useState/g)].map((m) => m[1]));
  if (owned.size > 0) {
    for (let i = 0; i < lines.length; i += 1) {
      if (!/\bawait\b/.test(lines[i])) continue;
      // Every write in the window, not just the first: stopping at the first one is
      // how useAcesUpGame's unguarded `catch` hid behind the guarded `try` above it.
      for (let j = i + 1; j < Math.min(lines.length, i + 14); j += 1) {
        // Stop at the end of the enclosing callback — otherwise the window spills
        // into the next function and flags its synchronous writes, which is how an
        // early version of this guard "found" setHint(null) in useAcesUpGame's
        // handleUndo. A guard that cries wolf gets switched off.
        if (/^\s{0,4}\}(,\s*\[|\)|;|$)/.test(lines[j])) break;
        const m = /\b(set[A-Z]\w*)\(/.exec(lines[j]);
        if (!m) continue;
        // The guard must be in the same branch as the write. A check inside `try`
        // does nothing for a `catch`/`finally` body, which is exactly how
        // useAcesUpGame's catch slipped past the first version of this guard: so
        // only look from the nearest intervening `} catch {` / `} finally {`.
        let from = i;
        for (let k = i + 1; k <= j; k += 1) {
          if (/\}\s*(catch|finally)\b/.test(lines[k])) from = k;
        }
        const between = lines.slice(from, j + 1).join('\n');
        if (owned.has(m[1]) && !GUARD_RE.test(between)) {
          unmountViolations.push({
            file: `src/hooks/${file}`,
            line: j + 1,
            detail: `${m[1]}() runs after the await on line ${i + 1} with no isMounted guard between them`,
          });
        }
      }
    }
  }
  // runReplay sleeps between steps, so it routinely outlives the page. Its callers
  // must pass `isMounted` or it cannot stop — and the writes it drives are theirs.
  for (const m of text.matchAll(/runReplay\(/g)) {
    const call = text.slice(m.index, m.index + 400);
    const body = call.slice(0, call.indexOf('});') + 1);
    if (!/\bisMounted\b/.test(body)) {
      unmountViolations.push({
        file: `src/hooks/${file}`,
        line: text.slice(0, m.index).split('\n').length,
        detail: 'runReplay() without `isMounted` — the replay keeps driving state after the page is gone',
      });
    }
  }
}

if (unmountViolations.length > 0) {
  console.error('\nUnmount-safety violations (state written after an await, with no useIsMounted guard):\n');
  for (const v of unmountViolations) {
    console.error(`  ${v.file}:${v.line}  ${v.detail}`);
  }
  console.error('\nUse useIsMounted() and `if (!isMounted()) return;` after the await. See issue #4447.');
  process.exit(1);
}

console.log(`unmount-safety: OK (${hookFiles.length} hooks; every await -> own-state write is guarded).`);

// --- Vacuous "not called" assertions in tests (issues #4439, #4451) ---------
//
// `expect(<apiMock>).not.toHaveBeenCalled()` placed straight after an interaction
// passes whether or not the action fired: `exec` dispatches through react-query's
// `mutateAsync`, which invokes the mutation function on a microtask. The assertion
// runs first and sees nothing, so it asserts nothing. #4443 fixed 148 of these and
// found a real 3-page bug doing so; #4451 fixed 76 more that its detector could not
// see, because that one keyed on `fireEvent` and missed the `act(...)` form used in
// hook tests. The fix is `await flushPendingDispatch()` before the assertion.
//
// Checked per site, and the interaction list is deliberately broad — three widenings
// (fireEvent -> act -> raw .click()) each turned up sites the previous pass called
// clean, so a narrow matcher here reads as coverage it does not provide.
const vacuousViolations = [];
const INTERACTION = /fireEvent|userEvent|\bact\(|\.click\(\)|\.type\(/;
async function collectTests(dir) {
  const out = [];
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) out.push(...(await collectTests(full)));
    else if (/\.test\.tsx?$/.test(entry.name)) out.push(full);
  }
  return out;
}
for (const file of await collectTests(SRC_DIR)) {
  const text = await readFile(file, 'utf8');
  // Only mocks of a game API — a plain `vi.fn()` prop callback IS synchronous and
  // needs no flush, so flagging those would be noise that gets the guard switched off.
  const apiMocks = new Set([...text.matchAll(/const (\w+) = vi\.mocked\((\w+Api)\.\w+\)/g)].map((m) => m[1]));
  if (apiMocks.size === 0) continue;
  const lines = text.split('\n');
  for (let i = 0; i < lines.length; i += 1) {
    const m = /^\s*expect\((\w+)\)\.not\.toHaveBeenCalled/.exec(lines[i]);
    if (!m || !apiMocks.has(m[1])) continue;
    for (let j = i - 1; j >= Math.max(0, i - 8); j -= 1) {
      // Already flushed. Advancing fake timers counts: usePokerGame's debounce test
      // drives the pending timer with `vi.advanceTimersByTime(300)` rather than an
      // await, and its assertion is meaningful precisely because of that — flagging
      // it would be the guard crying wolf.
      if (/\bawait\b|advanceTimersByTime|runAllTimers|runOnlyPendingTimers/.test(lines[j])) break;
      if (INTERACTION.test(lines[j])) {
        vacuousViolations.push({ file: relative(ROOT, file), line: i + 1, mock: m[1] });
        break;
      }
    }
  }
}

if (vacuousViolations.length > 0) {
  console.error('\nVacuous "not called" assertions (no await between the interaction and the assertion):\n');
  for (const v of vacuousViolations) {
    console.error(`  ${v.file}:${v.line}  expect(${v.mock}).not.toHaveBeenCalled() cannot fail here`);
  }
  console.error('\nAwait flushPendingDispatch() first. See issues #4439 and #4451.');
  process.exit(1);
}

console.log('vacuous-assertions: OK (every "not called" assertion on an api mock is awaited first).');
