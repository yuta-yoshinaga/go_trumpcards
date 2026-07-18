import { describe, expect, it } from 'vitest';
import type { Card, SpeedResponse } from '../../types/card';
import { SpeedPhase } from '../../types/phases';
import { getSpeedHint, isSpeedPlayable } from './speedHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<SpeedResponse> = {}): SpeedResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 4,
        cards: [card('HEART', 5), card('SPADE', 8), card('DIAMOND', 3), card('CLOVER', 10)],
        drawPileSize: 10,
      },
      { id: 1, isHuman: false, cardCount: 4, cards: [], drawPileSize: 10 },
    ],
    centerPiles: [card('HEART', 6), card('SPADE', 9)],
    phase: SpeedPhase.PLAY,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { cpuDifficulty: 0, autoFlip: true },
    message: '',
    ...overrides,
  };
}

describe('getSpeedHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getSpeedHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getSpeedHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null in non-play phase', () => {
    expect(getSpeedHint(makeState({ phase: SpeedPhase.STUCK }))).toBeNull();
  });

  it('suggests playing when playable cards exist', () => {
    // center: HEART 6, SPADE 9 → can play 5 (6-1) or 8 (9-1) or 10 (9+1)
    const result = getSpeedHint(makeState());
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.hasPlayable');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests waiting when no playable cards', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 1), card('SPADE', 13)];
    state.centerPiles = [card('HEART', 5), card('SPADE', 9)];
    const result = getSpeedHint(state);
    expect(result?.targetAction).toBe('wait');
    expect(result?.reason).toBe('hint.noPlayable');
    expect(result?.confidence).toBe('moderate');
  });

  it('handles wrap-around (Ace-King adjacency)', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 1)];
    state.centerPiles = [card('HEART', 13), card('SPADE', 5)];
    // Ace (1) is adjacent to King (13) in Speed (wrap-around)
    const result = getSpeedHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.hasPlayable');
    expect(result?.confidence).toBe('strong');
  });
});

describe('isSpeedPlayable', () => {
  const piles = [card('HEART', 6), card('SPADE', 9)];

  it('returns true for a card one below a pile top', () => {
    expect(isSpeedPlayable(5, piles)).toBe(true); // 6 - 1
  });

  it('returns true for a card one above a pile top', () => {
    expect(isSpeedPlayable(10, piles)).toBe(true); // 9 + 1
  });

  it('returns false for a non-adjacent card', () => {
    expect(isSpeedPlayable(3, piles)).toBe(false);
  });

  it('handles King-Ace wrap-around', () => {
    expect(isSpeedPlayable(1, [card('HEART', 13), card('SPADE', 5)])).toBe(true); // A next to K
    expect(isSpeedPlayable(13, [card('HEART', 1), card('SPADE', 5)])).toBe(true); // K next to A
  });

  it('ignores null piles safely', () => {
    expect(isSpeedPlayable(5, [null as unknown as Card, card('SPADE', 6)])).toBe(true);
    expect(isSpeedPlayable(5, [null as unknown as Card])).toBe(false);
  });
});
