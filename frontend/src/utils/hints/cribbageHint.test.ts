import { describe, expect, it } from 'vitest';
import type { Card, CribbageResponse } from '../../types/card';
import { CribbagePhase } from '../../types/phases';
import { getCribbageHint } from './cribbageHint';

function makeState(overrides: Partial<CribbageResponse> = {}): CribbageResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
      { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
    ],
    phase: CribbagePhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    crib: [],
    starter: null,
    pegCount: 0,
    pegPlayedCards: [],
    showPhaseStep: 0,
    handScoreDetails: [null, null],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 121 },
    ...overrides,
  };
}

function card(design: Card['design'], value: number): Card {
  return { design, value };
}

describe('getCribbageHint', () => {
  it('returns null when no human player', () => {
    const state = makeState({
      players: [{ id: 0, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 }],
    });
    expect(getCribbageHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    expect(getCribbageHint(state)).toBeNull();
  });

  it('returns null when gameEndFlag is true', () => {
    const state = makeState({
      gameEndFlag: true,
      players: [
        { id: 0, isHuman: true, cardCount: 1, cards: [card('HEART', 5)], roundScore: 0, cumulativeScore: 0 },
        { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getCribbageHint(state)).toBeNull();
  });

  it('returns null when not human turn', () => {
    const state = makeState({
      currentPlayerIdx: 1,
      players: [
        { id: 0, isHuman: true, cardCount: 6, cards: [card('HEART', 5)], roundScore: 0, cumulativeScore: 0 },
        { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getCribbageHint(state)).toBeNull();
  });

  it('returns null for show phase', () => {
    const state = makeState({
      phase: CribbagePhase.SHOW,
      players: [
        { id: 0, isHuman: true, cardCount: 4, cards: [card('HEART', 5)], roundScore: 0, cumulativeScore: 0 },
        { id: 1, isHuman: false, cardCount: 4, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getCribbageHint(state)).toBeNull();
  });

  it('returns null for round end phase', () => {
    const state = makeState({
      phase: CribbagePhase.ROUND_END,
      players: [
        { id: 0, isHuman: true, cardCount: 1, cards: [card('HEART', 5)], roundScore: 0, cumulativeScore: 0 },
        { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getCribbageHint(state)).toBeNull();
  });

  describe('discard phase', () => {
    it('suggests keeping fives and tens as non-dealer', () => {
      const state = makeState({
        phase: CribbagePhase.DISCARD,
        currentPlayerIdx: 0,
        dealerIdx: 1, // human is NOT dealer
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 6,
            cards: [
              card('HEART', 5),
              card('SPADE', 10),
              card('CLOVER', 3),
              card('DIAMOND', 7),
              card('HEART', 9),
              card('SPADE', 2),
            ],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint).not.toBeNull();
      expect(hint?.targetAction).toBe('discard');
      expect(hint?.reason).toBe('hint.keepFivesAndTens');
    });

    it('suggests discarding to crib as dealer with fives/tens', () => {
      const state = makeState({
        phase: CribbagePhase.DISCARD,
        currentPlayerIdx: 0,
        dealerIdx: 0, // human IS dealer
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 6,
            cards: [
              card('HEART', 5),
              card('SPADE', 10),
              card('CLOVER', 3),
              card('DIAMOND', 7),
              card('HEART', 9),
              card('SPADE', 2),
            ],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.reason).toBe('hint.discardToCribDealer');
    });

    it('suggests keep best hand when no fives or tens (dealer)', () => {
      const state = makeState({
        phase: CribbagePhase.DISCARD,
        currentPlayerIdx: 0,
        dealerIdx: 0,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 6,
            cards: [
              card('HEART', 3),
              card('SPADE', 4),
              card('CLOVER', 6),
              card('DIAMOND', 7),
              card('HEART', 8),
              card('SPADE', 9),
            ],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.reason).toBe('hint.keepBestHand');
    });

    it('suggests keep best hand when no fives or tens (non-dealer)', () => {
      const state = makeState({
        phase: CribbagePhase.DISCARD,
        currentPlayerIdx: 0,
        dealerIdx: 1,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 6,
            cards: [
              card('HEART', 3),
              card('SPADE', 4),
              card('CLOVER', 6),
              card('DIAMOND', 7),
              card('HEART', 8),
              card('SPADE', 9),
            ],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.reason).toBe('hint.keepBestHand');
    });

    it('suggests discard any when hand is already 4 cards', () => {
      const state = makeState({
        phase: CribbagePhase.DISCARD,
        currentPlayerIdx: 0,
        dealerIdx: 1,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 4,
            cards: [card('HEART', 3), card('SPADE', 5), card('CLOVER', 7), card('DIAMOND', 9)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 4, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.reason).toBe('hint.discardAny');
    });
  });

  describe('pegging phase', () => {
    it('suggests go when no card can be played', () => {
      const state = makeState({
        phase: CribbagePhase.PEGGING,
        currentPlayerIdx: 0,
        pegCount: 28,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('HEART', 5), card('SPADE', 8)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.targetAction).toBe('go');
      expect(hint?.reason).toBe('hint.mustGo');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests hitting fifteen', () => {
      const state = makeState({
        phase: CribbagePhase.PEGGING,
        currentPlayerIdx: 0,
        pegCount: 10,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('HEART', 5), card('SPADE', 3)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.targetAction).toBe('peg');
      expect(hint?.reason).toBe('hint.pegFifteen');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests hitting thirty-one', () => {
      const state = makeState({
        phase: CribbagePhase.PEGGING,
        currentPlayerIdx: 0,
        pegCount: 24,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('HEART', 7), card('SPADE', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.targetAction).toBe('peg');
      expect(hint?.reason).toBe('hint.pegThirtyOne');
    });

    it('suggests safe peg when avoiding dangerous counts', () => {
      const state = makeState({
        phase: CribbagePhase.PEGGING,
        currentPlayerIdx: 0,
        pegCount: 0,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('HEART', 3), card('SPADE', 7)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.targetAction).toBe('peg');
      expect(hint?.reason).toBe('hint.pegSafe');
    });

    it('suggests peg play when all moves lead to dangerous counts', () => {
      const state = makeState({
        phase: CribbagePhase.PEGGING,
        currentPlayerIdx: 0,
        pegCount: 0,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 1,
            cards: [card('HEART', 5)], // count becomes 5 (dangerous)
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.targetAction).toBe('peg');
      expect(hint?.reason).toBe('hint.pegPlay');
    });

    it('suggests go when peg count is 31', () => {
      const state = makeState({
        phase: CribbagePhase.PEGGING,
        currentPlayerIdx: 0,
        pegCount: 31,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 1,
            cards: [card('HEART', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });
      const hint = getCribbageHint(state);
      expect(hint?.targetAction).toBe('go');
      expect(hint?.reason).toBe('hint.mustGo');
    });
  });
});
