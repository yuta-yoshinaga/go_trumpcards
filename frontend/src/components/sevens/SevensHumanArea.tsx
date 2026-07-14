import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../../hooks/useCardDimensions';
import { focusRingWhite } from '../../styles/buttonStyles';
import { playableCardStyle } from '../../styles/cardStyles';
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
  const { cardWidth } = useCardDimensions();
  // Whether the human has at least one legal move this turn (cards or a joker placement).
  const anyPlayable =
    isCurrentTurn &&
    !loading &&
    (player.cards ?? []).some((card) =>
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
      ),
    );
  const conditionalClass = player.isFinished
    ? 'opacity-50'
    : isCurrentTurn
      ? 'border-2 border-game-status-active shadow-[0_0_12px_var(--color-game-status-active)]'
      : '';
  return (
    <div className={`${playerAreaClass}${conditionalClass ? ` ${conditionalClass}` : ''}`}>
      <div className="text-ds-text-primary font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && <StatusBadge variant="success">{t('rankLabel', { rank: player.rank })}</StatusBadge>}
      </div>
      {!player.isFinished && (
        <div className="text-game-text-muted text-xs mb-1">
          {t('cardCount', { count: player.cardCount })}
          {'　'}
          {t('passCount', {
            used: player.passesUsed,
            max: player.maxPasses === 0 ? t('passUnlimited') : player.maxPasses,
          })}
          {isCurrentTurn &&
            (anyPlayable ? (
              <span className="ml-2 text-game-text-highlight">{t('clickPlayable')}</span>
            ) : (
              <span className="ml-2 text-ds-warning" data-testid="sv-must-pass">
                {t('mustPass')}
              </span>
            ))}
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
              className={focusRingWhite}
              disabled={!playable}
              onClick={() => onPlay(i)}
              title={playable ? t('playTitle', { design: card.design, value: valueName(card.value) }) : undefined}
              // Number keys 1-9 (and 0 for the 10th card) directly play the matching
              // card (useCardKeyboardNav maps digit 0 → index 9); advertise the
              // shortcut only where it is a legal move this turn.
              aria-keyshortcuts={playable && i <= 9 ? String((i + 1) % 10) : undefined}
              data-testid={playable ? 'playable-card' : undefined}
              style={{
                background: 'none',
                padding: 0,
                cursor: playable ? 'pointer' : 'default',
                borderRadius: 8,
                ...playableCardStyle(playable),
                opacity: isCurrentTurn && !playable ? 0.5 : 1,
                boxSizing: 'border-box',
              }}
            >
              <CardImage card={card} width={cardWidth} />
            </button>
          );
        })}
      </div>
    </div>
  );
}

/** Renders the human player area for Sevens with playable card highlighting. */
export { HumanArea as SevensHumanArea };
