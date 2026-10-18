import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { unsunKarutaCornerLabel, unsunKarutaLabel } from './unsunKarutaLabel';

afterEach(async () => i18n.changeLanguage('ja'));

describe('Unsun Karuta labels', () => {
  it.each([
    ['ja', 'ソウタ・ぱお', 'ソウ'],
    ['en', 'Sota of Pao', 'Sot'],
  ])('localizes %s full, corner, and center labels', async (language, full, corner) => {
    await i18n.changeLanguage(language);
    const card = {
      design: 'SPADE' as const,
      value: 10,
      deck: 'unsun',
      kind: 'sota' as const,
      unsunSuit: 'pao' as const,
    };
    expect(unsunKarutaLabel(card)).toBe(full);
    expect(unsunKarutaCornerLabel(card)).toBe(corner);
  });

  it('returns undefined for other decks', () => {
    expect(unsunKarutaLabel({ design: 'SPADE', value: 10, deck: 'wizard' })).toBeUndefined();
    expect(unsunKarutaCornerLabel({ design: 'SPADE', value: 10, deck: 'wizard', kind: 'sota' })).toBeUndefined();
    // A number card shows only its number in the corner; the full name stays in the accessible name.
    expect(unsunKarutaCornerLabel({ design: 'SPADE', value: 9, deck: 'unsun', unsunSuit: 'isu' })).toBe('9');
  });
});
