import { useReducedMotion } from '../../hooks/useReducedMotion';
import type { Card } from '../../types/card';
import { CardImage } from '../CardImage';
import { AnimatedCard } from './AnimatedCard';

/** Cascade batch size: max concurrent animations per group. */
const CASCADE_BATCH_SIZE = 5;

/** Delay between cascade batches in seconds. */
const CASCADE_BATCH_DELAY = 0.4;

/** Stagger between cards within a batch in seconds. */
const CASCADE_CARD_STAGGER = 0.08;

/** Props for {@link AnimatedPile}. */
export interface AnimatedPileProps {
  /** Cards in this pile (bottom to top). */
  cards: Card[];
  /** Layout mode: stacked (overlapping) or fanned (spread). */
  layout: 'stacked' | 'fanned';
  /** Card width in pixels. */
  cardWidth?: number;
  /** Callback fired when each card lands in the pile. */
  onPlace?: () => void;
  /** Enable cascade mode for auto-complete sequences. */
  cascade?: boolean;
  /** Callback fired after all cards in cascade have landed. */
  onComplete?: () => void;
}

/** Computes the deal delay for a card in cascade mode, batching for performance. */
function getCascadeDelay(index: number): number {
  const batchIndex = Math.floor(index / CASCADE_BATCH_SIZE);
  const positionInBatch = index % CASCADE_BATCH_SIZE;
  return batchIndex * CASCADE_BATCH_DELAY + positionInBatch * CASCADE_CARD_STAGGER;
}

/**
 * Renders a pile of animated cards with stacked or fanned layout.
 * In cascade mode, batches animations in groups of 5 with inter-group delays
 * to stay within the mobile performance budget.
 */
export function AnimatedPile({ cards, layout, cardWidth, onPlace, cascade = false, onComplete }: AnimatedPileProps) {
  const reduced = useReducedMotion();

  if (reduced) {
    return (
      <div className="relative" data-testid="animated-pile">
        {cards.map((card, i) => (
          <div
            key={`${card.design}-${card.value}-${i}`}
            style={layout === 'fanned' ? { marginTop: i > 0 ? -(cardWidth ?? 0) * 0.6 : 0 } : undefined}
          >
            <CardImage card={card} width={cardWidth} />
          </div>
        ))}
      </div>
    );
  }

  const lastIndex = cards.length - 1;

  return (
    <div className="relative" data-testid="animated-pile">
      {cards.map((card, i) => {
        const delay = cascade ? getCascadeDelay(i) : i * 0.1;
        const isLast = i === lastIndex;

        return (
          <div
            key={`${card.design}-${card.value}-${i}`}
            style={layout === 'fanned' ? { marginTop: i > 0 ? -(cardWidth ?? 0) * 0.6 : 0 } : undefined}
          >
            <AnimatedCard
              card={card}
              width={cardWidth}
              dealDelay={delay}
              onDealComplete={() => {
                onPlace?.();
                if (isLast && cascade) {
                  onComplete?.();
                }
              }}
            />
          </div>
        );
      })}
    </div>
  );
}
