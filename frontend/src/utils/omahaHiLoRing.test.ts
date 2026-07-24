import { describe, expect, it } from 'vitest';
import { hiLoRingStyle } from './omahaHiLoRing';

describe('hiLoRingStyle', () => {
  it('returns a distinct dual ring for a card used by both Hi and Lo', () => {
    const { category, ring } = hiLoRingStyle(true, true);
    expect(category).toBe('both');
    // Dual ring conveys both attributes: outer green ring + inner blue offset.
    expect(ring).toContain('ring-ds-success');
    expect(ring).toContain('ring-offset-ds-info');
  });

  it('returns the green Hi ring for a Hi-only card (no blue)', () => {
    const { category, ring } = hiLoRingStyle(true, false);
    expect(category).toBe('hi');
    expect(ring).toContain('ring-ds-success');
    expect(ring).not.toContain('ring-ds-info');
    expect(ring).not.toContain('ring-offset-ds-info');
  });

  it('returns the blue Lo ring for a Lo-only card (no green)', () => {
    const { category, ring } = hiLoRingStyle(false, true);
    expect(category).toBe('lo');
    expect(ring).toContain('ring-ds-info');
    expect(ring).not.toContain('ring-ds-success');
  });

  it('returns no ring for a card used by neither half', () => {
    const { category, ring } = hiLoRingStyle(false, false);
    expect(category).toBe('none');
    expect(ring).toBe('');
  });
});
