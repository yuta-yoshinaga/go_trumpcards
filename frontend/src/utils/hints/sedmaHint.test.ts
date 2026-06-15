import { describe, expect, it } from 'vitest';
import { makeSedmaState } from '../../test/stateFactories';
import { getSedmaHint } from './sedmaHint';

describe('getSedmaHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getSedmaHint(makeSedmaState())).toBeNull();
    expect(getSedmaHint(makeSedmaState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeSedmaState({ hint: { cardIndices: [], reason: '' } });
    expect(getSedmaHint(state)).toBeNull();
  });

  it('maps a server capture hint into a HintResult', () => {
    const state = makeSedmaState({ hint: { cardIndices: [2], reason: 'capture' } });
    expect(getSedmaHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.capture',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeSedmaState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getSedmaHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeSedmaState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getSedmaHint(state)?.reason).toBe('hint.discard_low');
  });
});
