import { describe, expect, it } from 'vitest';
import { makeMariasState } from '../../test/stateFactories';
import { getMariasHint } from './mariasHint';

describe('getMariasHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getMariasHint(makeMariasState())).toBeNull();
    expect(getMariasHint(makeMariasState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeMariasState({ hint: { cardIndices: [], reason: '' } });
    expect(getMariasHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeMariasState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getMariasHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeMariasState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getMariasHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeMariasState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getMariasHint(state)?.reason).toBe('hint.discard_low');
  });
});
