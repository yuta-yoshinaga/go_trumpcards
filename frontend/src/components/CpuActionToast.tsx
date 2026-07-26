import { useTranslation } from 'react-i18next';
import { useTransientToast } from '../hooks/useTransientToast';
import { bettingActionName } from '../styles/gameConstants';
import { TOAST_DURATION } from '../styles/toastDurations';
import { Toast } from './Toast';

/** Props for {@link CpuActionToast}. */
export interface CpuActionToastProps {
  actions: { playerIdx: number; action: number; amount: number }[] | undefined;
}

/**
 * Toast notification for CPU betting actions. Shows when new actions arrive,
 * auto-dismisses after the long duration, and can be dismissed early via the
 * close button or Escape. A thin wrapper over the shared {@link Toast} banner +
 * {@link useTransientToast} lifecycle (issue #4313).
 */
export function CpuActionToast({ actions }: CpuActionToastProps) {
  const { t } = useTranslation('common');
  const len = actions?.length ?? 0;
  const { visible, dismiss } = useTransientToast(len, TOAST_DURATION.long, {
    active: len > 0,
    skipInitial: false,
  });

  if (!visible || !actions || actions.length === 0) return null;

  return (
    <Toast onDismiss={dismiss}>
      {actions.map((a, i) => (
        <div key={`${i}-${a.playerIdx}-${a.action}`}>
          {t('player.player', { idx: a.playerIdx })}: {bettingActionName(a.action)}
          {a.amount > 0 && ` (${a.amount})`}
        </div>
      ))}
    </Toast>
  );
}
