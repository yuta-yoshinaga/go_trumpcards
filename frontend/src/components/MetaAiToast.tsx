import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { StrategyStyle } from '../utils/metaAiAdaptation';

/** Props for the MetaAiToast component. */
export interface MetaAiToastProps {
  /** Current strategy style of the CPU meta-AI. */
  strategyStyle: StrategyStyle;
}

const DISMISS_MS = 3000;

/** Toast notification for CPU meta-AI strategy changes. Auto-dismisses after 3 seconds. */
export function MetaAiToast({ strategyStyle }: MetaAiToastProps) {
  const { t } = useTranslation('common');
  const [visible, setVisible] = useState(false);
  const prevRef = useRef<StrategyStyle | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const isFirstRender = useRef(true);

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      prevRef.current = strategyStyle;
      return;
    }
    if (strategyStyle !== prevRef.current) {
      prevRef.current = strategyStyle;
      setVisible(true);
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setVisible(false), DISMISS_MS);
    }
    return () => clearTimeout(timerRef.current);
  }, [strategyStyle]);

  if (!visible) return null;

  const strategyLabel = t(`metaAi.strategy.${strategyStyle}`);

  return (
    <div
      role="status"
      aria-live="polite"
      data-testid="meta-ai-toast"
      className="absolute top-0 left-0 right-0 z-20 mx-4 mt-1 animate-[slideDown_0.3s_ease-out] rounded bg-black/70 px-3 py-1.5 text-white text-xs shadow-lg"
    >
      {t('metaAi.strategyShift', { strategy: strategyLabel })}
    </div>
  );
}
