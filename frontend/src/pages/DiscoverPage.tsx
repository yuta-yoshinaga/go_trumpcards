import { useCallback, useEffect, useReducer } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { MoodQuestion } from '../components/discover/MoodQuestion';
import { SurveyProgress } from '../components/discover/SurveyProgress';
import { AXES, AXIS_KEYS, type AxisKey, TOTAL_QUESTIONS } from '../constants/discoverAxes';
import { useSurveyDraft } from '../hooks/useSurveyDraft';
import { focusRingWhite } from '../styles/buttonStyles';
import { encodeMood } from '../utils/urlMoodCodec';

interface SurveyState {
  /** Current step, 0..TOTAL_QUESTIONS (TOTAL means "submitted"). */
  readonly step: number;
}

type Action = { type: 'advance' } | { type: 'back' } | { type: 'reset' };

function reducer(state: SurveyState, action: Action): SurveyState {
  switch (action.type) {
    case 'advance':
      return { step: Math.min(state.step + 1, TOTAL_QUESTIONS) };
    case 'back':
      return { step: Math.max(state.step - 1, 0) };
    case 'reset':
      return { step: 0 };
  }
}

/** Map a step (0..7) to its axis + question index pair. */
function stepToAxisQuestion(step: number): { axis: AxisKey; qIdx: 0 | 1 } | null {
  if (step < 0 || step >= TOTAL_QUESTIONS) return null;
  const axisIdx = Math.floor(step / 2);
  const qIdx = (step % 2) as 0 | 1;
  return { axis: AXIS_KEYS[axisIdx], qIdx };
}

/** Find the first step whose answer is null — resume position after reload. */
function firstUnansweredStep(axes: ReturnType<typeof useSurveyDraft>['axes']): number {
  for (let step = 0; step < TOTAL_QUESTIONS; step++) {
    const m = stepToAxisQuestion(step);
    if (!m) continue;
    if (axes[m.axis][m.qIdx] === null) return step;
  }
  return TOTAL_QUESTIONS;
}

/**
 * Survey driver page. Owns the step pointer; the answers themselves
 * live in `useSurveyDraft` (localStorage-persisted). Keyboard support:
 * 1-N selects an option, Backspace goes back.
 */
export function DiscoverPage() {
  const { t } = useTranslation('discover');
  const navigate = useNavigate();
  const { axes, setAnswer, reset: resetDraft } = useSurveyDraft();
  const [state, dispatch] = useReducer(reducer, { step: 0 });

  // On mount only, resume to the first unanswered step from any stored draft.
  // biome-ignore lint/correctness/useExhaustiveDependencies: intentional mount-only restore
  useEffect(() => {
    dispatch({ type: 'reset' });
    const resumeStep = firstUnansweredStep(axes);
    for (let i = 0; i < resumeStep; i++) dispatch({ type: 'advance' });
  }, []);

  const current = stepToAxisQuestion(state.step);

  const handleSelect = useCallback(
    (optIdx: number) => {
      if (!current) return;
      setAnswer(current.axis, current.qIdx, optIdx);
      dispatch({ type: 'advance' });
    },
    [current, setAnswer],
  );

  const handleSkip = useCallback(() => {
    if (!current) return;
    setAnswer(current.axis, current.qIdx, null);
    dispatch({ type: 'advance' });
  }, [current, setAnswer]);

  const handleBack = useCallback(() => dispatch({ type: 'back' }), []);

  // Submit when the user finishes the last question.
  useEffect(() => {
    if (state.step < TOTAL_QUESTIONS) return;
    const query = encodeMood({
      mood: [axes.mood[0], axes.mood[1]],
      skill: [axes.skill[0], axes.skill[1]],
      social: [axes.social[0], axes.social[1]],
      theme: [axes.theme[0], axes.theme[1]],
    });
    resetDraft();
    navigate(`/discover/result?${query}`, { replace: false });
  }, [state.step, axes, navigate, resetDraft]);

  // Keyboard shortcuts: digit keys pick options; Backspace goes back.
  useEffect(() => {
    function onKey(ev: KeyboardEvent) {
      if (!current) return;
      const target = ev.target as HTMLElement | null;
      if (target?.tagName === 'INPUT' || target?.tagName === 'TEXTAREA') return;
      if (ev.key === 'Backspace' && state.step > 0) {
        ev.preventDefault();
        handleBack();
        return;
      }
      const n = Number.parseInt(ev.key, 10);
      const optCount = AXES[current.axis].options.length;
      if (Number.isInteger(n) && n >= 1 && n <= optCount) {
        ev.preventDefault();
        handleSelect(n - 1);
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [current, state.step, handleSelect, handleBack]);

  if (!current) {
    // Submitted — the effect above is navigating away; render an empty shell.
    return null;
  }

  return (
    <div className="flex-1 min-h-0 flex flex-col items-center justify-start px-4 py-8 gap-6">
      <SurveyProgress current={state.step + 1} />
      <div className="w-full max-w-md">
        <MoodQuestion
          axis={AXES[current.axis]}
          questionIndex={current.qIdx}
          selected={axes[current.axis][current.qIdx]}
          onSelect={handleSelect}
          onSkip={handleSkip}
          questionNumber={state.step + 1}
          totalQuestions={TOTAL_QUESTIONS}
        />
        {state.step > 0 && (
          <button
            type="button"
            onClick={handleBack}
            className={`mt-4 text-xs text-ds-text-muted hover:text-ds-text-primary underline ${focusRingWhite}`}
          >
            {t('action.back')}
          </button>
        )}
      </div>
    </div>
  );
}

export default DiscoverPage;
