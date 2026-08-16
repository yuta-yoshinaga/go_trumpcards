import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// vitest の root は `frontend/` なので cwd から解決する。パスを誤ると全ケースが
// 「無いファイルを spawn した」になり、落ちてほしいケースが落ちて見えるだけになる。
const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-cli-hint-reachable.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** Builds a fixture tree and runs the guard against it. */
function run({ pages = {}, commands = {} }) {
  const root = mkdtempSync(join(tmpdir(), 'cli-hint-guard-'));
  dirs.push(root);
  mkdirSync(join(root, 'src/pages'), { recursive: true });
  mkdirSync(join(root, 'src/utils/cli/commands'), { recursive: true });
  for (const [name, body] of Object.entries(pages)) {
    writeFileSync(join(root, 'src/pages', name), body);
  }
  for (const [name, body] of Object.entries(commands)) {
    writeFileSync(join(root, 'src/utils/cli/commands', name), body);
  }
  return spawnSync('bun', [GUARD], {
    encoding: 'utf8',
    env: { ...process.env, CLI_HINT_GUARD_ROOT: root },
  });
}

/** A page with a hint and a CLI, importing the named command module. */
const page = (mod, extra = '') => `
  import { parseXCommand } from '../utils/cli/commands/${mod}Commands';
  const { hint, hintEnabled } = useGameHint('x', state);
  const cfg = { parseCommand: parseXCommand, helpText: [] };
  ${extra}
`;

describe('check-cli-hint-reachable', () => {
  it('fails when a page has a hint but no way to ask for one', () => {
    const r = run({
      pages: { 'XPage.tsx': page('x') },
      commands: { 'xCommands.ts': "case 'play': return { args: ['play'] };" },
    });
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('XPage.tsx');
  });

  it('passes when the command module answers hint', () => {
    const r = run({
      pages: { 'XPage.tsx': page('x') },
      commands: { 'xCommands.ts': "case 'hint': return { args: ['hint'] };" },
    });
    expect(r.status).toBe(0);
  });

  it('passes when the page answers locally with hintCliText', () => {
    const r = run({
      pages: { 'XPage.tsx': page('x', 'localCommand: (i) => hintCliText(hint),') },
      commands: { 'xCommands.ts': "case 'play': return { args: ['play'] };" },
    });
    expect(r.status).toBe(0);
  });

  it('passes when the page parses inline and handles hint there', () => {
    const r = run({
      pages: {
        'XPage.tsx': `
          function parseXCommand(input) { switch (input) { case 'hint': return 1; } }
          const { hint } = useGameHint('x', state);
          const cfg = { parseCommand: parseXCommand, helpText: [] };
        `,
      },
    });
    expect(r.status).toBe(0);
  });

  it('passes when the module delegates to a shared parser that answers hint', () => {
    const r = run({
      pages: { 'XPage.tsx': page('x') },
      commands: {
        'sharedTrickCommands.ts': "case 'hint': return { command: 'hint' };",
        'xCommands.ts': "export * from './sharedTrickCommands';",
      },
    });
    expect(r.status).toBe(0);
  });

  it('ignores a page with no hint to expose', () => {
    const r = run({
      pages: { 'XPage.tsx': 'const cfg = { parseCommand: p, helpText: [] };' },
      commands: { 'xCommands.ts': "case 'play': return 1;" },
    });
    expect(r.status).toBe(0);
  });

  it('ignores a page with a hint but no CLI at all', () => {
    const r = run({
      pages: { 'XPage.tsx': "const { hint } = useGameHint('x', state);" },
    });
    expect(r.status).toBe(0);
  });
});
