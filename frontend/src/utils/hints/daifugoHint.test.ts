import { describe, expect, it } from 'vitest';
import type { Card, DaifugoConfig, DaifugoResponse } from '../../types/card';
import { getDaifugoHint } from './daifugoHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: DaifugoConfig = {
  jokerCount: 0,
  eightCutEnabled: false,
  suitLockMode: 0,
  elevenBackEnabled: false,
  sequenceEnabled: false,
  cardExchangeEnabled: false,
  blindExchangeEnabled: false,
  fiveSkipEnabled: false,
  fiveSkipCount: 0,
  sevenPassEnabled: false,
  tenDiscardEnabled: false,
  spadeThreeEnabled: false,
  capitalFallEnabled: false,
  nineReverseEnabled: false,
  coupDetatEnabled: false,
  numberLockEnabled: false,
  sandstormEnabled: false,
  emperorEnabled: false,
  sequenceRevolutionEnabled: false,
  sequenceLockEnabled: false,
  illegalFinishEnabled: false,
  queenBomberEnabled: false,
  cpuDifficulty: 0,
};

function makeState(overrides: Partial<DaifugoResponse> = {}): DaifugoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        rank: 0,
        cardCount: 5,
        cards: [card('HEART', 5), card('SPADE', 8), card('DIAMOND', 10), card('CLOVER', 3), card('HEART', 3)],
      },
      { id: 1, isHuman: false, isFinished: false, rank: 0, cardCount: 5, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    revolutionActive: false,
    elevenBackActive: false,
    suitLocked: false,
    lockedSuit: '',
    tableIsSequence: false,
    config: defaultConfig,
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    message: '',
    pendingAction: 'none',
    pendingActionTarget: 0,
    reverseDirection: false,
    numberLocked: false,
    sequenceLocked: false,
    sortMode: 0,
    playableCardIndices: null,
    ...overrides,
  };
}

describe('getDaifugoHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getDaifugoHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getDaifugoHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getDaifugoHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('suggests playing when table is empty (free turn)', () => {
    const result = getDaifugoHint(makeState({ tableCards: [] }));
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playLowest');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests revolution when 4 cards of same value', () => {
    const state = makeState({ tableCards: [] });
    state.players[0].cards = [
      card('HEART', 5),
      card('SPADE', 5),
      card('DIAMOND', 5),
      card('CLOVER', 5),
      card('HEART', 3),
    ];
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.revolutionChance');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests playing stronger card when table has single card', () => {
    const state = makeState({ tableCards: [card('HEART', 5)] });
    // Human has cards stronger than 5 (8, 10) in Daifugo ordering
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playStronger');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests passing when no stronger cards (normal order)', () => {
    const state = makeState({ tableCards: [card('HEART', 2)] });
    state.players[0].cards = [card('HEART', 3), card('SPADE', 5), card('DIAMOND', 7)];
    // 2 is the strongest card in Daifugo (3<4<...<K<A<2)
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('pass');
    expect(result?.reason).toBe('hint.passNoPlay');
    expect(result?.confidence).toBe('moderate');
  });

  it('uses correct Daifugo strength ordering (A > K, 2 > A)', () => {
    const state = makeState({ tableCards: [card('HEART', 13)] });
    state.players[0].cards = [card('HEART', 1)];
    // A (1) is stronger than K (13) in Daifugo
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playStronger');
  });

  it('checks multi-card play count requirement', () => {
    const state = makeState({ tableCards: [card('HEART', 5), card('SPADE', 5)] });
    // Table has pair of 5s; human needs pair of stronger value
    state.players[0].cards = [card('HEART', 8), card('DIAMOND', 3)];
    // Only one 8, can't follow a pair
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('pass');
    expect(result?.reason).toBe('hint.passNoPlay');
  });

  it('suggests play when enough cards of stronger value for multi-card', () => {
    const state = makeState({ tableCards: [card('HEART', 5), card('SPADE', 5)] });
    state.players[0].cards = [card('HEART', 8), card('SPADE', 8), card('DIAMOND', 3)];
    // Two 8s can follow a pair of 5s
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playStronger');
    expect(result?.confidence).toBe('strong');
  });

  it('reverses strength comparison during revolution', () => {
    const state = makeState({ tableCards: [card('HEART', 5)], revolutionActive: true });
    state.players[0].cards = [card('HEART', 3), card('SPADE', 4)];
    // During revolution, lower strength cards beat higher; 3 < 5 so 3 beats 5
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playStronger');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests passing during revolution when no weaker cards', () => {
    const state = makeState({ tableCards: [card('HEART', 3)], revolutionActive: true });
    state.players[0].cards = [card('HEART', 8), card('SPADE', 10)];
    // During revolution, need cards with lower strength than 3, but 8 and 10 are stronger in normal order
    const result = getDaifugoHint(state);
    expect(result?.targetAction).toBe('pass');
    expect(result?.reason).toBe('hint.passNoPlay');
    expect(result?.confidence).toBe('moderate');
  });

  it('returns null when human is finished', () => {
    const state = makeState();
    state.players[0].isFinished = true;
    expect(getDaifugoHint(state)).toBeNull();
  });

  it('returns null during pending action', () => {
    expect(getDaifugoHint(makeState({ pendingAction: 'sevenPass' }))).toBeNull();
  });
});
