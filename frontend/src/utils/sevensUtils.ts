import { playerAreaBase } from '../styles/gameStyles';
import type { CardDesign, SevensAction } from '../types/card';
import { suitName, valueName } from './cardUtils';
import { findPlayerName } from './playerUtils';

/** Map from card design name to numeric suit index. */
export const designToSuit: Record<CardDesign, number> = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
  JOKER: 0,
};

/** Suit display metadata for the Sevens board. */
export const SUITS = [
  { idx: 1, name: 'SPADE', label: '♠', color: '#e0e0e0' },
  { idx: 2, name: 'CLOVER', label: '♣', color: '#e0e0e0' },
  { idx: 3, name: 'HEART', label: '♥', color: '#f87171' },
  { idx: 4, name: 'DIAMOND', label: '♦', color: '#f87171' },
];

/** Check if a card position is already placed on the Sevens board. */
export function isPositionPlaced(tablePlaced: number[], suit: number, value: number): boolean {
  return (tablePlaced[suit] & (1 << value)) !== 0;
}

/** Build a human-readable description of a Sevens action. */
export function actionDesc(
  players: { id: number; isHuman: boolean }[],
  action: SevensAction,
  t: (key: string, opts?: Record<string, unknown>) => string,
): string {
  if (!action.playedCard) {
    const base = t('actionPassed', { name: findPlayerName(players, action.playerIdx) });
    return action.forcedPass ? t('actionForcedPass', { base }) : base;
  }
  const c = action.playedCard;
  if (c.design === 'JOKER' && action.targetSuit > 0) {
    return t('actionPlayedJoker', {
      name: findPlayerName(players, action.playerIdx),
      design: c.design,
      value: valueName(c.value),
      targetSuit: suitName(action.targetSuit),
      targetValue: valueName(action.targetValue),
    });
  }
  const played = t('actionPlayed', {
    name: findPlayerName(players, action.playerIdx),
    design: c.design,
    value: valueName(c.value),
  });
  // Reclaiming silently returns the joker to the hand: the card count goes up
  // with nothing on screen to account for it.
  return action.jokerReclaimed ? `${played} ${t('actionReclaimedJoker')}` : played;
}

/** CSS class for Sevens player area layout. */
export const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_180px] min-w-[150px]`;
