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
  const cardHeight = Math.round(cardWidth * 1.4);

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
        const cx = ringDiameter / 2 + (ringDiameter / 2) * Math.cos(angle);
        const cy = ringDiameter / 2 + (ringDiameter / 2) * Math.sin(angle);
        return (
          <button
            type="button"
            key={`ring-${i}`}
            onClick={onDrawCard}
            disabled={disabled}
            aria-label={`${drawAriaLabel} #${i + 1}`}
            data-testid={`circular-deck-card-${i}`}
            className="absolute p-0 m-0 border-0 bg-transparent disabled:opacity-40 disabled:cursor-not-allowed transition-transform hover:scale-110"
            style={{
              left: cx,
              top: cy,
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
