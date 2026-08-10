import { describe, expect, it } from 'vitest';
import type { Card, PageOnePlayerData, PageOneResponse } from '../../types/card';
import { PageOnePhase } from '../../types/phases';
import { getPageOneHint, isPageOnePlayable } from './pageoneHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function buildState(overrides: Partial<PageOneResponse> & { humanCards: Card[] }): PageOneResponse {
  const human: PageOnePlayerData = {
    id: 0,
    isHuman: true,
    cardCount: overrides.humanCards.length,
    cards: overrides.humanCards,
    roundScore: 0,
    cumulativeScore: 0,
    hasDeclared: false,
  };
  const cpu: PageOnePlayerData = {
    id: 1,
    isHuman: false,
    cardCount: 5,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    hasDeclared: false,
  };
  return {
    players: [human, cpu],
    phase: PageOnePhase.PLAY,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('SPADE', 5),
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 200 },
    ...overrides,
  };
}

describe('getPageOneHint', () => {
  it('returns null when the game has ended', () => {
    const state = buildState({ humanCards: [card('SPADE', 3)], gameEndFlag: true });
    expect(getPageOneHint(state)).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    const state = buildState({ humanCards: [card('SPADE', 3)], currentPlayerIdx: 1 });
    expect(getPageOneHint(state)).toBeNull();
  });

  it('suggests declare during MUST_DECLARE on the human turn', () => {
    const state = buildState({ humanCards: [card('SPADE', 3)], phase: PageOnePhase.MUST_DECLARE });
    expect(getPageOneHint(state)).toEqual({
      targetAction: 'declare',
      reason: 'hint.declare',
      confidence: 'strong',
    });
  });

  it('returns null during MUST_DECLARE when it is not the human turn', () => {
    const state = buildState({
      humanCards: [card('SPADE', 3)],
      phase: PageOnePhase.MUST_DECLARE,
      currentPlayerIdx: 1,
    });
    expect(getPageOneHint(state)).toBeNull();
  });

  it('suggests draw when no playable card exists', () => {
    const state = buildState({ humanCards: [card('HEART', 2), card('DIAMOND', 9)] });
    expect(getPageOneHint(state)).toEqual({
      targetAction: 'draw',
      reason: 'hint.drawCard',
      confidence: 'strong',
    });
  });

  it('suggests play with strong confidence when a high-value card is playable', () => {
    const state = buildState({
      humanCards: [card('HEART', 2), card('SPADE', 12)],
    });
    expect(getPageOneHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.playHighValue',
      confidence: 'strong',
    });
  });

  it('suggests play with strong confidence when a 10 is playable', () => {
    const state = buildState({
      humanCards: [card('HEART', 2), card('SPADE', 10)],
    });
    expect(getPageOneHint(state)).toMatchObject({ targetAction: 'play', confidence: 'strong' });
  });

  it('suggests play with moderate confidence when only low-value cards are playable', () => {
    const state = buildState({
      humanCards: [card('SPADE', 3), card('HEART', 7)],
    });
    expect(getPageOneHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.playMatching',
      confidence: 'moderate',
    });
  });

  it('returns null in non-PLAY/MUST_DECLARE phases', () => {
    const state = buildState({ humanCards: [card('SPADE', 3)], phase: PageOnePhase.ROUND_END });
    expect(getPageOneHint(state)).toBeNull();
  });

  it('returns null when discardTop is null', () => {
    const state = buildState({ humanCards: [card('SPADE', 3)], discardTop: null });
    expect(getPageOneHint(state)).toBeNull();
  });

  it('returns null when the response has no human player', () => {
    const cpu: PageOnePlayerData = {
      id: 1,
      isHuman: false,
      cardCount: 5,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      hasDeclared: false,
    };
    const state: PageOneResponse = {
      players: [cpu],
      phase: PageOnePhase.PLAY,
      roundNumber: 1,
      currentPlayerIdx: 0,
      discardTop: card('SPADE', 5),
      drawPileCount: 30,
      gameEndFlag: false,
      winnerIdx: -1,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 200 },
    };
    expect(getPageOneHint(state)).toBeNull();
  });
});

describe('isPageOnePlayable', () => {
  const top = { design: 'SPADE', value: 7 } as Card;

  it('accepts a matching suit', () => {
    expect(isPageOnePlayable({ design: 'SPADE', value: 3 } as Card, top)).toBe(true);
  });

  it('accepts a matching rank', () => {
    expect(isPageOnePlayable({ design: 'HEART', value: 7 } as Card, top)).toBe(true);
  });

  it('rejects a card matching neither', () => {
    expect(isPageOnePlayable({ design: 'HEART', value: 3 } as Card, top)).toBe(false);
  });

  it('accepts anything onto an empty discard, as the domain does', () => {
    // PageOne.isValidPlay returns true with no top; Start() always seeds one,
    // so this only matters if the two implementations are compared.
    expect(isPageOnePlayable({ design: 'HEART', value: 3 } as Card, null)).toBe(true);
  });
});
