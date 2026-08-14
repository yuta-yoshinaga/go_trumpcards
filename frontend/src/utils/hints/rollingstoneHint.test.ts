import { describe, expect, it } from 'vitest';
import type { RollingStoneResponse } from '../../types/card';
import { getRollingStoneHint } from './rollingstoneHint';

const state = (hint?: RollingStoneResponse['hint']): RollingStoneResponse => ({ hint }) as RollingStoneResponse;

describe('getRollingStoneHint', () => {
  it('returns null without a hint', () => {
    expect(getRollingStoneHint(state())).toBeNull();
  });

  // **引き取るしかない場面は選択の余地がない。**
  it('is certain when picking up is the only move', () => {
    expect(getRollingStoneHint(state({ reason: 'rollingstonePickUp' }))).toEqual({
      targetAction: 'pickup',
      reason: 'hint.rollingstonePickUp',
      confidence: 'strong',
    });
  });

  it.each(['rollingstoneLead', 'rollingstoneFollow'])('names a card for %s', (reason) => {
    expect(getRollingStoneHint(state({ cardIndex: 2, reason }))).toEqual({
      targetAction: 'card-2',
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  // cardIndex 0 は「札を指していない」ではない。
  it('treats card index zero as a card', () => {
    expect(getRollingStoneHint(state({ cardIndex: 0, reason: 'rollingstoneLead' }))?.targetAction).toBe('card-0');
  });
});
