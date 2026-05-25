import type { ReactNode } from 'react';

/** Props for the felt-textured shell wrapping `/discover` pages. */
export interface DiscoverShellProps {
  readonly children: ReactNode;
  /**
   * Optional test id forwarded to the outer wrapper so callers can find the
   * concierge container in DOM-level tests.
   */
  readonly testId?: string;
}

/**
 * Visual shell for the AI Game Concierge pages (DR-6).
 *
 * On desktop (≥1024px) renders the page content inside a max-w-[600px]
 * frame against a dark felt-green diagonal-stripe background, giving a
 * private-card-room feel and visually separating the concierge flow
 * from the standard game pages. On smaller widths the felt is dropped
 * and the frame spans the viewport so the content remains readable
 * on phones.
 */
export function DiscoverShell({ children, testId }: DiscoverShellProps) {
  return (
    <div
      data-testid={testId}
      className="discover-shell flex-1 min-h-0 flex flex-col bg-ds-bg lg:bg-game-bg-green-dark lg:bg-[image:repeating-linear-gradient(45deg,var(--color-game-bg-green-dark)_0px,var(--color-game-bg-green-dark)_2px,transparent_2px,transparent_6px)]"
    >
      <div className="flex-1 min-h-0 flex flex-col w-full lg:max-w-[600px] lg:mx-auto lg:bg-ds-bg lg:shadow-[0_0_64px_rgba(0,0,0,0.6)]">
        {children}
      </div>
    </div>
  );
}
