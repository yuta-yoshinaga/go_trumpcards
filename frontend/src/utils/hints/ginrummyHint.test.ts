import { describe, expect, it } from 'vitest';
import type { Card, GinRummyMeld, GinRummyResponse } from '../../types/card';
import { GinRummyPhase } from '../../types/phases';
import { calcDeadwood, getGinRummyHint } from './ginrummyHint';

function makeState(overrides: Partial<GinRummyResponse> = {}): GinRummyResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
      { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
    ],
    layoffTargets: [],
    phase: GinRummyPhase.DRAW,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: null,
    drawPileCount: 20,
    gameEndFlag: false,
    winnerIdx: -1,
    knockerIdx: -1,
    knockerMelds: [],
    knockerDeadwood: [],
    isGin: false,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 100 },
    ...overrides,
  };
}

function card(design: Card['design'], value: number): Card {
  return { design, value };
}

describe('getGinRummyHint', () => {
  it('returns null when no human player', () => {
    const state = makeState({
      players: [{ id: 0, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 }],
    });
    expect(getGinRummyHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    expect(getGinRummyHint(state)).toBeNull();
  });

  it('returns null when gameEndFlag is true', () => {
    const state = makeState({
      gameEndFlag: true,
      players: [
        { id: 0, isHuman: true, cardCount: 1, cards: [card('HEART', 5)], roundScore: 0, cumulativeScore: 0 },
        { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getGinRummyHint(state)).toBeNull();
  });

  it('returns null when not human turn', () => {
    const state = makeState({
      currentPlayerIdx: 1,
      players: [
        { id: 0, isHuman: true, cardCount: 1, cards: [card('HEART', 5)], roundScore: 0, cumulativeScore: 0 },
        { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getGinRummyHint(state)).toBeNull();
  });

  it('returns null for round end phase', () => {
    const state = makeState({
      phase: GinRummyPhase.ROUND_END,
      players: [
        { id: 0, isHuman: true, cardCount: 1, cards: [card('HEART', 5)], roundScore: 0, cumulativeScore: 0 },
        { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getGinRummyHint(state)).toBeNull();
  });

  describe('draw phase', () => {
    it('suggests drawing from discard when card fits hand (set)', () => {
      const state = makeState({
        phase: GinRummyPhase.DRAW,
        currentPlayerIdx: 0,
        discardTop: card('HEART', 7),
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 10,
            cards: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 2), card('DIAMOND', 9)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint).not.toBeNull();
      expect(hint?.targetAction).toBe('drawDiscard');
      expect(hint?.reason).toBe('hint.drawFromDiscard');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests drawing from discard when card fits hand (run)', () => {
      const state = makeState({
        phase: GinRummyPhase.DRAW,
        currentPlayerIdx: 0,
        discardTop: card('HEART', 5),
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 10,
            cards: [card('HEART', 4), card('HEART', 6), card('SPADE', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('drawDiscard');
    });

    it('suggests drawing from stock when discard does not fit', () => {
      const state = makeState({
        phase: GinRummyPhase.DRAW,
        currentPlayerIdx: 0,
        discardTop: card('HEART', 13),
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 10,
            cards: [card('SPADE', 2), card('CLOVER', 4), card('DIAMOND', 8)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('drawStock');
      expect(hint?.reason).toBe('hint.drawFromStock');
    });

    it('suggests drawing from stock when no discard top', () => {
      const state = makeState({
        phase: GinRummyPhase.DRAW,
        currentPlayerIdx: 0,
        discardTop: null,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 10,
            cards: [card('SPADE', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('drawStock');
    });
  });

  describe('discard phase', () => {
    it('suggests gin when deadwood is 0 after discard', () => {
      // 3 sets of 3 + 1 deadwood card => after discard deadwood = 0
      const state = makeState({
        phase: GinRummyPhase.DISCARD,
        currentPlayerIdx: 0,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 10,
            cards: [
              card('HEART', 3),
              card('SPADE', 3),
              card('CLOVER', 3),
              card('HEART', 7),
              card('SPADE', 7),
              card('CLOVER', 7),
              card('HEART', 10),
              card('SPADE', 10),
              card('CLOVER', 10),
              card('DIAMOND', 5), // deadwood to discard
            ],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('knock');
      expect(hint?.reason).toBe('hint.ginOpportunity');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests knock when deadwood is low after discard', () => {
      const state = makeState({
        phase: GinRummyPhase.DISCARD,
        currentPlayerIdx: 0,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 10,
            cards: [
              card('HEART', 3),
              card('SPADE', 3),
              card('CLOVER', 3),
              card('HEART', 7),
              card('SPADE', 7),
              card('CLOVER', 7),
              card('HEART', 10),
              card('SPADE', 10),
              card('CLOVER', 10),
              card('DIAMOND', 2), // 2 deadwood = discard this, remaining deadwood = 0 => gin
            ],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('knock');
    });

    it('suggests discard when deadwood is high', () => {
      const state = makeState({
        phase: GinRummyPhase.DISCARD,
        currentPlayerIdx: 0,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 10,
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
              card('DIAMOND', 10),
            ],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('discard');
      expect(hint?.reason).toBe('hint.discardDeadwood');
    });
  });

  describe('layoff phase', () => {
    it('suggests layoff when card fits knocker meld (set)', () => {
      const knockerMelds: GinRummyMeld[] = [{ cards: [card('HEART', 5), card('SPADE', 5), card('CLOVER', 5)] }];
      const state = makeState({
        phase: GinRummyPhase.LAYOFF,
        currentPlayerIdx: 0,
        knockerMelds,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 3,
            cards: [card('DIAMOND', 5), card('HEART', 9), card('SPADE', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('layoff');
      expect(hint?.reason).toBe('hint.layoffCards');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests layoff when card fits knocker meld (run)', () => {
      const knockerMelds: GinRummyMeld[] = [{ cards: [card('HEART', 3), card('HEART', 4), card('HEART', 5)] }];
      const state = makeState({
        phase: GinRummyPhase.LAYOFF,
        currentPlayerIdx: 0,
        knockerMelds,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('HEART', 6), card('SPADE', 12)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('layoff');
    });

    it('suggests skip layoff when no cards fit', () => {
      const knockerMelds: GinRummyMeld[] = [{ cards: [card('HEART', 5), card('SPADE', 5), card('CLOVER', 5)] }];
      const state = makeState({
        phase: GinRummyPhase.LAYOFF,
        currentPlayerIdx: 0,
        knockerMelds,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('DIAMOND', 9), card('SPADE', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('skipLayoff');
      expect(hint?.reason).toBe('hint.skipLayoff');
    });

    it('suggests skip layoff when knocker melds are empty', () => {
      const state = makeState({
        phase: GinRummyPhase.LAYOFF,
        currentPlayerIdx: 0,
        knockerMelds: [],
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('DIAMOND', 9), card('SPADE', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getGinRummyHint(state);
      expect(hint?.targetAction).toBe('skipLayoff');
    });
  });
});

describe('calcDeadwood', () => {
  it('returns 0 for a hand of all melds', () => {
    const hand = [
      card('HEART', 3),
      card('SPADE', 3),
      card('CLOVER', 3),
      card('HEART', 7),
      card('SPADE', 7),
      card('CLOVER', 7),
    ];
    expect(calcDeadwood(hand)).toBe(0);
  });

  it('calculates deadwood correctly with mixed melds and deadwood', () => {
    const hand = [
      card('HEART', 3),
      card('SPADE', 3),
      card('CLOVER', 3),
      card('DIAMOND', 9), // deadwood: 9
      card('HEART', 1), // deadwood: 1
    ];
    expect(calcDeadwood(hand)).toBe(10);
  });

  it('counts face cards as 10 points', () => {
    const hand = [card('HEART', 11), card('SPADE', 12), card('CLOVER', 13)]; // all deadwood: 10 + 10 + 10
    // But these are all different values, no melds => 30
    expect(calcDeadwood(hand)).toBe(30);
  });

  it('detects runs in same suit', () => {
    const hand = [
      card('HEART', 4),
      card('HEART', 5),
      card('HEART', 6),
      card('SPADE', 10), // deadwood: 10
    ];
    expect(calcDeadwood(hand)).toBe(10);
  });

  it('chooses runs-first when it yields lower deadwood than sets-first', () => {
    // Cards: 3H 4H 5H 5S 5C 9D
    // Sets-first: {5H,5S,5C} as set => remaining 3H,4H,9D = 3+4+9 = 16 deadwood
    // Runs-first: {3H,4H,5H} as run => remaining 5S,5C,9D = 5+5+9 = 19 deadwood
    // Actually sets-first is better here. Let's construct the opposite:
    // Cards: 3H 4H 5H 5S 5C 6H
    // Sets-first: {5H,5S,5C} as set => remaining 3H,4H,6H = 3+4+6 = 13 deadwood
    // Runs-first: {3H,4H,5H,6H} as run => remaining 5S,5C = 5+5 = 10 deadwood
    const hand = [
      card('HEART', 3),
      card('HEART', 4),
      card('HEART', 5),
      card('SPADE', 5),
      card('CLOVER', 5),
      card('HEART', 6),
    ];
    expect(calcDeadwood(hand)).toBe(10); // runs-first wins
  });
});
