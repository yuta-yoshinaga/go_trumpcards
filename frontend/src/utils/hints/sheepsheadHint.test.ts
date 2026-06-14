import { describe, expect, it } from 'vitest';
import { makeSheepsheadState } from '../../test/stateFactories';
import { getSheepsheadHint } from './sheepsheadHint';

describe('getSheepsheadHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getSheepsheadHint(makeSheepsheadState())).toBeNull();
    expect(getSheepsheadHint(makeSheepsheadState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeSheepsheadState({ hint: { cardIndices: [], suit: 0, pick: false, reason: '' } });
    expect(getSheepsheadHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeSheepsheadState({
      hint: { cardIndices: [2], suit: 0, pick: false, reason: 'follow_win' },
    });
    expect(getSheepsheadHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a pick hint reason verbatim', () => {
    const state = makeSheepsheadState({
      phase: 0,
      hint: { cardIndices: [], suit: 0, pick: true, reason: 'pick_take' },
    });
    expect(getSheepsheadHint(state)?.reason).toBe('hint.pick_take');
  });
});
