import i18n from '../i18n';
import niuniuMultipliersGolden from './__fixtures__/niuniuMultipliers.golden.json';

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

/**
 * Returns the payout multiplier for a given Niu Niu rank key.
 *
 * 値は internal/domain/NiuNiu.go の niuNiuMultiplier が正。
 * Go 側の golden テストが JSON と規則の一致を検査するので片方だけ変えると落ちる。
 */
export function niuniuMultiplier(rankKey: string): number {
  const multipliers: Record<string, number> = niuniuMultipliersGolden;
  return multipliers[rankKey] ?? 1;
}
