import { useTranslation } from 'react-i18next';
import { btnPrimary, btnSecondary } from '../../styles/buttonStyles';
import type { TutorialAdvanceOn, TutorialPlacement } from '../../types/tutorial';

/** Props for the TutorialTooltip component. */
export interface TutorialTooltipProps {
  /** The message to display in the tooltip. */
  message: string;
  /** Tooltip placement relative to the target element. */
  placement: TutorialPlacement;
  /** Zero-based index of the current step. */
  stepIndex: number;
  /** Total number of steps. */
  totalSteps: number;
  /** Called when advancing to the next step. */
  onNext: () => void;
  /** Called when the tutorial is skipped. */
  onSkip: () => void;
  /** How to advance: 'next' shows a next button, 'click' hides it. */
  advanceOn: TutorialAdvanceOn;
}

/** Renders a glass-panel tooltip with step indicator and navigation buttons for the tutorial. */
export function TutorialTooltip({ message, stepIndex, totalSteps, onNext, onSkip, advanceOn }: TutorialTooltipProps) {
  const { t } = useTranslation('tutorial');
  const isLastStep = stepIndex === totalSteps - 1;

  return (
    <div role="status" aria-live="polite" className="glass-panel rounded-lg shadow-xl p-4 max-w-xs">
      <p className="text-white text-sm mb-3">{message}</p>
      <div className="flex items-center justify-between">
        <span className="text-gray-300 text-xs">
          {stepIndex + 1} / {totalSteps}
        </span>
        <div className="flex gap-2">
          <button type="button" className={btnSecondary} onClick={onSkip}>
            {t('skip')}
          </button>
          {advanceOn === 'next' && (
            <button type="button" className={btnPrimary} onClick={onNext}>
              {isLastStep ? t('complete') : t('next')}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
