/**
 * Guard: both surfaces agree on how many cards complete a Samba canasta.
 *
 * The Web page subtracts `SAMBA_CANASTA_SIZE` from a meld's length to show "3 more to go",
 * and the CUI now does the same with `domain.SambaCanastaSize` (#6499). Two literals for the
 * same rule can drift apart with nothing failing: the two surfaces would then disagree about
 * how close a meld is to a canasta, which is the whole point of the readout.
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
    label: 'frontend/src/pages/SambaPage.tsx',
    file: path.join(ROOT, 'frontend', 'src', 'pages', 'SambaPage.tsx'),
    pattern: /const SAMBA_CANASTA_SIZE = (\d+);/,
  },
  {
    label: 'internal/domain/SambaPlayer.go',
    file: path.join(ROOT, 'internal', 'domain', 'SambaPlayer.go'),
    pattern: /const SambaCanastaSize = (\d+)/,
  },
];

const found = [];
for (const src of SOURCES) {
  const text = await readFile(src.file, 'utf8');
  const m = text.match(src.pattern);
  if (!m) {
    console.error(`samba-canasta-size: could not find the size in ${src.label}.`);
    process.exit(1);
  }
  found.push({ label: src.label, value: Number(m[1]) });
}

// **0 件で成功と読まれない.** A rename on either side must fail, not silently pass.
assertFloor('samba-canasta-size', found.length, 2, 'size literals read');

const [a, b] = found;
if (a.value !== b.value) {
  console.error(
    `samba-canasta-size: the two surfaces disagree — ${a.label} says ${a.value}, ${b.label} says ${b.value}.`,
  );
  process.exit(1);
}

console.log(`samba-canasta-size: OK (${found.length} size literals read, both ${a.value}).`);
