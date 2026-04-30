import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonCard } from './SkeletonCard';

/** Renders a loading skeleton placeholder for the Baker's Dozen page. */
export function BakersDozenSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-casino"
      footerClassName="bg-game-bg-casino border-white/20 px-4 py-2.5"
      footer={<div className="h-8 w-48 rounded bg-white/10 animate-pulse mx-auto" />}
    >
      {/* Foundation row */}
      <div className="flex gap-2 mb-3 items-start flex-wrap">
        {Array.from({ length: 4 }, (_, i) => (
          <SkeletonCard key={i} width={cardWidth} height={cardHeight} />
        ))}
      </div>
      {/* Tableau columns */}
      <div className="flex gap-2 mb-3">
        {Array.from({ length: 13 }, (_, i) => (
          <div key={i} className="flex-1 min-w-0">
            <SkeletonCard width={cardWidth} height={cardHeight} />
          </div>
        ))}
      </div>
    </GameSkeleton>
  );
}
