import { GameSkeleton } from './GameSkeleton';

/** Loading skeleton placeholder for the Fifty-one game page. */
export function FiftyOneSkeleton() {
  return (
    <GameSkeleton
      bgClass="bg-game-bg-green"
      footerClassName="bg-game-bg-green-dark border-white/20 px-4 py-2.5"
      footer={<div className="h-10 w-48 rounded bg-white/10 animate-pulse mx-auto" />}
    >
      <div className="flex flex-col items-center gap-4 my-8">
        <div className="flex gap-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-20 w-14 rounded bg-white/10 animate-pulse" />
          ))}
        </div>
        <div className="flex gap-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-20 w-14 rounded bg-white/10 animate-pulse" />
          ))}
        </div>
      </div>
    </GameSkeleton>
  );
}
