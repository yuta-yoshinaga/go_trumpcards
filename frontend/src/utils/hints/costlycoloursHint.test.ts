import { describe, expect, it } from 'vitest';
import { makeCostlyColoursState } from '../../test/stateFactories';
import { getCostlyColoursHint } from './costlycoloursHint';

describe('getCostlyColoursHint', () => {
  it('returns null when there is nothing to suggest', () => {
    expect(getCostlyColoursHint(makeCostlyColoursState())).toBeNull();
    expect(getCostlyColoursHint(makeCostlyColoursState({ hintReason: 'none' }))).toBeNull();
  });

  // **交換フェーズは札を指さない。** hintHandIdx が -1 でも助言は出す ──
  // ここを index で門番すると、mog の助言が丸ごと消える。
  it('still returns a hint when no card is named', () => {
    for (const reason of ['mog_accept', 'mog_refuse']) {
      const hint = getCostlyColoursHint(makeCostlyColoursState({ hintHandIdx: -1, hintReason: reason }));
      expect(hint).not.toBeNull();
      expect(hint?.reason).toBe(`hint.${reason}`);
      expect(hint?.targetAction).toBe('select');
    }
  });

  it('maps each play reason onto a hint.* key', () => {
    for (const reason of ['fifteen', 'twenty_five', 'thirty_one', 'safe', 'go']) {
      const hint = getCostlyColoursHint(makeCostlyColoursState({ hintHandIdx: 1, hintReason: reason }));
      expect(hint).not.toBeNull();
      expect(hint?.reason).toBe(`hint.${reason}`);
      expect(hint?.reason).not.toContain(':');
      expect(hint?.targetAction).toBe('play');
      expect(hint?.confidence).toBe('moderate');
    }
  });
});
