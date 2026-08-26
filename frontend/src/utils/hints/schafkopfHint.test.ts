import { describe, expect, it } from 'vitest';
import { makeSchafkopfState } from '../../test/stateFactories';
import { getSchafkopfHint } from './schafkopfHint';

describe('getSchafkopfHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getSchafkopfHint(makeSchafkopfState())).toBeNull();
    expect(getSchafkopfHint(makeSchafkopfState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeSchafkopfState({ hint: { cardIndices: [], suit: 0, pick: false, reason: '' } });
    expect(getSchafkopfHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeSchafkopfState({
      hint: { cardIndices: [2], suit: 0, pick: false, reason: 'follow_win' },
    });
    expect(getSchafkopfHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a pick hint reason verbatim', () => {
    const state = makeSchafkopfState({
      phase: 0,
      hint: { cardIndices: [], suit: 0, pick: true, reason: 'pick_take' },
    });
    expect(getSchafkopfHint(state)?.reason).toBe('hint.pick_take');
  });
});
