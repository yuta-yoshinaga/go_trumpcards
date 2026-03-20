import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Hearts page. */
export function HeartsSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-blue" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <div className="h-5 w-48 rounded bg-white/10 animate-pulse mx-auto mb-2" />
        {Array.from({ length: 3 }, (_, i) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
          <div key={i} className="mb-2 p-2 rounded bg-black/30">
            <div className="h-4 w-32 rounded bg-white/10 animate-pulse" />
          </div>
        ))}
        <div className="my-3 p-2 rounded bg-black/30">
          <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-1" />
          <div className="h-20 w-full rounded bg-white/10 animate-pulse" />
        </div>
      </div>
      <GameFooter className="bg-game-bg-blue-dark border-white/20 px-4 py-2.5">
        <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} className="mb-2" />
        <div className="h-8 w-32 rounded bg-white/10 animate-pulse" />
      </GameFooter>
    </div>
  );
}
