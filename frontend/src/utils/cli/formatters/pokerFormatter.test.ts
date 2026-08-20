import { describe, expect, it } from 'vitest';
import type { PokerResponse } from '../../../types/card';
import { formatPokerState } from './pokerFormatter';

function makeState(overrides?: Partial<PokerResponse>): PokerResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cards: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 13 },
        ],
        chips: 990,
        currentBet: 10,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        exchangeCount: -1,
        playStyleName: '',
      },
      {
        id: 1,
        isHuman: false,
        cards: [],
        chips: 990,
        currentBet: 10,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        exchangeCount: -1,
        playStyleName: 'Balanced',
      },
    ],
    pot: 20,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: 1,
    exchangeRead: false,
    gameEndFlag: false,
    lastBet: 10,
    minRaise: 20,
    ante: 10,
    jokerCount: 0,
    bettingLimit: 2,
    raiseCount: 0,
    maxBetAmount: 990,
    roundResults: [],
    cpuActions: [],
    cpuExchanges: [],
    isLowball: false,
    message: '',
    ...overrides,
  };
}

describe('formatPokerState', () => {
  it('formats basic state', () => {
    const output = formatPokerState(makeState());
    expect(output).toContain('Poker');
    expect(output).toContain('pot: 20');
    expect(output).toContain('chips=990');
  });

  it('shows lowball variant', () => {
    const output = formatPokerState(makeState({ isLowball: true }));
    expect(output).toContain('[2-7 Lowball]');
  });

  it('shows CPU actions', () => {
    const output = formatPokerState(makeState({ cpuActions: [{ playerIdx: 1, action: 2, amount: 10 }] }));
    expect(output).toContain('Call');
  });

  it('shows hand results at game end', () => {
    const output = formatPokerState(
      makeState({
        exchangeRead: false,
        gameEndFlag: true,
        roundResults: [{ playerIdx: 0, handName: 'Full House', wonAmount: 20, handRank: 6, kickers: '' }],
      }),
    );
    expect(output).toContain('Full House');
    expect(output).toContain('+20');
    expect(output).toContain('Game Over');
  });

  it('shows player cards indexed', () => {
    const output = formatPokerState(makeState());
    expect(output).toContain('[0]');
    expect(output).toContain('[1]');
  });
});
