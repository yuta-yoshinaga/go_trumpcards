import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { GameFooter } from '../GameFooter';
import { SkeletonBar } from './SkeletonBar';

/** Props for the shared game skeleton shell. */
export interface SkeletonShellProps {
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
export function SkeletonShell({
  bgClass,
  bodyClassName = 'pt-3 px-4',
  footerClassName,
  children,
  footer,
}: SkeletonShellProps) {
  const { t } = useTranslation('common');
  const loadingLabel = t('skeleton.loading');
  return (
    <div className={['flex-1', 'flex', 'flex-col', 'min-h-0', bgClass].join(' ')} role="status" data-testid="skeleton">
      <SkeletonBar />
      <div className={['flex-1', 'overflow-y-auto', bodyClassName].join(' ')}>{children}</div>
      <GameFooter className={footerClassName}>{footer}</GameFooter>
      <span className="sr-only">{loadingLabel}</span>
    </div>
  );
}
