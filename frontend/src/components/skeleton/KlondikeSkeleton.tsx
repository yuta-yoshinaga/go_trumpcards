import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonCard } from './SkeletonCard';

export function KlondikeSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Foundation + Stock/Waste row */}
        <div className="flex gap-2 mb-3 items-start flex-wrap">
          {Array.from({ length: 6 }, (_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
            <SkeletonCard key={i} width={cardWidth} height={cardHeight} />
          ))}
        </div>
        {/* Tableau columns */}
        <div className="flex gap-2 mb-3">
          {Array.from({ length: 7 }, (_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
            <div key={i} className="flex-1 min-w-0">
              <SkeletonCard width={cardWidth} height={cardHeight} />
            </div>
          ))}
        </div>
      </div>
      <GameFooter className="bg-game-bg-casino border-white/20 px-4 py-2.5">
        <div className="h-8 w-48 rounded bg-white/10 animate-pulse mx-auto" />
      </GameFooter>
    </div>
  );
}
