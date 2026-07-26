import { useTranslation } from 'react-i18next';
import { activeTurnClass, finishedPlayerClass } from '../styles/gameConstants';
import { playerName } from '../utils/playerUtils';
import { StatusBadge } from './StatusBadge';

/** Props for {@link CpuTurnArea}. */
export interface CpuTurnAreaProps {
  id?: string;
  playerId: number;
  isHuman: boolean;
  isCurrentTurn: boolean;
  isFinished: boolean;
  dimFinished?: boolean;
  finishedLabel?: string;
  className: string;
  nameClassName?: string;
  children?: React.ReactNode;
}

/** Renders a player area with name, turn indicator, and optional children. */
export function CpuTurnArea({
  id,
  playerId,
  isHuman,
  isCurrentTurn,
  isFinished,
  dimFinished = true,
  finishedLabel,
  className,
  nameClassName,
  children,
}: CpuTurnAreaProps) {
  const { t } = useTranslation('common');
  const conditionalClass = isFinished && dimFinished ? finishedPlayerClass : isCurrentTurn ? activeTurnClass : '';
  return (
    <div id={id} className={`${className}${conditionalClass ? ` ${conditionalClass}` : ''}`}>
      <div className={`text-ds-text-primary font-bold mb-1${nameClassName ? ` ${nameClassName}` : ''}`}>
        {playerName(playerId, isHuman)}
        {isFinished && finishedLabel && <StatusBadge variant="success">{finishedLabel}</StatusBadge>}
        {isCurrentTurn && !isFinished && <StatusBadge variant="warning">{t('status.thinking')}</StatusBadge>}
      </div>
      {children}
    </div>
  );
}
