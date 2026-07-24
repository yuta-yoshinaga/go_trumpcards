import { useEffect, useRef, useState } from 'react';
import { labellelucieApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import { LaBelleLuciePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { labelleLucieHasLegalMove } from '../utils/labelleLucieLegalMove';

/** La Belle Lucie tutorial step definitions. */
const LL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ll-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ll-fans"]', messageKey: 'tutorial.fan', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ll-redeal"]', messageKey: 'tutorial.redeal', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ll-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric La Belle Lucie phases to i18n phase-label keys. */
const LL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [LaBelleLuciePhase.PLAYING]: 'playing',
  [LaBelleLuciePhase.GAME_CLEAR]: 'gameClear',
  [LaBelleLuciePhase.GAME_OVER]: 'gameOver',
};

/** Renders the La Belle Lucie game page. */
export const LaBelleLuciePage = withTutorial(LaBelleLuciePageContent, 'labellelucie', LL_TUTORIAL_STEPS);

/** Inner content of the La Belle Lucie page, wrapped by TutorialProvider. */
function LaBelleLuciePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('labellelucie');
  const { state, loading, error, exec, retry } = useGameApi(labellelucieApi.exec);
  const [selected, setSelected] = useState<number | null>(null);
  // Whether the last hint's suggested move is currently highlighted on the board.
  // The move coordinates themselves come from `state.hint` (set by the server on a
  // `hint` command); this flag just gates the rings so they only show after the
  // player asks for a hint, then auto-dismiss.
  const [showHint, setShowHint] = useState(false);
  const hintTimerRef = useRef<number | null>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // Clear a stale source selection (and any hint highlight) whenever the board
  // changes (move, redeal, undo, auto-complete) so a selected/hinted index can't
  // point at a different fan.
  // biome-ignore lint/correctness/useExhaustiveDependencies: deps are the change-trigger, not read in the body.
  useEffect(() => {
    setSelected(null);
    setShowHint(false);
  }, [state?.moveCount, state?.redealsLeft]);

  // Cancel a pending hint auto-dismiss timer on unmount.
  useEffect(() => () => window.clearTimeout(hintTimerRef.current ?? undefined), []);

  const phaseNames = usePhaseNames('labellelucie', LL_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const w = Math.round(cardWidth * 0.6);

  if (!state) return <GameSkeleton gameKey="labellelucie" layout={{ kind: 'tableau', topRow: 4, tableau: 9 }} />;

  const isClear = state.phase === LaBelleLuciePhase.GAME_CLEAR;
  const isOver = state.phase === LaBelleLuciePhase.GAME_OVER;
  const isEnd = isClear || isOver;
  const canAct = !isEnd;
  const hasLegalMove = labelleLucieHasLegalMove(state.fans, state.foundation);
  // No legal move left but redeals remain: recommend a redeal before the
  // player wastes time hunting for a move that does not exist.
  const stuck = canAct && state.redealsLeft > 0 && !hasLegalMove;
  // No legal move left and redeals are exhausted: a true deadlock. Guide the
  // player to give up instead of hunting for a move that cannot exist.
  const deadlocked = canAct && state.redealsLeft <= 0 && !hasLegalMove;
  const phaseName = phaseNames[state.phase] ?? '';

  const handleReset = () => {
    hideActionLog();
    setSelected(null);
    // A reset keeps moveCount/redealsLeft unchanged from a fresh board, so the
    // board-change effect may not fire — clear any stale hint highlight here.
    setShowHint(false);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    exec('reset');
  };

  // Ask the server for a hint, then highlight the suggested move (source fan →
  // destination fan/foundation) for a few seconds. Does not execute the move.
  const handleHint = () => {
    exec('hint');
    setShowHint(true);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    hintTimerRef.current = window.setTimeout(() => setShowHint(false), 4000);
  };

  // The move to highlight, only while the highlight is active.
  const hint = showHint ? state.hint : undefined;
  const hintFoundation = hint?.toFoundation === true;

  // Picking a fan: first click selects the source; second click moves to that fan.
  const pickFan = (idx: number) => {
    if (!canAct) return;
    if (selected === null) {
      if ((state.fans[idx]?.length ?? 0) > 0) setSelected(idx);
      return;
    }
    if (selected === idx) {
      setSelected(null);
      return;
    }
    exec('mf', selected, idx);
    setSelected(null);
  };

  const sendToFoundation = () => {
    if (selected === null) return;
    exec('ff', selected);
    setSelected(null);
  };

  /** Renders a fan as a small overlapping vertical pile; only the top is interactive. */
  const renderFan = (fan: Card[], idx: number) => {
    const isHintSource = hint?.fromFan === idx;
    const isHintDest = hint !== undefined && !hintFoundation && hint.toFan === idx;
    // Source ring (info) and destination ring (success) are additive hint cues.
    // A selected fan keeps its own warning ring; hint rings take visual priority
    // via later class order when both would apply to the same fan.
    const hintRing = isHintSource
      ? ' ring-2 ring-ds-info motion-safe:animate-pulse'
      : isHintDest
        ? ' ring-2 ring-ds-success motion-safe:animate-pulse'
        : '';
    return (
      <button
        type="button"
        key={`fan-${idx}`}
        className={`relative flex flex-col items-center rounded p-1 ${selected === idx ? 'ring-2 ring-ds-warning' : ''}${hintRing} ${canAct ? 'cursor-pointer' : ''}`}
        style={{ minHeight: Math.round(w * 1.4) }}
        onClick={canAct ? () => pickFan(idx) : undefined}
        disabled={!canAct}
        data-testid={`fan-${idx}`}
        data-hint-source={isHintSource ? 'true' : undefined}
        data-hint-dest={isHintDest ? 'true' : undefined}
      >
        {fan.length === 0 ? (
          <div
            className="rounded border border-dashed border-white/25 bg-black/20"
            style={{ width: w, height: Math.round(w * 1.4) }}
            title={t('empty')}
          />
        ) : (
          fan.map((c, i) => (
            <div key={`fan-${idx}-${i}`} style={{ marginTop: i === 0 ? 0 : -Math.round(w * 1.0) }}>
              <CardImage card={c} width={w} />
            </div>
          ))
        )}
      </button>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.labellelucie')}
      gameThemeBg={gameTheme.labellelucie.bg}
      phaseName={phaseName}
      gamePath="/labellelucie"
      gameEndFlag={isEnd}
      winShow={isClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-3 lg:px-8">
        {/* Foundations */}
        <div className="mb-3" data-tutorial="ll-foundation">
          <span className="text-ds-text-muted text-[11px]">{t('foundation')}</span>
          <div
            className={`flex gap-1 mt-0.5 rounded${hintFoundation ? ' ring-2 ring-ds-success motion-safe:animate-pulse p-1' : ''}`}
            data-testid="ll-foundation-row"
            data-hint-foundation={hintFoundation ? 'true' : undefined}
          >
            {state.foundation.map((pile, i) => (
              <button
                type="button"
                key={`fnd-${i}`}
                className={`rounded ${selected !== null ? 'ring-1 ring-ds-success' : ''} ${canAct ? 'cursor-pointer' : ''}`}
                onClick={selected !== null ? sendToFoundation : undefined}
                disabled={selected === null}
                data-testid={`foundation-${i}`}
              >
                {pile.length > 0 ? (
                  <CardImage card={pile[pile.length - 1]} width={w} />
                ) : (
                  <div
                    className="rounded border border-dashed border-white/25 bg-black/20"
                    style={{ width: w, height: Math.round(w * 1.4) }}
                  />
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Fans */}
        <div className="grid grid-cols-6 sm:grid-cols-9 gap-1" data-tutorial="ll-fans">
          {state.fans.map((fan, i) => renderFan(fan, i))}
        </div>

        <div className="mt-2 text-ds-text-muted text-xs">
          {t('redealsLeft', { count: state.redealsLeft })} · {t('moveCount', { count: state.moveCount })}
        </div>
        {stuck && (
          <div
            className="mt-1 flex items-center gap-2 text-ds-warning text-sm font-medium"
            role="status"
            data-testid="ll-stuck-banner"
          >
            <span>{t('stuckRedeal')}</span>
            <span className="rounded-full bg-ds-warning/20 px-2 py-0.5 text-xs font-bold tabular-nums">
              {t('redealsLeftBadge', { count: state.redealsLeft })}
            </span>
          </div>
        )}
        {deadlocked && (
          <div
            className="mt-1 flex items-center gap-2 text-ds-danger text-sm font-medium"
            role="status"
            data-testid="ll-deadlock-banner"
          >
            <span>{t('stuckDeadlock')}</span>
          </div>
        )}
        {canAct && (
          <div className="mt-1 text-ds-text-primary text-xs">
            {selected === null ? t('selectSource') : t('selectDestination')}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        <ActionLogSection
          isEndPhase={isEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.labellelucie.footer} px-3 py-2.5`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex flex-wrap gap-2 items-center">
          {canAct && (
            <button
              type="button"
              className={`${btnWarning}${stuck ? ' motion-safe:animate-pulse' : ''}`}
              onClick={() => exec('rd')}
              disabled={loading || state.redealsLeft <= 0}
              data-tutorial="ll-redeal"
              data-testid="redeal-button"
            >
              {t('redeal')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnSuccess}
              onClick={() => exec('ac')}
              disabled={loading}
              data-testid="autocomplete-button"
            >
              {t('autoComplete')}
            </button>
          )}
          {canAct && state.canUndo && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('u')}
              disabled={loading}
              data-testid="undo-button"
            >
              {t('undo')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnPrimary}
              onClick={handleHint}
              disabled={loading}
              data-testid="hint-button"
            >
              {t('hint')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={`${btnSecondary}${deadlocked ? ' motion-safe:animate-pulse' : ''}`}
              onClick={() => exec('giveup')}
              disabled={loading}
              data-testid="giveup-button"
            >
              {t('giveup')}
            </button>
          )}
          <GameResetButton
            isGameEnd={isEnd}
            onReset={handleReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="ll-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
