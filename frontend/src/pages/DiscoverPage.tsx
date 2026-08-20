import { AnimatePresence, motion } from 'framer-motion';
import { useCallback, useEffect, useReducer, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { DiscoverShell } from '../components/discover/DiscoverShell';
import { DiscoverSkeleton } from '../components/discover/DiscoverSkeleton';
import { MoodQuestion } from '../components/discover/MoodQuestion';
import { SurveyProgress } from '../components/discover/SurveyProgress';
import { AXES, AXIS_KEYS, type AxisKey, TOTAL_QUESTIONS } from '../constants/discoverAxes';
import { useDiscoverI18nBundle } from '../hooks/useDiscoverI18nBundle';
import { useDocumentTitle } from '../hooks/useDocumentTitle';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { useSurveyDraft } from '../hooks/useSurveyDraft';
import { focusRingWhite } from '../styles/buttonStyles';
import { encodeMood } from '../utils/urlMoodCodec';

/** Direction of the most recent step change — drives the slide animation. */
type SlideDirection = 'forward' | 'backward';

interface SurveyState {
  /** Current step, 0..TOTAL_QUESTIONS (TOTAL means "submitted"). */
  readonly step: number;
  /** Direction of the last transition, so motion follows navigation. */
  readonly direction: SlideDirection;
}

type Action = { type: 'advance' } | { type: 'back' };

function reducer(state: SurveyState, action: Action): SurveyState {
  switch (action.type) {
    case 'advance':
      return { step: Math.min(state.step + 1, TOTAL_QUESTIONS), direction: 'forward' };
    case 'back':
      return { step: Math.max(state.step - 1, 0), direction: 'backward' };
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
  const { t: tc } = useTranslation('common');
  useDocumentTitle(tc('nav.discover'));
  const navigate = useNavigate();
  const bundleReady = useDiscoverI18nBundle();
  const reducedMotion = useReducedMotion();
  const { axes, setAnswer, reset: resetDraft } = useSurveyDraft();
  // `useSurveyDraft` lazy-initializes `axes` from localStorage on its first
  // render, so the `axes` we read here on first paint is already the
  // restored draft. We use the reducer's lazy-init form (3rd arg) to seed
  // the step pointer in one shot — no post-mount effect needed.
  const [state, dispatch] = useReducer(reducer, axes, (initialAxes) => ({
    step: firstUnansweredStep(initialAxes),
    direction: 'forward' as SlideDirection,
  }));

  const current = stepToAxisQuestion(state.step);

  // Drop a second answer fired before the next question's render — without
  // this guard, two clicks within the same React batch share the *same*
  // captured `current` and `dispatch({type:'advance'})` runs twice, silently
  // skipping the next sub-dimension. The lock releases whenever `state.step`
  // actually moves, so legitimate keyboard runs (one key per question)
  // are unaffected. See #1898.
  const lockedStepRef = useRef<number | null>(null);
  if (lockedStepRef.current !== null && lockedStepRef.current !== state.step) {
    lockedStepRef.current = null;
  }

  const handleSelect = useCallback(
    (optIdx: number) => {
      if (!current) return;
      if (lockedStepRef.current === state.step) return;
      lockedStepRef.current = state.step;
      setAnswer(current.axis, current.qIdx, optIdx);
      dispatch({ type: 'advance' });
    },
    [current, state.step, setAnswer],
  );

  const handleSkip = useCallback(() => {
    if (!current) return;
    if (lockedStepRef.current === state.step) return;
    lockedStepRef.current = state.step;
    setAnswer(current.axis, current.qIdx, null);
    dispatch({ type: 'advance' });
  }, [current, state.step, setAnswer]);

  const handleBack = useCallback(() => dispatch({ type: 'back' }), []);

  // Submit when the user finishes the last question.
  // `resetDraft()` clears `axes` to all-null, which is itself a dependency of
  // this effect — without the ref guard the effect re-fires with empty axes
  // and a second navigate() overwrites the good URL with `m=-,-&s=-,-&...`.
  const submittedRef = useRef(false);
  useEffect(() => {
    if (state.step < TOTAL_QUESTIONS) return;
    if (submittedRef.current) return;
    submittedRef.current = true;
    const query = encodeMood({
      mood: [axes.mood[0], axes.mood[1]],
      skill: [axes.skill[0], axes.skill[1]],
      social: [axes.social[0], axes.social[1]],
      theme: [axes.theme[0], axes.theme[1]],
    });
    navigate(`/discover/result?${query}`, { replace: false });
    resetDraft();
  }, [state.step, axes, navigate, resetDraft]);

  // Browser back integration (#1899). Each forward advance pushes a history
  // entry so a subsequent browser-back lands within /discover instead of
  // exiting the survey. The popstate listener catches that back press and
  // dispatches the in-survey back action, so the native button walks
  // through questions exactly like the on-page "← previous question" link.
  // We only push on `forward` transitions — a back-walk would otherwise
  // re-stack entries we just consumed.
  useEffect(() => {
    if (state.step === 0) return;
    if (state.direction !== 'forward') return;
    window.history.pushState({ discoverStep: state.step }, '');
  }, [state.step, state.direction]);

  useEffect(() => {
    function onPopState() {
      if (state.step > 0) {
        dispatch({ type: 'back' });
      }
    }
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, [state.step]);

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
      const optCount = AXES[current.axis].questions[current.qIdx].options.length;
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

  if (!bundleReady) {
    return (
      <DiscoverShell testId="discover-survey">
        <DiscoverSkeleton />
      </DiscoverShell>
    );
  }

  // DR-4: slide-in / slide-out on step change, reduced to a 50ms fade
  // when the user prefers reduced motion (DR-7 #3). The slide direction
  // follows the user's navigation — forward enters from the right and
  // exits to the left; backward reverses both sides so Back feels like
  // "rewinding" rather than another forward step.
  const dirSign = state.direction === 'forward' ? 1 : -1;
  const transition = reducedMotion ? { duration: 0.05 } : { duration: 0.2, ease: 'easeOut' as const };
  const initial = reducedMotion ? { opacity: 0 } : { opacity: 0, x: 24 * dirSign };
  const animate = reducedMotion ? { opacity: 1 } : { opacity: 1, x: 0 };
  const exit = reducedMotion ? { opacity: 0 } : { opacity: 0, x: -24 * dirSign };

  return (
    <DiscoverShell testId="discover-survey">
      <div className="flex-1 min-h-0 flex flex-col items-center justify-start px-4 py-8 gap-6">
        <SurveyProgress current={state.step + 1} />
        <div className="w-full max-w-md">
          <AnimatePresence mode="wait" initial={false}>
            <motion.div key={state.step} initial={initial} animate={animate} exit={exit} transition={transition}>
              <MoodQuestion
                axis={AXES[current.axis]}
                questionIndex={current.qIdx}
                selected={axes[current.axis][current.qIdx]}
                onSelect={handleSelect}
                onSkip={handleSkip}
                questionNumber={state.step + 1}
                totalQuestions={TOTAL_QUESTIONS}
              />
            </motion.div>
          </AnimatePresence>
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
    </DiscoverShell>
  );
}

export default DiscoverPage;
