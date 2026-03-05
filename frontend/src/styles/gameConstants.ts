import i18n from '../i18n';
import { PokerAction } from '../types/phases';

/**
 * Betting action display names shared by Poker and Holdem.
 * Both games use the same numeric action values (0-5),
 * so a single map is sufficient. Separate aliases are exported
 * for semantic clarity at call sites.
 */
export function bettingActionName(action: number): string {
  switch (action) {
    case PokerAction.FOLD:
      return i18n.t('action.fold');
    case PokerAction.CHECK:
      return i18n.t('action.check');
    case PokerAction.CALL:
      return i18n.t('action.call');
    case PokerAction.BET:
      return i18n.t('action.bet');
    case PokerAction.RAISE:
      return i18n.t('action.raise');
    case PokerAction.ALL_IN:
      return i18n.t('action.allIn');
    default:
      return i18n.t('action.unknown');
  }
}

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
