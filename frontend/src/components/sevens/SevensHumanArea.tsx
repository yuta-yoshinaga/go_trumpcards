import { useTranslation } from 'react-i18next';
import type { Card, SevensPlayerData } from '../../types/card';
import { valueName } from '../../utils/cardUtils';
import { playerName } from '../../utils/playerUtils';
import { isCardPlayable, playerAreaClass } from '../../utils/sevensUtils';
import { CardImage } from '../CardImage';
import { StatusBadge } from '../StatusBadge';

interface HumanAreaProps {
  player: SevensPlayerData;
  isCurrentTurn: boolean;
  tablePlaced: number[];
  tunnelEnabled: boolean;
  noJokerFinish: boolean;
  endStopEnabled: boolean;
  jokerConsecutiveBanned: boolean;
  loading: boolean;
  onPlay: (idx: number) => void;
}

function HumanArea({
  player,
  isCurrentTurn,
  tablePlaced,
  tunnelEnabled,
  noJokerFinish,
  endStopEnabled,
  jokerConsecutiveBanned,
  loading,
  onPlay,
}: HumanAreaProps) {
  const { t } = useTranslation('sevens');
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {};
  return (
    <div className={playerAreaClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && <StatusBadge variant="success">{t('rankLabel', { rank: player.rank })}</StatusBadge>}
      </div>
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em] mb-1">
          {t('cardCount', { count: player.cardCount })}
          {'　'}
          {t('passCount', {
            used: player.passesUsed,
            max: player.maxPasses === 0 ? t('passUnlimited') : player.maxPasses,
          })}
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>{t('clickPlayable')}</span>}
        </div>
      )}
      <div className="flex flex-wrap gap-1">
        {player.cards?.map((card: Card, i: number) => {
          const playable =
            isCurrentTurn &&
            !loading &&
            isCardPlayable(
              card,
              tablePlaced,
              tunnelEnabled,
              noJokerFinish,
              player.cards,
              endStopEnabled,
              jokerConsecutiveBanned,
              player.lastPlayedJoker,
            );
          return (
            <button
              key={`${card.design}-${card.value}`}
              type="button"
              disabled={!playable}
              onClick={() => onPlay(i)}
              title={playable ? t('playTitle', { design: card.design, value: valueName(card.value) }) : undefined}
              data-testid={playable ? 'playable-card' : undefined}
              style={{
                background: 'none',
                padding: 0,
                cursor: playable ? 'pointer' : 'default',
                borderRadius: 8,
                border: playable ? '3px solid #5cb85c' : '3px solid transparent',
                opacity: isCurrentTurn && !playable ? 0.5 : 1,
                boxSizing: 'border-box',
              }}
            >
              <CardImage card={card} width={52} />
            </button>
          );
        })}
      </div>
    </div>
  );
}

export { HumanArea as SevensHumanArea };
