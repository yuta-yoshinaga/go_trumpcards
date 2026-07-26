import { useTranslation } from 'react-i18next';
import { useTransientToast } from '../hooks/useTransientToast';
import { TOAST_DURATION } from '../styles/toastDurations';
import type { StrategyStyle } from '../utils/metaAiAdaptation';
import { Toast } from './Toast';

/** Props for the MetaAiToast component. */
export interface MetaAiToastProps {
  /** Current strategy style of the CPU meta-AI. Undefined when meta-AI is disabled. */
  strategyStyle: StrategyStyle | undefined;
}

/**
 * Toast notification for CPU meta-AI strategy changes. Shows when the strategy
 * changes, auto-dismisses after the medium duration, and can be dismissed early
 * via the close button or Escape. A thin wrapper over the shared {@link Toast}
 * banner + {@link useTransientToast} lifecycle (issue #4313).
 */
export function MetaAiToast({ strategyStyle }: MetaAiToastProps) {
  const { t } = useTranslation('common');
  const { visible, dismiss } = useTransientToast(strategyStyle, TOAST_DURATION.medium, {
    active: strategyStyle !== undefined,
  });

  if (!visible || !strategyStyle) return null;

  const strategyLabel = t(`metaAi.strategy.${strategyStyle}`);
  return (
    <Toast onDismiss={dismiss} testId="meta-ai-toast">
      {t('metaAi.strategyShift', { strategy: strategyLabel })}
    </Toast>
  );
}
