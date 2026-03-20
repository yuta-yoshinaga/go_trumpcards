import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

interface PhaseIndicatorProps {
  phaseName: string;
  isHumanTurn?: boolean;
  children?: ReactNode;
}

/** Renders the current game phase name and turn indicator. */
export function PhaseIndicator({ phaseName, isHumanTurn, children }: PhaseIndicatorProps) {
  const { t } = useTranslation('common');

  return (
    <div
      className="shrink-0 bg-black/40 text-white text-sm px-5 py-2 flex flex-wrap gap-x-6 gap-y-1 items-center"
      aria-live="polite"
      data-testid="phase-indicator"
    >
      <span>
        <strong>{phaseName}</strong>
      </span>
      {isHumanTurn !== undefined && (
        <span className={isHumanTurn ? 'text-green-400 animate-pulse' : 'text-game-text-muted'}>
          {isHumanTurn ? t('turnIndicator.yourTurn') : t('turnIndicator.waiting')}
        </span>
      )}
      {children}
    </div>
  );
}
