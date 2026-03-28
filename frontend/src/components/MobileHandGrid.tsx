import { useWindowWidth } from '../hooks/useCardDimensions';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import type { Card } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { AnimatedCard } from './motion/AnimatedCard';

/** Minimum number of cards before splitting into 2 rows. */
const TWO_ROW_THRESHOLD = 4;
/** Horizontal padding (px) on the hand container (px-4 = 16px each side). */
const CONTAINER_PADDING = 32;
/** Base width of each card button including border (cardWidth + 6px border). */
const BUTTON_EXTRA = 6;
/** Default gap (px) between cards when they fit without overlap. */
const DEFAULT_CARD_GAP = 2;
/** Maximum overlap ratio — each card stays at least 30% visible. */
const MAX_OVERLAP_RATIO = 0.7;

/** Props for the MobileHandGrid component. */
interface MobileHandGridProps {
  /** Cards to display. */
  cards: Card[];
  /** Indices of currently selected cards. */
  selectedIndices: number[];
  /** Callback when a card is tapped. */
  onToggle: (idx: number) => void;
  /** Card image width in pixels. */
  cardWidth: number;
  /** Optional data-tutorial attribute for tutorial system. */
  dataTutorial?: string;
}

/**
 * Renders the player's hand in a mobile-friendly 2-row grid layout.
 * Dynamically calculates negative overlap so all cards fit within the viewport.
 * Falls back to a single row when 3 or fewer cards are present.
 */
export function MobileHandGrid({ cards, selectedIndices, onToggle, cardWidth, dataTutorial }: MobileHandGridProps) {
  const viewportWidth = useWindowWidth();
  const buttonWidth = cardWidth + BUTTON_EXTRA;

  const useTwoRows = cards.length >= TWO_ROW_THRESHOLD;
  const splitAt = useTwoRows ? Math.ceil(cards.length / 2) : cards.length;
  const rows = useTwoRows ? [cards.slice(0, splitAt), cards.slice(splitAt)] : [cards];

  return (
    <div className="mb-2" data-tutorial={dataTutorial}>
      {rows.map((rowCards, rowIdx) => {
        const overlap = computeOverlap(rowCards.length, buttonWidth, viewportWidth);
        const startIdx = rowIdx === 0 ? 0 : splitAt;

        return (
          <div
            key={`row-${rowIdx}`}
            data-testid="hand-row"
            className="flex justify-center"
            style={{ marginBottom: rowIdx === 0 && rows.length > 1 ? 4 : 0 }}
          >
            {rowCards.map((card, i) => {
              const globalIdx = startIdx + i;
              const isSelected = selectedIndices.includes(globalIdx);
              return (
                <button
                  type="button"
                  key={`${card.design}-${card.value}-${globalIdx}`}
                  onClick={() => onToggle(globalIdx)}
                  aria-label={cardAlt(card)}
                  aria-pressed={isSelected}
                  className={`transition-transform ${focusRingCard}`}
                  style={{
                    background: 'none',
                    padding: 0,
                    borderRadius: 8,
                    ...selectedCardStyle(isSelected),
                    boxSizing: 'border-box',
                    marginLeft: i === 0 ? 0 : overlap,
                  }}
                >
                  <AnimatedCard card={card} width={cardWidth} />
                </button>
              );
            })}
          </div>
        );
      })}
    </div>
  );
}

/** Compute negative margin-left overlap so cards fit within the viewport. */
function computeOverlap(cardCount: number, buttonWidth: number, viewportWidth: number): number {
  if (cardCount <= 1) return 0;
  const availableWidth = viewportWidth - CONTAINER_PADDING;
  const totalNeeded = cardCount * buttonWidth;
  if (totalNeeded <= availableWidth) return DEFAULT_CARD_GAP;
  const rawOverlap = -((totalNeeded - availableWidth) / (cardCount - 1));
  return Math.max(rawOverlap, -buttonWidth * MAX_OVERLAP_RATIO);
}
