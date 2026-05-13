import { describe, expect, it } from 'vitest';
import type { PiquetResponse } from '../../../types/card';
import { PiquetDeclarationKind, PiquetExchangeTurn, PiquetPhase } from '../../../types/phases';
import { formatPiquetState } from './piquetFormatter';

function baseState(overrides: Partial<PiquetResponse> = {}): PiquetResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 12,
        cards: [],
        trickCount: 0,
        declScore: 0,
        trickScore: 0,
        bonusScore: 0,
        roundScore: 0,
        matchScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 12,
        cards: [],
        trickCount: 0,
        declScore: 0,
        trickScore: 0,
        bonusScore: 0,
        roundScore: 0,
        matchScore: 0,
      },
    ],
    phase: PiquetPhase.EXCHANGE,
    dealNumber: 1,
    dealsPerPartie: 6,
    elderIdx: 0,
    youngerIdx: 1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 0,
    tricksWon: [0, 0],
    exchangeTurn: PiquetExchangeTurn.ELDER,
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

describe('formatPiquetState', () => {
  it('renders the header and players', () => {
    const out = formatPiquetState(baseState());
    expect(out).toContain('Piquet');
    expect(out).toContain('deal: 1/6');
    expect(out).toContain('Elder');
    expect(out).toContain('Younger');
  });

  it('shows carte blanche when present', () => {
    const out = formatPiquetState(baseState({ carteBlanche: [true, false] }));
    expect(out).toContain('carte blanche');
  });

  it('shows exchange prompt for elder', () => {
    const out = formatPiquetState(baseState({ exchangeTurn: PiquetExchangeTurn.ELDER }));
    expect(out).toContain('Elder to exchange');
  });

  it('shows exchange prompt for younger', () => {
    const out = formatPiquetState(baseState({ exchangeTurn: PiquetExchangeTurn.YOUNGER }));
    expect(out).toContain('Younger to exchange');
  });

  it('shows declaration results when present', () => {
    const out = formatPiquetState(
      baseState({
        declResults: [{ kind: PiquetDeclarationKind.POINT, winner: 0, score: 5, scoredBy: 0 }],
      }),
    );
    expect(out).toContain('Point');
    expect(out).toContain('+5');
  });

  it('marks declaration tie when score=0', () => {
    const out = formatPiquetState(
      baseState({
        declResults: [{ kind: PiquetDeclarationKind.SEQUENCE, winner: -1, score: 0, scoredBy: -1 }],
      }),
    );
    expect(out).toContain('tied');
  });

  it('shows partie winner at game end', () => {
    const out = formatPiquetState(baseState({ phase: PiquetPhase.GAME_END, winnerIdx: 0 }));
    expect(out).toContain('Partie winner');
  });

  it('shows draw at game end with no winner', () => {
    const out = formatPiquetState(baseState({ phase: PiquetPhase.GAME_END, winnerIdx: -1 }));
    expect(out).toContain('draw');
  });
});
