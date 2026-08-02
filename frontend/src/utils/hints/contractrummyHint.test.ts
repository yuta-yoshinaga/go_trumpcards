import { describe, expect, it } from 'vitest';
import type { Card, ContractRummyResponse } from '../../types/card';
import { getContractRummyHint } from './contractrummyHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; contractMet?: boolean };

function base({
  hand = [card('SPADE', 4), card('HEART', 12)],
  contractMet = true,
  ...overrides
}: Partial<ContractRummyResponse> & Extra = {}) {
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
        cardCount: 10,
        cards: [],
        melds: [],
        contractMet: false,
        roundScore: 0,
        cumulativeScore: 0,
      },
    ],
    phase: 1,
    roundNumber: 1,
    totalRounds: 7,
    currentPlayerIdx: 0,
    discardTop: card('CLOVER', 9),
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinnerIdx: -1,
    contractSlots: [
      { kind: 0, size: 3 },
      { kind: 0, size: 3 },
    ],
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  } as unknown as ContractRummyResponse;
}

describe('getContractRummyHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getContractRummyHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getContractRummyHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getContractRummyHint(base({ phase: 2 }))).toBeNull();
  });

  it('takes a discard that connects with the hand', () => {
    const hand = [card('SPADE', 9), card('HEART', 12)];
    const s = base({ phase: 0, hand, discardTop: card('CLOVER', 9) });
    expect(getContractRummyHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('draws from the stock when the discard connects with nothing', () => {
    const hand = [card('SPADE', 2), card('HEART', 5)];
    const s = base({ phase: 0, hand, discardTop: card('CLOVER', 9) });
    expect(getContractRummyHint(s)?.targetAction).toBe('drawStock');
  });

  it('draws from the stock when there is nothing to take', () => {
    expect(getContractRummyHint(base({ phase: 0, discardTop: null }))?.targetAction).toBe('drawStock');
  });

  // **契約を満たすまでは崩さない。**一度に全部そろえる必要がある。
  it('aims at the contract while it is unmet', () => {
    expect(getContractRummyHint(base({ contractMet: false }))).toEqual({
      targetAction: 'meld',
      reason: 'frontendHint.contractrummyMeetContract',
      confidence: 'moderate',
    });
  });

  // 契約スロットが無い局面では、未達でも捨て札の助言に落ちる。
  it('falls back to a discard when the round names no contract', () => {
    const s = base({ contractMet: false, contractSlots: [] });
    expect(getContractRummyHint(s)?.reason).toBe('frontendHint.contractrummyDiscardHeavy');
  });

  it('discards the heaviest loose card once the contract is met', () => {
    const hand = [card('SPADE', 6), card('SPADE', 7), card('HEART', 13), card('DIAMOND', 2)];
    expect(getContractRummyHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a discard suggestion on card index 0', () => {
    const hand = [card('HEART', 13), card('SPADE', 6), card('SPADE', 7)];
    expect(getContractRummyHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  it('falls back to the heaviest card when everything connects', () => {
    const hand = [card('SPADE', 6), card('SPADE', 7), card('SPADE', 8)];
    expect(getContractRummyHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  it('stays quiet without a visible hand', () => {
    expect(getContractRummyHint(base({ hand: [] }))).toBeNull();
  });
});
