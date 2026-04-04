import { describe, expect, it } from 'vitest';
import {
  btnDanger,
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

  it('each button variant has a distinct background token', () => {
    const variants = [btnPrimary, btnWarning, btnSuccess, btnDanger, btnSecondary];
    const bgs = variants.map((v) => {
      const match = v.match(/bg-([^\s]+)/);
      return match?.[1];
    });
    const unique = new Set(bgs);
    expect(unique.size).toBe(variants.length);
  });
});
