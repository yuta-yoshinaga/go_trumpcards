import { useTranslation } from 'react-i18next';
import { useIsMobile } from '../hooks/useCardDimensions';

/** Props for {@link CpuAccordion}. */
export interface CpuAccordionProps {
  children: React.ReactNode;
  playerCount: number;
  dataTutorial?: string;
}

/** Collapsible wrapper for CPU player cards. Closed by default on mobile to reduce scroll. */
export function CpuAccordion({ children, playerCount, dataTutorial }: CpuAccordionProps) {
  const { t } = useTranslation('common');
  const isMobile = useIsMobile();

  return (
    <details
      open={!isMobile || undefined}
      className="mb-3 rounded-lg bg-black/20 border border-white/10"
      data-testid="cpu-accordion"
      {...(dataTutorial ? { 'data-tutorial': dataTutorial } : {})}
    >
      <summary className="cursor-pointer select-none px-2 py-1.5 text-ds-text-primary text-sm font-bold">
        {t('label.cpuOpponents', { count: playerCount })}
      </summary>
      <div className="px-1 pb-1">{children}</div>
    </details>
  );
}
