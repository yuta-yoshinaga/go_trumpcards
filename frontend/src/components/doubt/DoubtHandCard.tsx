import { useCardDimensions } from '../../hooks/useCardDimensions';
import { focusRingCard, selectedCardStyle } from '../../styles/cardStyles';
import type { Card } from '../../types/card';
import { CardImage } from '../CardImage';

/** Props for {@link DoubtHandCard}. */
export interface HandCardProps {
  card: Card;
  index: number;
  selected: boolean;
  selectable: boolean;
  onToggle: (idx: number) => void;
  onSwipeStart?: (idx: number) => void;
}

/** Renders a selectable hand card for Doubt with selection highlight. */
export function DoubtHandCard({ card, index, selected, selectable, onToggle, onSwipeStart }: HandCardProps) {
  const { cardWidth } = useCardDimensions();
  return (
    <button
      type="button"
      data-testid="hand-card"
      data-card-index={index}
      aria-pressed={selected}
      disabled={!selectable}
      className={`${focusRingCard} touch-none`}
      onClick={() => onToggle(index)}
      onPointerDown={selectable && onSwipeStart ? () => onSwipeStart(index) : undefined}
      style={{
        background: 'none',
        padding: 0,
        cursor: selectable ? 'pointer' : 'default',
        borderRadius: 8,
        ...selectedCardStyle(selected),
        opacity: !selectable ? 0.5 : 1,
        boxSizing: 'border-box',
      }}
    >
      <CardImage card={card} width={cardWidth} />
    </button>
  );
}
