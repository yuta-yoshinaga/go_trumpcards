import { SkeletonCard } from './SkeletonCard';

interface SkeletonHandProps {
  cardWidth: number;
  cardHeight: number;
  count?: number;
  className?: string;
}

export function SkeletonHand({ cardWidth, cardHeight, count = 5, className }: SkeletonHandProps) {
  return (
    <div className={`flex flex-wrap gap-1.5${className ? ` ${className}` : ''}`} aria-hidden="true">
      {Array.from({ length: count }, (_, i) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
        <SkeletonCard key={i} width={cardWidth} height={cardHeight} />
      ))}
    </div>
  );
}
