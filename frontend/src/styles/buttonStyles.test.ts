// @ts-expect-error Node typings are intentionally excluded from frontend app code.
import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import {
  btnDanger,
  btnOutline,
  btnPokerAccent,
  btnPokerAllIn,
  btnPokerMuted,
  btnPokerPrimary,
  btnPrimary,
  btnSecondary,
  btnSuccess,
  btnWarning,
  focusRingAccent,
  focusRingWhite,
} from './buttonStyles';

describe('buttonStyles', () => {
  it('btnPrimary uses design system accent token', () => {
    expect(btnPrimary).toContain('ds-accent');
    expect(btnPrimary).toContain('ds-text-on-accent');
  });

  it('btnWarning uses design system warning token', () => {
    expect(btnWarning).toContain('ds-warning');
  });

  it('btnSuccess uses design system success token', () => {
    expect(btnSuccess).toContain('ds-success');
  });

  it('btnDanger uses design system error token', () => {
    expect(btnDanger).toContain('ds-error');
  });

  it('btnSecondary uses design system surface token', () => {
    expect(btnSecondary).toContain('ds-surface-elevated');
    expect(btnSecondary).toContain('ds-text-primary');
  });

  it('base includes disabled:saturate-50', () => {
    expect(btnPrimary).toContain('disabled:saturate-50');
  });

  it('base includes disabled:opacity-70', () => {
    expect(btnPrimary).toContain('disabled:opacity-70');
  });

  it('base includes accent focus ring', () => {
    expect(btnPrimary).toContain('focus-visible:ring-ds-accent');
    expect(btnPrimary).toContain('focus-visible:ring-offset-2');
  });

  it('focusRingWhite includes focus-visible:ring-white/80', () => {
    expect(focusRingWhite).toContain('focus-visible:outline-none');
    expect(focusRingWhite).toContain('focus-visible:ring-2');
    expect(focusRingWhite).toContain('focus-visible:ring-white/80');
  });

  it('focusRingAccent includes accent ring with offset', () => {
    expect(focusRingAccent).toContain('focus-visible:outline-none');
    expect(focusRingAccent).toContain('focus-visible:ring-2');
    expect(focusRingAccent).toContain('focus-visible:ring-ds-accent');
    expect(focusRingAccent).toContain('focus-visible:ring-offset-2');
    expect(focusRingAccent).toContain('focus-visible:ring-offset-black');
  });

  it('btnOutline uses design system border and text tokens', () => {
    expect(btnOutline).toContain('ds-border-subtle');
    expect(btnOutline).toContain('ds-text-on-felt');
    expect(btnOutline).toContain('bg-transparent');
    expect(btnOutline).toContain('hover:bg-ds-surface-elevated');
  });

  it('each button variant has a distinct background token', () => {
    const variants = [btnPrimary, btnWarning, btnSuccess, btnDanger, btnSecondary];
    const bgs = variants.map((v) => {
      const match = v.match(/bg-([^\s]+)/);
      return match?.[1];
    });
    const unique = new Set(bgs);
    expect(unique.size).toBe(variants.length);
  });

  it('uses AA text contrast for colored action buttons from the CSS token values', () => {
    const css = readFileSync('src/index.css', 'utf8');
    const token = (name: string) => {
      if (name === 'white') return '#ffffff';
      const value = css.match(new RegExp(`--color-${name}:\\s*(#[0-9a-fA-F]{6})`))?.[1];
      if (!value) throw new Error(`Missing color token ${name}`);
      return value;
    };
    const luminance = (hex: string) => {
      const parts = hex.slice(1).match(/.{2}/g);
      if (parts?.length !== 3) throw new Error(`Invalid color ${hex}`);
      const channels = parts.map((part) => parseInt(part, 16) / 255);
      const linear = channels.map((channel) =>
        channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4,
      );
      const [red, green, blue] = linear;
      return 0.2126 * (red ?? 0) + 0.7152 * (green ?? 0) + 0.0722 * (blue ?? 0);
    };
    const contrast = (foreground: string, background: string) => {
      const [light, dark] = [luminance(foreground), luminance(background)].sort((a, b) => b - a);
      return (light + 0.05) / (dark + 0.05);
    };
    const cases = [
      [btnSuccess, 'ds-text-on-accent', 'ds-success'],
      [btnPokerPrimary, 'ds-text-on-accent', 'poker-call'],
      [btnPokerAccent, 'ds-text-on-accent', 'poker-raise'],
      [btnPokerAllIn, 'ds-text-on-accent', 'poker-allin'],
      [btnDanger, 'white', 'ds-error'],
      [btnPokerMuted, 'white', 'poker-fold'],
    ] as const;
    for (const [button, foreground, background] of cases) {
      expect(button).toContain(`text-${foreground}`);
      const hoverBackground = button.match(/hover:bg-([^\s]+)/)?.[1];
      expect(hoverBackground).toBeDefined();
      expect(contrast(token(foreground), token(background))).toBeGreaterThanOrEqual(4.5);
      expect(contrast(token(foreground), token(hoverBackground ?? ''))).toBeGreaterThanOrEqual(4.5);
    }
  });
});
