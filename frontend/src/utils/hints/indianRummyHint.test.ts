import { describe, expect, it } from 'vitest';
import type { Card, IndianRummyPlayer, IndianRummyResponse } from '../../types/card';
import { IndianRummyPhase } from '../../types/phases';
import { calcDeadwood, getIndianRummyHint, isWild } from './indianRummyHint';

function card(design: Card['design'], value: number): Card {
  return { design, value };
}

function player(overrides: Partial<IndianRummyPlayer> = {}): IndianRummyPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    deadwood: 0,
    hasPureSequence: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<IndianRummyResponse> = {}): IndianRummyResponse {
  return {
    players: [player(), player({ id: 1, isHuman: false })],
    phase: IndianRummyPhase.DRAW,
    roundNumber: 1,
    targetRounds: 3,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    discardTop: null,
    drawPileCount: 40,
    wildJoker: null,
    wildRank: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    declarerIdx: -1,
    declarationValid: false,
    message: '',
    config: { playerCount: 2, cpuDifficulty: 1, targetRounds: 3 },
    ...overrides,
  };
}

describe('getIndianRummyHint', () => {
  it('returns null when no human player', () => {
    expect(getIndianRummyHint(makeState({ players: [player({ isHuman: false })] }))).toBeNull();
  });

  it('returns null when human has no cards', () => {
    expect(getIndianRummyHint(makeState())).toBeNull();
  });

  it('returns null when gameEndFlag is true', () => {
    const state = makeState({ gameEndFlag: true, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getIndianRummyHint(state)).toBeNull();
  });

  it('returns null when not the human turn', () => {
    const state = makeState({ currentPlayerIdx: 1, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getIndianRummyHint(state)).toBeNull();
  });

  it('returns null for round-end phase', () => {
    const state = makeState({ phase: IndianRummyPhase.ROUND_END, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getIndianRummyHint(state)).toBeNull();
  });

  describe('draw phase', () => {
    it('suggests draw from discard when the top forms a set', () => {
      const state = makeState({
        phase: IndianRummyPhase.DRAW,
        discardTop: card('HEART', 7),
        players: [player({ cards: [card('SPADE', 7), card('DIAMOND', 2)] })],
      });
      const hint = getIndianRummyHint(state);
      expect(hint?.targetAction).toBe('drawDiscard');
      expect(hint?.reason).toBe('hint.drawFromDiscard');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests draw from discard when the top extends a run', () => {
      const state = makeState({
        phase: IndianRummyPhase.DRAW,
        discardTop: card('HEART', 5),
        players: [player({ cards: [card('HEART', 4), card('SPADE', 2)] })],
      });
      expect(getIndianRummyHint(state)?.targetAction).toBe('drawDiscard');
    });

    it('suggests draw from stock when the top does not fit', () => {
      const state = makeState({
        phase: IndianRummyPhase.DRAW,
        discardTop: card('HEART', 13),
        players: [player({ cards: [card('SPADE', 2), card('CLOVER', 4), card('DIAMOND', 8)] })],
      });
      const hint = getIndianRummyHint(state);
      expect(hint?.targetAction).toBe('drawStock');
      expect(hint?.reason).toBe('hint.drawFromStock');
    });

    it('suggests draw from stock when there is no discard top', () => {
      const state = makeState({
        phase: IndianRummyPhase.DRAW,
        discardTop: null,
        players: [player({ cards: [card('SPADE', 2)] })],
      });
      expect(getIndianRummyHint(state)?.targetAction).toBe('drawStock');
    });

    it('ignores a wild card on top of the discard pile', () => {
      const state = makeState({
        phase: IndianRummyPhase.DRAW,
        wildRank: 7,
        discardTop: card('HEART', 7),
        players: [player({ cards: [card('SPADE', 7)] })],
      });
      expect(getIndianRummyHint(state)?.targetAction).toBe('drawStock');
    });
  });

  describe('discard phase', () => {
    it('suggests declare when one discard clears all deadwood', () => {
      const state = makeState({
        phase: IndianRummyPhase.DISCARD,
        players: [
          player({
            cardCount: 14,
            cards: [
              card('SPADE', 3),
              card('SPADE', 4),
              card('SPADE', 5),
              card('HEART', 6),
              card('HEART', 7),
              card('HEART', 8),
              card('SPADE', 9),
              card('HEART', 9),
              card('CLOVER', 9),
              card('SPADE', 10),
              card('HEART', 10),
              card('CLOVER', 10),
              card('DIAMOND', 10),
              card('DIAMOND', 2), // lone deadwood card to discard
            ],
          }),
        ],
      });
      const hint = getIndianRummyHint(state);
      expect(hint?.targetAction).toBe('declare');
      expect(hint?.reason).toBe('hint.declareNow');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests discard when deadwood cannot be cleared', () => {
      const state = makeState({
        phase: IndianRummyPhase.DISCARD,
        players: [
          player({
            cardCount: 14,
            cards: [
              card('HEART', 2),
              card('SPADE', 5),
              card('CLOVER', 8),
              card('DIAMOND', 11),
              card('HEART', 13),
              card('SPADE', 9),
              card('CLOVER', 4),
              card('DIAMOND', 6),
              card('HEART', 1),
              card('SPADE', 12),
              card('DIAMOND', 3),
              card('CLOVER', 7),
              card('HEART', 10),
              card('SPADE', 1),
            ],
          }),
        ],
      });
      const hint = getIndianRummyHint(state);
      expect(hint?.targetAction).toBe('discard');
      expect(hint?.reason).toBe('hint.discardDeadwood');
    });
  });
});

describe('isWild', () => {
  it('treats a printed joker as wild', () => {
    expect(isWild(card('JOKER', 0), 0)).toBe(true);
  });

  it('treats a card of the wild rank as wild', () => {
    expect(isWild(card('HEART', 7), 7)).toBe(true);
    expect(isWild(card('SPADE', 7), 7)).toBe(true);
  });

  it('is not wild when the rank does not match and wildRank is 0', () => {
    expect(isWild(card('HEART', 7), 0)).toBe(false);
  });
});

describe('calcDeadwood', () => {
  it('returns 0 for a hand that fully melds', () => {
    const hand = [
      card('HEART', 3),
      card('SPADE', 3),
      card('CLOVER', 3),
      card('HEART', 7),
      card('SPADE', 7),
      card('CLOVER', 7),
    ];
    expect(calcDeadwood(hand, 0)).toBe(0);
  });

  it('counts face cards as 10 points', () => {
    const hand = [card('HEART', 11), card('SPADE', 12), card('CLOVER', 13)];
    expect(calcDeadwood(hand, 0)).toBe(30);
  });

  it('lets a printed joker cancel the highest unmatched card', () => {
    const hand = [card('JOKER', 0), card('SPADE', 5), card('CLOVER', 9)];
    expect(calcDeadwood(hand, 0)).toBe(5);
  });

  it('lets a wild-rank card cancel the highest unmatched card', () => {
    const hand = [card('HEART', 2), card('SPADE', 5), card('CLOVER', 9)];
    expect(calcDeadwood(hand, 2)).toBe(5);
  });

  it('detects runs in the same suit', () => {
    const hand = [card('HEART', 4), card('HEART', 5), card('HEART', 6), card('SPADE', 10)];
    expect(calcDeadwood(hand, 0)).toBe(10);
  });

  it('scores an unmatched Ace as 10 points, not 1 (matching the backend)', () => {
    const hand = [card('SPADE', 1), card('HEART', 4), card('CLOVER', 8)];
    expect(calcDeadwood(hand, 0)).toBe(22);
  });
});
