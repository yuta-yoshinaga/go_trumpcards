import { describe, expect, it } from 'vitest';
import type { LaBelleLucieResponse } from '../../types/card';
import { getLaBelleLucieHint } from './labellelucieHint';

function makeState(overrides?: Partial<LaBelleLucieResponse>): LaBelleLucieResponse {
  return {
    fans: [[], [], []],
    foundation: [[], [], [], []],
    redealsLeft: 2,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('getLaBelleLucieHint', () => {
  it('rates a foundation move as strong', () => {
    const hint = getLaBelleLucieHint(makeState({ hint: { fromFan: 2, toFan: -1, toFoundation: true } }));
    expect(hint?.targetAction).toBe('fan-2');
    expect(hint?.reason).toBe('frontendHint.labellelucieToFoundation');
    expect(hint?.confidence).toBe('strong');
  });

  it('rates a fan-to-fan move as moderate', () => {
    const hint = getLaBelleLucieHint(makeState({ hint: { fromFan: 1, toFan: 4, toFoundation: false } }));
    expect(hint?.reason).toBe('frontendHint.labellelucieToFan');
    expect(hint?.confidence).toBe('moderate');
  });

  // **台札行きは toFan が -1 でも成立する。**移動先が扇ではないため。
  it('does not require a target fan for a foundation move', () => {
    expect(getLaBelleLucieHint(makeState({ hint: { fromFan: 0, toFan: -1, toFoundation: true } }))).not.toBeNull();
  });

  it('returns null when a fan-to-fan move has no target', () => {
    expect(getLaBelleLucieHint(makeState({ hint: { fromFan: 0, toFan: -1, toFoundation: false } }))).toBeNull();
  });

  it('returns null without a hint', () => {
    expect(getLaBelleLucieHint(makeState())).toBeNull();
  });

  it('returns null once the game has ended', () => {
    expect(
      getLaBelleLucieHint(makeState({ phase: 1, hint: { fromFan: 0, toFan: 1, toFoundation: false } })),
    ).toBeNull();
  });
});
