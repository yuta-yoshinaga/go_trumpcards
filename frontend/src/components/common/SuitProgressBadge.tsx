/** The four card suits, drawn in canonical order to mark foundation progress. */
const SUIT_GLYPHS = ['♠', '♣', '♥', '♦'] as const;

/** Total number of suits a player must complete. */
const TOTAL_SUITS = SUIT_GLYPHS.length;

/** Props for {@link SuitProgressBadge}. */
export interface SuitProgressBadgeProps {
  /**
   * Number of completed suits (0–4). The first `completed` glyphs are filled
   * (`text-ds-success`); the rest are shown muted. The count is suit-agnostic —
   * games such as Scorpion/Wasp expose only a tally, so glyphs fill in canonical
   * ♠♣♥♦ order rather than by which specific suits finished.
   */
  completed: number;
  /** Optional label rendered before the glyphs (e.g. "Completed"). */
  label?: string;
}

/**
 * Renders four suit glyphs that fill as foundations complete, giving solitaire
 * games (Scorpion, Wasp) a visual "N of 4 suits done" progress indicator in
 * place of plain `N/4` text.
 */
export function SuitProgressBadge({ completed, label }: SuitProgressBadgeProps) {
  const filled = Math.max(0, Math.min(completed, TOTAL_SUITS));
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
          data-testid={i < filled ? 'suit-done' : 'suit-todo'}
          className={i < filled ? 'text-ds-success' : 'text-ds-text-muted'}
        >
          {glyph}
        </span>
      ))}
    </span>
  );
}
