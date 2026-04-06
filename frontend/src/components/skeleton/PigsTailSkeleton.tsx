import { GameSkeleton } from './GameSkeleton';

/** Renders a loading skeleton placeholder for the Pig's Tail page. */
export function PigsTailSkeleton() {
  return (
    <GameSkeleton
      bgClass="bg-game-bg-green"
      footerClassName="bg-game-bg-green-dark border-white/20 px-4 py-2.5"
      footer={<div className="h-10 w-48 rounded bg-white/10 animate-pulse mx-auto" />}
    >
      <div className="flex justify-center gap-8 mb-4">
        <div className="h-20 w-20 rounded-full bg-white/10 animate-pulse" />
        <div className="h-20 w-20 rounded-full bg-white/10 animate-pulse" />
      </div>
      {Array.from({ length: 4 }, (_, i) => (
        <div key={i} className="mb-2 p-2 rounded bg-black/30">
          <div className="h-4 w-32 rounded bg-white/10 animate-pulse" />
        </div>
      ))}
    </GameSkeleton>
  );
}
