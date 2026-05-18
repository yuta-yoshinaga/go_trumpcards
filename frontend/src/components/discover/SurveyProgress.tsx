import { useTranslation } from 'react-i18next';
import { TOTAL_QUESTIONS } from '../../constants/discoverAxes';

/** Props for the survey progress indicator (deck of card backs). */
export interface SurveyProgressProps {
  /** 1-indexed current question (1..TOTAL_QUESTIONS). */
  readonly current: number;
}

/**
 * Renders 8 card-back tiles. The current one is highlighted gold and
 * scaled up; completed ones are filled; future ones are dim outlines.
 * SR users get a polite live region with "Question X of Y".
 */
export function SurveyProgress({ current }: SurveyProgressProps) {
  const { t } = useTranslation('discover');
  const total = TOTAL_QUESTIONS;
  return (
    <div className="flex flex-col gap-2">
      <ul aria-hidden="true" className="flex gap-1.5 items-end justify-center">
        {Array.from({ length: total }, (_, i) => {
          const idx = i + 1;
          const state = idx < current ? 'done' : idx === current ? 'current' : 'future';
          return (
            <li
              key={idx}
              data-testid="survey-progress-card"
              data-state={state}
              className={
                state === 'current'
                  ? 'w-7 h-10 rounded-sm bg-ds-accent shadow-[0_0_12px_rgba(212,168,83,0.6)] scale-110 transition-transform'
                  : state === 'done'
                    ? 'w-5 h-7 rounded-sm bg-ds-accent/70'
                    : 'w-5 h-7 rounded-sm border border-ds-border/60'
              }
            />
          );
        })}
      </ul>
      <p aria-live="polite" aria-atomic="true" className="sr-only">
        {t('aria.progress', { current, total })}
      </p>
    </div>
  );
}
