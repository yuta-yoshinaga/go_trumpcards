import { describe, expect, it } from 'vitest';
import { makeGoStopState } from '../../test/stateFactories';
import { getGoStopHint } from './gostopHint';

describe('getGoStopHint', () => {
  it('returns null when there is no hint', () => {
    expect(getGoStopHint(makeGoStopState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint has no reason', () => {
    expect(getGoStopHint(makeGoStopState({ hint: { cardIndex: 0, fieldIndex: -1, go: -1, reason: '' } }))).toBeNull();
  });

  it('maps a Play-phase hint to the play action', () => {
    const res = getGoStopHint(
      makeGoStopState({ phase: 0, hint: { cardIndex: 1, fieldIndex: 0, go: -1, reason: 'capture' } }),
    );
    expect(res).toEqual({ targetAction: 'play', reason: 'hint.capture', confidence: 'moderate' });
  });

  it('maps a GoDecision-phase hint to the decide action', () => {
    const res = getGoStopHint(
      makeGoStopState({ phase: 1, hint: { cardIndex: -1, fieldIndex: -1, go: 1, reason: 'go_lowscore' } }),
    );
    expect(res).toEqual({ targetAction: 'decide', reason: 'hint.go_lowscore', confidence: 'moderate' });
  });
});
