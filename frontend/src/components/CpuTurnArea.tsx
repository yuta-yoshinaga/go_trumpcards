import { activeTurnStyle, finishedPlayerStyle } from '../styles/gameConstants';
import { playerName } from '../utils/playerUtils';
import { StatusBadge } from './StatusBadge';

interface CpuTurnAreaProps {
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
  const conditionalStyle: React.CSSProperties =
    isFinished && dimFinished ? finishedPlayerStyle : isCurrentTurn ? activeTurnStyle : {};
  return (
    <div id={id} className={className} style={conditionalStyle}>
      <div className={`text-white font-bold mb-1${nameClassName ? ` ${nameClassName}` : ''}`}>
        {playerName(playerId, isHuman)}
        {isFinished && finishedLabel && <StatusBadge variant="success">{finishedLabel}</StatusBadge>}
        {isCurrentTurn && !isFinished && <StatusBadge variant="warning">考え中...</StatusBadge>}
      </div>
      {children}
    </div>
  );
}
