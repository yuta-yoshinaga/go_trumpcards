import type { Card } from '../types/card'

function getImagePath(card: Card): string {
  const zeroPad = (n: number) => String(n).padStart(2, '0')
  const prefixMap: Record<string, string> = {
    SPADE: 's',
    CLOVER: 'c',
    HEART: 'h',
    DIAMOND: 'd',
    JOKER: 'x',
  }
  const prefix = prefixMap[card.design] ?? 'x'
  return `/images/${prefix}${zeroPad(card.value)}.png`
}

interface CardImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  card: Card
  style?: React.CSSProperties
  className?: string
}

export function CardImage({ card, style, className, ...props }: CardImageProps) {
  return (
    <img
      {...props}
      src={getImagePath(card)}
      alt={`${card.design} ${card.value}`}
      style={{ width: 80, borderRadius: 6, display: 'block', ...style }}
      className={className}
    />
  )
}

interface CardBackProps {
  style?: React.CSSProperties
  className?: string
  onClick?: () => void
}

export function CardBack({ style, className, onClick }: CardBackProps) {
  return (
    <img
      src="/images/z01.png"
      alt="card back"
      style={{ width: 80, borderRadius: 6, display: 'block', cursor: onClick ? 'pointer' : undefined, ...style }}
      className={className}
      onClick={onClick}
    />
  )
}
