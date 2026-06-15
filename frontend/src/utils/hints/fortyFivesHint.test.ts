import { describe, expect, it } from 'vitest';
import { makeFortyFivesState } from '../../test/stateFactories';
import { getFortyFivesHint } from './fortyFivesHint';

describe('getFortyFivesHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getFortyFivesHint(makeFortyFivesState())).toBeNull();
    expect(getFortyFivesHint(makeFortyFivesState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeFortyFivesState({ hint: { cardIndices: [], reason: '' } });
    expect(getFortyFivesHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeFortyFivesState({ hint: { cardIndices: [2], reason: 'take_trick' } });
    expect(getFortyFivesHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.take_trick',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeFortyFivesState({ hint: { cardIndices: [0], reason: 'lead_high' } });
    expect(getFortyFivesHint(state)?.reason).toBe('hint.lead_high');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeFortyFivesState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getFortyFivesHint(state)?.reason).toBe('hint.discard_low');
  });
});
