import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-action-log-detail.mjs');
if (!existsSync(GUARD)) throw new Error(`check-action-log-detail.mjs not found at ${GUARD}`);
const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function fixture(...files) {
  const root = mkdtempSync(join(tmpdir(), 'action-log-detail-'));
  dirs.push(root);
  const domain = join(root, 'internal', 'domain');
  mkdirSync(domain, { recursive: true });
  for (const [name, source] of files) writeFileSync(join(domain, name), source);
  return root;
}
function run(root) {
  return spawnSync('bun', [GUARD, root], { encoding: 'utf8' });
}

describe('check-action-log-detail', () => {
  it('counts free-text detailCode literals, but not valid or non-literal codes', () => {
    const r = run(
      fixture([
        'Details.go',
        'package domain\nfunc f(code string) {\n appendLog("step", "日本語", nil, nil)\n appendLog("step", "English", nil, nil)\n appendLog("step", "game.log.step", map[string]string{"x": "y"}, nil)\n appendLog("step", code, nil, nil)\n}\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('2 literal details');
  });

  it('counts fmt.Sprintf detailCode and ignores cards', () => {
    const r = run(
      fixture([
        'Formatted.go',
        'package domain\nfunc f(cards []*Card, detail string) {\n appendLog("step", fmt.Sprintf("value %d", 1), nil, cards)\n appendLog("step", detail, nil, []*Card{card})\n}\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('1 literal details');
  });

  it('counts literal details before appended and sliced card arguments', () => {
    const r = run(
      fixture([
        'CardArguments.go',
        'package domain\nfunc f(x []*Card, h *Hand) {\n appendLog("step", "appended detail", nil, append([]*Card(nil), x...))\n appendLog("step", "sliced detail", nil, h.communityCards[3:])\n}\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('2 literal details');
  });

  it('checks literal details passed through addLog wrappers', () => {
    const r = run(
      fixture([
        'Wrapper.go',
        'package domain\nfunc f() { addLog("step", "turn up", nil, nil) }\nfunc addLog(action, detail string, params map[string]string, cards []*Card) { appendLogCode(0, action, detail, params, cards) }\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('1 literal details');
  });

  it('fails when the fixture ceiling is exceeded', () => {
    const calls = Array.from({ length: 101 }, (_, i) => `func f${i}() { appendLog("step", "detail", nil, nil) }`).join(
      '\n',
    );
    const r = run(fixture(['TooMany.go', `package domain\n${calls}\n`]));
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('exceeds ceiling');
    expect(r.stderr).toContain('TooMany.go');
    expect(r.stderr).toContain('"detail"');
  });
});
