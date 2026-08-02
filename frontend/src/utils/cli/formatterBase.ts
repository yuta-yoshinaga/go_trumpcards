import i18n from '../../i18n';
import type { Card } from '../../types/card';
import { isRequestedHint } from '../hintRequest';

const SUIT_SYMBOLS: Record<string, string> = {
  SPADE: '\u2660',
  HEART: '\u2665',
  DIAMOND: '\u2666',
  CLOVER: '\u2663',
};

const VALUE_NAMES: Record<number, string> = {
  1: 'A',
  11: 'J',
  12: 'Q',
  13: 'K',
};

/** Format a single card as a Unicode symbol string (e.g., "♠5", "♥K", "🃏").
 * Procedural (non-52 deck) cards fall back to their descriptor glyph/label. */
export function formatCard(card: Card): string {
  if (card.label || card.glyph) return `${card.glyph ?? ''}${card.label ?? ''}`;
  if (card.design === 'JOKER') return '🃏';
  const suit = SUIT_SYMBOLS[card.design] ?? '?';
  const value = VALUE_NAMES[card.value] ?? String(card.value);
  return `${suit}${value}`;
}

/** Format cards as a comma-separated string. */
export function formatCardList(cards: Card[]): string {
  return cards.map(formatCard).join(', ');
}

/** Format cards with indices (e.g., "[0]♠5  [1]♥K"). */
export function formatIndexedCards(cards: Card[]): string {
  return cards.map((c, i) => `[${i}]${formatCard(c)}`).join('  ');
}

/** Return the standard separator line. */
export function formatSeparator(): string {
  return '==========';
}

/** Return a header block with title wrapped in separators. */
export function formatHeader(title: string): string {
  return `${formatSeparator()}\n${title}\n${formatSeparator()}`;
}

/** Format a player display name. */
export function formatPlayerName(id: number, isHuman: boolean): string {
  return isHuman ? i18n.t('player.you') : i18n.t('player.cpu', { id });
}

// 実体は `utils/hintRequest.ts`。CLI フォーマッタ 77 本がここから読んでいるので
// 再エクスポートで受ける (#4605)。
export { isRequestedHint };
