import { useCallback, useEffect } from 'react';
import { briscolaApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BriscolaResponse } from '../types/card';
import { BriscolaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';

/** Tutorial steps for the Briscola page. v1 ships without a guided tour. */
const BRISCOLA_TUTORIAL_STEPS: TutorialStep[] = [];

/**
 * Renders the Briscola game page: 2-player Italian trick-taking with a face-up
 * trump card, no must-follow rule, and per-card point values (A=11, 3=10,
 * K=4, Q=3, J=2). Players hold 3 cards, replenished from the stock after each
 * trick; the game ends when all 40 cards have been played and the player with
 * more than 60 points wins.
 */
/** Inner Briscola page content; wrapped by withTutorial below. */
function BriscolaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('briscola');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<BriscolaResponse, Parameters<typeof briscolaApi.exec>>(briscolaApi.exec);
  const { cardWidth } = useCardDimensions();

  // Initial reset on mount.
  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNext = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleViewLog = useCallback(() => {
    void showActionLog();
  }, [showActionLog]);

  if (!state) {
    return <GameSkeleton gameKey="briscola" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 3 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const cpu = state.players.find((p) => !p.isHuman);
  const isPlayPhase = state.phase === BriscolaPhase.PLAY;
  const isTrickEnd = state.phase === BriscolaPhase.TRICK_END;
  const isGameEnd = state.phase === BriscolaPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd ? t('phase.gameEnd') : isTrickEnd ? t('phase.trickEnd') : t('phase.play');

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const p0 = state.players[0]?.points ?? 0;
    const p1 = state.players[1]?.points ?? 0;
    const params = { p0: String(p0), p1: String(p1) };
    if (state.winnerIdx === 0) return t('result.youWin', params);
    if (state.winnerIdx === 1) return t('result.cpuWin', params);
    return t('result.tie', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.briscola')}
      gameThemeBg={gameTheme.briscola.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/briscola"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerIdx === 0}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        <div className="text-ds-text-primary text-center mb-3">
          <span className="mr-4">
            {t('header.trick')}: {state.trickNumber}
          </span>
          <span className="mr-4">
            {t('header.stock')}: {state.stockRemaining}
          </span>
          <span>
            {t('header.points')} — {t('header.you')}: {human?.points ?? 0} / {t('header.cpu')}: {cpu?.points ?? 0}
          </span>
        </div>

        {/* CPU info + trump card */}
        <div className="flex flex-wrap items-start gap-4 mb-4">
          <div className="p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {t('header.cpu')}: {cpu?.cardCount ?? 0} / {t('header.tricks')}: {cpu?.trickCount ?? 0}
          </div>
          <div className="p-2 rounded bg-black/30 flex items-center gap-2">
            <span className="text-ds-text-muted text-sm">
              {state.trumpCard ? t('header.trump') : t('header.trumpNone')}
            </span>
            {state.trumpCard && <CardImage card={state.trumpCard} width={Math.round(cardWidth * 0.7)} />}
          </div>
        </div>

        {/* Current trick */}
        <TrickDisplay
          currentTrick={state.currentTrick}
          players={state.players}
          cardWidth={cardWidth}
          label={tc('currentTrick', { defaultValue: 'Current trick' })}
        />

        {/* Result banner */}
        {resultBanner && (
          <div className="text-center text-xl my-4 text-ds-accent font-semibold" role="status">
            {resultBanner}
          </div>
        )}

        <GameMessageBox
          message={state.message}
          messageCode={state.messageCode}
          messageParams={state.messageParams}
          ns="briscola"
        />
        <ErrorAlert error={error} onRetry={retry} />

        {/* Human hand */}
        {human && human.cards.length > 0 && (
          <div className="mt-4">
            <div className="text-ds-text-muted text-sm mb-1">
              {t('header.you')}: {human.cardCount} / {t('header.tricks')}: {human.trickCount}
            </div>
            <div className="flex flex-wrap gap-2">
              {human.cards.map((card, idx) => (
                <button
                  key={`${card.design}-${card.value}-${idx}`}
                  type="button"
                  onClick={() => handlePlay(idx)}
                  disabled={loading || !isHumanTurn}
                  aria-label={`Play ${card.design} ${card.value}`}
                  className="disabled:opacity-50"
                >
                  <CardImage card={card} width={cardWidth} />
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Phase-specific controls */}
        <div className="mt-4 flex flex-wrap gap-2">
          {isTrickEnd && (
            <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
              {t('actions.next')}
            </button>
          )}
          <button type="button" className={btnPrimary} onClick={() => requestConfirm(handleReset)} disabled={loading}>
            {t('actions.reset')}
          </button>
        </div>
      </div>

      <GameFooter onViewLog={handleViewLog} disabled={loading} />
      <ActionLogSection actionLog={actionLog} onClose={hideActionLog} />
    </GamePageShell>
  );
}

/** Briscola page wrapped with TutorialProvider. */
export const BriscolaPage = withTutorial(BriscolaPageContent, 'briscola', BRISCOLA_TUTORIAL_STEPS);
