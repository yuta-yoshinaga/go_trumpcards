import { describe, expect, it } from 'vitest';
import type { Card, VideoPokerResponse } from '../../types/card';
import { VideoPokerPhase } from '../../types/phases';
import { getVideoPokerBaseHint } from './videoPokerBaseHint';

const noWild = () => false;
const deucesWild = (c: Card) => c.value === 2;
const jokerWild = (c: Card) => c.design === 'JOKER';

function makeCard(design: Card['design'], value: number): Card {
  return { design, value };
}

function makeState(overrides: Partial<VideoPokerResponse> = {}): VideoPokerResponse {
  return {
    hand: [
      makeCard('SPADE', 10),
      makeCard('HEART', 5),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 12),
    ],
    phase: VideoPokerPhase.DRAW,
    chips: 100,
    betAmount: 1,
    result: 0,
    payout: 0,
    handRank: 0,
    handName: '',
    heldIndices: [false, false, false, false, false],
    variantName: '',
    message: '',
    ...overrides,
  };
}

describe('getVideoPokerBaseHint', () => {
  it('returns null in BET phase', () => {
    expect(getVideoPokerBaseHint(makeState({ phase: VideoPokerPhase.BET }), noWild)).toBeNull();
  });

  it('returns null in RESULT phase', () => {
    expect(getVideoPokerBaseHint(makeState({ phase: VideoPokerPhase.RESULT }), noWild)).toBeNull();
  });

  it('returns null when hand is empty', () => {
    expect(getVideoPokerBaseHint(makeState({ hand: [] }), noWild)).toBeNull();
  });

  // Pair detection
  it('suggests holding a pair', () => {
    const hand = [
      makeCard('SPADE', 10),
      makeCard('HEART', 10),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 5),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdPair');
    expect(result?.targetAction).toBe('hold:0,1');
  });

  // Three of a kind
  it('suggests holding three of a kind', () => {
    const hand = [
      makeCard('SPADE', 7),
      makeCard('HEART', 7),
      makeCard('DIAMOND', 7),
      makeCard('CLOVER', 2),
      makeCard('SPADE', 5),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdTrips');
    expect(result?.targetAction).toBe('hold:0,1,2');
    expect(result?.confidence).toBe('strong');
  });

  // Four of a kind
  it('suggests holding four of a kind', () => {
    const hand = [
      makeCard('SPADE', 9),
      makeCard('HEART', 9),
      makeCard('DIAMOND', 9),
      makeCard('CLOVER', 9),
      makeCard('SPADE', 5),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdQuads');
    expect(result?.targetAction).toBe('hold:0,1,2,3');
  });

  // Flush draw
  it('suggests holding flush draw (4 same suit)', () => {
    const hand = [
      makeCard('HEART', 2),
      makeCard('HEART', 5),
      makeCard('HEART', 9),
      makeCard('HEART', 12),
      makeCard('SPADE', 3),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdFlushDraw');
    expect(result?.targetAction).toBe('hold:0,1,2,3');
  });

  // Straight draw
  it('suggests holding straight draw (4 sequential)', () => {
    const hand = [
      makeCard('SPADE', 5),
      makeCard('HEART', 6),
      makeCard('DIAMOND', 7),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 12),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdStraightDraw');
    expect(result?.targetAction).toBe('hold:0,1,2,3');
  });

  // High cards
  it('suggests holding high cards (J+)', () => {
    const hand = [
      makeCard('SPADE', 11),
      makeCard('HEART', 3),
      makeCard('DIAMOND', 4),
      makeCard('CLOVER', 6),
      makeCard('SPADE', 13),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdHighCards');
    expect(result?.targetAction).toBe('hold:0,4');
  });

  // Draw all
  it('suggests draw all when nothing to hold', () => {
    const hand = [
      makeCard('SPADE', 2),
      makeCard('HEART', 4),
      makeCard('DIAMOND', 6),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 10),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.drawAll');
    expect(result?.targetAction).toBe('draw-all');
  });

  // Wild card handling (Deuces Wild)
  it('holds wild cards (deuces) and reports holdWild', () => {
    const hand = [
      makeCard('SPADE', 2),
      makeCard('HEART', 5),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 10),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), deucesWild);
    expect(result?.reason).toBe('hint.holdWild');
    expect(result?.confidence).toBe('strong');
    expect(result?.targetAction).toContain('0');
  });

  it('holds wild cards with non-wild pairs', () => {
    const hand = [
      makeCard('SPADE', 2),
      makeCard('HEART', 7),
      makeCard('DIAMOND', 7),
      makeCard('CLOVER', 3),
      makeCard('SPADE', 10),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), deucesWild);
    expect(result?.reason).toBe('hint.holdWild');
    expect(result?.targetAction).toBe('hold:0,1,2');
  });

  // Joker wild
  it('holds joker wild cards', () => {
    const hand = [
      makeCard('JOKER', 0),
      makeCard('HEART', 5),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 10),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), jokerWild);
    expect(result?.reason).toBe('hint.holdWild');
    expect(result?.confidence).toBe('strong');
    expect(result?.targetAction).toContain('0');
  });
});
