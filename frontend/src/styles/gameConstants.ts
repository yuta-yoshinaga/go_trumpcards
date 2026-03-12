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

/** Badge classes for hand name display (e.g. "ツーペア"). */
export const handNameBadgeClass = 'bg-game-status-waiting text-game-text-strong';

/** Highlight classes for the active turn player area. */
export const activeTurnClass = 'border-2 border-game-status-waiting shadow-[0_0_12px_var(--color-game-status-waiting)]';

/** Dim classes for finished players. */
export const finishedPlayerClass = 'opacity-50';
