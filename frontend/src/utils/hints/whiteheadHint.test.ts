import { describe, expect, it } from 'vitest';
import type { Card, WhiteheadResponse } from '../../types/card';
import { WhiteheadPhase } from '../../types/phases';
import { getWhiteheadHint } from './whiteheadHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
function card(design: Card['design'], value: number) {
  return { design, value };
}

function makeState(overrides: Partial<WhiteheadResponse> = {}): WhiteheadResponse {
  return {
    tableau: [
      [{ card: card(S, 7), faceUp: true }],
      [
        { card: card(H, 3), faceUp: false },
        { card: card(D, 9), faceUp: true },
      ],
      [],
      [],
      [],
      [],
      [],
    ],
    stockCount: 20,
    waste: [],
    foundation: [[], [], [], []],
    phase: WhiteheadPhase.PLAYING,
    moveCount: 0,
    drawCount: 1,
    canUndo: false,
    isStalemate: false,
    score: 0,
    scoringMode: 0,
    message: '',
    ...overrides,
  };
}

describe('getWhiteheadHint', () => {
  it('returns null when game is cleared', () => {
    expect(getWhiteheadHint(makeState({ phase: WhiteheadPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getWhiteheadHint(makeState({ phase: WhiteheadPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests moving Ace to foundation', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 1), faceUp: true }]],
    });
    const hint = getWhiteheadHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests moving card to foundation when it fits', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 2), faceUp: true }]],
      foundation: [[card(S, 1)], [], [], []],
    });
    const hint = getWhiteheadHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('suggests foundation move from waste', () => {
    const state = makeState({
      tableau: [[{ card: card(H, 7), faceUp: true }]],
      waste: [card(S, 1)],
      foundation: [[], [], [], []],
    });
    const hint = getWhiteheadHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  // Klondike's "reveal a face-down card" case is gone: Whitehead deals every
  // card face up, so the branch it covered cannot fire.

  // **Any card may take an empty column**, not just a King. Keeping Klondike's
  // King-only case would have passed while the hint stayed silent for every
  // non-King -- so the fixture below deliberately offers no King at all.
  it('suggests filling an empty column with a non-King', () => {
    const state = makeState({
      tableau: [
        [
          { card: card(S, 3), faceUp: true },
          { card: card(H, 9), faceUp: true },
          { card: card(D, 8), faceUp: true },
        ],
        [],
      ],
      foundation: [[], [], [], []],
      stockCount: 0,
      waste: [],
    });
    const hint = getWhiteheadHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToEmptyColumn');
    expect(hint?.confidence).toBe('moderate');
  });

  it('does not suggest an empty-column move when every column holds one card', () => {
    const state = makeState({
      tableau: [[{ card: card(H, 9), faceUp: true }], [{ card: card(D, 4), faceUp: true }], []],
      foundation: [[], [], [], []],
      stockCount: 0,
      waste: [],
    });
    // Every occupied column holds exactly one card, so shifting one into the
    // empty column changes nothing. Neither 9 nor 4 can start a foundation, and
    // the stock is gone, so there is genuinely nothing to suggest.
    expect(getWhiteheadHint(state)).toBeNull();
  });

  it('suggests drawing from stock as fallback', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 7), faceUp: true }]],
      foundation: [[], [], [], []],
      stockCount: 5,
    });
    const hint = getWhiteheadHint(state);
    expect(hint?.reason).toBe('frontendHint.drawFromStock');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests drawing even with waste only', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 7), faceUp: true }]],
      foundation: [[], [], [], []],
      stockCount: 0,
      waste: [card(H, 10)],
    });
    const hint = getWhiteheadHint(state);
    expect(hint?.reason).toBe('frontendHint.drawFromStock');
  });

  it('returns null when no moves possible', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 7), faceUp: true }]],
      foundation: [[], [], [], []],
      stockCount: 0,
      waste: [],
    });
    expect(getWhiteheadHint(state)).toBeNull();
  });
});
