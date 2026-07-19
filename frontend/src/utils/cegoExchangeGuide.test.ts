import { describe, expect, it } from 'vitest';
import { cegoExchangeGuide } from './cegoExchangeGuide';

describe('cegoExchangeGuide', () => {
  it('is on step 1 with the full keep count remaining when nothing is selected', () => {
    const guide = cegoExchangeGuide(0, 1, 11);
    expect(guide.currentStep).toBe(1);
    expect(guide.totalSteps).toBe(2);
    expect(guide.remaining).toBe(1);
    expect(guide.ready).toBe(false);
    expect(guide.layDownCount).toBe(10);
  });

  it('advances to step 2 once the required keep card is selected', () => {
    const guide = cegoExchangeGuide(1, 1, 11);
    expect(guide.currentStep).toBe(2);
    expect(guide.remaining).toBe(0);
    expect(guide.ready).toBe(true);
    expect(guide.layDownCount).toBe(10);
  });

  it('clamps remaining at zero when more than the keep count is selected', () => {
    const guide = cegoExchangeGuide(3, 1, 11);
    expect(guide.remaining).toBe(0);
    expect(guide.ready).toBe(true);
    expect(guide.currentStep).toBe(2);
  });

  it('supports keep counts greater than one', () => {
    const guide = cegoExchangeGuide(1, 2, 11);
    expect(guide.remaining).toBe(1);
    expect(guide.ready).toBe(false);
    expect(guide.currentStep).toBe(1);
    expect(guide.layDownCount).toBe(9);
  });

  it('clamps layDownCount at zero when the hand is smaller than the keep count', () => {
    const guide = cegoExchangeGuide(0, 1, 0);
    expect(guide.layDownCount).toBe(0);
  });
});
