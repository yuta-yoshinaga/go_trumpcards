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

describe('felt color contrast', () => {
  const css = readFileSync('src/index.css', 'utf8');

  it('keeps every non-dark game felt, bare or under a glass panel, at AA for primary and on-felt text', () => {
    const backgroundColors = [...css.matchAll(/--color-game-bg-([\w-]+):\s*(#[0-9a-fA-F]{6})/g)]
      .filter(([, name]) => !name?.endsWith('-dark'))
      .map(([, name, hex]) => [name, hex] as const);
    expect(backgroundColors.length).toBeGreaterThan(0);

    const luminance = (hex: string) => {
      const channels = hex.slice(1).match(/.{2}/g);
      if (channels?.length !== 3) throw new Error(`Invalid color ${hex}`);
      const linear = channels.map((part) => {
        const channel = parseInt(part, 16) / 255;
        return channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
      });
      const [red, green, blue] = linear;
      return 0.2126 * (red ?? 0) + 0.7152 * (green ?? 0) + 0.0722 * (blue ?? 0);
    };
    // Glass panels (white at the --color-glass-bg alpha) sit on the felt and lighten it, so the felt
    // must keep AA for both primary and on-felt text underneath one as well (#11722).
    const glassAlpha = Number(css.match(/--color-glass-bg:\s*rgba\(255,\s*255,\s*255,\s*([\d.]+)\)/)?.[1]);
    expect(glassAlpha).toBeGreaterThan(0);
    const underGlass = (hex: string) =>
      `#${(hex.slice(1).match(/.{2}/g) ?? [])
        .map((part) => Math.round(255 * glassAlpha + Number.parseInt(part, 16) * (1 - glassAlpha)))
        .map((value) => value.toString(16).padStart(2, '0'))
        .join('')}`;
    const contrast = (a: string, b: string) => {
      const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
      return ((hi ?? 0) + 0.05) / ((lo ?? 0) + 0.05);
    };
    for (const [name, background] of backgroundColors) {
      for (const surface of [background ?? '', underGlass(background ?? '')]) {
        for (const text of ['#E8E0D4', '#DDE3EA']) {
          expect(contrast(text, surface), `${name} (${surface}) with ${text}`).toBeGreaterThanOrEqual(4.5);
        }
      }
    }
  });
});
