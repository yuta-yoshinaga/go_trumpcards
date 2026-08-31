/**
 * Guard: the "close to winning" threshold is the same on both surfaces.
 *
 * Tysiac highlights a player who is within striking distance of the target — the Web page
 * paints the progress bar amber, the CUI colours the line. **The threshold is a display
 * decision, so it lives in the display layer on each side** rather than in the domain; that
 * leaves two literals that can drift apart silently, and a drifted pair means the two
 * surfaces disagree about who is about to win (#6483).
 *
 * This guard reads both literals and fails when they differ.
 */
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const HERE = path.dirname(fileURLToPath(import.meta.url));
// The repo root by default; tests pass a fixture root as argv[2] so the guard can be
// exercised against a tree that deliberately disagrees.
const ROOT = process.argv[2] ? path.resolve(process.argv[2]) : path.resolve(HERE, '..', '..');

/** Each side of the pair: where the literal lives and how to read it. */
const SOURCES = [
  {
    label: 'frontend/src/pages/TysiacPage.tsx',
    file: path.join(ROOT, 'frontend', 'src', 'pages', 'TysiacPage.tsx'),
    pattern: /const NEAR_WIN_RATIO = ([0-9.]+);/,
  },
  {
    label: 'internal/adapter/presenter/TysiacCuiPresenter.go',
    file: path.join(ROOT, 'internal', 'adapter', 'presenter', 'TysiacCuiPresenter.go'),
    pattern: /const tysiacNearWinRatio = ([0-9.]+)/,
  },
];

const found = [];
for (const src of SOURCES) {
  const text = await readFile(src.file, 'utf8');
  const m = text.match(src.pattern);
  if (!m) {
    console.error(`near-win-threshold: could not find the threshold in ${src.label}.`);
    process.exit(1);
  }
  found.push({ label: src.label, value: Number(m[1]) });
}

// **0 件で成功と読まれない.** A rename on either side must fail, not silently pass.
assertFloor('near-win-threshold', found.length, 2, 'threshold literals read');

const [a, b] = found;
if (a.value !== b.value) {
  console.error(
    `near-win-threshold: the two surfaces disagree — ${a.label} says ${a.value}, ${b.label} says ${b.value}.`,
  );
  process.exit(1);
}

console.log(`near-win-threshold: OK (${found.length} threshold literals read, both ${a.value}).`);
