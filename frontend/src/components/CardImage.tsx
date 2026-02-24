import type { Card } from '../types/card';

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
}

export function CardImage({ card, width, style, className }: CardImageProps) {
  return (
    <img
      src={getImagePath(card)}
      alt={`${card.design} ${card.value}`}
      style={{ width: width ?? 80, borderRadius: 6, display: 'block', ...style }}
      className={className}
    />
  );
}

interface CardBackProps {
  width?: number;
  style?: React.CSSProperties;
  className?: string;
  onClick?: () => void;
}

export function CardBack({ width, style, className, onClick }: CardBackProps) {
  const img = (
    <img
      src="/images/z01.png"
      alt="card back"
      style={{ width: width ?? 80, borderRadius: 6, display: 'block', ...style }}
      className={className}
    />
  );
  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        style={{ background: 'none', border: 'none', padding: 0, cursor: 'pointer', lineHeight: 0 }}
      >
        {img}
      </button>
    );
  }
  return img;
}
