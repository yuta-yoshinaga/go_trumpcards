import { useTranslation } from 'react-i18next';
import type { GoFishPlayerData } from '../../types/card';
import { valueName } from '../../utils/cardUtils';
import { playerName } from '../../utils/playerUtils';
import { CpuActionBubble } from '../CpuActionBubble';

/** Optional transient "last ask" annotation rendered as a speech bubble. */
export interface GoFishAskAnnotation {
  /** The rank that was asked (1-13). */
  rank: number;
  /** Number of cards actually received. Zero means "Go Fish" miss. */
  receivedCount: number;
  /** Stable identity used to re-trigger the bubble animation. */
  triggerKey: string | number;
}

interface GoFishPlayerAreaProps {
  player: GoFishPlayerData;
  isSelected: boolean;
  onSelect: (idx: number) => void;
  disabled: boolean;
  /**
   * When set, a short-lived speech bubble is rendered above the player area
   * summarizing the last ask that targeted this player. Pass `undefined` to
   * hide the bubble.
   */
  askAnnotation?: GoFishAskAnnotation;
}

/** Renders an opponent's card count and book count, clickable to select as ask target. */
export function GoFishPlayerArea({ player, isSelected, onSelect, disabled, askAnnotation }: GoFishPlayerAreaProps) {
  const { t } = useTranslation('gofish');
  const name = playerName(player.id, player.isHuman);

  const bubbleMessage = askAnnotation
    ? askAnnotation.receivedCount > 0
      ? t('bubble.askHit', {
          rank: valueName(askAnnotation.rank),
          count: askAnnotation.receivedCount,
        })
      : t('bubble.askMiss', { rank: valueName(askAnnotation.rank) })
    : undefined;

  return (
    <div className="relative">
      {bubbleMessage && (
        <div className="absolute -top-2 right-2 z-10">
          <CpuActionBubble message={bubbleMessage} triggerKey={askAnnotation?.triggerKey} />
        </div>
      )}
      <button
        type="button"
        onClick={() => onSelect(player.id)}
        disabled={disabled}
        className={`w-full mb-2 p-2 rounded text-left transition-colors ${
          isSelected ? 'bg-ds-warning/30 ring-2 ring-ds-warning' : 'bg-black/30 hover:bg-black/40'
        } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
        aria-pressed={isSelected}
      >
        <div className="text-ds-text-muted text-sm">
          {name}: {t('deck', { count: player.cardCount })} | {t('books', { count: player.bookCount })}
        </div>
      </button>
    </div>
  );
}
