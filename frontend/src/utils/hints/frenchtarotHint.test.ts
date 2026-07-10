import { describe, expect, it } from 'vitest';
import { makeFrenchTarotState } from '../../test/stateFactories';
import { getFrenchTarotHint } from './frenchtarotHint';

describe('getFrenchTarotHint', () => {
  it('returns null when there is no hint', () => {
    expect(getFrenchTarotHint(makeFrenchTarotState({ hint: null }))).toBeNull();
    expect(getFrenchTarotHint(makeFrenchTarotState({ hint: undefined }))).toBeNull();
  });

  it('returns null when the hint has an empty reason', () => {
    expect(getFrenchTarotHint(makeFrenchTarotState({ hint: { cardIndices: [1], reason: '' } }))).toBeNull();
  });

  it('maps a play hint reason into a namespaced HintResult', () => {
    const result = getFrenchTarotHint(makeFrenchTarotState({ hint: { cardIndices: [2], reason: 'lead_high' } }));
    expect(result).toEqual({ targetAction: 'play', reason: 'hint.lead_high', confidence: 'moderate' });
  });

  it('maps a bid hint reason (carrying a bid value)', () => {
    const result = getFrenchTarotHint(makeFrenchTarotState({ hint: { bid: 2, cardIndices: [], reason: 'bid_take' } }));
    expect(result).toEqual({ targetAction: 'play', reason: 'hint.bid_take', confidence: 'moderate' });
  });
});
