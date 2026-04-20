import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../../hooks/useCardDimensions';
import { playerAreaBase } from '../../styles/gameStyles';
import type { OldMaidPlayerData } from '../../types/card';
import { cardAlt } from '../../utils/cardAlt';
import { playerName } from '../../utils/playerUtils';
import { CardBack, CardImage } from '../CardImage';
import { StatusBadge } from '../StatusBadge';

/** CSS class for Old Maid player area layout. */
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
  /** When true, non-target CPU players hide card back images to save space. */
  compactNonTarget?: boolean;
}

/** Renders a player area for Old Maid with draw targets, hand display, and reorder support. */
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
  compactNonTarget,
}: PlayerAreaProps) {
  const { t } = useTranslation('oldmaid');
  const { t: tc } = useTranslation('common');
  const { cardWidth } = useCardDimensions();
  const [focusedCardIdx, setFocusedCardIdx] = useState<number | null>(null);
  const [selectedForMove, setSelectedForMove] = useState<number | null>(null);
  const cardCount = player.cards?.length ?? 0;

  useEffect(() => {
    if (focusedCardIdx !== null && focusedCardIdx >= cardCount) {
      setFocusedCardIdx(null);
    }
    if (selectedForMove !== null && selectedForMove >= cardCount) {
      setSelectedForMove(null);
    }
  }, [cardCount, focusedCardIdx, selectedForMove]);

  const handleCardTap = (i: number) => {
    if (!onReorder || !player.cards || gameEndFlag) return;
    if (selectedForMove === null) {
      setSelectedForMove(i);
      return;
    }
    if (selectedForMove === i) {
      setSelectedForMove(null);
      return;
    }
    const indices = player.cards.map((_, idx) => idx);
    const [moved] = indices.splice(selectedForMove, 1);
    indices.splice(i, 0, moved);
    onReorder(indices);
    setSelectedForMove(null);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!onReorder || !player.cards || player.cards.length === 0) return;
    const max = player.cards.length - 1;

    const swapAndReorder = (index1: number, index2: number) => {
      // biome-ignore lint/style/noNonNullAssertion: guard on line 52 ensures player.cards is non-null
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
            className="ml-1 px-1.5 py-0.5 text-xs rounded bg-ds-error/60 hover:bg-ds-error text-white"
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
          player.cards?.map((card, i) => {
            if (!onReorder) {
              return (
                <CardImage
                  key={`${card.design}-${card.value}`}
                  card={card}
                  width={cardWidth}
                  className={focusedCardIdx === i ? 'ring-2 ring-ds-info' : undefined}
                />
              );
            }
            const isSelectedForMove = selectedForMove === i;
            const ringClass = isSelectedForMove
              ? 'ring-2 ring-ds-warning -translate-y-2'
              : focusedCardIdx === i
                ? 'ring-2 ring-ds-info'
                : '';
            const tapAriaLabel = (() => {
              if (selectedForMove === null) {
                return t('reorder.selectCardToMove', { card: cardAlt(card) });
              }
              if (selectedForMove === i) {
                return t('reorder.deselectCard', { card: cardAlt(card) });
              }
              // biome-ignore lint/style/noNonNullAssertion: guard above checks player.cards
              const movingCard = player.cards![selectedForMove];
              return t('reorder.moveHere', { card: cardAlt(movingCard), to: i + 1 });
            })();
            return (
              <button
                key={`${card.design}-${card.value}`}
                type="button"
                onClick={() => handleCardTap(i)}
                aria-pressed={isSelectedForMove}
                aria-label={tapAriaLabel}
                className={`bg-transparent p-0 border-0 cursor-pointer leading-none transition-transform ${ringClass}`}
                draggable={!gameEndFlag}
                onDragStart={(e: React.DragEvent) => {
                  e.dataTransfer.setData('oldmaidCardIndex', String(i));
                }}
                onDragOver={(e: React.DragEvent) => e.preventDefault()}
                onDrop={(e: React.DragEvent) => {
                  e.preventDefault();
                  const fromStr = e.dataTransfer.getData('oldmaidCardIndex');
                  if (!fromStr || !player.cards) return;
                  const from = Number(fromStr);
                  if (from === i) return;
                  const indices = player.cards.map((_, idx) => idx);
                  indices.splice(from, 1);
                  indices.splice(i, 0, from);
                  onReorder(indices);
                  setSelectedForMove(null);
                }}
              >
                <CardImage card={card} width={cardWidth} />
              </button>
            );
          })
        ) : showSelectable ? (
          <>
            {Array.from({ length: showCount }, (_, i) => {
              const isHighlighted = isTarget && !player.isHuman && i === highlightedCardIdx;
              const cardStyle: React.CSSProperties = {
                border: isHighlighted ? '2px solid var(--color-ds-accent)' : '2px solid transparent',
                borderRadius: 4,
                cursor: 'pointer',
                transition: 'transform 0.2s, border-color 0.2s, box-shadow 0.2s',
                ...(isHighlighted
                  ? {
                      transform: 'translateY(-8px)',
                      boxShadow: 'var(--shadow-ds-accent-glow)',
                    }
                  : {}),
              };
              return (
                <CardBack
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
        ) : compactNonTarget && !isTarget && !player.isHuman ? null : (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              <CardBack key={i} width={cardWidth} />
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
