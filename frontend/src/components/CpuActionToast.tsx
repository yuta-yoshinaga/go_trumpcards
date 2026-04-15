import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { bettingActionName } from '../styles/gameConstants';

interface CpuActionToastProps {
  actions: { playerIdx: number; action: number; amount: number }[] | undefined;
}

const DISMISS_MS = 5000;

/**
 * Toast notification for CPU betting actions. Auto-dismisses after 5 seconds,
 * but can be dismissed early via the close button or the Escape key. Respects
 * the user's `prefers-reduced-motion` setting.
 */
export function CpuActionToast({ actions }: CpuActionToastProps) {
  const { t } = useTranslation('common');
  const reduced = useReducedMotion();
  const [visible, setVisible] = useState(false);
  const prevLenRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const triggerRef = useRef<Element | null>(null);

  const len = actions?.length ?? 0;

  useEffect(() => {
    if (len > 0 && len !== prevLenRef.current) {
      triggerRef.current = document.activeElement;
      setVisible(true);
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setVisible(false), DISMISS_MS);
    }
    prevLenRef.current = len;
    return () => clearTimeout(timerRef.current);
  }, [len]);

  useEffect(() => {
    if (!visible) {
      if (triggerRef.current instanceof HTMLElement && document.contains(triggerRef.current)) {
        triggerRef.current.focus();
      }
      triggerRef.current = null;
      return;
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return;
      if (document.querySelector('[role="dialog"][aria-modal="true"]')) return;
      setVisible(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [visible]);

  if (!visible || !actions || actions.length === 0) return null;

  const animation = reduced ? '' : 'animate-[slideDown_0.3s_ease-out] ';

  return (
    <div
      role="status"
      aria-live="polite"
      className={`absolute top-0 left-0 right-0 z-20 mx-4 mt-1 ${animation}rounded bg-black/70 px-3 py-1.5 pr-8 text-white text-xs shadow-lg`}
    >
      {actions.map((a, i) => (
        <div key={`${i}-${a.playerIdx}-${a.action}`}>
          {t('player.player', { idx: a.playerIdx })}: {bettingActionName(a.action)}
          {a.amount > 0 && ` (${a.amount})`}
        </div>
      ))}
      <button
        type="button"
        onClick={() => setVisible(false)}
        aria-label={t('button.dismiss')}
        className="absolute top-0 right-1 px-2 py-0.5 text-white/70 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-ds-accent"
      >
        ×
      </button>
    </div>
  );
}
