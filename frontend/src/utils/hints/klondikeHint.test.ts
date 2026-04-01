import { describe, expect, it } from 'vitest';
import type { Card, KlondikeResponse } from '../../types/card';
import { KlondikePhase } from '../../types/phases';
import { getKlondikeHint } from './klondikeHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
function card(design: Card['design'], value: number) {
  return { design, value };
}

function makeState(overrides: Partial<KlondikeResponse> = {}): KlondikeResponse {
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
    phase: KlondikePhase.PLAYING,
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

describe('getKlondikeHint', () => {
  it('returns null when game is cleared', () => {
    expect(getKlondikeHint(makeState({ phase: KlondikePhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getKlondikeHint(makeState({ phase: KlondikePhase.GAME_OVER }))).toBeNull();
  });

  it('suggests moving Ace to foundation', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 1), faceUp: true }]],
    });
    const hint = getKlondikeHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests moving card to foundation when it fits', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 2), faceUp: true }]],
      foundation: [[card(S, 1)], [], [], []],
    });
    const hint = getKlondikeHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('suggests foundation move from waste', () => {
    const state = makeState({
      tableau: [[{ card: card(H, 7), faceUp: true }]],
      waste: [card(S, 1)],
      foundation: [[], [], [], []],
    });
    const hint = getKlondikeHint(state);
    expect(hint?.reason).toBe('frontendHint.moveToFoundation');
  });

  it('suggests revealing face-down card', () => {
    const state = makeState({
      tableau: [
        [
          { card: card(S, 5), faceUp: false },
          { card: card(H, 4), faceUp: true },
        ],
      ],
      foundation: [[], [], [], []],
    });
    const hint = getKlondikeHint(state);
    expect(hint?.reason).toBe('frontendHint.revealFaceDown');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests moving King to empty column', () => {
    const state = makeState({
      tableau: [
        [
          { card: card(S, 3), faceUp: true },
          { card: card(H, 13), faceUp: true },
          { card: card(D, 12), faceUp: true },
        ],
        [],
      ],
      foundation: [[], [], [], []],
      stockCount: 0,
      waste: [],
    });
    const hint = getKlondikeHint(state);
    expect(hint?.reason).toBe('frontendHint.moveKingToEmpty');
    expect(hint?.confidence).toBe('moderate');
  });

  it('does not suggest King move if King is already at base', () => {
    const state = makeState({
      tableau: [
        [
          { card: card(H, 13), faceUp: true },
          { card: card(D, 12), faceUp: true },
        ],
        [],
      ],
      foundation: [[], [], [], []],
      stockCount: 0,
      waste: [],
    });
    // King is at array index 0 (base of column) — i > 0 guard in hasKingForEmptyColumn prevents hint; no face-down cards, no stock → null
    expect(getKlondikeHint(state)).toBeNull();
  });

  it('suggests drawing from stock as fallback', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 7), faceUp: true }]],
      foundation: [[], [], [], []],
      stockCount: 5,
    });
    const hint = getKlondikeHint(state);
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
    const hint = getKlondikeHint(state);
    expect(hint?.reason).toBe('frontendHint.drawFromStock');
  });

  it('returns null when no moves possible', () => {
    const state = makeState({
      tableau: [[{ card: card(S, 7), faceUp: true }]],
      foundation: [[], [], [], []],
      stockCount: 0,
      waste: [],
    });
    expect(getKlondikeHint(state)).toBeNull();
  });
});
