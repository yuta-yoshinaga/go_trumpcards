import { describe, expect, it } from 'vitest';
import type { Card, PanPlayer, PanResponse } from '../../types/card';
import { PanPhase } from '../../types/phases';
import { getPanHint, hasMeld, panRankIndex } from './panHint';

function card(design: Card['design'], value: number): Card {
  return { design, value };
}

function player(overrides: Partial<PanPlayer> = {}): PanPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 10,
    cards: [],
    laidMelds: [],
    meldedCount: 0,
    chips: 50,
    handPoints: 0,
    roundScore: 0,
    cumulativeScore: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<PanResponse> = {}): PanResponse {
  return {
    players: [player(), player({ id: 1, isHuman: false })],
    phase: PanPhase.DRAW,
    roundNumber: 1,
    targetRounds: 3,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    discardTop: null,
    drawPileCount: 250,
    deckSize: 320,
    winMeldCount: 11,
    gameEndFlag: false,
    winnerIdx: -1,
    panDeclarerIdx: -1,
    message: '',
    config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
    ...overrides,
  };
}

describe('panRankIndex', () => {
  it('treats 7 and J as consecutive (no 8/9/10)', () => {
    expect(panRankIndex(11) - panRankIndex(7)).toBe(1);
  });

  it('returns -1 for a value not in the Pan deck', () => {
    expect(panRankIndex(9)).toBe(-1);
  });
});

describe('hasMeld', () => {
  it('detects a set of three matching ranks', () => {
    expect(hasMeld([card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)])).toBe(true);
  });

  it('detects a rope of three consecutive same-suit cards', () => {
    expect(hasMeld([card('SPADE', 5), card('SPADE', 6), card('SPADE', 7)])).toBe(true);
  });

  it('detects a rope spanning 7 to J (8/9/10 removed)', () => {
    expect(hasMeld([card('SPADE', 7), card('SPADE', 11), card('SPADE', 12)])).toBe(true);
  });

  it('returns false when no meld exists', () => {
    expect(hasMeld([card('SPADE', 2), card('HEART', 5), card('CLOVER', 12)])).toBe(false);
  });
});

describe('getPanHint', () => {
  it('returns null when there is no human player', () => {
    expect(getPanHint(makeState({ players: [player({ isHuman: false })] }))).toBeNull();
  });

  it('returns null when the human has no cards', () => {
    expect(getPanHint(makeState())).toBeNull();
  });

  it('returns null when the game is over', () => {
    const state = makeState({ gameEndFlag: true, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getPanHint(state)).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    const state = makeState({ currentPlayerIdx: 1, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getPanHint(state)).toBeNull();
  });

  it('returns null for the round-end phase', () => {
    const state = makeState({ phase: PanPhase.ROUND_END, players: [player({ cards: [card('HEART', 5)] })] });
    expect(getPanHint(state)).toBeNull();
  });

  describe('draw phase', () => {
    it('suggests drawing the discard when it forms a set with the hand', () => {
      const state = makeState({
        phase: PanPhase.DRAW,
        discardTop: card('HEART', 7),
        players: [player({ cards: [card('SPADE', 7), card('DIAMOND', 2)] })],
      });
      const hint = getPanHint(state);
      expect(hint?.targetAction).toBe('drawDiscard');
      expect(hint?.reason).toBe('hint.drawFromDiscard');
    });

    it('suggests drawing the discard when it extends a rope', () => {
      const state = makeState({
        phase: PanPhase.DRAW,
        discardTop: card('HEART', 5),
        players: [player({ cards: [card('HEART', 6), card('SPADE', 2)] })],
      });
      expect(getPanHint(state)?.targetAction).toBe('drawDiscard');
    });

    it('suggests drawing from stock when the discard does not fit', () => {
      const state = makeState({
        phase: PanPhase.DRAW,
        discardTop: card('HEART', 13),
        players: [player({ cards: [card('SPADE', 2), card('CLOVER', 4)] })],
      });
      const hint = getPanHint(state);
      expect(hint?.targetAction).toBe('drawStock');
      expect(hint?.reason).toBe('hint.drawFromStock');
    });

    it('suggests drawing from stock when there is no discard top', () => {
      const state = makeState({
        phase: PanPhase.DRAW,
        discardTop: null,
        players: [player({ cards: [card('SPADE', 2)] })],
      });
      expect(getPanHint(state)?.targetAction).toBe('drawStock');
    });
  });

  describe('play phase', () => {
    it('suggests melding when the hand contains a meld', () => {
      const state = makeState({
        phase: PanPhase.PLAY,
        players: [player({ cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] })],
      });
      const hint = getPanHint(state);
      expect(hint?.targetAction).toBe('meld');
      expect(hint?.reason).toBe('hint.meld');
    });

    it('suggests discarding when no meld is available', () => {
      const state = makeState({
        phase: PanPhase.PLAY,
        players: [player({ cards: [card('SPADE', 2), card('HEART', 5), card('CLOVER', 12)] })],
      });
      const hint = getPanHint(state);
      expect(hint?.targetAction).toBe('discard');
      expect(hint?.reason).toBe('hint.discard');
    });
  });
});
