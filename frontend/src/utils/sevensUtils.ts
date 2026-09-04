import { playerAreaBase } from '../styles/gameStyles';
import type { Card, CardDesign, SevensAction } from '../types/card';
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

/** Check if a position is blocked by the end-stop rule. */
export function isEndStopped(tablePlaced: number[], suit: number, value: number, endStopEnabled: boolean): boolean {
  if (!endStopEnabled) return false;
  if (value === 7) return false;
  if (value > 7 && isPositionPlaced(tablePlaced, suit, 1)) return true;
  if (value < 7 && isPositionPlaced(tablePlaced, suit, 13)) return true;
  return false;
}

/** Wrap a card value to the 1-13 range for tunnel mode. */
export function wrapValue(v: number): number {
  v = ((v - 1) % 13) + 1;
  if (v <= 0) v += 13;
  return v;
}

/** Check if a board position can accept a card placement. */
export function isPositionPlayable(
  tablePlaced: number[],
  suit: number,
  value: number,
  tunnelEnabled: boolean,
  endStopEnabled: boolean,
  tunnelSkipWidth = 0,
): boolean {
  if (isPositionPlaced(tablePlaced, suit, value)) return false;
  if (isEndStopped(tablePlaced, suit, value, endStopEnabled)) return false;
  if (isPositionPlaced(tablePlaced, suit, value + 1)) return true;
  if (isPositionPlaced(tablePlaced, suit, value - 1)) return true;
  if (tunnelEnabled) {
    if (value === 1 && isPositionPlaced(tablePlaced, suit, 13)) return true;
    if (value === 13 && isPositionPlaced(tablePlaced, suit, 1)) return true;
  }
  if (tunnelSkipWidth >= 2) {
    let low = value - tunnelSkipWidth;
    let high = value + tunnelSkipWidth;
    if (tunnelEnabled) {
      low = wrapValue(low);
      high = wrapValue(high);
    }
    if (low >= 1 && low <= 13 && isPositionPlaced(tablePlaced, suit, low)) return true;
    if (high >= 1 && high <= 13 && isPositionPlaced(tablePlaced, suit, high)) return true;
  }
  return false;
}

/** Check if any position on the board can accept a card placement. */
export function hasAnyPlayablePosition(
  tablePlaced: number[],
  tunnelEnabled: boolean,
  endStopEnabled: boolean,
  tunnelSkipWidth = 0,
): boolean {
  for (let suit = 1; suit <= 4; suit++) {
    for (let v = 1; v <= 13; v++) {
      if (isPositionPlayable(tablePlaced, suit, v, tunnelEnabled, endStopEnabled, tunnelSkipWidth)) return true;
    }
  }
  return false;
}

/**
 * Returns every (suit, value) coordinate where the joker can legally land.
 * Used by the page to auto-place the joker when there is exactly one option,
 * so the player avoids a redundant click on a single highlighted cell.
 */
export function listJokerPlacements(
  tablePlaced: number[],
  tunnelEnabled: boolean,
  endStopEnabled: boolean,
  tunnelSkipWidth = 0,
): Array<{ suit: number; value: number }> {
  const result: Array<{ suit: number; value: number }> = [];
  for (let suit = 1; suit <= 4; suit++) {
    for (let v = 1; v <= 13; v++) {
      if (isPositionPlayable(tablePlaced, suit, v, tunnelEnabled, endStopEnabled, tunnelSkipWidth)) {
        result.push({ suit, value: v });
      }
    }
  }
  return result;
}

/** Check if a hand contains only joker cards. */
export function hasOnlyJokers(cards: Card[]): boolean {
  return cards.length > 0 && cards.every((c) => c.design === 'JOKER');
}

/** Check if a specific card can be played in Sevens. */
export function isCardPlayable(
  card: Card,
  tablePlaced: number[],
  tunnelEnabled: boolean,
  noJokerFinish: boolean,
  allCards: Card[],
  endStopEnabled: boolean,
  jokerConsecutiveBanned: boolean,
  lastPlayedJoker: boolean,
  tunnelSkipWidth = 0,
): boolean {
  if (card.design === 'JOKER') {
    if (noJokerFinish && hasOnlyJokers(allCards)) return false;
    if (jokerConsecutiveBanned && lastPlayedJoker) return false;
    return hasAnyPlayablePosition(tablePlaced, tunnelEnabled, endStopEnabled, tunnelSkipWidth);
  }
  const suit = designToSuit[card.design];
  return isPositionPlayable(tablePlaced, suit, card.value, tunnelEnabled, endStopEnabled, tunnelSkipWidth);
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
