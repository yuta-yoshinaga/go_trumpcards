import { describe, expect, it } from 'vitest';
import type { BeggarMyNeighbourConfig, BeggarMyNeighbourResponse } from '../../types/card';
import { BeggarMyNeighbourPhase } from '../../types/phases';
import { getBeggarMyNeighbourHint } from './beggarmyneighbourHint';

const defaultConfig: BeggarMyNeighbourConfig = { maxRounds: 2000 };

function makeState(overrides: Partial<BeggarMyNeighbourResponse> = {}): BeggarMyNeighbourResponse {
  return {
    players: [
      { id: 0, isHuman: true, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
      { id: 1, isHuman: false, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
    ],
    phase: BeggarMyNeighbourPhase.PLAY,
    gameEndFlag: false,
    winnerIdx: -1,
    currentPlayerIdx: 0,
    penaltyOwnerIdx: -1,
    penaltyRemaining: 0,
    centralPileSize: 0,
    lastCardPlayed: null,
    roundsPlayed: 0,
    config: defaultConfig,
    message: '',
    ...overrides,
  };
}

describe('getBeggarMyNeighbourHint', () => {
  it('returns null when game ended', () => {
    expect(getBeggarMyNeighbourHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('suggests step to play card during Play phase', () => {
    const hint = getBeggarMyNeighbourHint(makeState({ phase: BeggarMyNeighbourPhase.PLAY }));
    expect(hint).toEqual({ targetAction: 'step', reason: 'hint.playCard', confidence: 'strong' });
  });

  it('suggests step to pay penalty during PayPenalty phase', () => {
    const hint = getBeggarMyNeighbourHint(
      makeState({ phase: BeggarMyNeighbourPhase.PAY_PENALTY, penaltyRemaining: 3 }),
    );
    expect(hint).toEqual({ targetAction: 'step', reason: 'hint.payPenalty', confidence: 'strong' });
  });

  it('suggests step to collect pile during Collect phase', () => {
    const hint = getBeggarMyNeighbourHint(makeState({ phase: BeggarMyNeighbourPhase.COLLECT, centralPileSize: 8 }));
    expect(hint).toEqual({ targetAction: 'step', reason: 'hint.collectPile', confidence: 'strong' });
  });

  it('returns playCard hint during GameEnd phase (via default branch)', () => {
    const hint = getBeggarMyNeighbourHint(makeState({ phase: BeggarMyNeighbourPhase.GAME_END, gameEndFlag: false }));
    expect(hint).toEqual({ targetAction: 'step', reason: 'hint.playCard', confidence: 'strong' });
  });
});
