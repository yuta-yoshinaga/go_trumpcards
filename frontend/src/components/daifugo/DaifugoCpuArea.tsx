import { useTranslation } from 'react-i18next';
import { playerAreaBase } from '../../styles/gameStyles';
import type { DaifugoPlayerData } from '../../types/card';
import { CpuTurnArea } from '../CpuTurnArea';
import { StatusBadge } from '../StatusBadge';

export const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_180px] min-w-[150px]`;

export function DaifugoCpuArea({ player, isCurrentTurn }: { player: DaifugoPlayerData; isCurrentTurn: boolean }) {
  const { t } = useTranslation('daifugo');
  const finishedLabel = player.isFinished ? t('finishedWithRank', { rank: t(`rank.${player.rank}`) }) : undefined;
  return (
    <CpuTurnArea
      id={`player-area-${player.id}`}
      playerId={player.id}
      isHuman={player.isHuman}
      isCurrentTurn={isCurrentTurn}
      isFinished={player.isFinished}
      finishedLabel={finishedLabel}
      className={playerAreaClass}
    >
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em]">{t('cardCount', { count: player.cardCount })}</div>
      )}
      {player.illegalFinishPenalty && <StatusBadge variant="danger">{t('badge.illegalFinishPenalty')}</StatusBadge>}
    </CpuTurnArea>
  );
}
