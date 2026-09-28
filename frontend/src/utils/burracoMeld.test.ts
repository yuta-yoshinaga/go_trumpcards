import { describe, expect, it } from 'vitest';
import type { Card } from '../types/common';
import type { BurracoMeldData } from '../types/games/burraco';
import { burracoMeldSelectionStatus } from './burracoMeld';

const card = (value: number, design: Card['design'] = 'HEART'): Card => ({ value, design });
const meld = (rank: number, cards: Card[]): BurracoMeldData => ({ cards, rank, isNatural: true, isBurraco: false });

describe('burracoMeldSelectionStatus', () => {
  it('returns select for an empty selection', () => {
    expect(burracoMeldSelectionStatus([], [])).toBe('select');
  });

  it('requires three cards for a new meld', () => {
    expect(burracoMeldSelectionStatus([card(7), card(7)], [])).toBe('invalid');
    expect(burracoMeldSelectionStatus([card(7), card(7), card(7)], [])).toBe('valid');
  });

  it('requires natural cards in a new meld to share a rank', () => {
    expect(burracoMeldSelectionStatus([card(7), card(7), card(8)], [])).toBe('invalid');
  });

  it('rejects black threes in a new meld', () => {
    expect(burracoMeldSelectionStatus([card(3, 'SPADE'), card(7), card(7)], [])).toBe('invalid');
  });

  it('requires at least two natural cards in a new meld', () => {
    expect(burracoMeldSelectionStatus([card(0, 'JOKER'), card(0, 'JOKER'), card(7)], [])).toBe('invalid');
  });

  it('allows at most three wild cards in a new meld', () => {
    expect(
      burracoMeldSelectionStatus(
        [card(0, 'JOKER'), card(0, 'JOKER'), card(0, 'JOKER'), card(0, 'JOKER'), card(7), card(7), card(7), card(7)],
        [],
      ),
    ).toBe('invalid');
  });

  it('does not allow more wild cards than natural cards in a new meld', () => {
    expect(
      burracoMeldSelectionStatus([card(0, 'JOKER'), card(0, 'JOKER'), card(0, 'JOKER'), card(7), card(7)], []),
    ).toBe('invalid');
  });

  it('checks each natural card against the rank of a new meld', () => {
    expect(burracoMeldSelectionStatus([card(7), card(8), card(8)], [])).toBe('invalid');
  });

  it('rejects a black three when adding cards to an existing meld', () => {
    expect(burracoMeldSelectionStatus([card(7), card(3, 'SPADE')], [meld(7, [card(7), card(7)])])).toBe('invalid');
  });

  it('allows an all-wild selection to form a new meld only when its conditions pass', () => {
    expect(burracoMeldSelectionStatus([card(0, 'JOKER'), card(2), card(2)], [])).toBe('invalid');
  });

  it('validates rank and the combined wild-card limit when adding to an existing meld', () => {
    const existing = [meld(7, [card(7), card(7), card(0, 'JOKER')])];
    expect(burracoMeldSelectionStatus([card(8)], existing)).toBe('invalid');
    expect(burracoMeldSelectionStatus([card(0, 'JOKER')], existing)).toBe('invalid');
    expect(burracoMeldSelectionStatus([card(7), card(0, 'JOKER')], existing)).toBe('valid');
    expect(burracoMeldSelectionStatus([card(7), card(0, 'JOKER'), card(0, 'JOKER'), card(0, 'JOKER')], existing)).toBe(
      'invalid',
    );
    expect(burracoMeldSelectionStatus([card(7)], existing)).toBe('valid');
  });

  it('rejects black threes when adding to an existing meld', () => {
    expect(burracoMeldSelectionStatus([card(3, 'CLOVER')], [meld(3, [card(3)])])).toBe('invalid');
  });
});
