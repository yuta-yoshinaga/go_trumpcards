import { describe, expect, it } from 'vitest';
import { makeDoppelkopfState } from '../../test/stateFactories';
import { getDoppelkopfHint } from './doppelkopfHint';

describe('getDoppelkopfHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getDoppelkopfHint(makeDoppelkopfState())).toBeNull();
    expect(getDoppelkopfHint(makeDoppelkopfState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeDoppelkopfState({ hint: { cardIndices: [], reason: '' } });
    expect(getDoppelkopfHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeDoppelkopfState({
      hint: { cardIndices: [2], reason: 'follow_win' },
    });
    expect(getDoppelkopfHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeDoppelkopfState({
      hint: { cardIndices: [0], reason: 'lead_low' },
    });
    expect(getDoppelkopfHint(state)?.reason).toBe('hint.lead_low');
  });
});
