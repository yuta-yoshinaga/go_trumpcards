import { useCardDimensions } from '../../hooks/useCardDimensions';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonHand } from './SkeletonHand';

/** Renders a loading skeleton placeholder for the Speed page. */
export function SpeedSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass="bg-ds-surface"
      footerClassName="border-white/20 px-4 py-2.5"
      footer={<div className="h-8 w-32 rounded bg-white/10 animate-pulse" />}
    >
      {/* CPU hand placeholder */}
      <div className="flex items-center justify-center gap-2 mb-3">
        <div className="h-4 w-16 rounded bg-white/10 animate-pulse" />
        <SkeletonHand cardWidth={cardWidth * 0.7} cardHeight={cardHeight * 0.7} count={4} />
      </div>
      {/* Center piles placeholder */}
      <div className="flex items-center justify-center gap-6 mb-3">
        <div
          className="rounded bg-white/10 animate-pulse"
          style={{ width: cardWidth * 1.2, height: cardHeight * 1.2 }}
        />
        <div
          className="rounded bg-white/10 animate-pulse"
          style={{ width: cardWidth * 1.2, height: cardHeight * 1.2 }}
        />
      </div>
      {/* Human hand placeholder */}
      <div className="flex flex-col items-center gap-1">
        <div className="h-4 w-24 rounded bg-white/10 animate-pulse" />
        <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={4} />
      </div>
    </GameSkeleton>
  );
}
