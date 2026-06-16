import { describe, expect, it } from 'vitest';
import { makeCourtPieceState } from '../../test/stateFactories';
import { getCourtPieceHint } from './courtPieceHint';

describe('getCourtPieceHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getCourtPieceHint(makeCourtPieceState())).toBeNull();
    expect(getCourtPieceHint(makeCourtPieceState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeCourtPieceState({ hint: { reason: '' } });
    expect(getCourtPieceHint(state)).toBeNull();
  });

  it('maps a server play hint into a play HintResult', () => {
    const state = makeCourtPieceState({ hint: { cardIndex: 2, reason: 'trump_cut' } });
    expect(getCourtPieceHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.trump_cut',
      confidence: 'moderate',
    });
  });

  it('maps a trump-declaration hint to the trump action', () => {
    const state = makeCourtPieceState({ hint: { trumpSuit: 1, reason: 'trump_longest' } });
    expect(getCourtPieceHint(state)).toEqual({
      targetAction: 'trump',
      reason: 'hint.trump_longest',
      confidence: 'moderate',
    });
  });

  it('maps a follow_suit hint reason verbatim', () => {
    const state = makeCourtPieceState({ hint: { cardIndex: 0, reason: 'follow_suit' } });
    expect(getCourtPieceHint(state)?.reason).toBe('hint.follow_suit');
  });

  it('maps a discard_high hint reason verbatim', () => {
    const state = makeCourtPieceState({ hint: { cardIndex: 3, reason: 'discard_high' } });
    expect(getCourtPieceHint(state)?.reason).toBe('hint.discard_high');
  });
});
