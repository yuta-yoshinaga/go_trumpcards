import { describe, expect, it } from 'vitest';
import type { AccordionResponse } from '../../types/card';
import { getAccordionHint } from './accordionHint';

function makeState(overrides: Partial<AccordionResponse> = {}): AccordionResponse {
  return {
    piles: [],
    pileCount: 0,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getAccordionHint', () => {
  it('returns null when the game is over (clear)', () => {
    const state = makeState({ phase: 1, hint: { fromIdx: 3, toIdx: 0 } });
    expect(getAccordionHint(state)).toBeNull();
  });

  it('returns null when phase is game over (2)', () => {
    const state = makeState({ phase: 2, hint: { fromIdx: 3, toIdx: 0 } });
    expect(getAccordionHint(state)).toBeNull();
  });

  it('returns null when there is no backend hint', () => {
    expect(getAccordionHint(makeState())).toBeNull();
  });

  it('returns mergeOffset3 for an offset-3 hint', () => {
    const state = makeState({ hint: { fromIdx: 3, toIdx: 0 } });
    const hint = getAccordionHint(state);
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.accordionOffset3');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns mergeOffset1 for an offset-1 hint', () => {
    const state = makeState({ hint: { fromIdx: 2, toIdx: 1 } });
    const hint = getAccordionHint(state);
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.accordionOffset1');
  });
});
