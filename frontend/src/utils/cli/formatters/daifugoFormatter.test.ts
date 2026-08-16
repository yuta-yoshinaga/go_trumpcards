import { describe, expect, it } from 'vitest';
import type { DaifugoResponse } from '../../../types/card';
import { formatDaifugoState } from './daifugoFormatter';

function makeState(overrides?: Partial<DaifugoResponse>): DaifugoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        rank: 0,
        cardCount: 5,
        cards: [{ design: 'SPADE', value: 3 }],
      },
      { id: 1, isHuman: false, isFinished: false, rank: 0, cardCount: 5, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    revolutionActive: false,
    elevenBackActive: false,
    suitLocked: false,
    lockedSuit: '',
    tableIsSequence: false,
    config: {} as DaifugoResponse['config'],
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    message: '',
    pendingAction: 'none',
    pendingActionTarget: -1,
    reverseDirection: false,
    numberLocked: false,
    sequenceLocked: false,
    sortMode: 0,
    playableCardIndices: null,
    ...overrides,
  };
}

describe('formatDaifugoState', () => {
  it('formats basic state', () => {
    const output = formatDaifugoState(makeState());
    expect(output).toContain('Daifugo');
    expect(output).toContain('5 cards');
    expect(output).toContain('table: empty');
  });

  it('shows active rules', () => {
    const output = formatDaifugoState(makeState({ revolutionActive: true, elevenBackActive: true }));
    expect(output).toContain('Revolution');
    expect(output).toContain('11-Back');
  });

  it('shows table cards', () => {
    const output = formatDaifugoState(makeState({ tableCards: [{ design: 'HEART', value: 5 }], lastPlayPlayerIdx: 1 }));
    expect(output).toContain('table:');
    expect(output).not.toContain('empty');
  });

  it('shows game over', () => {
    const output = formatDaifugoState(makeState({ gameEndFlag: true }));
    expect(output).toContain('Game Over');
  });
});
