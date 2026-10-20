import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

// Read the route declarations as text so specs avoid loading React modules.
const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../../src/constants/gameRoutes.ts'),
  'utf8',
);

export const allGameRoutes: readonly string[] = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(
  ([, path]) => path,
);
