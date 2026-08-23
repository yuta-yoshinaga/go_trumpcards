import { describe, expect, it } from 'vitest';
import { makeQuadrilleState } from '../../test/stateFactories';
import { getQuadrilleHint } from './quadrilleHint';

describe('getQuadrilleHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getQuadrilleHint(makeQuadrilleState())).toBeNull();
    expect(getQuadrilleHint(makeQuadrilleState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeQuadrilleState({ hint: { cardIndices: [], reason: '' } });
    expect(getQuadrilleHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeQuadrilleState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getQuadrilleHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead-high hint reason verbatim', () => {
    const state = makeQuadrilleState({ hint: { cardIndices: [0], reason: 'lead_high' } });
    expect(getQuadrilleHint(state)?.reason).toBe('hint.lead_high');
  });

  it('maps a give-partner hint reason verbatim', () => {
    const state = makeQuadrilleState({ hint: { cardIndices: [0], reason: 'give_partner' } });
    expect(getQuadrilleHint(state)?.reason).toBe('hint.give_partner');
  });

  it('maps an entrar bid hint reason verbatim', () => {
    const state = makeQuadrilleState({ hint: { cardIndices: [], reason: 'bid_entrar' } });
    expect(getQuadrilleHint(state)?.reason).toBe('hint.bid_entrar');
  });
});
