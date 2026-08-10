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
    config: { cpuDifficulty: 1, pairCount: 26 },
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

  // **記憶がある時だけ強く言う (#4775)。**位置を名指しできない一般論のまま
  // confidence を上げると、当てずっぽうを高信頼と称することになる。
  it('names the remembered square and raises confidence in FLIP2', () => {
    const hint = getMemoryHint(makeState({ phase: MemoryPhase.FLIP2 }), 6);
    expect(hint?.reason).toBe('frontendHint.knownMatch');
    expect(hint?.reasonParams).toEqual({ position: '7' });
    expect(hint?.confidence).toBe('strong');
  });

  it('falls back to the generic advice when nothing is remembered', () => {
    const hint = getMemoryHint(makeState({ phase: MemoryPhase.FLIP2 }), null);
    expect(hint?.reason).toBe('frontendHint.findMatch');
    expect(hint?.confidence).toBe('moderate');
  });

  // 1枚目をめくる前は、記憶があっても指す相手がいない。
  it('ignores a remembered match while still on the first flip', () => {
    const hint = getMemoryHint(makeState({ phase: MemoryPhase.FLIP1 }), 6);
    expect(hint?.reason).toBe('frontendHint.flipAny');
  });

  it('returns null in RESULT phase', () => {
    expect(getMemoryHint(makeState({ phase: MemoryPhase.RESULT }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getMemoryHint(makeState({ phase: MemoryPhase.GAME_END }))).toBeNull();
  });
});
