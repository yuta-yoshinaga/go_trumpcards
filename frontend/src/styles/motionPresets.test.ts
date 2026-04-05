import { describe, expect, it } from 'vitest';
import { dealSpring, EXPANSION_GAP_PX, flipSpring, hoverLift, selectLift, staggerTiming } from './motionPresets';

describe('motionPresets', () => {
  it('dealSpring has correct shape', () => {
    expect(dealSpring).toEqual({ type: 'spring', stiffness: 300, damping: 25 });
  });

  it('flipSpring has correct shape', () => {
    expect(flipSpring).toEqual({ type: 'spring', stiffness: 400, damping: 30 });
  });

  it('selectLift has correct values', () => {
    expect(selectLift).toEqual({ y: -8, scale: 1.02 });
  });

  it('hoverLift has correct values', () => {
    expect(hoverLift).toEqual({ y: -4, scale: 1.05 });
  });

  it('staggerTiming has correct value', () => {
    expect(staggerTiming).toEqual({ staggerChildren: 0.12 });
  });

  it('EXPANSION_GAP_PX is 12', () => {
    expect(EXPANSION_GAP_PX).toBe(12);
  });
});
