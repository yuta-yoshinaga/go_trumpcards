#!/usr/bin/env bun
// Guard that presenters do not expose raw error codes for games using coded domain errors.

import { readdirSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

const IS_FIXTURE = Boolean(process.argv[2]);
const ROOT = IS_FIXTURE ? resolve(process.argv[2]) : resolve(new URL('../..', import.meta.url).pathname);
const DOMAIN_DIR = join(ROOT, 'internal', 'domain');
const PRESENTER_DIR = join(ROOT, 'internal', 'adapter', 'presenter');

const goFiles = (directory) =>
  readdirSync(directory).filter((file) => file.endsWith('.go') && !file.endsWith('_test.go'));

const gameName = (file) => file.replace(/(CuiPresenter|WebPresenter|Presenter)\.go$/, '');

const codeGames = new Set(
  goFiles(DOMAIN_DIR)
    .filter((file) => readFileSync(join(DOMAIN_DIR, file), 'utf8').includes('NewDomainErrorCode'))
    .map((file) => file.slice(0, -'.go'.length)),
);

const pairs = goFiles(PRESENTER_DIR)
  .map((file) => ({ file, game: gameName(file) }))
  .filter(({ game }) => codeGames.has(game));

// A presenter is exempt when it names one of the shared resolvers. That list is the
// guard's weak point: extracting the resolution into a new helper would remove both
// names from the callers and make this guard fire on the files it is meant to bless.
// Add any new shared resolver here in the same change that introduces it.
const RESOLVERS = ['ErrorMessageCode', 'cuiErrorBlock'];

const violations = pairs.filter(({ file }) => {
  const source = readFileSync(join(PRESENTER_DIR, file), 'utf8');
  return source.includes('lastErr.Error()') && !RESOLVERS.some((name) => source.includes(name));
});

if (IS_FIXTURE) {
  assertFloor('error-code-presenters', pairs.length, 1, 'game/presenter pairs inspected');
} else {
  assertFloor('error-code-presenters', pairs.length, 64, 'game/presenter pairs inspected');
}

if (violations.length > 0) {
  console.error(`error-code-presenters: ${violations.length} violation(s).`);
  for (const { file } of violations) console.error(`  ${file}`);
  process.exit(1);
}

console.log(`error-code-presenters: OK (${pairs.length} game/presenter pairs inspected).`);
