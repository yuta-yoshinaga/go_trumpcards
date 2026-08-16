import i18n from '../i18n';

/**
 * Renders a Niu Niu rank key as text in the current locale.
 *
 * The server sends a locale-independent key (`"none"`, `"niuniu"`, `"n1"`..`"n9"`)
 * rather than a display string. It used to send the Japanese label itself, so an
 * English-locale player saw 「牛牛」 on every hand and in the round-over line
 * (#5567).
 *
 * Returns an empty string while the hand is hidden, when the key is empty, and
 * for anything unrecognised -- the caller renders nothing rather than a raw key.
 */
export function niuniuRankText(rankKey: string): string {
  if (rankKey === 'none') return i18n.t('niuniu:rankNone');
  if (rankKey === 'niuniu') return i18n.t('niuniu:rankNiuNiu');
  const n = /^n([1-9])$/.exec(rankKey);
  if (n) return i18n.t('niuniu:rankN', { n: n[1] });
  return '';
}

/**
 * Renders the round-over headline ("Banker: Niu Niu") for the given rank key.
 *
 * Empty when the rank key is not one the server can send, so a round that has
 * not settled shows no headline instead of a stray label.
 */
export function niuniuBankerResultText(rankKey: string): string {
  const rank = niuniuRankText(rankKey);
  return rank === '' ? '' : i18n.t('niuniu:bankerResult', { rank });
}
