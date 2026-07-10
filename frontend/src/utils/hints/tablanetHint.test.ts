import { describe, expect, it } from 'vitest';
import { makeTablanetState } from '../../test/stateFactories';
import { getTablanetHint } from './tablanetHint';

describe('getTablanetHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getTablanetHint(makeTablanetState())).toBeNull();
    expect(getTablanetHint(makeTablanetState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeTablanetState({ hint: { cardIndices: [], reason: '' } });
    expect(getTablanetHint(state)).toBeNull();
  });

  it('maps a server capture hint into a HintResult', () => {
    const state = makeTablanetState({ hint: { cardIndices: [0], tableIndices: [0], reason: 'capture' } });
    expect(getTablanetHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.capture',
      confidence: 'moderate',
    });
  });

  it('maps a tabla sweep hint reason verbatim', () => {
    const state = makeTablanetState({ hint: { cardIndices: [1], tableIndices: [0, 1], reason: 'tabla_sweep' } });
    expect(getTablanetHint(state)?.reason).toBe('hint.tabla_sweep');
  });

  it('maps a jack sweep hint reason verbatim', () => {
    const state = makeTablanetState({ hint: { cardIndices: [2], reason: 'jack_sweep' } });
    expect(getTablanetHint(state)?.reason).toBe('hint.jack_sweep');
  });

  it('maps a trail hint reason verbatim', () => {
    const state = makeTablanetState({ hint: { cardIndices: [3], reason: 'trail_low' } });
    expect(getTablanetHint(state)?.reason).toBe('hint.trail_low');
  });
});
