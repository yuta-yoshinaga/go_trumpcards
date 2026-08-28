import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Vitest serves test modules under a non-file URL, so resolve from cwd (the
// vitest root is `frontend/`) and assert the script is there — a wrong cwd would
// otherwise turn every case below into a spawn of a missing file rather than a
// visible failure.
const GUARD = join(process.cwd(), 'scripts', 'check-hint-live-region.mjs');
if (!existsSync(GUARD)) throw new Error(`check-hint-live-region.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** A page whose hint sits inside an always-mounted live region. */
const compliant = (name) => `export function ${name}() {
  return (
    <div data-testid="${name.toLowerCase()}-hint-live" role="status" aria-live="polite">
      {hint && (
        <div className="text-ds-warning">
          {t('hintAvailable')}: {t(\`hint.\${hint.reason}\`)}
        </div>
      )}
    </div>
  );
}
`;

/** A page whose hint is rendered under a key other than `hintAvailable`. */
const compliantOtherKey = (name) => `export function ${name}() {
  return (
    <div data-testid="${name.toLowerCase()}-hint-live" role="status" aria-live="polite">
      {hint && (
        <div className="text-ds-warning">
          {t('hintPlay')}: [{hint.cardIndex}] ({t(\`hintReason.\${hint.reason}\`)})
        </div>
      )}
    </div>
  );
}
`;

/**
 * Build a pages fixture.
 *
 * The guard floors its walk at 50 pages rendering `hintAvailable` and 15 pages
 * rendering a hint under another key, so every fixture ships enough of both —
 * otherwise a case would fail on a floor rather than on the rule under test.
 */
function fixture(extra = {}) {
  const dir = mkdtempSync(join(tmpdir(), 'hint-live-'));
  dirs.push(dir);
  for (let i = 0; i < 55; i += 1) {
    writeFileSync(join(dir, `Fine${i}Page.tsx`), compliant(`Fine${i}`));
  }
  for (let i = 0; i < 20; i += 1) {
    writeFileSync(join(dir, `Other${i}Page.tsx`), compliantOtherKey(`Other${i}`));
  }
  for (const [name, body] of Object.entries(extra)) writeFileSync(join(dir, name), body);
  return dir;
}

const run = (dir) => spawnSync('bun', [GUARD, dir], { encoding: 'utf8' });

describe('check-hint-live-region', () => {
  it('accepts hints wrapped in an always-mounted live region', () => {
    const r = run(fixture());
    expect(r.stdout).toContain('hint-live-region: OK');
    expect(r.status).toBe(0);
  });

  it('rejects a hint that is in no live region at all', () => {
    const r = run(
      fixture({
        'RoguePage.tsx': `export function Rogue() {
  return (
    <div data-testid="rogue-hint">
      {hint && <div>{t('hintAvailable')}: {hint.reason}</div>}
    </div>
  );
}
`,
      }),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('RoguePage.tsx');
    expect(r.stderr).toContain('no live region');
  });

  // The bug this guard exists for: the attributes are present, but on the
  // element that only exists while a hint is set, so nothing is announced.
  it('rejects a live region that is itself conditional', () => {
    const r = run(
      fixture({
        'LatePage.tsx': `export function Late() {
  return (
    <div>
      {hint && (
        <div role="status" aria-live="polite">
          {t('hintAvailable')}: {hint.reason}
        </div>
      )}
    </div>
  );
}
`,
      }),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('LatePage.tsx');
    expect(r.stderr).toContain('already filled');
  });

  it('rejects aria-live without a role', () => {
    const r = run(
      fixture({
        'RolelessPage.tsx': `export function Roleless() {
  return (
    <div aria-live="polite">
      {hint && <div>{t('hintAvailable')}: {hint.reason}</div>}
    </div>
  );
}
`,
      }),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('RolelessPage.tsx');
    expect(r.stderr).toContain('role="status"');
  });

  // The bug in #6663: the walk only entered on `t('hintAvailable')`, so a page
  // rendering its hint under any other key was never looked at — 20 of them had
  // no aria-live at all while the guard printed OK.
  it('rejects a hint rendered under another key with no live region', () => {
    const r = run(
      fixture({
        'OtherKeyRoguePage.tsx': `export function OtherKeyRogue() {
  return (
    <div className="text-ds-warning">
      {hint && (
        <div>{t('hintPlay')}: [{hint.cardIndex}]</div>
      )}
    </div>
  );
}
`,
      }),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('OtherKeyRoguePage.tsx');
    expect(r.stderr).toContain('no live region');
  });

  it('rejects an other-key hint whose live region is inside the gate', () => {
    const r = run(
      fixture({
        'OtherKeyLatePage.tsx': `export function OtherKeyLate() {
  return (
    <div>
      {hint && (
        <div role="status" aria-live="polite">
          {t('hintPlay')}: [{hint.cardIndex}]
        </div>
      )}
    </div>
  );
}
`,
      }),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('OtherKeyLatePage.tsx');
    expect(r.stderr).toContain('already filled');
  });

  // Negative control for the case above: the same page with the region hoisted
  // out of the gate must pass, or the rule would just be "no aria-live here".
  it('accepts an other-key hint once the region is hoisted out of the gate', () => {
    const r = run(fixture({ 'OtherKeyFixedPage.tsx': compliantOtherKey('OtherKeyFixed') }));
    expect(r.stdout).toContain('hint-live-region: OK');
    expect(r.status).toBe(0);
  });

  // The floor that would have caught #6663: a walk entering on `hintAvailable`
  // alone finds zero other-key pages.
  it('fails when no page rendering a hint under another key is found', () => {
    const dir = mkdtempSync(join(tmpdir(), 'hint-live-onekey-'));
    dirs.push(dir);
    for (let i = 0; i < 55; i += 1) writeFileSync(join(dir, `Fine${i}Page.tsx`), compliant(`Fine${i}`));
    const r = run(dir);
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('hint-live-region-other-keys');
  });

  // A walk that finds almost nothing reports no violations for the wrong reason.
  it('fails when the walk finds far fewer pages than the floor', () => {
    const dir = mkdtempSync(join(tmpdir(), 'hint-live-empty-'));
    dirs.push(dir);
    writeFileSync(join(dir, 'OnlyPage.tsx'), compliant('Only'));
    const r = run(dir);
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('hint-live-region');
  });
});
