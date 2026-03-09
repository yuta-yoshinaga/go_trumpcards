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
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {};
  return (
    <div id={`player-area-${player.id}`} className={playerAreaClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && (
          <StatusBadge variant="success">{t('finishedWithRank', { rank: t(`rank.${player.rank}`) })}</StatusBadge>
        )}
        {player.illegalFinishPenalty && <StatusBadge variant="danger">{t('badge.illegalFinishPenalty')}</StatusBadge>}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {t('cardCount', { count: player.cardCount })}
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>{t('selectToPlay')}</span>}
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
              border: selectedIndices.includes(i) ? '3px solid #f0ad4e' : '3px solid transparent',
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
