import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Tonk page. */
export function TonkSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-game-bg-blue"
      footerClassName="bg-game-bg-blue-dark border-white/20 px-4 py-2.5"
      footer={
        <>
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={5} className="mb-2" />
          <div className="h-8 w-32 rounded bg-white/10 animate-pulse" />
        </>
      }
    >
      <div className="h-5 w-48 rounded bg-white/10 animate-pulse mx-auto mb-2" />
      <div className="my-3 p-2 rounded bg-black/30 flex items-center gap-3">
        <div className="h-16 w-12 rounded bg-white/10 animate-pulse" />
        <div className="h-4 w-24 rounded bg-white/10 animate-pulse" />
      </div>
      <div className="mb-2 p-2 rounded bg-black/30">
        <div className="h-4 w-32 rounded bg-white/10 animate-pulse" />
      </div>
    </GameSkeleton>
  );
}
