import { describe, expect, it } from 'vitest';
import type { WarConfig, WarResponse } from '../../types/card';
import { getWarHint } from './warHint';

const defaultConfig: WarConfig = { maxRounds: 500 };

function makeState(overrides: Partial<WarResponse> = {}): WarResponse {
  return {
    players: [
      { id: 0, isHuman: true, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
      { id: 1, isHuman: false, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
    ],
    phase: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    playerRevealed: null,
    cpuRevealed: null,
    warPotSize: 0,
    lastWinnerIdx: -1,
    lastBurialCount: 0,
    roundsPlayed: 0,
    config: defaultConfig,
    message: '',
    ...overrides,
  };
}

describe('getWarHint', () => {
  it('returns null when game ended', () => {
    expect(getWarHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('suggests step when no cards revealed', () => {
    const hint = getWarHint(makeState());
    expect(hint).toEqual({ targetAction: 'step', reason: 'hint.flipCard', confidence: 'strong' });
  });

  it('suggests step when cards are revealed (resolve round)', () => {
    const hint = getWarHint(
      makeState({ playerRevealed: { design: 'HEART', value: 10 }, cpuRevealed: { design: 'SPADE', value: 5 } }),
    );
    expect(hint).toEqual({ targetAction: 'step', reason: 'hint.resolveRound', confidence: 'strong' });
  });

  it('suggests step during war pot', () => {
    const hint = getWarHint(makeState({ warPotSize: 6 }));
    expect(hint).toEqual({ targetAction: 'step', reason: 'hint.flipCard', confidence: 'strong' });
  });
});
