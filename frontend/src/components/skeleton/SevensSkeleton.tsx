import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonGrid } from './SkeletonGrid';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Sevens page. */
export function SevensSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <div className="flex gap-2.5 flex-wrap mb-2.5">
          {Array.from({ length: 3 }, (_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
            <div key={i} className="p-2 rounded bg-black/30">
              <div className="h-4 w-16 rounded bg-white/10 animate-pulse" />
            </div>
          ))}
        </div>
        <div className="bg-black/30 rounded p-2 my-2">
          <SkeletonGrid count={52} cols="grid-cols-13" aspectRatio="aspect-square" />
        </div>
      </div>
      <GameFooter className="bg-game-bg-green-dark border-white/20 px-4 py-2.5">
        <div className="mb-2">
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
        </div>
        <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
      </GameFooter>
    </div>
  );
}
