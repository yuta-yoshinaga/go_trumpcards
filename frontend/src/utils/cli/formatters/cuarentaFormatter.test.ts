import { describe, expect, it } from 'vitest';
import type { Card, CuarentaPlayer, CuarentaResponse } from '../../../types/card';
import { formatCuarentaState } from './cuarentaFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<CuarentaPlayer> = {}): CuarentaPlayer {
  return {
    id: 1,
    team: 1,
    isHuman: false,
    cardCount: 5,
    cards: [],
    capturedCount: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<CuarentaResponse> = {}): CuarentaResponse {
  return {
    players: [
      makePlayer({ id: 0, team: 0, isHuman: true, cards: [card('SPADE', 5), card('HEART', 11)] }),
      makePlayer({ id: 1, team: 1 }),
      makePlayer({ id: 2, team: 0 }),
      makePlayer({ id: 3, team: 1 }),
    ],
    currentTurn: 0,
    tableCards: [card('CLOVER', 7)],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 0,
    teamScores: [12, 8],
    remainingDeck: 16,
    roundWinners: [],
    cpuActions: [],
    humanAction: null,
    lastRoundDetail: null,
    config: { targetScore: 40, cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

describe('formatCuarentaState', () => {
  it('includes the header, phase, stock and team scores', () => {
    const out = formatCuarentaState(makeState());
    expect(out).toContain('Cuarenta');
    expect(out).toContain('phase: Play');
    expect(out).toContain('stock: 16');
    expect(out).toContain('Team A: 12 / 40');
    expect(out).toContain('Team B: 8 / 40');
  });

  it('renders player team lines, the table and the turn prompt', () => {
    const out = formatCuarentaState(makeState());
    expect(out).toContain('captured 0');
    expect(out).toContain('table:');
    expect(out).toContain('your turn');
    expect(out).toContain('your hand');
  });

  it('shows an empty table as a dash', () => {
    const out = formatCuarentaState(makeState({ tableCards: [] }));
    expect(out).toContain('table: -');
  });

  it('renders capture results with caída/ronda/limpia tags', () => {
    const out = formatCuarentaState(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7), card('HEART', 7), card('SPADE', 7)],
          isCaida: true,
          isLimpia: true,
          rondaBonus: 1,
        },
      }),
    );
    expect(out).toContain('Caída!+2');
    expect(out).toContain('Ronda!+1');
    expect(out).toContain('Limpia!+1');
  });

  it('renders a laid-down card with no captures', () => {
    const out = formatCuarentaState(
      makeState({
        cpuActions: [
          {
            playerIdx: 1,
            playedCard: card('DIAMOND', 2),
            capturedCards: [],
            isCaida: false,
            isLimpia: false,
            rondaBonus: 0,
          },
        ],
      }),
    );
    expect(out).toContain('laid');
  });

  it('announces the winning team and scores on game end', () => {
    const out = formatCuarentaState(
      makeState({ phase: 2, gameEndFlag: true, currentTurn: -1, roundWinners: [0], teamScores: [40, 31] }),
    );
    expect(out).toContain('Game Over');
    expect(out).toContain('Team A: 40 pts');
    expect(out).toContain('Winner: Team A');
  });
});
