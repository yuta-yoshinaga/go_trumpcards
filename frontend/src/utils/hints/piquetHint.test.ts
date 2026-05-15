import { describe, expect, it } from 'vitest';
import type { PiquetResponse } from '../../types/card';
import { PiquetDeclarationKind, PiquetExchangeTurn, PiquetPhase } from '../../types/phases';
import { getPiquetHint } from './piquetHint';

function base(overrides: Partial<PiquetResponse> = {}): PiquetResponse {
  return {
    players: [],
    phase: PiquetPhase.PLAY,
    dealNumber: 1,
    dealsPerPartie: 6,
    elderIdx: 0,
    youngerIdx: 1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 0,
    tricksWon: [0, 0],
    exchangeTurn: PiquetExchangeTurn.DONE,
    elderExchangedCnt: 0,
    youngerExchangedCnt: 0,
    elderTalon: [],
    youngerTalon: [],
    elderRevealedTalon: [],
    youngerRevealedTalon: [],
    carteBlanche: [false, false],
    declStage: PiquetDeclarationKind.POINT,
    declResults: [],
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, dealsPerPartie: 6 },
    ...overrides,
  };
}

describe('getPiquetHint', () => {
  it('returns null when no hint is present', () => {
    expect(getPiquetHint(base())).toBeNull();
  });

  it('returns play.card hint in play phase', () => {
    const result = getPiquetHint(base({ hint: { cardIndex: 3, reason: 'lowest' } }));
    expect(result).toEqual({ targetAction: 'play.card', reason: 'piquet.hint.play', confidence: 'moderate' });
  });

  it('returns play.exchange hint in exchange phase', () => {
    const result = getPiquetHint(
      base({ phase: PiquetPhase.EXCHANGE, hint: { discardIndices: [0, 1, 2], reason: 'lowest' } }),
    );
    expect(result).toEqual({
      targetAction: 'play.exchange',
      reason: 'piquet.hint.discard',
      confidence: 'moderate',
    });
  });

  it('returns null when hint is empty', () => {
    const result = getPiquetHint(base({ hint: { reason: 'none' } }));
    expect(result).toBeNull();
  });
});
