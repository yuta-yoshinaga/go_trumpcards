import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';
import { SkeletonGrid } from './SkeletonGrid';

/** Renders a loading skeleton placeholder for the Memory page. */
export function MemorySkeleton() {
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-blue" aria-busy="true" data-testid="skeleton">
      <SkeletonBar />
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <div className="my-2 p-2 rounded bg-black/30">
          <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-2" />
          {Array.from({ length: 4 }, (_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton placeholders never reorder
            <div key={i} className="h-4 w-full rounded bg-white/10 animate-pulse mb-1" />
          ))}
        </div>
        <div className="my-3 p-2 rounded bg-black/40">
          <SkeletonGrid count={52} cols="grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-13" />
        </div>
      </div>
      <GameFooter className="bg-game-bg-blue-dark border-white/20 px-4 py-2.5">
        <div className="h-8 w-24 rounded bg-white/10 animate-pulse" />
      </GameFooter>
    </div>
  );
}
