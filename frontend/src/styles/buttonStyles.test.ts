import { describe, expect, it } from 'vitest';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning, focusRing } from './buttonStyles';

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

  it('focusRing includes focus:ring-2 and focus:outline-none', () => {
    expect(focusRing).toContain('focus:outline-none');
    expect(focusRing).toContain('focus:ring-2');
    expect(focusRing).toContain('focus:ring-white/70');
  });
});
