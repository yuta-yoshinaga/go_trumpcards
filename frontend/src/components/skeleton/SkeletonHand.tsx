import { SkeletonCard } from './SkeletonCard';

/** Props for {@link SkeletonHand}. */
export interface SkeletonHandProps {
  cardWidth: number;
  cardHeight: number;
  count?: number;
  className?: string;
}

/** Renders an animated skeleton hand of card placeholders. */
export function SkeletonHand({ cardWidth, cardHeight, count = 5, className }: SkeletonHandProps) {
  return (
    <div className={`flex flex-wrap gap-1.5${className ? ` ${className}` : ''}`} aria-hidden="true">
      {Array.from({ length: count }, (_, i) => (
        <SkeletonCard key={i} width={cardWidth} height={cardHeight} />
      ))}
    </div>
  );
}
