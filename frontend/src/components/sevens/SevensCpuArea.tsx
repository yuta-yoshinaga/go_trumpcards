import { useTranslation } from 'react-i18next';
import type { SevensPlayerData } from '../../types/card';
import { playerAreaClass } from '../../utils/sevensUtils';
import { CpuTurnArea } from '../CpuTurnArea';

function SevCpuArea({ player, isCurrentTurn }: { player: SevensPlayerData; isCurrentTurn: boolean }) {
  const { t } = useTranslation('sevens');
  return (
    <CpuTurnArea
      playerId={player.id}
      isHuman={player.isHuman}
      isCurrentTurn={isCurrentTurn}
      isFinished={player.isFinished}
      finishedLabel={player.isFinished ? t('rankLabel', { rank: player.rank }) : undefined}
      className={playerAreaClass}
    >
      {!player.isFinished && (
        <div className="text-game-text-muted text-xs">
          {t('cardCount', { count: player.cardCount })}
          {'　'}
          {t('passCount', {
            used: player.passesUsed,
            max: player.maxPasses === 0 ? t('passUnlimited') : player.maxPasses,
          })}
        </div>
      )}
    </CpuTurnArea>
  );
}

/** Renders a CPU player area for Sevens with card count and pass info. */
export { SevCpuArea as SevensCpuArea };
