import { CardBack } from './CardImage';

/** Props for the {@link CircularDeck} component. */
export interface CircularDeckProps {
  /** Number of face-down cards laid out around the ring. */
  count: number;
  /** Card width in pixels. */
  cardWidth: number;
  /** Diameter of the imaginary circle (px). */
  diameter?: number;
  /** Tap handler invoked when any card on the ring is clicked. */
  onDrawCard: () => void;
  /** Disables tap interaction. */
  disabled?: boolean;
  /** Localised aria-label for the tap targets, e.g. "Draw a card". */
  drawAriaLabel: string;
  /** Optional data-tutorial attribute for tutorial highlighting. */
  dataTutorial?: string;
}

const MIN_DIAMETER = 120;
const MAX_VISIBLE_CARDS = 26;
/** Card aspect ratio (height / width). Matches `CARD_NATURAL_HEIGHT / CARD_NATURAL_WIDTH` in `CardImage.tsx`. */
const CARD_ASPECT = 1.5;
/** WCAG 2.5.5 AAA minimum tap-target dimension (px). Per `frontend/CLAUDE.md`. */
const MIN_TAP_TARGET_PX = 44;

/**
 * Lays out `count` face-down cards in a ring so the player can tap any of them
 * to draw "that" card. The drawn card is decided by the backend; the ring is
 * purely an affordance to bring back the physical feel of the game.
 *
 * When the deck has fewer cards than the visual cap (`MAX_VISIBLE_CARDS`) the
 * ring renders one per card. Beyond the cap the cap itself is used to avoid
 * overlapping tap targets.
 */
export function CircularDeck({
  count,
  cardWidth,
  diameter = MIN_DIAMETER,
  onDrawCard,
  disabled = false,
  drawAriaLabel,
  dataTutorial,
}: CircularDeckProps) {
  const visible = Math.max(0, Math.min(count, MAX_VISIBLE_CARDS));
  const ringDiameter = Math.max(diameter, MIN_DIAMETER);
  const cardHeight = Math.round(cardWidth * CARD_ASPECT);

  if (visible === 0) {
    return (
      <div
        className="text-ds-text-muted text-sm text-center"
        data-tutorial={dataTutorial}
        data-testid="circular-deck-empty"
      >
        —
      </div>
    );
  }

  return (
    <div
      className="relative mx-auto"
      style={{ width: ringDiameter + cardWidth, height: ringDiameter + cardHeight }}
      data-tutorial={dataTutorial}
      data-testid="circular-deck"
    >
      {Array.from({ length: visible }, (_, i) => {
        const angle = (i / visible) * 2 * Math.PI - Math.PI / 2;
        // Shift the ring centre by half a card so cards at every cardinal
        // direction stay fully inside the container (the original calc placed
        // the centre at `ringDiameter/2` which clipped the left and top cards).
        const cx = cardWidth / 2 + (ringDiameter / 2) * (1 + Math.cos(angle));
        const cy = cardHeight / 2 + (ringDiameter / 2) * (1 + Math.sin(angle));
        return (
          <button
            type="button"
            key={`ring-${i}`}
            onClick={onDrawCard}
            disabled={disabled}
            aria-label={`${drawAriaLabel} #${i + 1}`}
            data-testid={`circular-deck-card-${i}`}
            className="absolute p-0 m-0 border-0 bg-transparent disabled:opacity-40 disabled:cursor-not-allowed transition-transform hover:scale-110 flex items-center justify-center"
            style={{
              left: cx,
              top: cy,
              // Pad the interactive area to meet WCAG 2.5.5 AAA (44×44 minimum)
              // even when the visible card art is smaller.
              minWidth: MIN_TAP_TARGET_PX,
              minHeight: MIN_TAP_TARGET_PX,
              transform: `translate(-50%, -50%) rotate(${(angle * 180) / Math.PI + 90}deg)`,
            }}
          >
            <CardBack width={cardWidth} />
          </button>
        );
      })}
    </div>
  );
}
