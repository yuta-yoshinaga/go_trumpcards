import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { labelleLucieHasLegalMove, labelleLucieMovableFans } from './labelleLucieLegalMove';

const c = (design: string, value: number): Card => ({ design, value }) as Card;

describe('labelleLucieHasLegalMove', () => {
  it('returns true when a fan top can go to an empty foundation (Ace)', () => {
    const fans = [[c('SPADE', 5), c('SPADE', 1)]]; // top is ♠A
    const foundation = [[], [], [], []];
    expect(labelleLucieHasLegalMove(fans, foundation)).toBe(true);
  });

  it('returns true when a fan top builds up a foundation', () => {
    const fans = [[c('HEART', 6)]]; // top ♥6
    const foundation = [[c('HEART', 5)], [], [], []]; // ♥5 on top -> ♥6 fits
    expect(labelleLucieHasLegalMove(fans, foundation)).toBe(true);
  });

  it('returns true when a fan top can stack on another fan (same suit, one lower)', () => {
    const fans = [[c('CLOVER', 7)], [c('CLOVER', 8)]]; // ♣7 onto ♣8
    const foundation = [[], [], [], []];
    expect(labelleLucieHasLegalMove(fans, foundation)).toBe(true);
  });

  it('returns false when no fan top has any move', () => {
    const fans = [[c('SPADE', 5)], [c('HEART', 9)], [c('CLOVER', 2)]];
    const foundation = [[], [], [], []]; // no Aces, no builds, no same-suit stacks
    expect(labelleLucieHasLegalMove(fans, foundation)).toBe(false);
  });

  it('ignores empty fans', () => {
    const fans = [[], [c('DIAMOND', 9)]];
    const foundation = [[], [], [], []];
    expect(labelleLucieHasLegalMove(fans, foundation)).toBe(false);
  });
});

describe('labelleLucieMovableFans', () => {
  // #5678: 盤面全体に手があるかではなく、**どの扇が動かせるか**。ヒントを押さ
  // ないと分からなかった情報を常時出すために要る。
  it('names only the fans whose top can move', () => {
    // **積むのは1ランク下の札。**♠4 が ♠5 の上に乗るので、動かせるのは 1 の方。
    const fans = [
      [c('SPADE', 5)], // 0: ♠6 が無いので行き先なし
      [c('SPADE', 4)], // 1: 0 の ♠5 の上へ置ける
      [c('HEART', 1)], // 2: 空のファウンデーションへ
    ];
    expect(labelleLucieMovableFans(fans, [[]])).toEqual(new Set([1, 2]));
  });

  it('is empty when nothing can move', () => {
    const fans = [[c('SPADE', 5)], [c('HEART', 9)]];
    expect(labelleLucieMovableFans(fans, [[c('CLOVER', 1)]])).toEqual(new Set());
  });

  it('skips empty fans', () => {
    expect(labelleLucieMovableFans([[], [c('HEART', 1)]], [[]])).toEqual(new Set([1]));
  });

  // 既存の真偽値 API はこの集合から答える。両者が食い違わないことを踏む。
  it('agrees with labelleLucieHasLegalMove', () => {
    const stuck = [[c('SPADE', 5)], [c('HEART', 9)]];
    const foundation = [[c('CLOVER', 1)]];
    expect(labelleLucieHasLegalMove(stuck, foundation)).toBe(false);
    expect(labelleLucieMovableFans(stuck, foundation).size).toBe(0);

    const open = [[c('HEART', 1)]];
    expect(labelleLucieHasLegalMove(open, [[]])).toBe(true);
    expect(labelleLucieMovableFans(open, [[]]).size).toBe(1);
  });
});
