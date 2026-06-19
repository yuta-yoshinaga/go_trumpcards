import { describe, expect, it } from 'vitest';
import type { KempsPlayer, KempsResponse } from '../../../types/card';
import { formatKempsState } from './kempsFormatter';

function makePlayer(overrides: Partial<KempsPlayer> = {}): KempsPlayer {
  return {
    name: 'CPU',
    isHuman: false,
    team: 1,
    handSize: 4,
    hand: [],
    hasFourOfAKind: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<KempsResponse> = {}): KempsResponse {
  return {
    phase: 0,
    gameEndFlag: false,
    winnerTeam: -1,
    currentPlayerIdx: 0,
    isHumanTurn: true,
    teamScores: [1, 2],
    field: [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 6 },
    ],
    signalType: 0,
    partnerSignaling: false,
    opponentSignaling: false,
    fourHolderIdx: -1,
    roundResult: 0,
    roundWinnerTeam: -1,
    roundNumber: 1,
    cpuDifficulty: 1,
    targetScore: 5,
    message: '',
    players: [
      makePlayer({
        name: 'You',
        isHuman: true,
        team: 0,
        hand: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 1 },
        ],
      }),
      makePlayer({ team: 1 }),
      makePlayer({ team: 0 }),
      makePlayer({ team: 1 }),
    ],
    ...overrides,
  };
}

describe('formatKempsState', () => {
  it('includes the header, round, phase, scores and signal', () => {
    const out = formatKempsState(makeState());
    expect(out).toContain('Kemps');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: Exchange');
    expect(out).toContain('Team A: 1');
    expect(out).toContain('Team B: 2');
    expect(out).toContain('your signal: Sound');
  });

  it('renders the field and the human hand with indices', () => {
    const out = formatKempsState(makeState());
    expect(out).toContain('field:');
    expect(out).toContain('[0]');
    expect(out).toContain('[1]');
  });

  it('shows the swap prompt on the human turn during exchange', () => {
    const out = formatKempsState(makeState());
    expect(out).toContain('swap a card');
  });

  it('shows the declare prompts and partner signal cue', () => {
    const out = formatKempsState(makeState({ phase: 1, isHumanTurn: false, partnerSignaling: true, fourHolderIdx: 2 }));
    expect(out).toContain('declare window');
    expect(out).toContain('partner is signaling');
  });

  it('shows the opponent signal cue', () => {
    const out = formatKempsState(
      makeState({ phase: 1, isHumanTurn: false, opponentSignaling: true, fourHolderIdx: 1 }),
    );
    expect(out).toContain('opponent may be signaling');
  });

  it('announces the winner on game end', () => {
    const out = formatKempsState(makeState({ phase: 3, gameEndFlag: true, winnerTeam: 0 }));
    expect(out).toContain('Game Over');
    expect(out).toContain('Team A');
  });
});
