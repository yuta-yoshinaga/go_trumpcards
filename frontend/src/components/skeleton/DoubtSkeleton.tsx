import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Doubt page. */
export function DoubtSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-green"
      footerClassName="bg-game-bg-green-dark border-white/20 px-4 py-2.5"
      footer={
        <>
          <div className="mb-2">
            <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} />
          </div>
          <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
        </>
      }
    >
      <div className="flex gap-2 flex-wrap mb-3">
        {Array.from({ length: 3 }, (_, i) => (
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
    </GameSkeleton>
  );
}
