import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Vitest serves test modules under a non-file URL, so resolve from cwd (the
// vitest root is `frontend/`) and assert the script is there — a wrong cwd would
// otherwise turn every case below into a spawn of a missing file rather than a
// visible failure.
const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-modal-focus-trap.mjs');
if (!existsSync(GUARD)) throw new Error(`check-modal-focus-trap.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/**
 * Build a src fixture. The guard floors its walk at 2 components declaring
 * aria-modal, so every fixture ships at least that many compliant ones —
 * otherwise a case would fail on the floor rather than on the rule under test.
 */
function fixture(extra = {}) {
  const dir = mkdtempSync(join(tmpdir(), 'modal-trap-'));
  dirs.push(dir);
  const files = {
    'Modal.tsx': `import { useFocusTrap } from '../hooks/useFocusTrap';
export function Modal() {
  useFocusTrap(ref, open, onClose);
  return <div role="dialog" aria-modal="true" />;
}
`,
    'ManualModal.tsx': `import { useFocusTrap } from '../hooks/useFocusTrap';
export function ManualModal() {
  useFocusTrap(ref, open, onClose);
  return <div role="dialog" aria-modal="true" />;
}
`,
    ...extra,
  };
  for (const [name, body] of Object.entries(files)) {
    const full = join(dir, name);
    mkdirSync(join(full, '..'), { recursive: true });
    writeFileSync(full, body);
  }
  return dir;
}

const run = (dir) => spawnSync('bun', [GUARD, dir], { encoding: 'utf8' });

describe('check-modal-focus-trap', () => {
  it('accepts components that use useFocusTrap', () => {
    const r = run(fixture());
    expect(r.stdout).toContain('modal-focus-trap: OK');
    expect(r.status).toBe(0);
  });

  it('accepts a component that delegates to <Modal> instead of declaring its own trap', () => {
    const r = run(
      fixture({
        'Suggest.tsx': `import { Modal } from './Modal';
export function Suggest() {
  return <Modal open onClose={x} aria-modal="true"><p>hi</p></Modal>;
}
`,
      }),
    );
    expect(r.status).toBe(0);
  });

  it('rejects a component declaring aria-modal with a hand-rolled trap', () => {
    const r = run(
      fixture({
        'Rogue.tsx': `export function Rogue() {
  useEffect(() => {
    dialog.addEventListener('keydown', handleTab);
  }, []);
  return <div role="dialog" aria-modal="true" />;
}
`,
      }),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('Rogue.tsx');
    expect(r.stderr).toContain('aria-modal');
  });

  // The bug that motivated stripping comments: ActionLogPanel's source explains
  // why it has NO aria-modal, and a raw substring search flagged it for saying so.
  it('does not flag a component that only mentions aria-modal in a comment', () => {
    const r = run(
      fixture({
        'Panel.tsx': `export function Panel() {
  // This panel is a landmark role="region", not a dialog — no aria-modal="true" here.
  /* Nor in a block comment: aria-modal="true". */
  return <section aria-labelledby="t" />;
}
`,
      }),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('2 components declare aria-modal');
  });

  it('fails on an empty walk rather than reporting a clean tree', () => {
    const empty = mkdtempSync(join(tmpdir(), 'modal-trap-empty-'));
    dirs.push(empty);
    const r = run(empty);
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('modal-focus-trap');
  });

  it('ignores test files', () => {
    const r = run(
      fixture({
        'Rogue.test.tsx': `it('x', () => { render(<div role="dialog" aria-modal="true" />); });
`,
      }),
    );
    expect(r.status).toBe(0);
  });
});
