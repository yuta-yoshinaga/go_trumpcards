import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonCard } from './SkeletonCard';

/** Renders a loading skeleton placeholder for the TriPeaks page. */
export function TriPeaksSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-casino"
      footerClassName="bg-game-bg-casino border-white/20 px-4 py-2.5"
      footer={<div className="h-8 w-48 rounded bg-white/10 animate-pulse mx-auto" />}
    >
      {/* TriPeaks rows placeholder: 3, 6, 9, 10 cards */}
      {[3, 6, 9, 10].map((count, row) => (
        <div key={row} className="flex justify-center gap-1 mb-1">
          {Array.from({ length: count }, (_, col) => (
            <SkeletonCard key={col} width={cardWidth} height={cardHeight} />
          ))}
        </div>
      ))}
      {/* Stock/Waste placeholder */}
      <div className="flex gap-2 mt-3">
        <SkeletonCard width={cardWidth} height={cardHeight} />
        <SkeletonCard width={cardWidth} height={cardHeight} />
      </div>
    </GameSkeleton>
  );
}
