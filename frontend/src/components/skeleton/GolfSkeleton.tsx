import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonCard } from './SkeletonCard';

/** Renders a loading skeleton placeholder for the Golf Solitaire page. */
export function GolfSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-casino"
      footerClassName="bg-game-bg-casino border-white/20 px-4 py-2.5"
      footer={<div className="h-8 w-48 rounded bg-white/10 animate-pulse mx-auto" />}
    >
      {/* Golf: 7 columns × 5 rows placeholder */}
      <div className="flex gap-1 justify-center">
        {Array.from({ length: 7 }, (_, col) => (
          <div key={col} className="flex flex-col gap-1">
            {Array.from({ length: 5 }, (_, row) => (
              <SkeletonCard key={row} width={cardWidth} height={cardHeight} />
            ))}
          </div>
        ))}
      </div>
      {/* Stock/Waste placeholder */}
      <div className="flex gap-2 mt-3">
        <SkeletonCard width={cardWidth} height={cardHeight} />
        <SkeletonCard width={cardWidth} height={cardHeight} />
      </div>
    </GameSkeleton>
  );
}
