import { describe, expect, it } from 'vitest';
import type { CariocaResponse } from '../../../types/card';
import { formatCariocaState } from './cariocaFormatter';

/** Minimal Carioca state; the page's own fixture lives in its test file. */
function makeCariocaState(overrides?: Partial<CariocaResponse>): CariocaResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [
          { design: 'SPADE', value: 5 },
          { design: 'HEART', value: 9 },
        ],
        melds: [],
        contractMet: false,
        roundScore: 0,
        cumulativeScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 11,
        cards: [],
        melds: [],
        contractMet: false,
        roundScore: 0,
        cumulativeScore: 0,
      },
    ],
    phase: 0,
    roundNumber: 1,
    totalRounds: 6,
    currentPlayerIdx: 0,
    discardTop: { design: 'CLOVER', value: 7 },
    drawPileCount: 40,
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinnerIdx: -1,
    contractSlots: [
      { kind: 0, size: 3 },
      { kind: 0, size: 3 },
    ],
    config: { playerCount: 3, cpuDifficulty: 1, failContractPenalty: 0 },
    message: '',
    ...overrides,
  } as CariocaResponse;
}

describe('formatCariocaState', () => {
  it('renders the header, round, phase and contract', () => {
    const out = formatCariocaState(
      makeCariocaState({
        roundNumber: 2,
        totalRounds: 6,
        contractSlots: [
          { kind: 0, size: 3 },
          { kind: 1, size: 4 },
        ],
      }),
    );
    expect(out).toContain('Carioca');
    expect(out).toContain('round: 2/6');
    expect(out).toContain('phase: DRAW');
    expect(out).toContain('contract: trio x3 + run x4');
  });

  it('shows the empty discard placeholder when there is no top card', () => {
    expect(formatCariocaState(makeCariocaState({ discardTop: null }))).toContain('[  ]');
  });

  it('marks a player who has met the contract', () => {
    const base = makeCariocaState();
    const out = formatCariocaState({
      ...base,
      players: [{ ...base.players[0], contractMet: true }, ...base.players.slice(1)],
    });
    expect(out).toContain('[contract met]');
  });

  it('numbers the table melds so layoff can name them', () => {
    const base = makeCariocaState();
    const out = formatCariocaState({
      ...base,
      players: [
        {
          ...base.players[0],
          melds: [
            {
              cards: [
                { design: 'SPADE', value: 5 },
                { design: 'HEART', value: 5 },
                { design: 'CLOVER', value: 5 },
              ],
            },
          ],
        },
        ...base.players.slice(1),
      ],
    });
    expect(out).toContain('[0] ');
  });

  it('announces the winner once the game ends', () => {
    const out = formatCariocaState(makeCariocaState({ gameEndFlag: true, winnerIdx: 0 }));
    expect(out).toContain('Game Over!');
    // 終了後は手番行を出さない。
    expect(out).not.toContain('turn:');
  });
});
