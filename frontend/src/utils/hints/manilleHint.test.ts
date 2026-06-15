import { describe, expect, it } from 'vitest';
import { makeManilleState } from '../../test/stateFactories';
import { getManilleHint } from './manilleHint';

describe('getManilleHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getManilleHint(makeManilleState())).toBeNull();
    expect(getManilleHint(makeManilleState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeManilleState({ hint: { cardIndices: [], reason: '' } });
    expect(getManilleHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeManilleState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getManilleHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeManilleState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getManilleHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeManilleState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getManilleHint(state)?.reason).toBe('hint.discard_low');
  });
});
