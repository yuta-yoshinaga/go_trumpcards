import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

/** Props for {@link PhaseIndicator}. */
export interface PhaseIndicatorProps {
  phaseName: string;
  isHumanTurn?: boolean;
  children?: ReactNode;
}

/** Renders the current game phase name and turn indicator. */
export function PhaseIndicator({ phaseName, isHumanTurn, children }: PhaseIndicatorProps) {
  const { t } = useTranslation('common');

  const turnLabel =
    isHumanTurn === undefined ? '' : isHumanTurn ? t('turnIndicator.yourTurn') : t('turnIndicator.waiting');
  const announcementText = turnLabel
    ? t('turnIndicator.announcement', { phase: phaseName, turn: turnLabel })
    : t('turnIndicator.phaseOnly', { phase: phaseName });

  return (
    <div
      className="shrink-0 glass-panel text-ds-text-primary text-sm px-3 sm:px-5 py-2 flex flex-wrap gap-x-2 sm:gap-x-6 gap-y-1 items-center tabular-nums"
      data-testid="phase-indicator"
    >
      <span>
        <strong>{phaseName}</strong>
      </span>
      {isHumanTurn !== undefined && (
        <span
          className={
            isHumanTurn
              ? 'relative inline-flex text-ds-success font-bold text-base bg-ds-surface px-2 py-0.5 rounded-full'
              : 'text-game-text-muted'
          }
        >
          {isHumanTurn && (
            <span
              aria-hidden="true"
              className="absolute inset-0 rounded-full ring-2 ring-ds-success/40 motion-safe:animate-pulse pointer-events-none"
            />
          )}
          <span className={isHumanTurn ? 'relative' : undefined}>{turnLabel}</span>
        </span>
      )}
      {children}
      <span aria-live="polite" aria-atomic="true" className="sr-only" data-testid="phase-announcement">
        {announcementText}
      </span>
    </div>
  );
}
