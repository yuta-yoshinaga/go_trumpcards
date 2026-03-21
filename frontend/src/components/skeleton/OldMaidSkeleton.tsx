import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Old Maid page. */
export function OldMaidSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <div className="flex gap-2 flex-wrap mb-2 justify-center">
          {Array.from({ length: 3 }, (_, i) => (
            <div key={i} className="p-2 rounded bg-black/30">
              <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-2" />
              <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={3} />
            </div>
          ))}
        </div>
      </div>
      <GameFooter className="bg-game-bg-green-dark border-white/20 px-4 py-2.5">
        <div className="mb-2">
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
        </div>
        <div className="h-8 w-48 rounded bg-white/10 animate-pulse mx-auto" />
      </GameFooter>
    </div>
  );
}
