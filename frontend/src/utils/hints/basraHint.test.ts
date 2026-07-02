import { describe, expect, it } from 'vitest';
import { makeBasraState } from '../../test/stateFactories';
import { getBasraHint } from './basraHint';

describe('getBasraHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getBasraHint(makeBasraState())).toBeNull();
    expect(getBasraHint(makeBasraState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeBasraState({ hint: { cardIndices: [], reason: '' } });
    expect(getBasraHint(state)).toBeNull();
  });

  it('maps a server capture hint into a HintResult', () => {
    const state = makeBasraState({ hint: { cardIndices: [0], tableIndices: [0], reason: 'capture' } });
    expect(getBasraHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.capture',
      confidence: 'moderate',
    });
  });

  it('maps a basra sweep hint reason verbatim', () => {
    const state = makeBasraState({ hint: { cardIndices: [1], tableIndices: [0, 1], reason: 'basra_sweep' } });
    expect(getBasraHint(state)?.reason).toBe('hint.basra_sweep');
  });

  it('maps a jack sweep hint reason verbatim', () => {
    const state = makeBasraState({ hint: { cardIndices: [2], reason: 'jack_sweep' } });
    expect(getBasraHint(state)?.reason).toBe('hint.jack_sweep');
  });

  it('maps a trail hint reason verbatim', () => {
    const state = makeBasraState({ hint: { cardIndices: [3], reason: 'trail_low' } });
    expect(getBasraHint(state)?.reason).toBe('hint.trail_low');
  });
});
