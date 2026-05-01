import { describe, expect, it } from 'vitest';
import type { BadugiResponse } from '../../../types/card';
import { formatBadugiState } from './badugiFormatter';

function makeState(overrides?: Partial<BadugiResponse>): BadugiResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cards: [],
        chips: 1000,
        currentBet: 0,
        folded: false,
        allIn: false,
        handSize: 0,
        handName: '',
        drawCount: 0,
        totalDraws: 0,
        playStyleName: 'tight',
      },
    ],
    pot: 0,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: 2,
    drawIndex: 0,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 10,
    ante: 0,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    cpuExchanges: [],
    message: '',
    ...overrides,
  };
}

describe('formatBadugiState', () => {
  it('renders header and phase', () => {
    const result = formatBadugiState(makeState());
    expect(result).toContain('Badugi');
    expect(result).toContain('phase: BET');
  });

  it('renders human cards with indices', () => {
    const result = formatBadugiState(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cards: [
              { design: 'SPADE', value: 1 },
              { design: 'HEART', value: 5 },
            ],
            chips: 1000,
            currentBet: 0,
            folded: false,
            allIn: false,
            handSize: 0,
            handName: '',
            drawCount: 0,
            totalDraws: 0,
            playStyleName: 'tight',
          },
        ],
      }),
    );
    expect(result).toContain('[0]');
    expect(result).toContain('[1]');
  });

  it('shows folded/allin status', () => {
    const result = formatBadugiState(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cards: [],
            chips: 0,
            currentBet: 0,
            folded: true,
            allIn: false,
            handSize: 0,
            handName: '',
            drawCount: 0,
            totalDraws: 0,
            playStyleName: 'tight',
          },
        ],
      }),
    );
    expect(result).toContain('FOLD');
  });

  it('shows round results in showdown', () => {
    const result = formatBadugiState(
      makeState({
        phase: 4,
        roundResults: [{ playerIdx: 0, handSize: 4, handName: 'Badugi', wonAmount: 100 }],
      }),
    );
    expect(result).toContain('results');
    expect(result).toContain('Badugi');
    expect(result).toContain('100');
  });

  it('shows game over flag', () => {
    const result = formatBadugiState(makeState({ gameEndFlag: true }));
    expect(result).toContain('Game Over');
  });

  it('shows UNKNOWN for unexpected phase', () => {
    const result = formatBadugiState(makeState({ phase: 99 as 0 }));
    expect(result).toContain('UNKNOWN');
  });
});
