import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Omaha Hold'em page. */
export function OmahaSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green-poker" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        <div className="mb-4">
          <div className="h-5 w-32 rounded bg-white/10 animate-pulse mb-1.5" />
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
        </div>
        {Array.from({ length: 3 }, (_, i) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
          <div key={i} className="mb-3 p-2 rounded bg-black/30">
            <div className="h-4 w-20 rounded bg-white/10 animate-pulse mb-2" />
            <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={4} />
          </div>
        ))}
      </div>
      <GameFooter className="bg-game-bg-green-poker-dark border-white/20 px-5 py-3">
        <div className="mb-2">
          <div className="h-5 w-20 rounded bg-white/10 animate-pulse mb-1" />
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={4} />
        </div>
        <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
      </GameFooter>
    </div>
  );
}
