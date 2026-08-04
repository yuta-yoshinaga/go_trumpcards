import { describe, expect, it } from 'vitest';
import { sweepCelebration } from './sweepCelebration';

const nobody = () => false;
const everybody = () => true;

describe('sweepCelebration', () => {
  it('stays quiet on the first update, seeding the baseline', () => {
    expect(sweepCelebration(null, [0, 0, 0, 0], nobody)).toEqual({ kind: 'none' });
  });

  it('stays quiet when nothing changed', () => {
    expect(sweepCelebration([1, 0], [1, 0], nobody)).toEqual({ kind: 'none' });
  });

  it('fires when any counter rises', () => {
    expect(sweepCelebration([0, 0], [0, 1], nobody)).toEqual({ kind: 'fire', own: false });
  });

  it('marks the viewer’s own sweep', () => {
    expect(sweepCelebration([0, 0], [1, 0], (i) => i === 0)).toEqual({ kind: 'fire', own: true });
  });

  it('does not mark someone else’s sweep as the viewer’s', () => {
    expect(sweepCelebration([0, 0], [0, 1], (i) => i === 0)).toEqual({ kind: 'fire', own: false });
  });

  it('clears on a drop, which is a new round rather than a sweep', () => {
    expect(sweepCelebration([2, 1], [0, 0], everybody)).toEqual({ kind: 'clear' });
  });

  it('prefers firing when one counter rises while another resets', () => {
    // A simultaneous deal boundary must not swallow a sweep that just landed.
    expect(sweepCelebration([0, 3], [1, 0], (i) => i === 0)).toEqual({ kind: 'fire', own: true });
  });

  it('clears when the player count changes rather than firing on noise', () => {
    expect(sweepCelebration([0, 0], [0, 0, 0], nobody)).toEqual({ kind: 'clear' });
  });
});
