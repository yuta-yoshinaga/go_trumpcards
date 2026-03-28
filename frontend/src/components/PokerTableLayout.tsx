import type { ReactNode } from 'react';
import { useIsLargeDesktop } from '../hooks/useCardDimensions';

interface PokerTableLayoutProps {
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

  if (isLargeDesktop) {
    return (
      <>
        <div className="grid grid-cols-3 gap-3 mb-4" {...(cpuAreaTutorial && { 'data-tutorial': cpuAreaTutorial })}>
          {cpuPlayers}
        </div>
        <div className="mb-4" {...(communityCardsTutorial && { 'data-tutorial': communityCardsTutorial })}>
          {communityCards}
        </div>
      </>
    );
  }

  return (
    <>
      <div className="mb-4" {...(communityCardsTutorial && { 'data-tutorial': communityCardsTutorial })}>
        {communityCards}
      </div>
      <div {...(cpuAreaTutorial && { 'data-tutorial': cpuAreaTutorial })}>{cpuPlayers}</div>
    </>
  );
}
