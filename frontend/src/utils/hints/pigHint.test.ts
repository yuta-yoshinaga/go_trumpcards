import { describe, expect, it } from 'vitest';
import type { PigResponse } from '../../types/card';
import { getPigHint } from './pigHint';

const state = (hint?: PigResponse['hint']): PigResponse => ({ hint }) as PigResponse;

describe('getPigHint', () => {
  it('returns null without a hint', () => {
    expect(getPigHint(state())).toBeNull();
  });

  // **合図が出た場面は選択の余地がない。** 遅れることだけが負け。
  it('is certain once a signal is out', () => {
    expect(getPigHint(state({ reason: 'pigSignal' }))).toEqual({
      targetAction: 'signal',
      reason: 'hint.pigSignal',
      confidence: 'strong',
    });
  });

  it.each(['pigDiscardOdd', 'pigNoSingleton'])('names a card for %s', (reason) => {
    expect(getPigHint(state({ cardIndex: 2, reason }))).toEqual({
      targetAction: 'card-2',
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  // cardIndex 0 は「札を指していない」ではない。
  it('treats card index zero as a card', () => {
    expect(getPigHint(state({ cardIndex: 0, reason: 'pigDiscardOdd' }))?.targetAction).toBe('card-0');
  });
});
