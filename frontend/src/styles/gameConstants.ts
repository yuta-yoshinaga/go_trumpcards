import { HoldemAction, PokerAction } from '../types/phases';

/** Poker action display names (Poker/Holdem share the same numeric values). */
export const POKER_ACTION_NAMES: Record<number, string> = {
  [PokerAction.FOLD]: 'フォールド',
  [PokerAction.CHECK]: 'チェック',
  [PokerAction.CALL]: 'コール',
  [PokerAction.BET]: 'ベット',
  [PokerAction.RAISE]: 'レイズ',
  [PokerAction.ALL_IN]: 'オールイン',
};

/** Holdem action display names. */
export const HOLDEM_ACTION_NAMES: Record<number, string> = {
  [HoldemAction.FOLD]: 'フォールド',
  [HoldemAction.CHECK]: 'チェック',
  [HoldemAction.CALL]: 'コール',
  [HoldemAction.BET]: 'ベット',
  [HoldemAction.RAISE]: 'レイズ',
  [HoldemAction.ALL_IN]: 'オールイン',
};

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
