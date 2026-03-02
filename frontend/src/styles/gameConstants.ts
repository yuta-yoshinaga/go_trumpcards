import { PokerAction } from '../types/phases';

/**
 * Betting action display names shared by Poker and Holdem.
 * PokerAction and HoldemAction have the same numeric values (0-5),
 * so a single map is sufficient. Separate aliases are exported
 * for semantic clarity at call sites.
 */
export const BETTING_ACTION_NAMES: Record<number, string> = {
  [PokerAction.FOLD]: 'フォールド',
  [PokerAction.CHECK]: 'チェック',
  [PokerAction.CALL]: 'コール',
  [PokerAction.BET]: 'ベット',
  [PokerAction.RAISE]: 'レイズ',
  [PokerAction.ALL_IN]: 'オールイン',
};

/** @see {@link BETTING_ACTION_NAMES} */
export const POKER_ACTION_NAMES = BETTING_ACTION_NAMES;

/** @see {@link BETTING_ACTION_NAMES} */
export const HOLDEM_ACTION_NAMES = BETTING_ACTION_NAMES;

/** Badge style for hand name display (e.g. "ツーペア"). */
export const handNameBadgeStyle: React.CSSProperties = {
  background: '#f0ad4e',
  color: '#222',
};

/** Highlight style for the active turn player area. */
export const activeTurnStyle: React.CSSProperties = {
  border: '2px solid #f0ad4e',
  boxShadow: '0 0 12px #f0ad4e',
};

/** Dim style for finished players. */
export const finishedPlayerStyle: React.CSSProperties = {
  opacity: 0.5,
};
