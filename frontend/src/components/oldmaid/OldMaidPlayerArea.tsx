import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../../hooks/useCardDimensions';
import { playerAreaBase } from '../../styles/gameStyles';
import type { OldMaidPlayerData } from '../../types/card';
import { playerName } from '../../utils/playerUtils';
import { CardBack, CardImage } from '../CardImage';
import { StatusBadge } from '../StatusBadge';

export const playerAreaClass = `${playerAreaBase} p-2 flex-[1_1_140px] min-w-[120px]`;

interface PlayerAreaProps {
  player: OldMaidPlayerData;
  isTarget: boolean;
  isHumanTurn: boolean;
  gameEndFlag: boolean;
  loading: boolean;
  highlightedCardIdx: number;
  isSuspect?: boolean;
  onToggleSuspect?: () => void;
  onDraw: (drawIdx: number) => void;
  onReorder?: (indices: number[]) => void;
}

export function OldMaidPlayerArea({
  player,
  isTarget,
  isHumanTurn,
  gameEndFlag,
  loading,
  highlightedCardIdx,
  isSuspect,
  onToggleSuspect,
  onDraw,
  onReorder,
}: PlayerAreaProps) {
  const { t } = useTranslation('oldmaid');
  const { t: tc } = useTranslation('common');
  const { cardWidth } = useCardDimensions();
  const [focusedCardIdx, setFocusedCardIdx] = useState<number | null>(null);
  const cardCount = player.cards?.length ?? 0;

  useEffect(() => {
    if (focusedCardIdx !== null && focusedCardIdx >= cardCount) {
      setFocusedCardIdx(null);
    }
  }, [cardCount, focusedCardIdx]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!onReorder || !player.cards || player.cards.length === 0) return;
    const max = player.cards.length - 1;

    const swapAndReorder = (index1: number, index2: number) => {
      const indices = Array.from(player.cards!.keys());
      [indices[index1], indices[index2]] = [indices[index2], indices[index1]];
      onReorder(indices);
      setFocusedCardIdx(index2);
    };

    switch (e.key) {
      case 'Escape':
        setFocusedCardIdx(null);
        break;
      case 'ArrowLeft':
        e.preventDefault();
        if (e.shiftKey) {
          if (focusedCardIdx !== null && focusedCardIdx > 0) {
            swapAndReorder(focusedCardIdx, focusedCardIdx - 1);
          }
        } else {
          setFocusedCardIdx((prev) => (prev === null ? 0 : Math.max(0, prev - 1)));
        }
        break;
      case 'ArrowRight':
        e.preventDefault();
        if (e.shiftKey) {
          if (focusedCardIdx !== null && focusedCardIdx < max) {
            swapAndReorder(focusedCardIdx, focusedCardIdx + 1);
          }
        } else {
          setFocusedCardIdx((prev) => (prev === null ? 0 : Math.min(max, prev + 1)));
        }
        break;
      default:
        break;
    }
  };

  const conditionalClass = player.isFinished
    ? 'opacity-50'
    : isSuspect
      ? 'border-2 border-game-status-out shadow-[0_0_12px_var(--color-game-status-out)]'
      : isTarget && !gameEndFlag
        ? 'border-2 border-game-status-waiting shadow-[0_0_12px_var(--color-game-status-waiting)]'
        : '';

  const showSelectable = isHumanTurn && !loading && isTarget && !player.isFinished && !player.isHuman && !gameEndFlag;
  const showCount = Math.min(player.cardCount, 10);

  return (
    <div
      id={`player-area-${player.id}`}
      className={`${playerAreaClass}${conditionalClass ? ` ${conditionalClass}` : ''}`}
    >
      <div className="text-white font-bold mb-1 text-sm">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && <StatusBadge variant="success">{tc('status.finished')}</StatusBadge>}
        {isTarget && !player.isHuman && !player.isFinished && !gameEndFlag && (
          <StatusBadge variant="warning">{t('drawTarget')}</StatusBadge>
        )}
        {isSuspect && <StatusBadge variant="danger">{t('suspect.badge')}</StatusBadge>}
        {onToggleSuspect && !player.isFinished && !gameEndFlag && (
          <button
            type="button"
            className="ml-1 px-1.5 py-0.5 text-xs rounded bg-red-700/60 hover:bg-red-700 text-white"
            onClick={onToggleSuspect}
          >
            {isSuspect ? t('suspect.unpin') : t('suspect.pin')}
          </button>
        )}
      </div>
      {!player.isFinished && (
        <div className="text-game-text-muted text-xs mb-1">{t('cardCount', { count: player.cardCount })}</div>
      )}
      {showSelectable && !player.isFinished && <div className="text-game-text-highlight text-xs mb-1">{t('draw')}</div>}
      <div
        className="flex flex-wrap gap-0.5 justify-center"
        {...(player.isHuman && !player.isFinished && onReorder
          ? { tabIndex: 0, onKeyDown: handleKeyDown, 'data-testid': 'human-card-container' }
          : {})}
      >
        {player.isFinished ? null : player.isHuman ? (
          player.cards?.map((card, i) => (
            <CardImage
              key={`${card.design}-${card.value}`}
              card={card}
              width={cardWidth}
              className={focusedCardIdx === i && onReorder ? 'ring-2 ring-blue-500' : undefined}
              draggable={!gameEndFlag && !!onReorder}
              onDragStart={(e: React.DragEvent) => {
                e.dataTransfer.setData('oldmaidCardIndex', String(i));
              }}
              onDragOver={(e: React.DragEvent) => e.preventDefault()}
              onDrop={(e: React.DragEvent) => {
                e.preventDefault();
                const fromStr = e.dataTransfer.getData('oldmaidCardIndex');
                if (!fromStr || !onReorder || !player.cards) return;
                const from = Number(fromStr);
                if (from === i) return;
                const indices = player.cards.map((_, idx) => idx);
                indices.splice(from, 1);
                indices.splice(i, 0, from);
                onReorder(indices);
              }}
            />
          ))
        ) : showSelectable ? (
          <>
            {Array.from({ length: showCount }, (_, i) => {
              const isHighlighted = isTarget && !player.isHuman && i === highlightedCardIdx;
              const cardStyle: React.CSSProperties = {
                border: '2px solid transparent',
                borderRadius: 4,
                cursor: 'pointer',
                ...(isHighlighted ? { transform: 'translateY(-8px)', transition: 'transform 0.2s' } : {}),
              };
              return (
                <CardBack
                  // biome-ignore lint/suspicious/noArrayIndexKey: placeholder array with no card identity
                  key={i}
                  width={cardWidth}
                  style={cardStyle}
                  onClick={() => onDraw(i)}
                  ariaLabel={t('drawCardAriaLabel', { idx: i + 1 })}
                />
              );
            })}
            {player.cardCount > 10 && (
              <span className="text-white self-center ml-0.5 text-xs">+{player.cardCount - 10}</span>
            )}
          </>
        ) : (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: placeholder array with no card identity
              <CardBack key={i} width={40} />
            ))}
            {player.cardCount > 10 && (
              <span className="text-white self-center ml-0.5 text-xs">+{player.cardCount - 10}</span>
            )}
          </>
        )}
      </div>
    </div>
  );
}
