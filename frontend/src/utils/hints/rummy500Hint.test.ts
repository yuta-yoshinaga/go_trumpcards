import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, Rummy500Response } from '../../types/card';
import { Rummy500Phase } from '../../types/phases';
import { getRummy500Hint } from './rummy500Hint';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseState = (overrides: Partial<Rummy500Response> = {}): Rummy500Response => ({
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 0,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      laidMelds: [],
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 0,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      laidMelds: [],
    },
  ],
  layoffTargets: [],
  phase: Rummy500Phase.DRAW,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardPile: [],
  drawPileCount: 25,
  gameEndFlag: false,
  winnerIdx: -1,
  roundEnderIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 500 },
  ...overrides,
});

describe('getRummy500Hint', () => {
  it('returns null when game has ended', () => {
    expect(getRummy500Hint(baseState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when no human player', () => {
    const s = baseState();
    s.players[0].isHuman = false;
    expect(getRummy500Hint(s)).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getRummy500Hint(baseState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends discard top during Draw when pile non-empty', () => {
    const hint = getRummy500Hint(baseState({ discardPile: [card('SPADE', 7)] }));
    expect(hint?.reason).toBe('rummy500.hint.drawDiscardTop');
  });

  it('recommends stock during Draw when pile empty', () => {
    const hint = getRummy500Hint(baseState({ discardPile: [] }));
    expect(hint?.reason).toBe('rummy500.hint.drawStock');
  });

  it('recommends melding triples during Play', () => {
    const s = baseState({
      phase: Rummy500Phase.PLAY,
    });
    s.players[0].cards = [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7)];
    const hint = getRummy500Hint(s);
    expect(hint?.reason).toBe('rummy500.hint.meldSet');
  });

  it('recommends discarding high card when no triple available', () => {
    const s = baseState({
      phase: Rummy500Phase.PLAY,
    });
    s.players[0].cards = [card('SPADE', 5), card('HEART', 8), card('CLOVER', 11)];
    const hint = getRummy500Hint(s);
    expect(hint?.reason).toBe('rummy500.hint.discardHighCard');
  });

  it('returns null during RoundEnd', () => {
    expect(getRummy500Hint(baseState({ phase: Rummy500Phase.ROUND_END }))).toBeNull();
  });
});
