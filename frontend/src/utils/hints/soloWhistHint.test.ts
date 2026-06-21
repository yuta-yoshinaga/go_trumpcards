import { describe, expect, it } from 'vitest';
import { makeSoloWhistState } from '../../test/stateFactories';
import { getSoloWhistHint } from './soloWhistHint';

describe('getSoloWhistHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getSoloWhistHint(makeSoloWhistState())).toBeNull();
    expect(getSoloWhistHint(makeSoloWhistState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeSoloWhistState({ hint: { cardIndices: [], reason: '' } });
    expect(getSoloWhistHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeSoloWhistState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getSoloWhistHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeSoloWhistState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getSoloWhistHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeSoloWhistState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getSoloWhistHint(state)?.reason).toBe('hint.discard_low');
  });
});
