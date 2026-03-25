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
import { PyramidSkeleton } from '../components/skeleton/PyramidSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePyramidGame } from '../hooks/usePyramidGame';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnDanger, btnPrimary, btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import { PyramidPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

/** Pyramid tutorial step definitions. */
const PY_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="py-pyramid"]',
    messageKey: 'tutorial.pyramid',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Pyramid tutorial configuration. */
const PY_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'pyramid',
  steps: PY_TUTORIAL_STEPS,
};

/** Renders the Pyramid Solitaire game page with pyramid, stock/waste, and controls. */
export function PyramidPage() {
  const { t: tPy } = useTranslation('pyramid');
  return (
    <TutorialProvider config={PY_TUTORIAL_CONFIG} translateMessage={tPy}>
      <PyramidPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Pyramid page, wrapped by TutorialProvider. */
function PyramidPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pyramid');
  const {
    state,
    loading,
    error,
    hintError,
    selectedCard,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleSelectCard,
  } = usePyramidGame();
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();

  const isPlayingForKbd = state?.phase === PyramidPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw },
      { key: 'h', action: handleHint },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleDraw, handleHint, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <PyramidSkeleton />;

  const isPlaying = state.phase === PyramidPhase.PLAYING;
  const isGameClear = state.phase === PyramidPhase.GAME_CLEAR;
  const isGameOver = state.phase === PyramidPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSelected = (zone: 'pyramid' | 'waste', row?: number, col?: number) =>
    selectedCard !== null && selectedCard.zone === zone && selectedCard.row === row && selectedCard.col === col;

  // Calculate pyramid layout dimensions
  const maxCols = 7; // bottom row has 7 cards
  const cardGap = 4;
  /** Fraction of card height used for vertical overlap between rows (less on mobile for bigger tap targets) */
  const ROW_OVERLAP_RATIO = isMobile ? 0.3 : 0.35;
  const rowOverlap = cardHeight * ROW_OVERLAP_RATIO;
  const pyramidWidth = maxCols * (cardWidth + cardGap) - cardGap;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.pyramid')} />
      {/* Phase indicator */}
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
        <TutorialButton />
      </PhaseIndicator>

      <LandscapeBanner message={t('landscapeBanner')} />

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        {/* Pyramid */}
        <div data-tutorial="py-pyramid" className="flex flex-col items-center mb-3">
          {state.pyramid.map((row, rowIdx) => {
            const cols = row.length;
            const rowWidth = cols * (cardWidth + cardGap) - cardGap;
            const offsetX = (pyramidWidth - rowWidth) / 2;
            return (
              <div
                key={`row-${rowIdx.toString()}`}
                className="relative"
                style={{
                  height: rowIdx < state.pyramid.length - 1 ? cardHeight - rowOverlap : cardHeight,
                  width: pyramidWidth,
                }}
              >
                {row.map((pc, colIdx) => {
                  const left = offsetX + colIdx * (cardWidth + cardGap);
                  if (pc.removed) {
                    return (
                      <div
                        key={`pc-${rowIdx.toString()}-${colIdx.toString()}`}
                        className="absolute"
                        style={{ left, width: cardWidth, height: cardHeight }}
                      />
                    );
                  }
                  if (!pc.card) return null;
                  const exposed = pc.exposed;
                  return (
                    <div key={`pc-${rowIdx.toString()}-${colIdx.toString()}`} className="absolute" style={{ left }}>
                      <button
                        type="button"
                        onClick={() => {
                          if (!exposed || !pc.card) return;
                          handleSelectCard({ zone: 'pyramid', row: rowIdx, col: colIdx }, pc.card.value);
                        }}
                        disabled={!isPlaying || loading || !exposed}
                        aria-label={cardAlt(pc.card)}
                        aria-pressed={isSelected('pyramid', rowIdx, colIdx)}
                        className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                          isSelected('pyramid', rowIdx, colIdx) ? 'ring-2 ring-yellow-400' : ''
                        } ${!exposed ? 'opacity-60' : ''}`}
                      >
                        <AnimatedCard card={pc.card} width={cardWidth} />
                      </button>
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>

        {/* Stock + Waste */}
        <div className="flex gap-4 justify-center mb-3" data-tutorial="py-stock-waste">
          {/* Stock */}
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">
              {t('stock')} ({state.stockCount})
            </div>
            {state.stockCount > 0 ? (
              <AnimatedCardBack width={cardWidth} onClick={isPlaying ? handleDraw : undefined} ariaLabel={t('draw')} />
            ) : (
              <div
                style={{ width: cardWidth, height: cardHeight }}
                className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
              >
                {t('empty')}
              </div>
            )}
          </div>

          {/* Waste */}
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
            {state.waste.length > 0 ? (
              <button
                type="button"
                onClick={() => {
                  const topCard = state.waste[state.waste.length - 1];
                  handleSelectCard({ zone: 'waste' }, topCard.value);
                }}
                disabled={!isPlaying || loading}
                aria-label={cardAlt(state.waste[state.waste.length - 1])}
                aria-pressed={isSelected('waste')}
                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                  isSelected('waste') ? 'ring-2 ring-yellow-400' : ''
                }`}
              >
                <AnimatedCard card={state.waste[state.waste.length - 1]} width={cardWidth} />
              </button>
            ) : (
              <div
                style={{ width: cardWidth, height: cardHeight }}
                className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
              >
                {t('empty')}
              </div>
            )}
          </div>
        </div>

        {/* Hint display */}
        <div data-tutorial="py-hint-display">
          {hint && (
            <div className="text-yellow-300 text-sm mb-2 text-center">
              {t('hintAvailable')}: {t(`hintType.${hint.type}`)}
            </div>
          )}
        </div>

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
      <GameFooter className="bg-game-bg-casino-dark border-white/20 px-4 py-2.5">
        <ErrorAlert message={error ?? hintError} />
        <div className="flex gap-2 items-center flex-wrap">
          {isPlaying && (
            <div data-tutorial="py-controls">
              <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                {t('draw')}
              </button>
              <button type="button" className={btnPrimary} onClick={handleUndo} disabled={loading || !state.canUndo}>
                {t('undo')}
              </button>
              <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                {t('hint')}
              </button>
              <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                {t('giveup')}
              </button>
            </div>
          )}
          <div data-tutorial="py-reset-button">
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
      <WinCelebration show={state.phase === PyramidPhase.GAME_CLEAR} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
