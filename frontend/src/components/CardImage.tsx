import { useTranslation } from 'react-i18next';
import { focusRingWhite } from '../styles/buttonStyles';
import type { Card } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { CardFace } from './CardFace';

/** True when a card carries a non-52 deck descriptor and must render procedurally. */
function isProceduralCard(card: Card): boolean {
  return Boolean(card.deck || card.glyph || card.label);
}

function getImagePath(card: Card): string {
  const zeroPad = (n: number) => String(n).padStart(2, '0');
  const prefixMap: Record<string, string> = {
    SPADE: 's',
    CLOVER: 'c',
    HEART: 'h',
    DIAMOND: 'd',
    JOKER: 'x',
  };
  const prefix = prefixMap[card.design] ?? 'x';
  return `/images/${prefix}${zeroPad(card.value)}.png`;
}

/** Props for {@link CardImage}. */
export interface CardImageProps {
  card: Card;
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  onDragOver?: (e: React.DragEvent) => void;
  onDrop?: (e: React.DragEvent) => void;
}

/** Natural card image dimensions (250x375 PNG assets, 2:3 aspect ratio). */
const CARD_NATURAL_WIDTH = 250;
const CARD_NATURAL_HEIGHT = 375;

/** Suppresses iOS Safari long-press callout and text selection on card images. */
const noCalloutStyle = {
  WebkitTouchCallout: 'none',
  WebkitUserSelect: 'none',
  userSelect: 'none',
} as React.CSSProperties;

/** Renders a face-up playing card image. */
export function CardImage({
  card,
  width,
  style,
  className,
  draggable,
  onDragStart,
  onDragOver,
  onDrop,
}: CardImageProps) {
  const w = width ?? 80;
  if (isProceduralCard(card)) {
    return (
      <CardFace
        card={card}
        width={width}
        style={style}
        className={className}
        draggable={draggable}
        onDragStart={onDragStart}
        onDragOver={onDragOver}
        onDrop={onDrop}
      />
    );
  }
  return (
    <img
      src={getImagePath(card)}
      alt={cardAlt(card)}
      width={CARD_NATURAL_WIDTH}
      height={CARD_NATURAL_HEIGHT}
      loading="lazy"
      style={{
        width: w,
        maxWidth: '100%',
        borderRadius: 6,
        display: 'block',
        ...noCalloutStyle,
        ...style,
      }}
      className={className}
      draggable={draggable}
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDrop={onDrop}
    />
  );
}

/** Props for {@link CardBack}. */
export interface CardBackProps {
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  onClick?: () => void;
  /** Only applies when onClick is provided (button mode). */
  ariaLabel?: string;
}

/** Renders a face-down card back image, optionally as a clickable button. */
export function CardBack({ width, style, className, onClick, ariaLabel }: CardBackProps) {
  const { t } = useTranslation('common');
  const effectiveAriaLabel = onClick ? ariaLabel || t('card.back') : undefined;
  const w = width ?? 80;
  const img = (
    <img
      src="/images/z01.png"
      alt={onClick ? '' : t('card.back')}
      width={CARD_NATURAL_WIDTH}
      height={CARD_NATURAL_HEIGHT}
      loading="lazy"
      style={{
        width: w,
        maxWidth: '100%',
        borderRadius: 6,
        display: 'block',
        border: '1px solid var(--color-ds-card-back-border)',
        boxShadow: 'var(--shadow-ds-card-back)',
        ...noCalloutStyle,
        ...style,
      }}
      className={className}
    />
  );
  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        aria-label={effectiveAriaLabel}
        className={`${focusRingWhite} rounded-md`}
        style={{ background: 'none', border: 'none', padding: 0, cursor: 'pointer', lineHeight: 0 }}
      >
        {img}
      </button>
    );
  }
  return img;
}
