import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { labelleLucieHasLegalMove } from './labelleLucieLegalMove';

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
