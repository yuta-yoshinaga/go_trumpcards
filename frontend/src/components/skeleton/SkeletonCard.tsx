interface SkeletonCardProps {
  width: number;
  height: number;
  className?: string;
}

export function SkeletonCard({ width, height, className }: SkeletonCardProps) {
  return (
    <div
      className={`rounded bg-white/20 animate-pulse${className ? ` ${className}` : ''}`}
      style={{ width, height }}
      aria-hidden="true"
    />
  );
}
