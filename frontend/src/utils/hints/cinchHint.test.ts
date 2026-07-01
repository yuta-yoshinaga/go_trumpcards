import { describe, expect, it } from 'vitest';
import { makeCinchState } from '../../test/stateFactories';
import { getCinchHint } from './cinchHint';

describe('getCinchHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getCinchHint(makeCinchState())).toBeNull();
    expect(getCinchHint(makeCinchState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeCinchState({ hint: { cardIndices: [], reason: '' } });
    expect(getCinchHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeCinchState({ hint: { cardIndices: [2], reason: 'trump_cut' } });
    expect(getCinchHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.trump_cut',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeCinchState({ hint: { cardIndices: [0], reason: 'lead_strong' } });
    expect(getCinchHint(state)?.reason).toBe('hint.lead_strong');
  });

  it('maps a follow-suit hint reason verbatim', () => {
    const state = makeCinchState({ hint: { cardIndices: [1], reason: 'follow_suit' } });
    expect(getCinchHint(state)?.reason).toBe('hint.follow_suit');
  });

  it('maps a bid hint reason verbatim', () => {
    const state = makeCinchState({ hint: { cardIndices: [], bid: 8, reason: 'bid_strong' } });
    expect(getCinchHint(state)?.reason).toBe('hint.bid_strong');
  });

  it('maps a name-trump hint reason verbatim', () => {
    const state = makeCinchState({ hint: { cardIndices: [], trumpSuit: 1, reason: 'name_trump' } });
    expect(getCinchHint(state)?.reason).toBe('hint.name_trump');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeCinchState({ hint: { cardIndices: [3], reason: 'discard_low' } });
    expect(getCinchHint(state)?.reason).toBe('hint.discard_low');
  });
});
