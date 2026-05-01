import { useTranslation } from 'react-i18next';
import { activeTurnClass, finishedPlayerClass } from '../../styles/gameConstants';
import type { DaifugoPlayerData } from '../../types/card';
import { playerName } from '../../utils/playerUtils';
import { StatusBadge } from '../StatusBadge';

/** Renders a compact single-row CPU summary for Daifugo (used on mobile to avoid vertical overflow). */
export function DaifugoCpuCompact({ player, isCurrentTurn }: { player: DaifugoPlayerData; isCurrentTurn: boolean }) {
  const { t } = useTranslation('daifugo');
  const conditionalClass = player.isFinished
    ? `${finishedPlayerClass} border-2 border-transparent`
    : isCurrentTurn
      ? activeTurnClass
      : 'border-2 border-white/10';
  return (
    <div
      id={`player-area-${player.id}`}
      className={`flex flex-shrink-0 items-center gap-1.5 rounded-[8px] px-2 py-1 text-xs whitespace-nowrap bg-black/20 ${conditionalClass}`}
    >
      <span className="text-ds-text-primary font-bold">{playerName(player.id, player.isHuman)}</span>
      {player.isFinished ? (
        <span className="text-game-text-muted">{t(`rank.${player.rank}`)}</span>
      ) : (
        <span className="text-game-text-muted">{t('cardCount', { count: player.cardCount })}</span>
      )}
      {player.illegalFinishPenalty && <StatusBadge variant="danger">{t('badge.illegalFinishPenalty')}</StatusBadge>}
    </div>
  );
}
