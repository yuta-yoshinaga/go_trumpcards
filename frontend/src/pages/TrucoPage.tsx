import { useCallback, useEffect, useMemo } from 'react';
import { trucoApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TrucoResponse } from '../types/card';
import { TrucoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

/** Tutorial steps for the Truco page. */
const TRUCO_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="truco-score"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="truco-rankref"]', messageKey: 'tutorial.rankRef', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="truco-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="truco-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="truco-call"]', messageKey: 'tutorial.call', placement: 'top', advanceOn: 'next' },
];

/** Betting-level i18n key suffixes indexed by level (0=none .. 3=Vale Cuatro). */
const LEVEL_KEYS = ['none', 'truco', 'retruco', 'valecuatro'] as const;

/**
 * Inner content for the Truco page (wrapped by `withTutorial` below).
 *
 * Renders the 2-player South-American bluffing trick-taking game: a 40-card
 * deck with no must-follow rule, best-of-3 bazas per hand, and the "Truco"
 * stake escalation (Truco → Retruco → Vale Cuatro) that the opponent may
 * accept, decline, or re-raise. First to the match target (default 15) wins.
 */
function TrucoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('truco');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<TrucoResponse, Parameters<typeof trucoApi.exec>>(trucoApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('truco', state);

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handlePlay = useCallback((idx: number) => void dispatch('play', idx), [dispatch]);
  const handleTruco = useCallback(() => void dispatch('truco'), [dispatch]);
  const handleAccept = useCallback(() => void dispatch('accept'), [dispatch]);
  const handleDecline = useCallback(() => void dispatch('decline'), [dispatch]);
  const handleNext = useCallback(() => void dispatch('next'), [dispatch]);

  // Keyboard shortcuts (must run before the early return): 1/2/3 play a card,
  // t declares truco, a/d accept/decline a truco call, n advances at a boundary.
  const kbHumanIdx = state?.players.findIndex((p) => p.isHuman) ?? -1;
  const kbHumanCardCount = state?.players[kbHumanIdx]?.cards?.length ?? 0;
  const kbIsHumanPlayTurn = state?.phase === TrucoPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const kbIsHumanRespond = state?.phase === TrucoPhase.RESPOND && state.responderIdx === kbHumanIdx;
  const kbCanTruco = !!state?.canDeclareTruco;
  const kbIsBoundary = state?.phase === TrucoPhase.TRICK_END || state?.phase === TrucoPhase.HAND_END;
  const actionBindings = useMemo(
    () => [
      { key: '1', action: () => handlePlay(0), enabled: kbIsHumanPlayTurn && kbHumanCardCount >= 1 },
      { key: '2', action: () => handlePlay(1), enabled: kbIsHumanPlayTurn && kbHumanCardCount >= 2 },
      { key: '3', action: () => handlePlay(2), enabled: kbIsHumanPlayTurn && kbHumanCardCount >= 3 },
      { key: 't', action: handleTruco, enabled: (kbIsHumanPlayTurn || kbIsHumanRespond) && kbCanTruco },
      { key: 'a', action: handleAccept, enabled: kbIsHumanRespond },
      { key: 'd', action: handleDecline, enabled: kbIsHumanRespond },
      { key: 'n', action: handleNext, enabled: kbIsBoundary },
    ],
    [
      handlePlay,
      handleTruco,
      handleAccept,
      handleDecline,
      handleNext,
      kbIsHumanPlayTurn,
      kbHumanCardCount,
      kbIsHumanRespond,
      kbCanTruco,
      kbIsBoundary,
    ],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) {
    return <GameSkeleton gameKey="truco" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 3 }} />;
  }

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const human = state.players[humanIdx];
  const cpu = state.players.find((p) => !p.isHuman);
  const isPlay = state.phase === TrucoPhase.PLAY;
  const isRespond = state.phase === TrucoPhase.RESPOND;
  const isTrickEnd = state.phase === TrucoPhase.TRICK_END;
  const isHandEnd = state.phase === TrucoPhase.HAND_END;
  const isGameEnd = state.phase === TrucoPhase.GAME_END || state.gameEndFlag;
  const isHumanPlayTurn = isPlay && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanRespond = isRespond && state.responderIdx === humanIdx;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isHandEnd
      ? t('phase.handEnd')
      : isTrickEnd
        ? t('phase.trickEnd')
        : isRespond
          ? t('phase.respond')
          : t('phase.play');

  const levelLabel = (level: number) => t(`level.${LEVEL_KEYS[level] ?? 'none'}`);

  const youPoints = state.matchPoints[humanIdx] ?? 0;
  const cpuPoints = state.matchPoints[humanIdx === 0 ? 1 : 0] ?? 0;

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const params = { p0: String(youPoints), p1: String(cpuPoints) };
    return state.winnerIdx === humanIdx ? t('result.youWin', params) : t('result.cpuWin', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.truco')}
      gameThemeBg={gameTheme.truco.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanPlayTurn || isHumanRespond}
      gamePath="/truco"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        <div className="text-ds-text-primary text-center mb-3" data-tutorial="truco-score">
          <span className="mr-4">
            {t('header.match')} — {t('header.you')}: {youPoints} / {t('header.cpu')}: {cpuPoints} ({t('header.target')}:{' '}
            {state.matchTarget})
          </span>
        </div>
        <div className="text-ds-text-muted text-center text-sm mb-3">
          <span className="mr-4">
            {t('header.hand')}: {state.handNumber}
          </span>
          <span className="mr-4">
            {t('header.baza')}: {state.trickNumber}
          </span>
          <span data-testid="truco-stake">
            {t('header.stake')}: {state.handStake} ({levelLabel(state.acceptedLevel)})
          </span>
        </div>

        <details
          className="my-3 p-2 rounded bg-black/30 max-w-md mx-auto"
          data-testid="truco-rank-ref"
          data-tutorial="truco-rankref"
        >
          <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('rankRef.title')}</summary>
          <div className="mt-2 text-ds-text-muted text-xs space-y-1">
            <div>
              <span className="font-semibold text-ds-text-primary">{t('rankRef.matadores')}</span>
              <div className="mt-0.5">{t('rankRef.matadoresOrder')}</div>
            </div>
            <div>
              <span className="font-semibold text-ds-text-primary">{t('rankRef.common')}</span>
              <div className="mt-0.5">{t('rankRef.commonOrder')}</div>
            </div>
            <div className="pt-1">{t('rankRef.note')}</div>
          </div>
        </details>

        <div className="flex flex-wrap items-start gap-4 mb-4">
          <div className="p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {t('header.cpu')}: {cpu?.cardCount ?? 0} / {t('header.tricks')}: {cpu?.trickCount ?? 0}
          </div>
        </div>

        <TrickDisplay
          currentTrick={state.currentTrick}
          players={state.players}
          cardWidth={cardWidth}
          label={t('currentTrick')}
          dataTutorial="truco-trick"
        />

        {resultBanner && (
          <div className="text-center text-xl my-4 text-ds-accent font-semibold" role="status">
            {resultBanner}
          </div>
        )}

        {isHumanRespond && state.trucoCallerIdx >= 0 && (
          <div className="text-center text-lg my-2 text-ds-accent" role="status">
            {t('respondPrompt', { name: t('header.cpu'), level: levelLabel(state.pendingLevel) })}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
        <ErrorAlert message={error} onRetry={retry} />

        {human && human.cards.length > 0 && (
          <div className="mt-4" data-tutorial="truco-hand">
            <div className="text-ds-text-muted text-sm mb-1">
              {t('header.you')}: {human.cardCount} / {t('header.tricks')}: {human.trickCount}
            </div>
            <div className="flex flex-wrap gap-2">
              {human.cards.map((card, idx) => (
                <button
                  key={`${card.design}-${card.value}-${idx}`}
                  type="button"
                  onClick={() => handlePlay(idx)}
                  disabled={loading || !isHumanPlayTurn}
                  aria-label={tc('card.play', { card: cardAlt(card) })}
                  className="disabled:opacity-50"
                >
                  <CardImage card={card} width={cardWidth} />
                </button>
              ))}
            </div>
          </div>
        )}

        <div className="mt-4 flex flex-wrap gap-2">
          {isHumanPlayTurn && state.canDeclareTruco && (
            <button
              type="button"
              className={btnWarning}
              onClick={handleTruco}
              disabled={loading}
              data-tutorial="truco-call"
            >
              {t('actions.truco')}
            </button>
          )}
          {isHumanRespond && (
            <>
              <button type="button" className={btnSuccess} onClick={handleAccept} disabled={loading}>
                {t('actions.accept')}
              </button>
              <button type="button" className={btnDanger} onClick={handleDecline} disabled={loading}>
                {t('actions.decline')}
              </button>
              {state.canDeclareTruco && (
                <button type="button" className={btnWarning} onClick={handleTruco} disabled={loading}>
                  {t('actions.raise')}
                </button>
              )}
            </>
          )}
          {(isTrickEnd || isHandEnd) && (
            <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
              {t('actions.next')}
            </button>
          )}
          <button type="button" className={btnPrimary} onClick={() => requestConfirm(handleReset)} disabled={loading}>
            {t('actions.reset')}
          </button>
          <label className="flex items-center gap-1 text-ds-text-primary text-sm" data-testid="truco-hint-toggle">
            <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
            {tc('hint')}
          </label>
        </div>
      </div>

      <ActionLogSection
        isEndPhase={isGameEnd}
        actionLog={actionLog}
        showActionLog={showActionLog}
        hideActionLog={hideActionLog}
      />
    </GamePageShell>
  );
}

/** Truco page wrapped with TutorialProvider. */
export const TrucoPage = withTutorial(TrucoPageContent, 'truco', TRUCO_TUTORIAL_STEPS);
