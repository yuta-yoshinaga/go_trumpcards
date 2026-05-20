import { useTranslation } from 'react-i18next';
import type { AxisDef } from '../../constants/discoverAxes';
import { focusRingWhite } from '../../styles/buttonStyles';

/** Props for one survey question card. */
export interface MoodQuestionProps {
  /** The axis being asked. */
  readonly axis: AxisDef;
  /** Index into `axis.questions` (0 = first prompt, 1 = second). */
  readonly questionIndex: 0 | 1;
  /** Currently selected option index within `axis.questions[questionIndex].options`, or null if unanswered. */
  readonly selected: number | null;
  /** Called when the user picks an option. */
  readonly onSelect: (optionIndex: number) => void;
  /** Called when the user skips this question. */
  readonly onSkip: () => void;
  /** Total question count for the SR label (e.g. "Question 3 of 8"). */
  readonly questionNumber: number;
  /** Total question count. */
  readonly totalQuestions: number;
}

/**
 * One survey question with two-to-four options rendered as large tap
 * targets. Each option exposes its number key (1-N) as a 22px gold
 * circle — keyboard activation is handled by the parent page so one
 * page-level listener serves the whole survey.
 */
export function MoodQuestion({
  axis,
  questionIndex,
  selected,
  onSelect,
  onSkip,
  questionNumber,
  totalQuestions,
}: MoodQuestionProps) {
  const { t } = useTranslation('discover');
  const subQuestion = axis.questions[questionIndex];
  return (
    <section
      aria-label={t('aria.question', { current: questionNumber, total: totalQuestions })}
      className="flex flex-col gap-4"
    >
      <p className="text-xs uppercase tracking-[0.18em] text-ds-accent">
        {t('eyebrow.question', { current: questionNumber, total: totalQuestions })}
      </p>
      <h2 className="font-serif text-2xl text-ds-text-primary leading-tight">{t(subQuestion.questionI18nKey)}</h2>
      <ul className="flex flex-col gap-2">
        {subQuestion.options.map((opt, idx) => {
          const isSelected = selected === idx;
          return (
            <li key={opt.key}>
              <button
                type="button"
                onClick={() => onSelect(idx)}
                aria-pressed={isSelected}
                className={`w-full min-h-[44px] flex items-center gap-3 px-4 py-3 rounded-md border text-left text-sm transition-colors ${focusRingWhite} ${
                  isSelected
                    ? 'border-ds-accent bg-[rgba(212,168,83,0.18)] text-ds-text-primary'
                    : 'border-ds-border bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover'
                }`}
              >
                <span
                  aria-hidden="true"
                  className="inline-flex h-[22px] w-[22px] shrink-0 items-center justify-center rounded-full bg-ds-accent text-[11px] font-semibold text-ds-text-on-accent"
                >
                  {idx + 1}
                </span>
                <span>{t(opt.i18nKey)}</span>
              </button>
            </li>
          );
        })}
      </ul>
      <button
        type="button"
        onClick={onSkip}
        className={`self-end text-xs text-ds-text-muted hover:text-ds-text-primary underline ${focusRingWhite}`}
      >
        {t('action.skip')}
      </button>
    </section>
  );
}
