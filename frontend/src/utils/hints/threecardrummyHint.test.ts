import { describe, expect, it } from 'vitest';
import type { ThreeCardRummyResponse } from '../../types/card';
import { ThreeCardRummyPhase } from '../../types/phases';
import { getThreeCardRummyHint } from './threecardrummyHint';

const state = (playerScore: number, phase: number = ThreeCardRummyPhase.ACTION): ThreeCardRummyResponse =>
  ({ playerScore, phase }) as ThreeCardRummyResponse;

describe('getThreeCardRummyHint', () => {
  it('advises nothing outside the action phase', () => {
    expect(getThreeCardRummyHint(state(0, ThreeCardRummyPhase.BET))).toBeNull();
    expect(getThreeCardRummyHint(state(0, ThreeCardRummyPhase.END))).toBeNull();
  });

  it('plays a meld with strong confidence', () => {
    // 0 点は役 = 最強。**低いほど強い**ので、ここが最も自信のある play。
    expect(getThreeCardRummyHint(state(0))).toEqual({
      targetAction: 'play',
      reason: 'hintReason.strongHand',
      confidence: 'strong',
    });
  });

  it('draws the strong line at ten, not below it', () => {
    // 境界の両側。10 点は「強い」、11 点はクオリファイ上限より下なだけの play。
    expect(getThreeCardRummyHint(state(10))?.confidence).toBe('strong');
    expect(getThreeCardRummyHint(state(11))).toEqual({
      targetAction: 'play',
      reason: 'hintReason.lowEnoughToPlay',
      confidence: 'moderate',
    });
  });

  it('folds at the dealer qualifying limit but plays one point under it', () => {
    // 19 点は play、20 点 (ディーラーがクオリファイする上限そのもの) は fold。
    expect(getThreeCardRummyHint(state(19))?.targetAction).toBe('play');
    expect(getThreeCardRummyHint(state(20))).toEqual({
      targetAction: 'fold',
      reason: 'hintReason.weakHand',
      confidence: 'moderate',
    });
    expect(getThreeCardRummyHint(state(30))?.targetAction).toBe('fold');
  });
});
