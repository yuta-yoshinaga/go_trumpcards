import { describe, expect, it } from 'vitest';
import type { TienLenResponse } from '../../types/card';
import { getTienLenHint } from './tienlenHint';

function fixture(gameEnd: boolean): TienLenResponse {
  return {
    players: [],
    currentTurn: 0,
    tableCards: [],
    tablePlayType: 0,
    lastPlayPlayerIdx: -1,
    gameEndFlag: gameEnd,
    cpuActions: [],
    humanAction: null,
    config: {
      cpuDifficulty: 1,
    },
    message: '',
  };
}

describe('getTienLenHint', () => {
  it('returns null during play', () => {
    expect(getTienLenHint(fixture(false))).toBeNull();
  });

  it('returns null after game end', () => {
    expect(getTienLenHint(fixture(true))).toBeNull();
  });
});
