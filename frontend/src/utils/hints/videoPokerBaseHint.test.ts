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

  // Two pair — hold all 4 cards
  it('suggests holding two pair', () => {
    const hand = [
      makeCard('SPADE', 10),
      makeCard('HEART', 10),
      makeCard('DIAMOND', 5),
      makeCard('CLOVER', 5),
      makeCard('SPADE', 3),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdPair');
    expect(result?.targetAction).toBe('hold:0,1,2,3');
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

  // Made straight (5 sequential)
  it('suggests holding a made straight (5 cards)', () => {
    const hand = [
      makeCard('SPADE', 5),
      makeCard('HEART', 6),
      makeCard('DIAMOND', 7),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 9),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdStraightDraw');
    expect(result?.targetAction).toBe('hold:0,1,2,3,4');
  });

  // Ace-low wheel straight draw (A-2-3-4)
  it('suggests holding Ace-low wheel straight draw', () => {
    const hand = [
      makeCard('SPADE', 14),
      makeCard('HEART', 2),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 4),
      makeCard('SPADE', 10),
    ];
    const result = getVideoPokerBaseHint(makeState({ hand }), noWild);
    expect(result?.reason).toBe('hint.holdStraightDraw');
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
    expect(result?.reason).toBe('hint.holdWildAndPair');
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

  // **配当のつかない低いペアが、強いドローを潰していた (#4691)。**ペア判定が
  // 最初に無条件でヒットするため、4枚ロイヤル・4枚フラッシュが同居しても
  // 常に弱いペアが推奨されていた。
  //
  // 順序は Jacks or Better の標準戦略に合わせる:
  //   4枚ロイヤル > 4枚フラッシュ > 低いペア > 4枚ストレート
  // **issue は「3つのドローすべてを低ペアより優先」としているが、それは誤り。**
  // 標準戦略では低ペアは4枚ストレートより上に来る。
  describe('draw versus low pair ordering (Jacks or Better)', () => {
    const hint = (cards: Card[]) => getVideoPokerBaseHint(makeState({ hand: cards }), noWild);

    it('prefers a four-card royal draw over a non-paying pair', () => {
      // ♠10 J Q K + ♥10 → 10 のペアがあるが、4枚ロイヤルの方が遥かに強い。
      const h = hint([
        makeCard('SPADE', 10),
        makeCard('SPADE', 11),
        makeCard('SPADE', 12),
        makeCard('SPADE', 13),
        makeCard('HEART', 10),
      ]);
      expect(h?.reason).toBe('hint.holdRoyalDraw');
    });

    it('prefers a four-card flush over a non-paying pair', () => {
      // ♠3 5 8 K + ♥3 → 3 のペアより 4枚フラッシュ。
      const h = hint([
        makeCard('SPADE', 3),
        makeCard('SPADE', 5),
        makeCard('SPADE', 8),
        makeCard('SPADE', 13),
        makeCard('HEART', 3),
      ]);
      expect(h?.reason).toBe('hint.holdFlushDraw');
    });

    // **逆側。**低ペアは4枚ストレートより上。ここを一緒くたに「ドロー優先」と
    // すると、標準戦略から外れる方向に壊れる。
    it('keeps a low pair over a four-card straight', () => {
      // ♠4 5 6 7 + ♥4 → 4枚ストレートより 4 のペア。
      const h = hint([
        makeCard('SPADE', 4),
        makeCard('SPADE', 5),
        makeCard('HEART', 6),
        makeCard('CLOVER', 7),
        makeCard('HEART', 4),
      ]);
      expect(h?.reason).toBe('hint.holdPair');
    });

    // 配当のつくペア (J 以上) は据え置き。4枚フラッシュより上。
    it('keeps a paying high pair over a four-card flush', () => {
      const h = hint([
        makeCard('SPADE', 12),
        makeCard('SPADE', 5),
        makeCard('SPADE', 8),
        makeCard('SPADE', 3),
        makeCard('HEART', 12),
      ]);
      expect(h?.reason).toBe('hint.holdPair');
    });

    // 3カード以上は無条件で据え置き (ドローより強い)。
    it('keeps trips over any draw', () => {
      const h = hint([
        makeCard('SPADE', 4),
        makeCard('HEART', 4),
        makeCard('CLOVER', 4),
        makeCard('SPADE', 8),
        makeCard('SPADE', 9),
      ]);
      expect(h?.reason).toBe('hint.holdTrips');
    });
  });
});
