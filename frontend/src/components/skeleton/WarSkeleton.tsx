import { GameSkeleton } from './GameSkeleton';

/** Renders a loading skeleton placeholder for the War page. */
export function WarSkeleton() {
  return (
    <GameSkeleton
      bgClass="bg-game-bg-green"
      footerClassName="bg-game-bg-green-dark border-white/20 px-4 py-2.5"
      footer={<div className="h-10 w-48 rounded bg-white/10 animate-pulse mx-auto" />}
    >
      <div className="flex justify-center gap-8 my-8">
        <div className="h-24 w-16 rounded bg-white/10 animate-pulse" />
        <div className="h-24 w-16 rounded bg-white/10 animate-pulse" />
      </div>
    </GameSkeleton>
  );
}
