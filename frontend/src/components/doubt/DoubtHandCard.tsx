import type { Card } from '../../types/card';
import { CardImage } from '../CardImage';

interface HandCardProps {
  card: Card;
  index: number;
  selected: boolean;
  selectable: boolean;
  onToggle: (idx: number) => void;
}

export function DoubtHandCard({ card, index, selected, selectable, onToggle }: HandCardProps) {
  return (
    <button
      type="button"
      data-testid="hand-card"
      aria-pressed={selected}
      disabled={!selectable}
      onClick={() => onToggle(index)}
      style={{
        background: 'none',
        padding: 0,
        cursor: selectable ? 'pointer' : 'default',
        borderRadius: 8,
        border: selected ? '3px solid var(--color-game-status-active)' : '3px solid transparent',
        transform: selected ? 'translateY(-8px)' : 'none',
        transition: 'transform 0.15s, border 0.15s',
        opacity: !selectable ? 0.5 : 1,
        boxSizing: 'border-box',
      }}
    >
      <CardImage card={card} width={52} />
    </button>
  );
}
