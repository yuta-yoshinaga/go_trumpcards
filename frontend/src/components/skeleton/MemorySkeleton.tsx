import { gameTheme } from '../../styles/gameTheme';
import { GameSkeleton } from './GameSkeleton';
import { SkeletonGrid } from './SkeletonGrid';

/** Renders a loading skeleton placeholder for the Memory page. */
export function MemorySkeleton() {
  return (
    <GameSkeleton
      bgClass={gameTheme.memory.bg}
      footerClassName={`${gameTheme.memory.footer} px-4 py-2.5`}
      footer={<div className="h-8 w-24 rounded bg-white/10 animate-pulse" />}
    >
      <div className="my-2 p-2 rounded bg-black/30">
        <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-2" />
        {Array.from({ length: 4 }, (_, i) => (
          <div key={i} className="h-4 w-full rounded bg-white/10 animate-pulse mb-1" />
        ))}
      </div>
      <div className="my-3 p-2 rounded bg-black/40">
        <SkeletonGrid count={52} cols="grid-cols-6 md:grid-cols-8 lg:grid-cols-13" />
      </div>
    </GameSkeleton>
  );
}
