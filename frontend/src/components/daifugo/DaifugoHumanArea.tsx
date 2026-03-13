import { useTranslation } from 'react-i18next';
import type { DaifugoPlayerData } from '../../types/card';
import { playerName } from '../../utils/playerUtils';
import { CardImage } from '../CardImage';
import { StatusBadge } from '../StatusBadge';
import { playerAreaClass } from './DaifugoCpuArea';

interface HumanPlayerAreaProps {
  player: DaifugoPlayerData;
  selectedIndices: number[];
  onToggle: (idx: number) => void;
  isCurrentTurn: boolean;
  onDragCard: (idx: number) => void;
}

export function DaifugoHumanArea({
  player,
  selectedIndices,
  onToggle,
  isCurrentTurn,
  onDragCard,
}: HumanPlayerAreaProps) {
  const { t } = useTranslation('daifugo');
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
      <div className="text-white font-bold mb-1">
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
            aria-pressed={selectedIndices.includes(i)}
            disabled={!isCurrentTurn}
            draggable={isCurrentTurn}
            onClick={() => onToggle(i)}
            onDragStart={(e) => {
              e.dataTransfer.setData('cardIndex', String(i));
              onDragCard(i);
            }}
            style={{
              background: 'none',
              padding: 0,
              cursor: isCurrentTurn ? 'pointer' : 'default',
              borderRadius: 8,
              border: selectedIndices.includes(i)
                ? '3px solid var(--color-game-status-waiting)'
                : '3px solid transparent',
              boxSizing: 'border-box',
            }}
          >
            <CardImage card={card} width={52} />
          </button>
        ))}
      </div>
    </div>
  );
}
