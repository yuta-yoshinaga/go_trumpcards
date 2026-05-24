/** Props for {@link CardRoleBadge}. */
export interface CardRoleBadgeProps {
  /** Card index in the surrounding hand — feeds the `data-testid`. */
  idx: number;
  /** Short visual glyph rendered inside the badge (emoji or letter). */
  glyph: string;
  /** Tooltip text (also exposed via `title`). */
  title: string;
}

/** Small top-left corner badge identifying a card's special role within a
 * game (e.g. Mighty's 👑 marker on ♠A). Shared between MobileHandGrid and
 * any desktop hand renderer so both layouts get the same overlay. */
export function CardRoleBadge({ idx, glyph, title }: CardRoleBadgeProps) {
  return (
    <span
      data-testid={`card-role-badge-${idx}`}
      title={title}
      className="absolute top-0 left-0 bg-black/70 text-white rounded-br rounded-tl px-1 text-[10px] leading-tight pointer-events-none"
    >
      {glyph}
    </span>
  );
}
