import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { omahaLiveHandKey, omahaLivePreviewKey } from './livePokerPreview';

const c = (design: CardDesign, value: number): Card => ({ design, value });

describe('omahaLiveHandKey', () => {
  it('is null before the flop, when no five cards exist yet', () => {
    expect(omahaLiveHandKey([c('SPADE', 1), c('SPADE', 2), c('HEART', 3), c('HEART', 4)], [])).toBeNull();
  });

  it('is null with no hole cards', () => {
    expect(omahaLiveHandKey([], [c('SPADE', 5), c('SPADE', 6), c('SPADE', 7)])).toBeNull();
  });

  it('names the hand once the flop is out', () => {
    // Two hole spades plus three board spades is a flush under the must-use-2 rule.
    const hole = [c('SPADE', 2), c('SPADE', 4), c('HEART', 9), c('DIAMOND', 11)];
    const board = [c('SPADE', 6), c('SPADE', 8), c('SPADE', 10)];
    expect(omahaLiveHandKey(hole, board)).toBe('flush');
  });

  it('obeys the must-use-exactly-two rule rather than picking any five', () => {
    // Four board spades would be a flush in Hold'em, but Omaha forces exactly two
    // hole cards, and the only spade in hand is the 2 — so no flush is available.
    const hole = [c('SPADE', 2), c('HEART', 5), c('HEART', 7), c('DIAMOND', 9)];
    const board = [c('SPADE', 6), c('SPADE', 8), c('SPADE', 10), c('SPADE', 12), c('CLOVER', 3)];
    expect(omahaLiveHandKey(hole, board)).not.toBe('flush');
  });
});

describe('omahaLivePreviewKey', () => {
  const hole = [c('SPADE', 2), c('SPADE', 4), c('HEART', 9), c('DIAMOND', 11)];
  const board = [c('SPADE', 6), c('SPADE', 8), c('SPADE', 10)];
  const live = { isActive: true, isShowdown: false };

  it('previews a live, unfolded hand', () => {
    expect(omahaLivePreviewKey({ cards: hole }, board, live)).toBe('flush');
  });

  it('is silent before the hand is live', () => {
    expect(omahaLivePreviewKey({ cards: hole }, board, { isActive: false, isShowdown: false })).toBeNull();
  });

  it('is silent at showdown, where the final hand is already highlighted', () => {
    expect(omahaLivePreviewKey({ cards: hole }, board, { isActive: true, isShowdown: true })).toBeNull();
  });

  it('is silent for a folded player', () => {
    expect(omahaLivePreviewKey({ cards: hole, folded: true }, board, live)).toBeNull();
  });

  it('is silent with no seated player', () => {
    expect(omahaLivePreviewKey(null, board, live)).toBeNull();
    expect(omahaLivePreviewKey(undefined, board, live)).toBeNull();
  });

  it('treats a player with no cards as having nothing to preview', () => {
    expect(omahaLivePreviewKey({}, board, live)).toBeNull();
  });
});
