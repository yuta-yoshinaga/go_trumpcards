import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Short Deck Hold'em page. */
export function ShortDeckSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-green-poker"
      bodyClassName="pt-4 px-5"
      footerClassName="bg-game-bg-green-poker-dark border-white/20 px-5 py-3"
      footer={
        <>
          <div className="mb-2">
            <div className="h-5 w-20 rounded bg-white/10 animate-pulse mb-1" />
            <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={2} />
          </div>
          <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
        </>
      }
    >
      <div className="mb-4">
        <div className="h-5 w-32 rounded bg-white/10 animate-pulse mb-1.5" />
        <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
      </div>
      {Array.from({ length: 3 }, (_, i) => (
        <div key={i} className="mb-3 p-2 rounded bg-black/30">
          <div className="h-4 w-20 rounded bg-white/10 animate-pulse mb-2" />
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={2} />
        </div>
      ))}
    </GameSkeleton>
  );
}
