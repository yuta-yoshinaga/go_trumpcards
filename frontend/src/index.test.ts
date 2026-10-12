// @ts-expect-error Node typings are intentionally excluded from frontend app code.
import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

describe('felt muted text CSS', () => {
  const css = readFileSync('src/index.css', 'utf8');

  it('uses on-felt muted text on felt and restores it only on opaque surfaces', () => {
    expect(css).toMatch(/\[data-felt\]\s*\{[^}]*--color-ds-text-muted:\s*var\(--color-ds-text-on-felt\)/s);
    expect(css).toMatch(
      /\[data-felt\]\s*:is\(\.bg-ds-surface,\s*\.bg-ds-surface-elevated,\s*\.bg-ds-bg,\s*\[data-surface\]\)\s*\{[^}]*--color-ds-text-muted:\s*var\(--ds-text-muted-surface\)/s,
    );
    expect(css).not.toMatch(/\[data-felt\]\s*:is\([^)]*\.glass-panel/);
    expect(css).not.toMatch(/\[data-felt\]\s*:is\([^)]*\.glass-panel--dark/);
    expect(css.match(/--ds-text-muted-surface:\s*#8b9aaf/gi)).toHaveLength(1);
    expect(css).toMatch(/--color-ds-text-muted:\s*var\(--ds-text-muted-surface\)/);
  });
});
