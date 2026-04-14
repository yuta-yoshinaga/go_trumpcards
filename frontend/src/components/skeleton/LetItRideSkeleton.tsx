import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Let It Ride page. */
export function LetItRideSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-casino"
      footerClassName="bg-game-bg-casino-dark border-white/20 px-4 pt-3"
      footer={
        <div className="flex flex-col items-center gap-2 pb-2">
          <div className="h-8 w-32 rounded bg-white/10 animate-pulse" />
          <div className="h-8 w-24 rounded bg-white/10 animate-pulse" />
        </div>
      }
    >
      <div className="mb-4">
        <div className="h-5 w-24 rounded bg-white/10 animate-pulse mx-auto mb-1" />
        <div className="flex justify-center gap-2">
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={3} />
        </div>
      </div>
      <div className="mb-4">
        <div className="h-5 w-36 rounded bg-white/10 animate-pulse mx-auto mb-1" />
        <div className="flex justify-center gap-2">
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={2} />
        </div>
      </div>
    </GameSkeleton>
  );
}
