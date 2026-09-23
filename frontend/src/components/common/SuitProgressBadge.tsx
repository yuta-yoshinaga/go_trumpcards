/** The four card suits, drawn in canonical order to mark foundation progress. */
const SUIT_GLYPHS = ['♠', '♣', '♥', '♦'] as const;

/** Total number of suits a player must complete. */
const TOTAL_SUITS = SUIT_GLYPHS.length;

/** Props for {@link SuitProgressBadge}. */
export interface SuitProgressBadgeProps {
  /** Bit mask of completed suits, using the canonical ♠♣♥♦ bit order. */
  completedMask: number;
  /** Optional label rendered before the glyphs (e.g. "Completed"). */
  label?: string;
}

/**
 * Renders four suit glyphs that fill as foundations complete, giving solitaire
 * games (Scorpion, Wasp) a visual "N of 4 suits done" progress indicator in
 * place of plain `N/4` text.
 */
export function SuitProgressBadge({ completedMask, label }: SuitProgressBadgeProps) {
  const normalizedMask = completedMask & ((1 << TOTAL_SUITS) - 1);
  const filled = SUIT_GLYPHS.filter((_, i) => normalizedMask & (1 << i)).length;
  return (
    <span
      className="inline-flex items-center gap-1"
      data-testid="suit-progress"
      role="img"
      aria-label={`${filled}/${TOTAL_SUITS}`}
    >
      {label && <span className="text-ds-text-primary">{label}:</span>}
      {SUIT_GLYPHS.map((glyph, i) => (
        <span
          // Fixed-length canonical list; index is a stable key.
          key={glyph}
          data-testid={normalizedMask & (1 << i) ? 'suit-done' : 'suit-todo'}
          className={normalizedMask & (1 << i) ? 'text-ds-success' : 'text-ds-text-muted'}
        >
          {glyph}
        </span>
      ))}
    </span>
  );
}
