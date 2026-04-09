import { useTranslation } from 'react-i18next';
import type { AdaptationLevel, StrategyStyle } from '../utils/metaAiAdaptation';

/** Props for the MetaAiIndicator component. */
export interface MetaAiIndicatorProps {
  /** Current adaptation level of the CPU meta-AI. */
  adaptationLevel: AdaptationLevel;
  /** Current strategy style derived from behavior rates. */
  strategyStyle: StrategyStyle;
}

const levelColorClass: Record<AdaptationLevel, string> = {
  learning: 'text-gray-400',
  adapting: 'text-ds-warning',
  adapted: 'text-ds-success',
};

/** Renders a small inline indicator showing CPU meta-AI adaptation level and strategy style. */
export function MetaAiIndicator({ adaptationLevel, strategyStyle }: MetaAiIndicatorProps) {
  const { t } = useTranslation('common');
  return (
    <span className={`ml-2 text-xs ${levelColorClass[adaptationLevel]}`} data-testid="meta-ai-indicator">
      {t(`metaAi.${adaptationLevel}`)} [{t(`metaAi.strategy.${strategyStyle}`)}]
    </span>
  );
}
