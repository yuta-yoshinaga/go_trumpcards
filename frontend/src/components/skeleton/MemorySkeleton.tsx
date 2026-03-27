import { gameTheme } from '../../styles/gameTheme';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonGrid } from './SkeletonGrid';

/** Renders a loading skeleton placeholder for the Memory page. */
export function MemorySkeleton() {
  return (
    <GameSkeleton
      bgClass={gameTheme.memory.bg}
      bodyClassName="pt-3 lg:pt-1 px-4 lg:px-8 lg:flex lg:flex-col lg:overflow-hidden"
      footerClassName={`${gameTheme.memory.footer} px-4 py-2.5`}
      footer={<div className="h-8 w-24 rounded bg-white/10 animate-pulse" />}
    >
      <div className="my-1 px-2 py-1 rounded bg-black/30 flex gap-3 lg:shrink-0">
        {Array.from({ length: 4 }, (_, i) => (
          <div key={i} className="h-4 w-20 rounded bg-white/10 animate-pulse" />
        ))}
      </div>
      <div className="my-3 lg:my-1 p-2 lg:p-1 rounded bg-black/40 lg:flex-1 lg:min-h-0 lg:overflow-hidden">
        <SkeletonGrid
          count={52}
          cols="grid-cols-6 md:grid-cols-8 lg:grid-cols-13"
          aspectRatio="aspect-[2/3] lg:aspect-auto"
          gridClassName="lg:grid-rows-4 lg:h-full"
        />
      </div>
    </GameSkeleton>
  );
}
