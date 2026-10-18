import i18n from '../i18n';
import type { Card } from '../types/card';

const RANKS = ['sota', 'uma', 'kiri', 'un', 'sun', 'robai'] as const;
const SUITS = ['pao', 'isu', 'kotsu', 'oru', 'kuru'] as const;

/** Return the localized full Unsun Karuta card name, or undefined for other cards. */
export function unsunKarutaLabel(card: Card): string | undefined {
  if (card.deck !== 'unsun' || !card.unsunSuit || !SUITS.includes(card.unsunSuit)) return undefined;
  const rank =
    card.kind && RANKS.includes(card.kind as (typeof RANKS)[number])
      ? i18n.t(`common:unsunkaruta.rank.${card.kind}`)
      : String(card.value);
  const suit = i18n.t(`common:unsunkaruta.suit.${card.unsunSuit}`);
  return i18n.t('common:unsunkaruta.cardLabel', { rank, suit });
}

/**
 * Return the compact localized rank shown in an Unsun Karuta card corner: the abbreviated face
 * name for a face card, the bare number for a number card (the suit is the centre glyph).
 */
export function unsunKarutaCornerLabel(card: Card): string | undefined {
  if (card.deck !== 'unsun') return undefined;
  if (card.kind && RANKS.includes(card.kind as (typeof RANKS)[number])) {
    return i18n.t(`common:unsunkaruta.corner.${card.kind}`);
  }
  return String(card.value);
}
