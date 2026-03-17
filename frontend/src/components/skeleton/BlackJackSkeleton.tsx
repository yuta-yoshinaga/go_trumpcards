import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonHand } from './SkeletonHand';

export function BlackJackSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green-bright" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto p-4">
        <div className="h-5 w-24 rounded bg-white/10 animate-pulse mb-2" />
        <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
      </div>
      <GameFooter className="bg-game-bg-green-bright-dark border-white/15 px-4 py-3">
        <div className="mb-2">
          <div className="h-5 w-20 rounded bg-white/10 animate-pulse mb-1" />
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
        </div>
        <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
      </GameFooter>
    </div>
  );
}
