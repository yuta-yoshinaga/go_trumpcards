/** Props for {@link SkeletonCard}. */
export interface SkeletonCardProps {
  width: number;
  height: number;
  className?: string;
}

/** Renders an animated skeleton card placeholder. */
export function SkeletonCard({ width, height, className }: SkeletonCardProps) {
  return (
    <div
      className={`rounded bg-white/20 animate-pulse${className ? ` ${className}` : ''}`}
      style={{ width, height }}
      aria-hidden="true"
    />
  );
}
