import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the BlackJack page. */
export function BlackJackSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-green-bright"
      bodyClassName="p-4"
      footerClassName="bg-game-bg-green-bright-dark border-white/20 px-4 py-3"
      footer={
        <>
          <div className="mb-2">
            <div className="h-5 w-20 rounded bg-white/10 animate-pulse mb-1" />
            <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
          </div>
          <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
        </>
      }
    >
      <div className="h-5 w-24 rounded bg-white/10 animate-pulse mb-2" />
      <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
    </GameSkeleton>
  );
}
