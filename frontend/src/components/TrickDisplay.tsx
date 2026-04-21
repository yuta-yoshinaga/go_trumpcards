import type { Card } from '../types/card';
import { playerName } from '../utils/playerUtils';
import { AnimatedCard } from './motion/AnimatedCard';

/** One card played into the current trick. Matches all trick-taking game TrickCard shapes. */
export interface TrickDisplayCard {
  playerIdx: number;
  card: Card;
}

/** Minimal player reference used for display-name resolution. */
export interface TrickDisplayPlayer {
  id: number;
  isHuman: boolean;
}

/** Props for {@link TrickDisplay}. */
export interface TrickDisplayProps {
  /** Cards currently on the trick. When empty, the component renders nothing. */
  currentTrick: TrickDisplayCard[];
  /** All players; indexed by `trickCard.playerIdx`. */
  players: TrickDisplayPlayer[];
  /** Card width in px forwarded to {@link AnimatedCard}. */
  cardWidth: number;
  /** Localised label, e.g. `t('currentTrick')`. */
  label: string;
  /** Value for the `data-tutorial` attribute (e.g. `"ht-trick-display"`). */
  dataTutorial?: string;
  /** Forwarded to {@link AnimatedCard#onDealComplete}; typically plays a deal sound. */
  onCardDealComplete?: () => void;
}

/**
 * Shared trick-display area for trick-taking games (Hearts, Spades, Euchre, Bridge,
 * Napoleon, OhHell, TwoTenJack, Whist). Renders each played card with the player's
 * name underneath, or nothing when the trick is empty.
 */
export function TrickDisplay({
  currentTrick,
  players,
  cardWidth,
  label,
  dataTutorial,
  onCardDealComplete,
}: TrickDisplayProps) {
  if (currentTrick.length === 0) {
    return null;
  }
  return (
    <div className="my-3 p-3 rounded bg-black/40" data-tutorial={dataTutorial}>
      <div className="text-ds-text-muted text-sm mb-1">{label}</div>
      <div className="flex gap-2">
        {currentTrick.map((trickCard) => (
          <div key={`trick-${trickCard.playerIdx}`} className="text-center">
            <AnimatedCard card={trickCard.card} width={cardWidth} onDealComplete={onCardDealComplete} />
            <div className="text-game-text-muted text-xs mt-1">
              {playerName(
                players[trickCard.playerIdx]?.id ?? trickCard.playerIdx,
                players[trickCard.playerIdx]?.isHuman ?? false,
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
