import { useTranslation } from 'react-i18next';
import { bettingActionName } from '../styles/gameConstants';

interface CpuActionLogProps {
  actions: { playerIdx: number; action: number; amount: number }[] | undefined;
}

export function CpuActionLog({ actions }: CpuActionLogProps) {
  const { t } = useTranslation('common');
  if (!actions || actions.length === 0) return null;
  return (
    <div className="bg-black/30 rounded p-2 mb-3 text-white text-xs">
      <div className="font-bold mb-1">{t('label.cpuAction')}</div>
      {actions.map((a, i) => (
        <div key={`${i}-${a.playerIdx}-${a.action}`}>
          {t('player.player', { idx: a.playerIdx })}: {bettingActionName(a.action)}
          {a.amount > 0 && ` (${a.amount})`}
        </div>
      ))}
    </div>
  );
}
