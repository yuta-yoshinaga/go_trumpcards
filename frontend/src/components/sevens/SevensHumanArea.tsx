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
  tunnelSkipWidth: number;
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
  tunnelSkipWidth,
  noJokerFinish,
  endStopEnabled,
  jokerConsecutiveBanned,
  loading,
  onPlay,
}: HumanAreaProps) {
  const { t } = useTranslation('sevens');
  const conditionalClass = player.isFinished
    ? 'opacity-50'
    : isCurrentTurn
      ? 'border-2 border-game-status-active shadow-[0_0_12px_var(--color-game-status-active)]'
      : '';
  return (
    <div className={`${playerAreaClass}${conditionalClass ? ` ${conditionalClass}` : ''}`}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && <StatusBadge variant="success">{t('rankLabel', { rank: player.rank })}</StatusBadge>}
      </div>
      {!player.isFinished && (
        <div className="text-game-text-muted text-[0.85em] mb-1">
          {t('cardCount', { count: player.cardCount })}
          {'　'}
          {t('passCount', {
            used: player.passesUsed,
            max: player.maxPasses === 0 ? t('passUnlimited') : player.maxPasses,
          })}
          {isCurrentTurn && <span className="ml-2 text-game-text-highlight">{t('clickPlayable')}</span>}
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
              tunnelSkipWidth,
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
                border: playable ? '3px solid var(--color-game-status-active)' : '3px solid transparent',
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
