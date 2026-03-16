import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonHand } from './SkeletonHand';

export function BaccaratSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {Array.from({ length: 2 }, (_, i) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
          <div key={i} className="mb-4">
            <div className="h-5 w-24 rounded bg-white/10 animate-pulse mx-auto mb-1" />
            <div className="flex justify-center gap-2">
              <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={2} />
            </div>
          </div>
        ))}
      </div>
      <GameFooter className="bg-gray-800 px-4 pt-3">
        <div className="flex flex-col items-center gap-2 pb-2">
          <div className="h-8 w-32 rounded bg-white/10 animate-pulse" />
          <div className="h-8 w-24 rounded bg-white/10 animate-pulse" />
        </div>
      </GameFooter>
    </div>
  );
}
