/**
 * Layout-preserving skeleton shown while the lazy `discover` i18n bundle
 * loads. Renders an 8-card deck progress placeholder, a question-text
 * block, and four option rows so the survey doesn't shift layout when
 * the real content swaps in (DR-3).
 */
export function DiscoverSkeleton() {
  return (
    <div
      role="status"
      aria-busy="true"
      aria-label="Loading"
      className="flex-1 min-h-0 flex flex-col items-center justify-start px-4 py-8 gap-6"
    >
      <ul aria-hidden="true" className="flex gap-1.5 items-end justify-center">
        {Array.from({ length: 8 }, (_, i) => (
          <li
            key={`skeleton-card-${i + 1}`}
            className={
              i === 0 ? 'w-7 h-10 rounded-sm bg-ds-surface-elevated' : 'w-5 h-7 rounded-sm bg-ds-surface-elevated/60'
            }
          />
        ))}
      </ul>
      <div className="w-full max-w-md flex flex-col gap-4">
        <div className="h-3 w-24 rounded bg-ds-surface-elevated/80" />
        <div className="h-6 w-3/4 rounded bg-ds-surface-elevated" />
        <ul className="flex flex-col gap-2">
          {Array.from({ length: 4 }, (_, i) => (
            <li key={`skeleton-opt-${i + 1}`} className="h-14 w-full rounded-md bg-ds-surface-elevated/70" />
          ))}
        </ul>
      </div>
    </div>
  );
}
