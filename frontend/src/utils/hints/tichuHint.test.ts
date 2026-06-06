import { describe, expect, it } from 'vitest';
import type { TichuResponse } from '../../types/card';
import { getTichuHint } from './tichuHint';

function baseState(overrides: Partial<TichuResponse> = {}): TichuResponse {
  return {
    players: [
      { id: 0, isHuman: true, isFinished: false, team: 0, rank: 0, declType: 0, cardCount: 14, cards: [] },
      { id: 1, isHuman: false, isFinished: false, team: 1, rank: 0, declType: 0, cardCount: 14, cards: [] },
      { id: 2, isHuman: false, isFinished: false, team: 0, rank: 0, declType: 0, cardCount: 14, cards: [] },
      { id: 3, isHuman: false, isFinished: false, team: 1, rank: 0, declType: 0, cardCount: 14, cards: [] },
    ],
    phase: 'play',
    currentTurn: 0,
    tableCards: [],
    tableCombo: '',
    lastPlayIdx: -1,
    startLeader: 0,
    finishOrder: [],
    scores: [0, 0],
    isOneTwo: false,
    bombCount: 0,
    gameEndFlag: false,
    config: { cpuDifficulty: 0 },
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  } as TichuResponse;
}

describe('getTichuHint', () => {
  it('returns null when game ended', () => {
    expect(getTichuHint(baseState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not the human turn', () => {
    expect(getTichuHint(baseState({ currentTurn: 1 }))).toBeNull();
  });

  it('returns null during declaration phase', () => {
    expect(getTichuHint(baseState({ phase: 'declare' }))).toBeNull();
  });

  it('suggests playing lowest when leading', () => {
    const hint = getTichuHint(baseState({ tableCards: [] }));
    expect(hint?.reason).toBe('hintReason.playLowest');
  });

  it('suggests passing with high confidence when teammate controls the table', () => {
    const hint = getTichuHint(
      baseState({
        tableCards: [{ design: 'SPADE', value: 9 }],
        lastPlayIdx: 2, // teammate (team 0)
      }),
    );
    expect(hint?.reason).toBe('hintReason.pass');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests passing with moderate confidence when an opponent controls the table', () => {
    const hint = getTichuHint(
      baseState({
        tableCards: [{ design: 'SPADE', value: 9 }],
        lastPlayIdx: 1, // opponent (team 1)
      }),
    );
    expect(hint?.reason).toBe('hintReason.pass');
    expect(hint?.confidence).toBe('moderate');
  });
});
