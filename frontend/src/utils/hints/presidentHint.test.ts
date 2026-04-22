import { describe, expect, it } from 'vitest';
import type { PresidentResponse } from '../../types/card';
import { getPresidentHint } from './presidentHint';

function fixture(gameEnd: boolean): PresidentResponse {
  return {
    players: [],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: gameEnd,
    revolutionActive: false,
    config: {
      revolutionEnabled: true,
      cardExchangeEnabled: true,
      passFieldFlushEnabled: true,
      cpuDifficulty: 1,
    },
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    message: '',
  };
}

describe('getPresidentHint', () => {
  it('returns null during play', () => {
    expect(getPresidentHint(fixture(false))).toBeNull();
  });

  it('returns null after game end', () => {
    expect(getPresidentHint(fixture(true))).toBeNull();
  });
});
