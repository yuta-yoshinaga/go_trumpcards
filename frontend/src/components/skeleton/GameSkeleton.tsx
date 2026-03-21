import type { ReactNode } from 'react';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';

/** Props for the shared game skeleton shell. */
interface GameSkeletonProps {
  /** Background color class for the outer wrapper (e.g. "bg-game-bg-green-bright"). */
  bgClass: string;
  /** Classes for the scrollable body area. Defaults to "pt-3 px-4". */
  bodyClassName?: string;
  /** Classes for the GameFooter (background, border, padding). */
  footerClassName: string;
  /** Body content rendered inside the scrollable area. */
  children: ReactNode;
  /** Footer content rendered inside GameFooter. */
  footer: ReactNode;
}

/** Renders the shared outer shell for game skeleton placeholders. */
export function GameSkeleton({
  bgClass,
  bodyClassName = 'pt-3 px-4',
  footerClassName,
  children,
  footer,
}: GameSkeletonProps) {
  return (
    <div
      className={['flex-1', 'flex', 'flex-col', 'min-h-0', bgClass].join(' ')}
      aria-busy={true}
      data-testid="skeleton"
    >
      <SkeletonBar />
      <div className={['flex-1', 'overflow-y-auto', bodyClassName].join(' ')}>{children}</div>
      <GameFooter className={footerClassName}>{footer}</GameFooter>
    </div>
  );
}
