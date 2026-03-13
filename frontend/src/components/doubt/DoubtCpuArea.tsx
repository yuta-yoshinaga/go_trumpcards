import { useTranslation } from 'react-i18next';
import { playerAreaBase } from '../../styles/gameStyles';
import type { DoubtPlayerData } from '../../types/card';
import { CpuTurnArea } from '../CpuTurnArea';

export const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_150px] min-w-[120px]`;

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
        <span className="animate-sweat-drop text-lg" role="img" aria-label={t('tell')}>
          💧
        </span>
      )}
    </CpuTurnArea>
  );
}
