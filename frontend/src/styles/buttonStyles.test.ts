import { describe, expect, it } from 'vitest';
import {
  btnDanger,
  btnPrimary,
  btnSecondary,
  btnSuccess,
  btnWarning,
  focusRingBlue,
  focusRingWhite,
} from './buttonStyles';

describe('buttonStyles', () => {
  it('btnPrimary includes blue background', () => {
    expect(btnPrimary).toContain('bg-blue-600');
  });

  it('btnWarning includes yellow background', () => {
    expect(btnWarning).toContain('bg-yellow-400');
  });

  it('btnSuccess includes green background', () => {
    expect(btnSuccess).toContain('bg-green-600');
  });

  it('btnDanger includes red background', () => {
    expect(btnDanger).toContain('bg-red-600');
  });

  it('btnSecondary includes gray background', () => {
    expect(btnSecondary).toContain('bg-gray-600');
  });

  it('focusRingWhite includes focus-visible:ring-white/80', () => {
    expect(focusRingWhite).toContain('focus-visible:outline-none');
    expect(focusRingWhite).toContain('focus-visible:ring-2');
    expect(focusRingWhite).toContain('focus-visible:ring-white/80');
  });

  it('focusRingBlue includes focus-visible:ring-blue-400 with offset', () => {
    expect(focusRingBlue).toContain('focus-visible:outline-none');
    expect(focusRingBlue).toContain('focus-visible:ring-2');
    expect(focusRingBlue).toContain('focus-visible:ring-blue-400');
    expect(focusRingBlue).toContain('focus-visible:ring-offset-2');
    expect(focusRingBlue).toContain('focus-visible:ring-offset-black');
  });
});
