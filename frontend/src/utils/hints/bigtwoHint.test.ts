import { describe, expect, it } from 'vitest';
import type { BigTwoResponse } from '../../types/card';
import { getBigTwoHint } from './bigtwoHint';

function fixture(gameEnd: boolean): BigTwoResponse {
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

describe('getBigTwoHint', () => {
  it('returns null during play', () => {
    expect(getBigTwoHint(fixture(false))).toBeNull();
  });

  it('returns null after game end', () => {
    expect(getBigTwoHint(fixture(true))).toBeNull();
  });
});
