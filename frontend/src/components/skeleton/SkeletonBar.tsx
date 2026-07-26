/** Props for {@link SkeletonBar}. */
export interface SkeletonBarProps {
  height?: string;
  className?: string;
}

/** Renders an animated skeleton loading bar placeholder. */
export function SkeletonBar({ height = 'h-9', className }: SkeletonBarProps) {
  return (
    <div
      className={`shrink-0 bg-black/40 animate-pulse ${height}${className ? ` ${className}` : ''}`}
      aria-hidden="true"
    />
  );
}
