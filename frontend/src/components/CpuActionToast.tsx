import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { bettingActionName } from '../styles/gameConstants';

interface CpuActionToastProps {
  actions: { playerIdx: number; action: number; amount: number }[] | undefined;
}

const DISMISS_MS = 3000;

/** Toast notification for CPU betting actions. Auto-dismisses after 3 seconds. */
export function CpuActionToast({ actions }: CpuActionToastProps) {
  const { t } = useTranslation('common');
  const [visible, setVisible] = useState(false);
  const prevLenRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const len = actions?.length ?? 0;

  useEffect(() => {
    if (len > 0 && len !== prevLenRef.current) {
      setVisible(true);
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setVisible(false), DISMISS_MS);
    }
    prevLenRef.current = len;
    return () => clearTimeout(timerRef.current);
  }, [len]);

  if (!visible || !actions || actions.length === 0) return null;

  return (
    <div
      role="status"
      aria-live="polite"
      className="absolute top-0 left-0 right-0 z-20 mx-4 mt-1 animate-[slideDown_0.3s_ease-out] rounded bg-black/70 px-3 py-1.5 text-white text-xs shadow-lg"
    >
      {actions.map((a, i) => (
        <div key={`${i}-${a.playerIdx}-${a.action}`}>
          {t('player.player', { idx: a.playerIdx })}: {bettingActionName(a.action)}
          {a.amount > 0 && ` (${a.amount})`}
        </div>
      ))}
    </div>
  );
}
