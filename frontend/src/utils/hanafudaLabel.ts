import i18n from '../i18n';
import type { Card } from '../types/card';

/** Return the current-language hanafuda label when the card has complete hanafuda identity. */
export function hanafudaLabel(card: Card): string | undefined {
  if (card.deck !== 'hanafuda' || card.month === undefined || card.kind === undefined) return undefined;
  return i18n.t('common:hanafuda.cardLabel', {
    month: i18n.t(`common:hanafuda.month.${card.month}`),
    kind: i18n.t(`common:hanafuda.kind.${card.kind}`),
  });
}

/** Return the compact current-language label shown in a hanafuda card corner. */
export function hanafudaCornerLabel(card: Card): string | undefined {
  if (card.deck !== 'hanafuda' || card.month === undefined || card.kind === undefined) return undefined;
  return i18n.t('common:hanafuda.cornerLabel', {
    month: i18n.t(`common:hanafuda.monthShort.${card.month}`),
    kind: i18n.t(`common:hanafuda.kind.${card.kind}`),
  });
}
