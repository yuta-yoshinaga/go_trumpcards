import type { Card } from '../types/card';

/**
 * Maps a backend color token to a concrete card-ink hex color. Procedural
 * cards use a white face (like the PNG art) with colored ink, so they read as
 * cards alongside the standard 52-card images in either theme.
 */
const INK_COLORS: Record<string, string> = {
  red: '#B83A3A',
  black: '#1A1A1A',
  purple: '#7C3AED',
  green: '#2E7D46',
  blue: '#1A2C5C',
  gold: '#B8892E',
};

/** Resolve a color token to an ink hex, defaulting to near-black. */
function inkColor(color?: string): string {
  return (color && INK_COLORS[color]) || INK_COLORS.black;
}

/** Props for {@link CardFace}. */
export interface CardFaceProps {
  card: Card;
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  onDragOver?: (e: React.DragEvent) => void;
  onDrop?: (e: React.DragEvent) => void;
}

/** Card aspect ratio (2:3), matching the 250x375 PNG assets. */
const CARD_ASPECT = 1.5;

/** Suppresses iOS Safari long-press callout and text selection. */
const noCalloutStyle = {
  WebkitTouchCallout: 'none',
  WebkitUserSelect: 'none',
  userSelect: 'none',
} as React.CSSProperties;

/**
 * Renders a face-up card procedurally (CSS/SVG) from its self-describing
 * descriptor (`glyph`/`label`/`color`), for non-52 decks that have no PNG art
 * (tarot, hanafuda, kabu, Wizard, …). White card face with colored ink so it
 * harmonizes with the standard PNG cards. See ADR-0033 and `CardImage`, which
 * dispatches here when a card carries a `deck`.
 */
export function CardFace({ card, width, style, className, draggable, onDragStart, onDragOver, onDrop }: CardFaceProps) {
  const w = width ?? 80;
  const h = w * CARD_ASPECT;
  const ink = inkColor(card.color);
  const label = card.label ?? String(card.value);
  const glyph = card.glyph ?? label.charAt(0);
  const ariaLabel = card.glyph ? `${label} ${card.glyph}` : label;

  const cornerBase: React.CSSProperties = {
    position: 'absolute',
    fontSize: Math.max(9, w * 0.19),
    fontWeight: 700,
    lineHeight: 1,
    letterSpacing: '-0.02em',
  };

  return (
    <div
      role="img"
      aria-label={ariaLabel}
      className={className}
      draggable={draggable}
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDrop={onDrop}
      style={{
        position: 'relative',
        width: w,
        height: h,
        maxWidth: '100%',
        borderRadius: 6,
        background: '#FFFFFF',
        border: `1px solid ${ink}33`,
        boxShadow: '0 1px 2px rgba(0,0,0,0.25)',
        color: ink,
        display: 'block',
        overflow: 'hidden',
        fontFamily: "'Fraunces', Georgia, serif",
        ...noCalloutStyle,
        ...style,
      }}
    >
      <span style={{ ...cornerBase, top: w * 0.06, left: w * 0.08 }}>{label}</span>
      <span
        aria-hidden="true"
        style={{
          position: 'absolute',
          inset: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: Math.max(18, w * 0.5),
          lineHeight: 1,
        }}
      >
        {glyph}
      </span>
      <span
        style={{
          ...cornerBase,
          bottom: w * 0.06,
          right: w * 0.08,
          transform: 'rotate(180deg)',
        }}
      >
        {label}
      </span>
    </div>
  );
}
