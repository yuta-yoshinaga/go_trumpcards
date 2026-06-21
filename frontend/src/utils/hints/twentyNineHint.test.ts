import { describe, expect, it } from 'vitest';
import { makeTwentyNineState } from '../../test/stateFactories';
import { getTwentyNineHint } from './twentyNineHint';

describe('getTwentyNineHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getTwentyNineHint(makeTwentyNineState())).toBeNull();
    expect(getTwentyNineHint(makeTwentyNineState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeTwentyNineState({ hint: { cardIndices: [], reason: '' } });
    expect(getTwentyNineHint(state)).toBeNull();
  });

  it('maps a server follow_win hint into a HintResult', () => {
    const state = makeTwentyNineState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getTwentyNineHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead_low hint reason verbatim', () => {
    const state = makeTwentyNineState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getTwentyNineHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a follow_duck hint reason verbatim', () => {
    const state = makeTwentyNineState({ hint: { cardIndices: [1], reason: 'follow_duck' } });
    expect(getTwentyNineHint(state)?.reason).toBe('hint.follow_duck');
  });

  it('maps a discard_low hint reason verbatim', () => {
    const state = makeTwentyNineState({ hint: { cardIndices: [3], reason: 'discard_low' } });
    expect(getTwentyNineHint(state)?.reason).toBe('hint.discard_low');
  });
});
