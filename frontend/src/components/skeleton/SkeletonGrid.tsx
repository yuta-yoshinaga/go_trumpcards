/** Props for {@link SkeletonGrid}. */
export interface SkeletonGridProps {
  count: number;
  cols: string;
  aspectRatio?: string;
  /** Extra classes applied to the grid container. */
  gridClassName?: string;
}

/** Renders an animated skeleton grid of card placeholders. */
export function SkeletonGrid({ count, cols, aspectRatio = 'aspect-[2/3]', gridClassName = '' }: SkeletonGridProps) {
  return (
    <div className={`grid ${cols} gap-1 ${gridClassName}`.trim()} aria-hidden="true">
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className={`rounded bg-white/20 animate-pulse ${aspectRatio}`} />
      ))}
    </div>
  );
}
