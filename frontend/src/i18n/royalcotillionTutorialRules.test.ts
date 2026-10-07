import { describe, expect, it } from 'vitest';
import en from './locales/en/royalcotillion.json';
import ja from './locales/ja/royalcotillion.json';

describe('Royal Cotillion tutorial initial deal', () => {
  it('explains the tableau, reserve, and stock counts in Japanese', () => {
    expect(ja.tutorial.tableau).toContain('場札16枠に1枚ずつ');
    expect(ja.tutorial.tableau).toContain('リザーブ4山に3枚ずつ');
    expect(ja.tutorial.tableau).toContain('残り76枚は山札');
  });

  it('explains the tableau, reserve, and stock counts in English', () => {
    expect(en.tutorial.tableau).toContain('Sixteen tableau spaces are dealt one card each');
    expect(en.tutorial.tableau).toContain('four reserve piles are dealt three cards each');
    expect(en.tutorial.tableau).toContain('the remaining 76 cards are the stock');
  });
});

describe('Royal Cotillion tutorial rules', () => {
  it('explains the skipping foundation sequences in Japanese', () => {
    expect(ja.tutorial.foundation).toContain('A→3→5');
    expect(ja.tutorial.foundation).toContain('2つ飛ばし');
  });

  it('explains the skipping foundation sequences in English', () => {
    expect(en.tutorial.foundation).toContain('A-3-5');
    expect(en.tutorial.foundation).toContain('skipping one rank');
  });

  it('says tableau spaces cannot be stacked in both languages', () => {
    expect(ja.tutorial.tableau).toContain('重ねられず');
    expect(en.tutorial.tableau).toContain('cards cannot be stacked');
  });
});
