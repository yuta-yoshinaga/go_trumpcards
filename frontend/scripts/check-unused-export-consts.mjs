import { readFile } from 'node:fs/promises';
import { relative, resolve } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

const ROOT = resolve(process.cwd());

/** Find exported uppercase constants whose identifier appears only at its declaration. */
export function findUnusedExportConsts(files) {
  const declarations = [];
  for (const [file, source] of files) {
    for (const match of source.matchAll(/^export const ([A-Z][A-Z0-9_]*)\s*=/gm)) {
      declarations.push({ name: match[1], file });
    }
  }

  return declarations.filter(({ name }) => {
    const occurrences = new RegExp(`\\b${name}\\b`, 'g');
    let count = 0;
    for (const [, source] of files) {
      count += [...source.matchAll(occurrences)].length;
      if (count > 1) return false;
    }
    return count === 1;
  });
}

if (process.argv[1]?.endsWith('check-unused-export-consts.mjs')) {
  const paths = await Array.fromAsync(new Bun.Glob('{src,e2e}/**/*.{ts,tsx}').scan({ cwd: ROOT }));
  const files = await Promise.all(paths.map(async (path) => [path, await readFile(resolve(ROOT, path), 'utf8')]));
  const declarations = files.reduce(
    (total, [, source]) => total + [...source.matchAll(/^export const [A-Z][A-Z0-9_]*\s*=/gm)].length,
    0,
  );
  assertFloor('unused-export-consts', declarations, 300, 'exported constants scanned');
  const unused = findUnusedExportConsts(files);
  if (unused.length) {
    console.error(`Unused exported constants (${unused.length}):`);
    for (const { name, file } of unused) console.error(`  ${relative(ROOT, resolve(ROOT, file))}: ${name}`);
    process.exitCode = 1;
  } else {
    console.log(`unused-export-consts: OK (${declarations} exported constants scanned).`);
  }
}
