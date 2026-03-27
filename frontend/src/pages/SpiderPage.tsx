import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { SpiderSkeleton } from '../components/skeleton/SpiderSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSpiderGame } from '../hooks/useSpiderGame';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnDanger, btnPrimary, btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { SpiderPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

/** Spider Solitaire tutorial step definitions. */
const SPD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="spd-stock-pile"]',
    messageKey: 'tutorial.stockPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spd-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spd-completed-suits"]',
    messageKey: 'tutorial.completedSuits',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spd-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spd-difficulty"]',
    messageKey: 'tutorial.difficulty',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spd-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Spider Solitaire tutorial configuration. */
const SPD_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'spider',
  steps: SPD_TUTORIAL_STEPS,
};

/** Renders the Spider Solitaire game page with 10 tableau columns and stock. */
export function SpiderPage() {
  const { t: tSpd } = useTranslation('spider');
  return (
    <TutorialProvider config={SPD_TUTORIAL_CONFIG} translateMessage={tSpd}>
      <SpiderPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Spider page, wrapped by TutorialProvider. */
function SpiderPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spider');
  const {
    state,
    loading,
    error,
    hintError,
    selectedSource,
    hint,
    handleDeal,
    handleReset,
    handleResetWithConfig,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleSelectSource,
    handleSelectTarget,
  } = useSpiderGame();
  const { cardHeight, cardOverlap, cardWidth, isMobile, solitaireMinColWidth } = useCardDimensions();

  const isPlayingForKbd = state?.phase === SpiderPhase.PLAYING;

  const currentDifficulty = state?.difficulty ?? 1;

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDeal },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleDeal, handleHint, handleAutoComplete, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <SpiderSkeleton />;

  const isPlaying = state.phase === SpiderPhase.PLAYING;
  const isGameClear = state.phase === SpiderPhase.GAME_CLEAR;
  const isGameOver = state.phase === SpiderPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === 'tableau' &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  const dealsRemaining = Math.floor(state.stockCount / 10);

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.spider.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.spider')} />
      {/* Phase indicator */}
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
        <span className="ml-3">
          {t('score')}: {state.score}
        </span>
        <TutorialButton />
        <span className="ml-3" data-tutorial="spd-completed-suits">
          {t('completed')}: {state.completedSuits}/8
        </span>
      </PhaseIndicator>

      <LandscapeBanner message={t('landscapeBanner')} />

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        {/* Stock row */}
        <div className="flex gap-2 mb-3 items-start">
          {/* Stock */}
          <div className="text-center" data-tutorial="spd-stock-pile">
            <div className="text-game-text-muted text-xs mb-1">
              {t('stock')} ({state.stockCount})
            </div>
            {state.stockCount > 0 ? (
              <AnimatedCardBack width={cardWidth} onClick={isPlaying ? handleDeal : undefined} ariaLabel={t('deal')} />
            ) : (
              <div
                style={{ width: cardWidth, height: cardHeight }}
                className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
              >
                {t('empty')}
              </div>
            )}
            {state.stockCount > 0 && (
              <div className="text-game-text-muted text-xs mt-1">{t('dealsRemaining', { count: dealsRemaining })}</div>
            )}
          </div>
        </div>

        {/* Tableau (10 columns) */}
        <div className="relative">
          <div className="flex gap-0.5 sm:gap-1 mb-3 overflow-x-auto" data-tutorial="spd-tableau">
            {state.tableau.map((col, colIdx) => (
              <div
                key={`col-${colIdx.toString()}`}
                className={isMobile ? 'flex-shrink-0' : 'flex-1'}
                style={isMobile ? { width: solitaireMinColWidth } : undefined}
              >
                <div className="text-game-text-muted text-xs text-center mb-1">{colIdx}</div>
                <div className="relative" style={{ minHeight: cardHeight }}>
                  {col.length === 0 ? (
                    <button
                      type="button"
                      onClick={() => handleSelectTarget({ zone: 'tableau', col: colIdx })}
                      disabled={!isPlaying || loading || !selectedSource}
                      style={{ height: cardHeight }}
                      className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                    >
                      {t('empty')}
                    </button>
                  ) : (
                    col.map((tc, cardIdx) => (
                      <div
                        key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                        className="absolute left-0 right-0"
                        style={{ top: cardIdx * cardOverlap }}
                      >
                        {tc.faceUp && tc.card ? (
                          <button
                            type="button"
                            onClick={() => {
                              if (selectedSource) {
                                // If clicking a different column, treat as move target
                                // If clicking the same column, switch source selection
                                if (selectedSource.col !== colIdx) {
                                  handleSelectTarget({ zone: 'tableau', col: colIdx });
                                } else {
                                  handleSelectSource({ zone: 'tableau', col: colIdx, cardIndex: cardIdx });
                                }
                              } else {
                                handleSelectSource({ zone: 'tableau', col: colIdx, cardIndex: cardIdx });
                              }
                            }}
                            disabled={!isPlaying || loading}
                            aria-label={cardAlt(tc.card)}
                            aria-pressed={isSourceSelected(colIdx, cardIdx)}
                            className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected(colIdx, cardIdx) ? 'ring-2 ring-yellow-400' : ''}`}
                          >
                            <AnimatedCard card={tc.card} width={cardWidth} style={{ width: '100%' }} />
                          </button>
                        ) : (
                          <AnimatedCardBack width={cardWidth} className="w-full" />
                        )}
                      </div>
                    ))
                  )}
                  {col.length > 0 && <div style={{ height: (col.length - 1) * cardOverlap + cardHeight }} />}
                </div>
              </div>
            ))}
          </div>
          {isMobile && <ScrollFadeHint />}
        </div>

        {/* Hint display */}
        {hint && (
          <div className="text-yellow-300 text-sm mb-2">
            {t('hintAvailable')}: {t('tableau')} {hint.fromCol} [{hint.cardIndex}] → {t('tableau')} {hint.toCol}
          </div>
        )}

        {/* Message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Action log */}
        <ActionLogSection
          isEndPhase={isEnded}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      {/* Footer */}
      <GameFooter className={`${gameTheme.spider.footer} px-4 py-2.5`}>
        <ErrorAlert message={error ?? hintError} />
        <div className="flex gap-2 items-center flex-wrap">
          {isPlaying && (
            <div data-tutorial="spd-controls">
              <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
                {t('deal')}
              </button>
              <button type="button" className={btnPrimary} onClick={handleUndo} disabled={loading || !state.canUndo}>
                {t('undo')}
              </button>
              <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                {t('hint')}
              </button>
              <button type="button" className={btnSuccess} onClick={handleAutoComplete} disabled={loading}>
                {t('autoComplete')}
              </button>
              <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                {t('giveup')}
              </button>
            </div>
          )}
          {/* Difficulty selector */}
          <div data-tutorial="spd-difficulty">
            <select
              value={currentDifficulty}
              onChange={(e) => {
                handleResetWithConfig({ difficulty: Number(e.target.value) });
              }}
              className="bg-gray-700 text-white text-sm rounded px-2 py-1"
              aria-label={t('difficulty')}
            >
              <option value={1}>{t('difficulty1')}</option>
              <option value={2}>{t('difficulty2')}</option>
              <option value={4}>{t('difficulty4')}</option>
            </select>
          </div>
          <div data-tutorial="spd-reset-button">
            <button
              type="button"
              className={btnWarning}
              onClick={() =>
                requestConfirm(() => {
                  hideActionLog();
                  return handleReset();
                })
              }
              disabled={loading}
            >
              {tc('button.reset')}
            </button>
          </div>
        </div>
      </GameFooter>
      <WinCelebration show={state.phase === SpiderPhase.GAME_CLEAR} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
