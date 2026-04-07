import { useWindowWidth } from '../hooks/useCardDimensions';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { expansionMargin, focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import type { Card } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { CardImage } from './CardImage';

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
/** Minimum visible width (px) per card to meet WCAG 2.5.8 tap-target guidelines. */
const MIN_CARD_EXPOSURE_PX = 44;

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
  const reduced = useReducedMotion();
  const buttonWidth = cardWidth + BUTTON_EXTRA;

  const useTwoRows = cards.length >= TWO_ROW_THRESHOLD;
  const splitAt = useTwoRows ? Math.ceil(cards.length / 2) : cards.length;
  const rows = useTwoRows ? [cards.slice(0, splitAt), cards.slice(splitAt)] : [cards];

  return (
    <div className="mb-2" data-tutorial={dataTutorial}>
      {rows.map((rowCards, rowIdx) => {
        const { overlap, useScroll } = computeOverlap(rowCards.length, buttonWidth, viewportWidth);
        const startIdx = rowIdx === 0 ? 0 : splitAt;

        return (
          <div
            key={`row-${rowIdx}`}
            data-testid="hand-row"
            className={useScroll ? 'flex overflow-x-auto' : 'flex justify-center'}
            style={{ marginBottom: rowIdx === 0 && rows.length > 1 ? 4 : 0 }}
          >
            {rowCards.map((card, i) => {
              const globalIdx = startIdx + i;
              const isSelected = selectedIndices.includes(globalIdx);
              const isExpanded =
                selectedIndices.includes(globalIdx) || (i > 0 && selectedIndices.includes(globalIdx - 1));
              const ml = i === 0 ? 0 : isExpanded ? expansionMargin(true, overlap) : overlap;
              return (
                <button
                  type="button"
                  key={`${card.design}-${card.value}-${globalIdx}`}
                  onClick={() => onToggle(globalIdx)}
                  aria-label={cardAlt(card)}
                  aria-pressed={isSelected}
                  className={focusRingCard}
                  style={{
                    background: 'none',
                    padding: 0,
                    borderRadius: 8,
                    ...selectedCardStyle(isSelected),
                    transition: 'transform 0.15s, border 0.15s, box-shadow 0.15s, margin-left 0.15s',
                    boxSizing: 'border-box',
                    marginLeft: ml,
                    ...(useScroll ? { flexShrink: 0 } : {}),
                  }}
                >
                  <CardImage
                    card={card}
                    width={cardWidth}
                    className={reduced ? undefined : 'animate-card-deal-in'}
                    style={reduced ? undefined : { animationDelay: `${i * 0.12}s` }}
                  />
                </button>
              );
            })}
          </div>
        );
      })}
    </div>
  );
}

/** Result of overlap computation including scroll-fallback flag. */
interface OverlapResult {
  /** Negative margin-left in px (or positive gap). */
  overlap: number;
  /** True when cards cannot fit with minimum tap-target exposure. */
  useScroll: boolean;
}

/** Compute negative margin-left overlap so cards fit within the viewport. */
function computeOverlap(cardCount: number, buttonWidth: number, viewportWidth: number): OverlapResult {
  if (cardCount <= 1) return { overlap: 0, useScroll: false };
  const availableWidth = viewportWidth - CONTAINER_PADDING;
  const totalNeeded = cardCount * buttonWidth;
  if (totalNeeded <= availableWidth) return { overlap: DEFAULT_CARD_GAP, useScroll: false };
  const rawOverlap = -((totalNeeded - availableWidth) / (cardCount - 1));
  const maxNegative = -buttonWidth * MAX_OVERLAP_RATIO;
  const minExposureOverlap = -(buttonWidth - MIN_CARD_EXPOSURE_PX);
  if (rawOverlap < minExposureOverlap) {
    return { overlap: DEFAULT_CARD_GAP, useScroll: true };
  }
  return { overlap: Math.max(rawOverlap, maxNegative), useScroll: false };
}
