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

  it('maps state text utilities on felt without overriding semantic background tokens', () => {
    const feltRules = [...css.matchAll(/\[data-felt\][^{]*\{[^}]*\}/gs)].map(([rule]) => rule).join('\n');
    const surfaceRule = css.match(/\[data-felt\]\s*:is\(([^)]*)\)\s*\{([^}]*)\}/s);
    expect(surfaceRule?.[1]).toContain('.bg-ds-surface');
    expect(surfaceRule?.[1]).toContain('.bg-ds-surface-elevated');
    expect(surfaceRule?.[1]).toContain('.bg-ds-bg');
    expect(surfaceRule?.[1]).toContain('[data-surface]');
    expect(css).toMatch(/\[data-felt\] \.bg-white,[\s\S]*?\.bg-ds-error\s*\{[^}]*--ds-warning-text:/);
    for (const token of ['warning', 'info', 'accent', 'success', 'error-text']) {
      const textProperty = token === 'error-text' ? 'error-text' : `${token}-text`;
      expect(css).toMatch(
        new RegExp(`\\[data-felt\\]\\s*\\{[^}]*--ds-${textProperty}:\\s*var\\(--color-ds-${token}-on-felt\\)`, 's'),
      );
      expect(css).toMatch(
        new RegExp(
          `\\[data-felt\\]\\s*:is\\([^)]*\\.bg-ds-surface[^)]*\\[data-surface\\][^)]*\\)[^{]*\\{[^}]*--ds-${textProperty}:\\s*var\\(--color-ds-${token}\\)`,
        ),
      );
      expect(css).toMatch(
        new RegExp(`\\[data-felt\\]\\s+\\.text-ds-${token}\\s*\\{\\s*color:\\s*var\\(--ds-${textProperty}\\)`),
      );
      expect(css).toMatch(
        new RegExp(
          `\\[data-felt\\]\\s+\\.hover\\\\:text-ds-${token}:hover[^{}]*\\{[^}]*color:\\s*var\\(--ds-${textProperty}\\)`,
        ),
      );
      expect(feltRules).not.toMatch(new RegExp(`--color-ds-${token}:`));
    }
  });

  it('uses the same opaque-surface selectors for muted and state text', () => {
    const selectors = [...css.matchAll(/\[data-felt\]\s*:is\(([^)]*)\)\s*\{/gs)].map(([, list]) =>
      list?.replace(/\s+/g, ' ').trim(),
    );
    expect(selectors).toHaveLength(2);
    expect(selectors[0]).toBe(selectors[1]);
  });

  it('maps hover color variants on felt and restores their hover colors on opaque surfaces', () => {
    for (const [variant, onFelt, original] of [
      ['accent-hover', 'accent-text', 'accent-hover'],
      ['warning-hover', 'warning-text', 'warning-hover'],
    ]) {
      expect(css).toMatch(
        new RegExp(
          `\\[data-felt\\]\\s+\\.hover\\\\:text-ds-${variant}:hover\\s*\\{\\s*color:\\s*var\\(--ds-${onFelt}\\)`,
        ),
      );
      expect(css).toMatch(
        new RegExp(
          `\\[data-felt\\]\\s*:is\\([^)]*\\)\\s*\\.hover\\\\:text-ds-${variant}:hover\\s*\\{\\s*color:\\s*var\\(--color-ds-${original}\\)`,
        ),
      );
    }
  });

  it('keeps every non-dark game felt, bare or under a glass panel, at AA for primary and on-felt text', () => {
    const declarations = [...css.matchAll(/--color-game-bg-([\w-]+):\s*([^;]+);/g)].filter(
      ([, name]) => !name?.endsWith('-dark'),
    );
    const backgroundColors = declarations
      .filter(([, , value]) => /^#[0-9a-fA-F]{6}$/.test(value?.trim() ?? ''))
      .map(([, name, value]) => [name, value?.trim()] as const);
    expect(backgroundColors.length, 'every non-dark game background must be a 6-digit hex color').toBe(
      declarations.length,
    );
    expect(backgroundColors.length).toBeGreaterThan(0);

    const readTextColor = (token: string) => {
      const value = css.match(new RegExp(`${token}:\\s*(#[0-9a-fA-F]{6})\\s*;`))?.[1];
      expect(value, `${token} must be defined as a 6-digit hex color`).toBeDefined();
      return value as string;
    };
    const textColors = [
      readTextColor('--color-ds-text-primary'),
      readTextColor('--color-ds-text-on-felt'),
      ...['warning', 'info', 'accent', 'success', 'error-text'].map((token) =>
        readTextColor(`--color-ds-${token}-on-felt`),
      ),
    ];

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
        for (const text of textColors) {
          expect(contrast(text, surface), `${name} (${surface}) with ${text}`).toBeGreaterThanOrEqual(4.5);
        }
      }
    }
  });

  it('keeps on-felt accent text at AA over Pitch’s translucent accent panel', () => {
    const readHex = (token: string) => {
      const value = css.match(new RegExp(`${token}:\\s*(#[0-9a-fA-F]{6})\\s*;`))?.[1];
      expect(value, `${token} must be defined as a 6-digit hex color`).toBeDefined();
      return value as string;
    };
    const pitchFelt = readHex('--color-game-bg-blue');
    const accent = readHex('--color-ds-accent');
    // bg-ds-accent/20 composites 20% accent over 80% Pitch felt.
    const panelBackground = `#${(pitchFelt.slice(1).match(/.{2}/g) ?? [])
      .map((part, index) => {
        const accentChannel = Number.parseInt(accent.slice(1).match(/.{2}/g)?.[index] ?? '00', 16);
        return Math.round(accentChannel * 0.2 + Number.parseInt(part, 16) * 0.8)
          .toString(16)
          .padStart(2, '0');
      })
      .join('')}`;
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
    const [hi, lo] = [luminance(readHex('--color-ds-accent-on-felt')), luminance(panelBackground)].sort(
      (a, b) => b - a,
    );
    expect(((hi ?? 0) + 0.05) / ((lo ?? 0) + 0.05)).toBeGreaterThanOrEqual(4.5);
  });
});
