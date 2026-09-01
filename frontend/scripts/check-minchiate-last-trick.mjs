/**
 * Guard: both surfaces agree on what the last trick is worth in Minchiate.
 *
 * The round-result panel prints "+3 to team N" from a page constant while the settlement
 * itself adds `MinchiateLastTrickBonus` in the domain (#6512). Two literals for one rule can
 * drift apart with nothing failing, and a drifted pair means the breakdown the player uses to
 * check the score disagrees with the score.
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
    label: 'frontend/src/pages/MinchiatePage.tsx',
    file: path.join(ROOT, 'frontend', 'src', 'pages', 'MinchiatePage.tsx'),
    pattern: /const MINCHIATE_LAST_TRICK_BONUS = (\d+);/,
  },
  {
    label: 'internal/domain/Minchiate.go',
    file: path.join(ROOT, 'internal', 'domain', 'Minchiate.go'),
    pattern: /const MinchiateLastTrickBonus = (\d+)/,
  },
];

const found = [];
for (const src of SOURCES) {
  const text = await readFile(src.file, 'utf8');
  const m = text.match(src.pattern);
  if (!m) {
    console.error(`minchiate-last-trick: could not find the bonus in ${src.label}.`);
    process.exit(1);
  }
  found.push({ label: src.label, value: Number(m[1]) });
}

// **0 件で成功と読まれない.** A rename on either side must fail, not silently pass.
assertFloor('minchiate-last-trick', found.length, 2, 'bonus literals read');

const [a, b] = found;
if (a.value !== b.value) {
  console.error(
    `minchiate-last-trick: the two surfaces disagree — ${a.label} says ${a.value}, ${b.label} says ${b.value}.`,
  );
  process.exit(1);
}

console.log(`minchiate-last-trick: OK (${found.length} bonus literals read, both ${a.value}).`);
