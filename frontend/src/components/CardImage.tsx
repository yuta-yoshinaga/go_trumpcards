import { useTranslation } from 'react-i18next';
import { focusRingWhite } from '../styles/buttonStyles';
import type { Card } from '../types/card';
import { cardAlt } from '../utils/cardAlt';

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

interface CardImageProps {
  card: Card;
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  onDragOver?: (e: React.DragEvent) => void;
  onDrop?: (e: React.DragEvent) => void;
}

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
  return (
    <img
      src={getImagePath(card)}
      alt={cardAlt(card)}
      style={{ width: width ?? 80, maxWidth: '100%', borderRadius: 6, display: 'block', ...style }}
      className={className}
      draggable={draggable}
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDrop={onDrop}
    />
  );
}

interface CardBackProps {
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  onClick?: () => void;
  /** Only applies when onClick is provided (button mode). */
  ariaLabel?: string;
}

export function CardBack({ width, style, className, onClick, ariaLabel }: CardBackProps) {
  const { t } = useTranslation('common');
  const effectiveAriaLabel = onClick ? ariaLabel || t('card.back') : undefined;
  const img = (
    <img
      src="/images/z01.png"
      alt={onClick ? '' : t('card.back')}
      style={{ width: width ?? 80, maxWidth: '100%', borderRadius: 6, display: 'block', ...style }}
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
