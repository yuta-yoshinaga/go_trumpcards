import { describe, expect, it } from 'vitest';
import { ColourWhistContract } from '../types/phases';
import { colourWhistHasPartner } from './colourWhistPartner';

describe('colourWhistHasPartner', () => {
  it('is true for the two-against-two contracts', () => {
    expect(colourWhistHasPartner(ColourWhistContract.SAMEN)).toBe(true);
    // Troel は競りで選べず、配りで自動的に成立する。相方は 4 枚目のエース保持者。
    expect(colourWhistHasPartner(ColourWhistContract.TROEL)).toBe(true);
  });

  it('is false for the solo contracts and before anyone declares', () => {
    expect(colourWhistHasPartner(ColourWhistContract.ALLEEN)).toBe(false);
    expect(colourWhistHasPartner(ColourWhistContract.MISERIE)).toBe(false);
    expect(colourWhistHasPartner(ColourWhistContract.NONE)).toBe(false);
  });

  // ドメインの契約は 0..4 の 5 種類しかない。範囲外を true にすると、
  // 未知の契約が増えたときに「隠れた味方がいる」と出てしまう。
  it('is false for anything outside the domain range', () => {
    for (const c of [-1, 5, 99]) expect(colourWhistHasPartner(c)).toBe(false);
  });
});
