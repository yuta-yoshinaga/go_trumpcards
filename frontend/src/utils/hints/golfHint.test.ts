import { describe, expect, it } from 'vitest';
import type { Card, GolfCard, GolfResponse } from '../../types/card';
import { GolfPhase } from '../../types/phases';
import { getGolfHint } from './golfHint';

const card = (value: number): Card => ({ design: 'SPADE', value });
const exposed = (value: number): GolfCard => ({ card: card(value), removed: false, exposed: true });
const buried = (value: number): GolfCard => ({ card: card(value), removed: false, exposed: false });

const baseState = (overrides: Partial<GolfResponse> = {}): GolfResponse => ({
  layout: [[buried(9), exposed(5)], [exposed(11)]],
  stockCount: 10,
  waste: [card(2), card(7)],
  phase: GolfPhase.PLAYING,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  ...overrides,
});

describe('getGolfHint', () => {
  it('recommends removing when one exposed card is adjacent to the waste top', () => {
    // Waste top 7; col 1's exposed 6 is adjacent.
    const hint = getGolfHint(baseState({ layout: [[buried(9), exposed(5)], [exposed(6)]] }));
    expect(hint).toEqual({ targetAction: 'remove', reason: 'frontendHint.canRemove', confidence: 'strong' });
  });

  it('points out chain opportunities when several columns are removable', () => {
    // Waste top 7; exposed 6 and exposed 8 are both adjacent.
    const hint = getGolfHint(baseState({ layout: [[exposed(8)], [exposed(6)]] }));
    expect(hint).toEqual({ targetAction: 'remove', reason: 'frontendHint.multipleRemovable', confidence: 'strong' });
  });

  it('treats K and A as adjacent (wraparound)', () => {
    const hint = getGolfHint(baseState({ waste: [card(13)], layout: [[exposed(1)]] }));
    expect(hint?.targetAction).toBe('remove');
  });

  it('ignores buried and removed cards when scanning for removable columns', () => {
    const removable: GolfCard = { card: card(6), removed: true, exposed: true };
    const hint = getGolfHint(baseState({ layout: [[buried(6)], [removable]] }));
    expect(hint).toEqual({ targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' });
  });

  it('recommends drawing when nothing is removable and stock remains', () => {
    const hint = getGolfHint(baseState());
    expect(hint).toEqual({ targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' });
  });

  it('returns null when nothing is removable and the stock is empty', () => {
    expect(getGolfHint(baseState({ stockCount: 0 }))).toBeNull();
  });

  it('recommends drawing when the waste is empty but stock remains', () => {
    expect(getGolfHint(baseState({ waste: [] }))?.targetAction).toBe('draw');
  });

  it('returns null outside the playing phase', () => {
    expect(getGolfHint(baseState({ phase: GolfPhase.GAME_CLEAR }))).toBeNull();
    expect(getGolfHint(baseState({ phase: GolfPhase.GAME_OVER }))).toBeNull();
  });
});
