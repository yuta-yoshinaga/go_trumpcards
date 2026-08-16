import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../../hooks/useCardDimensions';
import { focusRingCard, playableRingStyle, selectedCardStyle } from '../../styles/cardStyles';
import type { DaifugoPlayerData } from '../../types/card';
import { playerName } from '../../utils/playerUtils';
import { CardImage } from '../CardImage';
import { StatusBadge } from '../StatusBadge';
import { playerAreaClass } from './DaifugoCpuArea';

/** Props for {@link DaifugoHumanArea}. */
export interface HumanPlayerAreaProps {
  player: DaifugoPlayerData;
  selectedIndices: number[];
  onToggle: (idx: number) => void;
  isCurrentTurn: boolean;
  onDragCard: (idx: number) => void;
  onSwipeStart?: (idx: number) => void;
  /**
   * Hand indices the server says are in at least one legal play, mirroring the
   * `*` marks the CUI prints. `null`/`undefined` means the server could not
   * decide, and **no card is marked** -- marking none is honest, marking the
   * wrong ones is not.
   */
  playableIndices?: number[] | null;
}

/** Renders the human player's hand area for Daifugo with card selection and drag support. */
export function DaifugoHumanArea({
  player,
  selectedIndices,
  onToggle,
  isCurrentTurn,
  onDragCard,
  onSwipeStart,
  playableIndices,
}: HumanPlayerAreaProps) {
  const { t } = useTranslation('daifugo');
  const { cardWidth } = useCardDimensions();
  // Only mark during the human's turn: the indices are computed for whoever is
  // to move, and a stale set on a CPU turn would point at the wrong cards.
  const playable = isCurrentTurn && playableIndices ? new Set(playableIndices) : null;
  const conditionalClass = player.isFinished
    ? 'opacity-50'
    : isCurrentTurn
      ? 'border-2 border-game-status-active shadow-[0_0_12px_var(--color-game-status-active)]'
      : '';
  return (
    <div
      id={`player-area-${player.id}`}
      className={`${playerAreaClass}${conditionalClass ? ` ${conditionalClass}` : ''}`}
    >
      <div className="text-ds-text-primary font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && (
          <StatusBadge variant="success">{t('finishedWithRank', { rank: t(`rank.${player.rank}`) })}</StatusBadge>
        )}
        {player.illegalFinishPenalty && <StatusBadge variant="danger">{t('badge.illegalFinishPenalty')}</StatusBadge>}
      </div>
      {!player.isFinished && (
        <div className="text-game-text-muted text-xs mb-1">
          {t('cardCount', { count: player.cardCount })}
          {isCurrentTurn && <span className="ml-2 text-game-text-highlight">{t('selectToPlay')}</span>}
        </div>
      )}
      <div className="flex flex-wrap gap-1">
        {player.cards?.map((card, i) => (
          <button
            key={`${card.design}-${card.value}`}
            type="button"
            data-card-index={i}
            data-playable={playable?.has(i) ? 'true' : undefined}
            aria-pressed={selectedIndices.includes(i)}
            disabled={!isCurrentTurn}
            draggable={isCurrentTurn}
            className={`${focusRingCard} touch-none`}
            onClick={() => onToggle(i)}
            onPointerDown={isCurrentTurn && onSwipeStart ? () => onSwipeStart(i) : undefined}
            onDragStart={(e) => {
              e.dataTransfer.setData('cardIndex', String(i));
              onDragCard(i);
            }}
            style={{
              background: 'none',
              padding: 0,
              cursor: isCurrentTurn ? 'pointer' : 'default',
              borderRadius: 8,
              ...selectedCardStyle(selectedIndices.includes(i)),
              // The ring uses `outline`, so it stacks on the selection border
              // rather than replacing it -- a playable card can also be selected.
              ...(playable?.has(i) ? playableRingStyle() : {}),
              boxSizing: 'border-box',
            }}
          >
            <CardImage card={card} width={cardWidth} />
          </button>
        ))}
      </div>
    </div>
  );
}
