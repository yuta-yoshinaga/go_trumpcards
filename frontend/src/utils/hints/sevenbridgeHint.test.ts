import { describe, expect, it } from 'vitest';
import type { Card, SevenBridgePlayerData, SevenBridgeResponse } from '../../types/card';
import { SevenBridgePhase } from '../../types/phases';
import { getSevenbridgeHint } from './sevenbridgeHint';

const c = (design: string, value: number): Card => ({ design, value }) as Card;

const humanPlayer = (cards: Card[]): SevenBridgePlayerData => ({
  id: 0,
  isHuman: true,
  cardCount: cards.length,
  cards,
  melds: [],
  roundScore: 0,
  cumulativeScore: 0,
});

const base = (phase: number, discardTop: Card | null, cards: Card[]): SevenBridgeResponse => ({
  players: [humanPlayer(cards), { ...humanPlayer([]), id: 1, isHuman: false }],
  phase,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop,
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100 },
});

describe('getSevenbridgeHint', () => {
  it('returns null when game has ended', () => {
    const state = base(SevenBridgePhase.DRAW, c('SPADE', 5), []);
    state.gameEndFlag = true;
    expect(getSevenbridgeHint(state)).toBeNull();
  });

  it('returns null for round end / game end phases', () => {
    const state = base(SevenBridgePhase.ROUND_END, c('SPADE', 5), []);
    expect(getSevenbridgeHint(state)).toBeNull();
  });

  it('returns null if no human player', () => {
    const state = base(SevenBridgePhase.DRAW, c('SPADE', 5), []);
    state.players = [];
    expect(getSevenbridgeHint(state)).toBeNull();
  });

  it('returns pon hint when 2+ same-rank cards are available', () => {
    const state = base(SevenBridgePhase.DRAW, c('HEART', 9), [c('SPADE', 9), c('CLOVER', 9), c('DIAMOND', 3)]);
    expect(getSevenbridgeHint(state)).toEqual({
      targetAction: 'pon',
      reason: 'frontendHint.sevenbridgePon',
      confidence: 'strong',
    });
  });

  it('returns chi hint when suited run can be formed', () => {
    const state = base(SevenBridgePhase.DRAW, c('SPADE', 5), [c('SPADE', 6), c('SPADE', 7)]);
    expect(getSevenbridgeHint(state)).toEqual({
      targetAction: 'chi',
      reason: 'frontendHint.sevenbridgeChi',
      confidence: 'moderate',
    });
  });

  it('returns null when draw phase has neither pon nor chi', () => {
    const state = base(SevenBridgePhase.DRAW, c('SPADE', 5), [c('HEART', 3), c('CLOVER', 10)]);
    expect(getSevenbridgeHint(state)).toBeNull();
  });

  it('returns null when no discard top in draw phase', () => {
    const state = base(SevenBridgePhase.DRAW, null, [c('SPADE', 9), c('CLOVER', 9)]);
    expect(getSevenbridgeHint(state)).toBeNull();
  });

  it('returns meld hint when 3+ of same rank held in play phase', () => {
    const state = base(SevenBridgePhase.PLAY, c('HEART', 2), [
      c('SPADE', 3),
      c('CLOVER', 3),
      c('HEART', 3),
      c('DIAMOND', 10),
    ]);
    expect(getSevenbridgeHint(state)).toEqual({
      targetAction: 'meld',
      reason: 'frontendHint.sevenbridgeMeld',
      confidence: 'strong',
    });
  });

  it('returns null in play phase when no 3-card set', () => {
    const state = base(SevenBridgePhase.PLAY, c('HEART', 2), [c('SPADE', 3), c('CLOVER', 4), c('HEART', 9)]);
    expect(getSevenbridgeHint(state)).toBeNull();
  });

  it('accepts chi with lower bounded run (top-2, top-1)', () => {
    const state = base(SevenBridgePhase.DRAW, c('SPADE', 5), [c('SPADE', 3), c('SPADE', 4)]);
    expect(getSevenbridgeHint(state)?.targetAction).toBe('chi');
  });

  it('accepts chi with upper bounded run (top+1, top+2)', () => {
    const state = base(SevenBridgePhase.DRAW, c('SPADE', 5), [c('SPADE', 6), c('SPADE', 7)]);
    expect(getSevenbridgeHint(state)?.targetAction).toBe('chi');
  });
});
