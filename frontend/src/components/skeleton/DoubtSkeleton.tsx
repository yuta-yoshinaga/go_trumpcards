import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Doubt page. */
export function DoubtSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-blue" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <div className="flex gap-2 flex-wrap mb-3">
          {Array.from({ length: 3 }, (_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
            <div key={i} className="p-2 rounded bg-black/30">
              <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-2" />
              <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={4} />
            </div>
          ))}
        </div>
        <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
          <div className="h-4 w-20 rounded bg-white/10 animate-pulse mb-1" />
          <div className="h-4 w-32 rounded bg-white/10 animate-pulse" />
        </div>
      </div>
      <GameFooter className="bg-game-bg-blue-dark border-white/20 px-4 py-2.5">
        <div className="mb-2">
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
        </div>
        <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
      </GameFooter>
    </div>
  );
}
