import type { TFunction } from 'i18next';

/** The play-history fields the badge reports. */
export interface PokerStats {
  /** Hands played so far. The badge renders nothing while this is 0. */
  totalHands: number;
  /** Voluntarily put money in pot, as a percentage. */
  vpip: number;
  /** Pre-flop raise, as a percentage. */
  pfr: number;
  /** Three-bet frequency, as a percentage. */
  threeBet: number;
  /** Aggression factor, pre-formatted by the backend (may be "∞"). */
  af: string;
}

/**
 * Renders the VPIP / PFR / 3Bet / AF summary for one player.
 *
 * **CUI (`FiveCardStudCuiPresenter`) は `fivecardstud.playerStats` でこの4つを
 * 全プレイヤー分出しているのに、Web はプレイスタイル名しか出していなかった。**
 * 名前だけでは相手の実際の傾向が読めず、同じレスポンスに乗っている数字が
 * Web だけ捨てられている状態だった (#4730)。
 *
 * `totalHands` が 0 のあいだは何も描画しない — 統計として意味を持たないうえ、
 * CUI 側も同じ条件で行ごと出していない。
 */
export function PokerStatsBadge({ stats, t }: { stats: PokerStats; t: TFunction }) {
  if (stats.totalHands <= 0) {
    return null;
  }
  return (
    <span className="ml-2 text-xs text-ds-text-muted" data-testid={'fcs-stats'}>
      {t('stats.badge', {
        vpip: stats.vpip,
        pfr: stats.pfr,
        tb: stats.threeBet,
        af: stats.af,
      })}
    </span>
  );
}
