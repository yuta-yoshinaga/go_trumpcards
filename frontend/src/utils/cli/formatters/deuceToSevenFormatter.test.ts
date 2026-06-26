import { describe, expect, it } from 'vitest';
import type { DeuceToSevenResponse } from '../../../types/card';
import { formatDeuceToSevenState } from './deuceToSevenFormatter';

function makeState(overrides?: Partial<DeuceToSevenResponse>): DeuceToSevenResponse {
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
        handRank: 0,
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

describe('formatDeuceToSevenState', () => {
  it('renders header and phase', () => {
    const result = formatDeuceToSevenState(makeState());
    expect(result).toContain('2-7 Triple Draw');
    expect(result).toContain('phase: BET');
  });

  it('renders human cards with indices', () => {
    const result = formatDeuceToSevenState(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cards: [
              { design: 'SPADE', value: 7 },
              { design: 'HEART', value: 5 },
            ],
            chips: 1000,
            currentBet: 0,
            folded: false,
            allIn: false,
            handRank: 0,
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
    const result = formatDeuceToSevenState(
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
            handRank: 0,
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
    const result = formatDeuceToSevenState(
      makeState({
        phase: 4,
        roundResults: [{ playerIdx: 0, handRank: 0, handName: 'High Card', wonAmount: 100 }],
      }),
    );
    expect(result).toContain('results');
    expect(result).toContain('High Card');
    expect(result).toContain('100');
  });

  it('shows game over flag', () => {
    const result = formatDeuceToSevenState(makeState({ gameEndFlag: true }));
    expect(result).toContain('Game Over');
  });

  it('shows UNKNOWN for unexpected phase', () => {
    const result = formatDeuceToSevenState(makeState({ phase: 99 as 0 }));
    expect(result).toContain('UNKNOWN');
  });

  function drawHuman(cards: DeuceToSevenResponse['players'][number]['cards']) {
    return makeState({
      phase: 3,
      currentTurn: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          cards,
          chips: 1000,
          currentBet: 0,
          folded: false,
          allIn: false,
          handRank: 0,
          handName: '',
          drawCount: 0,
          totalDraws: 0,
          playStyleName: 'tight',
        },
      ],
    });
  }

  it('flags a made pat low and recommends standing in the draw phase', () => {
    const result = formatDeuceToSevenState(
      drawHuman([
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 5 },
        { design: 'CLOVER', value: 4 },
        { design: 'DIAMOND', value: 3 },
        { design: 'SPADE', value: 2 },
      ]),
    );
    expect(result).toContain('Made low:');
    expect(result).toContain('stand recommended');
  });

  it('lists the cards worth keeping when the low is not yet made', () => {
    const result = formatDeuceToSevenState(
      drawHuman([
        { design: 'SPADE', value: 9 },
        { design: 'HEART', value: 5 },
        { design: 'CLOVER', value: 4 },
        { design: 'DIAMOND', value: 3 },
        { design: 'SPADE', value: 2 },
      ]),
    );
    expect(result).toContain('Best to keep:');
    expect(result).toContain('draw the rest');
  });

  it('does not show stand-pat guidance outside the draw phase', () => {
    const result = formatDeuceToSevenState(
      makeState({
        phase: 2,
        currentTurn: 0,
        players: [
          {
            id: 0,
            isHuman: true,
            cards: [
              { design: 'SPADE', value: 7 },
              { design: 'HEART', value: 5 },
              { design: 'CLOVER', value: 4 },
              { design: 'DIAMOND', value: 3 },
              { design: 'SPADE', value: 2 },
            ],
            chips: 1000,
            currentBet: 0,
            folded: false,
            allIn: false,
            handRank: 0,
            handName: '',
            drawCount: 0,
            totalDraws: 0,
            playStyleName: 'tight',
          },
        ],
      }),
    );
    expect(result).not.toContain('Made low:');
    expect(result).not.toContain('Best to keep:');
  });
});
