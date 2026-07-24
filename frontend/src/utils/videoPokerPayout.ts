/**
 * Video Poker payout tables, ported 1:1 from the Go domain
 * (`internal/domain/VideoPokerVariant.go`) so the on-screen paytable grid shows
 * exactly what the server pays. Each variant lists its hands top-down with the
 * per-coin multiplier; the Royal Flush row pays a non-linear jackpot at max bet
 * (250x for 1-4 coins, a flat 4000 at 5 coins), matching the server's special case.
 */

/** Maximum bet in coins (mirrors `VideoPokerMaxBet`). */
export const VIDEO_POKER_MAX_BET = 5;

/** The flat Royal Flush jackpot paid at max bet (Go returns multiplier 800 x 5 coins). */
const ROYAL_FLUSH_JACKPOT = 4000;

/** Supported variants (matches the shared component's `gameName`). */
export type VideoPokerVariant = 'videopoker' | 'deuceswild' | 'jokerpoker';

/** One paytable row: an i18n hand-name key plus its per-coin multiplier. */
export interface VideoPokerPayoutRow {
  /** i18n key under `payoutTable.name.*` */
  key: string;
  /** Coins paid per coin bet for this hand. */
  perCoin: number;
  /** When true, the max-bet (5-coin) cell pays the flat Royal jackpot instead of perCoin x 5. */
  royalJackpot?: boolean;
}

const JACKS_OR_BETTER_ROWS: VideoPokerPayoutRow[] = [
  { key: 'royalFlush', perCoin: 250, royalJackpot: true },
  { key: 'straightFlush', perCoin: 50 },
  { key: 'fourOfAKind', perCoin: 25 },
  { key: 'fullHouse', perCoin: 9 },
  { key: 'flush', perCoin: 6 },
  { key: 'straight', perCoin: 4 },
  { key: 'threeOfAKind', perCoin: 3 },
  { key: 'twoPair', perCoin: 2 },
  { key: 'jacksOrBetter', perCoin: 1 },
];

const DEUCES_WILD_ROWS: VideoPokerPayoutRow[] = [
  { key: 'naturalRoyalFlush', perCoin: 250, royalJackpot: true },
  { key: 'fourDeuces', perCoin: 200 },
  { key: 'wildRoyalFlush', perCoin: 25 },
  { key: 'fiveOfAKind', perCoin: 15 },
  { key: 'straightFlush', perCoin: 9 },
  { key: 'fourOfAKind', perCoin: 5 },
  { key: 'fullHouse', perCoin: 3 },
  { key: 'flush', perCoin: 2 },
  { key: 'straight', perCoin: 2 },
  { key: 'threeOfAKind', perCoin: 1 },
];

const JOKER_POKER_ROWS: VideoPokerPayoutRow[] = [
  { key: 'naturalRoyalFlush', perCoin: 250, royalJackpot: true },
  { key: 'fiveOfAKind', perCoin: 200 },
  { key: 'wildRoyalFlush', perCoin: 100 },
  { key: 'straightFlush', perCoin: 50 },
  { key: 'fourOfAKind', perCoin: 20 },
  { key: 'fullHouse', perCoin: 7 },
  { key: 'flush', perCoin: 5 },
  { key: 'straight', perCoin: 3 },
  { key: 'threeOfAKind', perCoin: 2 },
  { key: 'twoPair', perCoin: 1 },
  { key: 'kingsOrBetter', perCoin: 1 },
];

const PAYOUT_ROWS: Record<VideoPokerVariant, VideoPokerPayoutRow[]> = {
  videopoker: JACKS_OR_BETTER_ROWS,
  deuceswild: DEUCES_WILD_ROWS,
  jokerpoker: JOKER_POKER_ROWS,
};

/** Returns the ordered paytable rows for a variant. */
export function videoPokerPayoutRows(variant: VideoPokerVariant): VideoPokerPayoutRow[] {
  return PAYOUT_ROWS[variant];
}

/**
 * Returns the coins paid for `row` at the given bet (1..5). The Royal Flush row
 * pays its flat jackpot at max bet; every other cell is linear (perCoin x bet).
 */
export function videoPokerPayoutCell(row: VideoPokerPayoutRow, bet: number): number {
  if (row.royalJackpot && bet === VIDEO_POKER_MAX_BET) {
    return ROYAL_FLUSH_JACKPOT;
  }
  return row.perCoin * bet;
}

/**
 * Maps a server-produced hand name (e.g. "Natural Royal Flush", "Jacks or Better")
 * to its paytable row key, or `null` when the hand pays nothing / is not on the
 * table. Used as a fallback for states that predate the `handKey` field.
 */
export function videoPokerHandNameToRowKey(handName: string): string | null {
  return HAND_NAME_TO_ROW_KEY[handName] ?? null;
}

/**
 * Resolves the winning paytable row key for a result. Prefers the stable
 * server-supplied `handKey`; falls back to reverse-looking up the English
 * `handName` for older responses that omit the key. Returns `null` for a losing
 * hand (empty key and unmatched name).
 */
export function videoPokerRowKey(handKey: string | undefined, handName: string): string | null {
  if (handKey) {
    return handKey;
  }
  return videoPokerHandNameToRowKey(handName);
}

const HAND_NAME_TO_ROW_KEY: Record<string, string> = {
  'Royal Flush': 'royalFlush',
  'Natural Royal Flush': 'naturalRoyalFlush',
  'Wild Royal Flush': 'wildRoyalFlush',
  'Four Deuces': 'fourDeuces',
  'Five of a Kind': 'fiveOfAKind',
  'Straight Flush': 'straightFlush',
  'Four of a Kind': 'fourOfAKind',
  'Full House': 'fullHouse',
  Flush: 'flush',
  Straight: 'straight',
  'Three of a Kind': 'threeOfAKind',
  'Two Pair': 'twoPair',
  'Jacks or Better': 'jacksOrBetter',
  'Kings or Better': 'kingsOrBetter',
};
