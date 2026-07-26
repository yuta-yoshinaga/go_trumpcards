import type { ReactNode } from 'react';
import { useIsLargeDesktop } from '../hooks/useCardDimensions';

/** Props for {@link PokerTableLayout}. */
export interface PokerTableLayoutProps {
  communityCards: ReactNode;
  cpuPlayers: ReactNode;
  cpuAreaTutorial?: string;
  communityCardsTutorial?: string;
}

/** Renders community cards and CPU players in a poker-table-style layout on large desktops. */
export function PokerTableLayout({
  communityCards,
  cpuPlayers,
  cpuAreaTutorial,
  communityCardsTutorial,
}: PokerTableLayoutProps) {
  const isLargeDesktop = useIsLargeDesktop();

  const communityCardsEl = (
    <div className="mb-4" data-tutorial={communityCardsTutorial}>
      {communityCards}
    </div>
  );

  const cpuPlayersEl = (
    <div className={isLargeDesktop ? 'grid grid-cols-3 gap-3 mb-4' : undefined} data-tutorial={cpuAreaTutorial}>
      {cpuPlayers}
    </div>
  );

  if (isLargeDesktop) {
    return (
      <>
        {cpuPlayersEl}
        {communityCardsEl}
      </>
    );
  }

  return (
    <>
      {communityCardsEl}
      {cpuPlayersEl}
    </>
  );
}
