import { describe, expect, it } from 'vitest';
import type { SergeantMajorResponse } from '../../types/card';
import { getSergeantMajorHint } from './sergeantmajorHint';

const base = (hint?: SergeantMajorResponse['hint']): SergeantMajorResponse =>
  ({ hint }) as unknown as SergeantMajorResponse;

describe('getSergeantMajorHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getSergeantMajorHint(base())).toBeNull();
  });

  // **宣言の助言は札ではなくスートを指す。**
  it('names the suit when it recommends a trump call', () => {
    expect(getSergeantMajorHint(base({ reason: 'sergeantmajorSelectTrump', suit: 3, indices: [] }))).toEqual({
      targetAction: 'trump-3',
      reason: 'hint.sergeantmajorSelectTrump',
      confidence: 'moderate',
    });
  });

  // **捨て札の助言は複数の札を指す。**
  it('names every card when it recommends a discard', () => {
    expect(getSergeantMajorHint(base({ reason: 'sergeantmajorDiscard', suit: 0, indices: [0, 2, 5, 7] }))).toEqual({
      targetAction: 'discard-0-2-5-7',
      reason: 'hint.sergeantmajorDiscard',
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getSergeantMajorHint(base({ cardIndex: 0, reason: 'sergeantmajorPressOn', suit: 0, indices: [] }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.sergeantmajorPressOn',
      confidence: 'moderate',
    });
  });

  // **ノルマに届いていないときだけ強く勧める。**
  it('is strong only while short of the target', () => {
    expect(
      getSergeantMajorHint(base({ cardIndex: 4, reason: 'sergeantmajorWinTrick', suit: 0, indices: [] }))?.confidence,
    ).toBe('strong');
    expect(
      getSergeantMajorHint(base({ cardIndex: 4, reason: 'sergeantmajorPressOn', suit: 0, indices: [] }))?.confidence,
    ).toBe('moderate');
  });
});
