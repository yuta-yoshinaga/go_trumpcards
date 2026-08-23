import { describe, expect, it } from 'vitest';
import type { BauernschnapsenResponse, Card } from '../../types/card';
import { BauernschnapsenPhase } from '../../types/phases';
import { getBauernschnapsenHint } from './bauernschnapsenHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<BauernschnapsenResponse> = {}): BauernschnapsenResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 1), card('HEART', 10), card('CLOVER', 11)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: BauernschnapsenPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    contract: 1,
    declarerIdx: 0,
    validPlayIndices: [0, 1, 2, 3, 4],
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundMarriage: [0, 0],
    marriageIndices: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 24 },
    ...overrides,
  };
}

describe('getBauernschnapsenHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getBauernschnapsenHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getBauernschnapsenHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getBauernschnapsenHint(makeState({ phase: BauernschnapsenPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getBauernschnapsenHint(makeState({ phase: BauernschnapsenPhase.ROUND_END }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getBauernschnapsenHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  describe('marriage', () => {
    it('suggests marriage when leading with a declarable index', () => {
      const state = makeState({ currentTrick: [], marriageIndices: [0] });
      const hint = getBauernschnapsenHint(state);
      expect(hint?.targetAction).toBe('marriage');
      expect(hint?.reason).toBe('hint.marriage');
      expect(hint?.confidence).toBe('strong');
    });
  });

  describe('lead', () => {
    it('suggests lead trump when holding trump and trick empty', () => {
      const state = makeState({ trumpSuit: 1 });
      state.players[0].cards = [card('SPADE', 1), card('HEART', 11)];
      const hint = getBauernschnapsenHint(state);
      expect(hint?.reason).toBe('hint.leadTrump');
    });

    it('suggests lead value when no trump but a high-value card', () => {
      const state = makeState({ trumpSuit: 1 });
      state.players[0].cards = [card('HEART', 1), card('CLOVER', 11)];
      const hint = getBauernschnapsenHint(state);
      expect(hint?.reason).toBe('hint.leadValue');
    });

    it('suggests lead low when no trump and only low cards', () => {
      const state = makeState({ trumpSuit: 1 });
      state.players[0].cards = [card('HEART', 11), card('CLOVER', 11)];
      const hint = getBauernschnapsenHint(state);
      expect(hint?.reason).toBe('hint.leadLow');
    });
  });

  describe('follow', () => {
    it('suggests follow win when holding the led suit', () => {
      const state = makeState({
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('CLOVER', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 13), card('HEART', 10)];
      const hint = getBauernschnapsenHint(state);
      expect(hint?.reason).toBe('hint.followWin');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests cut when void of led suit but holding trump', () => {
      const state = makeState({
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('SPADE', 13), card('CLOVER', 11)];
      const hint = getBauernschnapsenHint(state);
      expect(hint?.reason).toBe('hint.followCut');
    });

    it('suggests dump when void of led suit and no trump', () => {
      const state = makeState({
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 11), card('DIAMOND', 11)];
      const hint = getBauernschnapsenHint(state);
      expect(hint?.reason).toBe('hint.followDump');
    });
  });
});

describe('getBauernschnapsenHint under Bettel', () => {
  // ベテルの宣言者は 1 トリックも取らない契約なので、「切り札でリード」も
  // 「トリックを取る」も契約を落とす助言になる。
  it('tells the Bettel declarer to duck instead of to win', () => {
    const hint = getBauernschnapsenHint(makeState({ contract: 3, declarerIdx: 0, trumpSuit: -1 }));
    expect(hint?.reason).toBe('hint.duck');
  });

  // **負のコントロール。** 同じ盤面でも通常契約なら通常の助言に戻る。
  it('keeps the normal advice under a trump contract', () => {
    const hint = getBauernschnapsenHint(makeState({ contract: 1, declarerIdx: 0 }));
    expect(hint?.reason).not.toBe('hint.duck');
  });

  // 宣言者でなければ (相手側やパートナー) 普通に取りに行ってよい。
  it('keeps the normal advice for a seat that did not declare the Bettel', () => {
    const hint = getBauernschnapsenHint(makeState({ contract: 3, declarerIdx: 2, trumpSuit: -1 }));
    expect(hint?.reason).not.toBe('hint.duck');
  });
});
