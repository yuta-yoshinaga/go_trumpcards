import { useTranslation } from 'react-i18next';
import { playerAreaBase } from '../../styles/gameStyles';
import type { DoubtPlayerData } from '../../types/card';
import { CpuTurnArea } from '../CpuTurnArea';

/** CSS class for Doubt player area layout. */
export const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_150px] min-w-[120px]`;

/** Renders a CPU player area for Doubt with card count and tell indicator. */
export function DoubtCpuArea({
  player,
  isCurrentTurn,
  hasTell,
}: {
  player: DoubtPlayerData;
  isCurrentTurn: boolean;
  hasTell: boolean;
}) {
  const { t } = useTranslation('doubt');
  const { t: tc } = useTranslation('common');
  return (
    <CpuTurnArea
      playerId={player.id}
      isHuman={player.isHuman}
      isCurrentTurn={isCurrentTurn}
      isFinished={player.isFinished}
      dimFinished={false}
      finishedLabel={player.isFinished ? tc('status.finished') : undefined}
      className={playerAreaClass}
      nameClassName="text-sm"
    >
      <div className="text-game-text-muted text-xs">{t('cardCount', { count: player.cardCount })}</div>
      {hasTell && (
        <div
          className="mt-0.5 flex items-center gap-1"
          role="img"
          aria-label={t('tell')}
          data-testid="doubt-tell-indicator"
        >
          <span aria-hidden="true" className="animate-sweat-drop text-lg leading-none">
            💧
          </span>
          <span aria-hidden="true" className="animate-eye-dart text-lg leading-none">
            👀
          </span>
          <span className="rounded-full bg-ds-warning/30 px-1.5 py-0 text-[10px] font-semibold leading-tight text-ds-warning motion-safe:animate-pulse">
            {t('tellBadge')}
          </span>
        </div>
      )}
    </CpuTurnArea>
  );
}
