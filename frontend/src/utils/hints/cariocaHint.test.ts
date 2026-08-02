import { describe, expect, it } from 'vitest';
import type { Card, CariocaResponse } from '../../types/card';
import { getCariocaHint } from './cariocaHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const PHASE_DRAW = 0;
const PHASE_PLAY = 1;
const PHASE_ROUND_END = 2;

type Extra = { hand?: Card[]; contractMet?: boolean };

function base({
  hand = [card('SPADE', 4), card('HEART', 12)],
  contractMet = true,
  ...overrides
}: Partial<CariocaResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: hand.length,
        cards: hand,
        melds: [],
        contractMet,
        roundScore: 0,
        cumulativeScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 8,
        cards: [],
        melds: [],
        contractMet: false,
        roundScore: 0,
        cumulativeScore: 0,
      },
    ],
    phase: PHASE_PLAY,
    roundNumber: 1,
    totalRounds: 6,
    currentPlayerIdx: 0,
    discardTop: card('CLOVER', 9),
    drawPileCount: 40,
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinnerIdx: -1,
    contractSlots: [
      { kind: 0, size: 3 },
      { kind: 0, size: 3 },
    ],
    config: { playerCount: 3, cpuDifficulty: 1, failContractPenalty: 0 },
    ...overrides,
  } as CariocaResponse;
}

describe('getCariocaHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getCariocaHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getCariocaHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getCariocaHint(base({ phase: PHASE_ROUND_END }))).toBeNull();
  });

  it('takes a discard that pairs a card in hand', () => {
    const s = base({
      phase: PHASE_DRAW,
      hand: [card('SPADE', 9), card('HEART', 12)],
      discardTop: card('CLOVER', 9),
    });
    expect(getCariocaHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('takes a discard that neighbours a card of the same suit', () => {
    const s = base({
      phase: PHASE_DRAW,
      hand: [card('SPADE', 8), card('HEART', 12)],
      discardTop: card('SPADE', 9),
    });
    expect(getCariocaHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('does not take a neighbour of another suit', () => {
    const s = base({
      phase: PHASE_DRAW,
      hand: [card('SPADE', 8), card('HEART', 12)],
      discardTop: card('CLOVER', 9),
    });
    expect(getCariocaHint(s)?.targetAction).toBe('drawStock');
  });

  it('always takes a discarded joker, which fills any slot', () => {
    const s = base({ phase: PHASE_DRAW, hand: [card('SPADE', 8)], discardTop: card('JOKER', 0) });
    expect(getCariocaHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('does not let a joker pose as an ace when judging the discard', () => {
    // **ジョーカーの value は 1 と 2** (`TrumpCards.go:51`)。除かずに比べると
    // 手札のジョーカーが A の相方に見え、拾う理由が無いのに拾わせてしまう。
    const s = base({
      phase: PHASE_DRAW,
      hand: [card('JOKER', 1), card('HEART', 12)],
      discardTop: card('CLOVER', 1),
    });
    expect(getCariocaHint(s)?.targetAction).toBe('drawStock');
  });

  it('keeps the joker even when every other card is material', () => {
    // 全部が材料なら候補は手札全体に戻る。そこでジョーカーを外していないと、
    // 25 点で一番高くつくジョーカーが選ばれてしまう。
    const s = base({
      hand: [card('JOKER', 1), card('SPADE', 6), card('HEART', 6), card('DIAMOND', 6)],
    });
    expect(getCariocaHint(s)?.targetAction).not.toBe('card-0');
  });

  it('draws from the stock when the pile is empty', () => {
    expect(getCariocaHint(base({ phase: PHASE_DRAW, discardTop: null }))?.targetAction).toBe('drawStock');
  });

  it('aims at the contract while it is unmet', () => {
    expect(getCariocaHint(base({ contractMet: false }))?.targetAction).toBe('meld');
  });

  it('names the all-sets contract differently from a mixed one', () => {
    const sets = base({ contractMet: false, contractSlots: [{ kind: 0, size: 3 }] });
    expect(getCariocaHint(sets)?.reason).toBe('frontendHint.cariocaMeetSets');

    const mixed = base({
      contractMet: false,
      contractSlots: [
        { kind: 0, size: 3 },
        { kind: 1, size: 4 },
      ],
    });
    expect(getCariocaHint(mixed)?.reason).toBe('frontendHint.cariocaMeetContract');
  });

  it('discards by the penalty table, not by rank', () => {
    // A は 15 点で K の 10 点より高くつく。ランクで比べると K を選んでしまう。
    const s = base({ hand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 5), card('CLOVER', 6)] });
    expect(getCariocaHint(s)?.targetAction).toBe('card-0');
  });

  it('never suggests discarding the joker, the costliest card of all', () => {
    const s = base({ hand: [card('JOKER', 0), card('HEART', 13), card('DIAMOND', 4)] });
    expect(getCariocaHint(s)?.targetAction).toBe('card-1');
  });

  it('keeps a discard suggestion on card index 0', () => {
    const s = base({ hand: [card('HEART', 13), card('SPADE', 6), card('SPADE', 7)] });
    expect(getCariocaHint(s)?.targetAction).toBe('card-0');
  });

  it('falls back to the costliest card when everything is material', () => {
    const s = base({ hand: [card('SPADE', 6), card('HEART', 6), card('DIAMOND', 6)] });
    expect(getCariocaHint(s)?.targetAction).toBe('card-0');
  });

  it('stays quiet without a visible hand', () => {
    expect(getCariocaHint(base({ hand: [] }))).toBeNull();
  });
});
