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
  /** Ranks this player has publicly asked for and not yet booked. */
  knownRanks?: number[];
  /** Ranks in `knownRanks` that the human can satisfy — rendered with extra emphasis. */
  matchedRanks?: number[];
}

/** Renders an opponent's card count and book count, clickable to select as ask target. */
export function GoFishPlayerArea({
  player,
  isSelected,
  onSelect,
  disabled,
  askAnnotation,
  knownRanks,
  matchedRanks,
}: GoFishPlayerAreaProps) {
  const { t } = useTranslation('gofish');
  const name = playerName(player.id, player.isHuman);
  const matchedSet = new Set(matchedRanks ?? []);

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
      {/* Always mount CpuActionBubble so its sr-only live region stays in the
          DOM across renders. Without this, unmounting would cause screen
          readers to miss announcements that land between mounts. */}
      <div className="absolute -top-2 right-2 z-10">
        <CpuActionBubble message={bubbleMessage} triggerKey={askAnnotation?.triggerKey} />
      </div>
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
        {knownRanks && knownRanks.length > 0 && (
          <div
            data-testid={`known-ranks-${player.id}`}
            className="mt-1 text-xs text-ds-text-muted flex flex-wrap gap-1 items-center"
          >
            <span aria-hidden="true">👀</span>
            <span className="sr-only">{t('knownRanksLabel')}</span>
            {knownRanks.map((r) => {
              const matched = matchedSet.has(r);
              return (
                <span
                  key={r}
                  data-rank={r}
                  data-matched={matched ? 'true' : undefined}
                  className={`px-1.5 py-0.5 rounded font-medium ${
                    matched ? 'bg-ds-warning text-ds-text-on-accent animate-pulse' : 'bg-white/10 text-ds-text-primary'
                  }`}
                >
                  {valueName(r)}
                </span>
              );
            })}
          </div>
        )}
        {player.books.length > 0 && (
          <div
            data-testid={`book-ranks-${player.id}`}
            className="mt-1 text-xs text-ds-text-muted flex flex-wrap gap-1 items-center"
          >
            <span aria-hidden="true">📚</span>
            <span className="sr-only">{t('bookRanksLabel')}</span>
            {player.books.map((book) => (
              <span
                key={book.rank}
                data-rank={book.rank}
                className="px-1.5 py-0.5 rounded font-medium bg-ds-success/50 text-white"
              >
                {valueName(book.rank)}
              </span>
            ))}
          </div>
        )}
      </button>
    </div>
  );
}
