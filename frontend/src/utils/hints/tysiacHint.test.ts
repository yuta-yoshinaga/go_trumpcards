import { describe, expect, it } from 'vitest';
import { makeTysiacState } from '../../test/stateFactories';
import { getTysiacHint } from './tysiacHint';

describe('getTysiacHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getTysiacHint(makeTysiacState())).toBeNull();
    expect(getTysiacHint(makeTysiacState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeTysiacState({ hint: { cardIndices: [], reason: '' } });
    expect(getTysiacHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeTysiacState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getTysiacHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeTysiacState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getTysiacHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a marriage lead hint reason verbatim', () => {
    const state = makeTysiacState({ hint: { cardIndices: [0], reason: 'lead_marriage' } });
    expect(getTysiacHint(state)?.reason).toBe('hint.lead_marriage');
  });

  it('maps a bid hint reason verbatim', () => {
    const state = makeTysiacState({ hint: { cardIndices: [], reason: 'bid_raise' } });
    expect(getTysiacHint(state)?.reason).toBe('hint.bid_raise');
  });

  it('maps a talon-discard hint reason verbatim', () => {
    const state = makeTysiacState({ hint: { cardIndices: [1], reason: 'talon_discard' } });
    expect(getTysiacHint(state)?.reason).toBe('hint.talon_discard');
  });
});
