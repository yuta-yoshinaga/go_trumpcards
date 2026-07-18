import { describe, expect, it } from 'vitest';
import type { MemoryResponse } from '../../types/card';
import { MemoryPhase } from '../../types/phases';
import { getMemoryHint } from './memoryHint';

function makeState(overrides: Partial<MemoryResponse> = {}): MemoryResponse {
  return {
    players: [
      { id: 0, isHuman: true, pairCount: 0, pairs: [] },
      { id: 1, isHuman: false, pairCount: 0, pairs: [] },
    ],
    board: [
      { card: { design: 'SPADE', value: 1 }, faceUp: false, taken: false },
      { card: { design: 'HEART', value: 1 }, faceUp: false, taken: false },
    ],
    phase: MemoryPhase.FLIP1,
    currentPlayerIdx: 0,
    firstFlipPos: -1,
    secondFlipPos: -1,
    lastMatchResult: false,
    gameEndFlag: false,
    winnerIdx: -1,
    turnNumber: 1,
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  };
}

describe('getMemoryHint', () => {
  it('returns null when game has ended', () => {
    expect(getMemoryHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not human turn', () => {
    expect(getMemoryHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('suggests flipping any card in FLIP1 phase', () => {
    const hint = getMemoryHint(makeState());
    expect(hint?.targetAction).toBe('flip');
    expect(hint?.reason).toBe('frontendHint.flipAny');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests finding match in FLIP2 phase', () => {
    const hint = getMemoryHint(makeState({ phase: MemoryPhase.FLIP2 }));
    expect(hint?.targetAction).toBe('flip');
    expect(hint?.reason).toBe('frontendHint.findMatch');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null in RESULT phase', () => {
    expect(getMemoryHint(makeState({ phase: MemoryPhase.RESULT }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getMemoryHint(makeState({ phase: MemoryPhase.GAME_END }))).toBeNull();
  });
});
