import { useCallback, useMemo } from 'react';
import type { tripeaksApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { TriPeaksSkeleton } from '../components/skeleton/TriPeaksSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { useTriPeaksGame } from '../hooks/useTriPeaksGame';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TriPeaksResponse } from '../types/card';
import { TriPeaksPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTripeaksCommand, TRIPEAKS_HELP } from '../utils/cli/commands/tripeaksCommands';
import { formatTripeaksState } from '../utils/cli/formatters/tripeaksFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Valid column positions per row in the TriPeaks tableau. */
const VALID_COLS: readonly number[][] = [
  [0, 3, 6],
  [0, 1, 3, 4, 6, 7],
  [0, 1, 2, 3, 4, 5, 6, 7, 8],
  [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
];

/** TriPeaks tutorial step definitions. */
const TP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tp-peaks"]',
    messageKey: 'tutorial.peaks',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the TriPeaks Solitaire game page with three peaks, stock/waste, and controls. */
export function TriPeaksPage() {
  return (
    <TutorialWrapper gameName="tripeaks" steps={TP_TUTORIAL_STEPS}>
      <TriPeaksPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the TriPeaks page, wrapped by TutorialProvider. */
function TriPeaksPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tripeaks');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleSelectCard,
  } = useTriPeaksGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tripeaks', state);
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tripeaks');
  const cliConfig: CliGameConfig<TriPeaksResponse, Parameters<typeof tripeaksApi.exec>> = useMemo(
    () => ({
      gameName: 'tripeaks',
      parseCommand: parseTripeaksCommand,
      formatResponse: formatTripeaksState,
      helpText: TRIPEAKS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === TriPeaksPhase.PLAYING;

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

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  // Issue #1609: warn before tab close / reload while a round is in progress.

  useGameRoundGuard(!!state && !state.gameEndFlag);

  if (!state) return <TriPeaksSkeleton />;

  const isPlaying = state.phase === TriPeaksPhase.PLAYING;
  const isGameClear = state.phase === TriPeaksPhase.GAME_CLEAR;
  const isGameOver = state.phase === TriPeaksPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  // Calculate layout dimensions
  const maxCols = 10;
  const cardGap = 4;
  const ROW_OVERLAP_RATIO = isMobile ? 0.3 : 0.35;
  const rowOverlap = cardHeight * ROW_OVERLAP_RATIO;
  // px-4 on the scrollable container = 16px * 2 = 32px total horizontal padding
  const CONTAINER_PADDING = 32;
  const effectiveCardWidth = isMobile
    ? Math.floor((windowWidth - CONTAINER_PADDING - cardGap * (maxCols - 1)) / maxCols)
    : cardWidth;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.tripeaks.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.tripeaks')} />
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/tripeaks" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Tableau */}
            <div data-tutorial="tp-peaks" className="flex flex-col items-center mb-3">
              {VALID_COLS.map((cols, rowIdx) => {
                const rowWidth = maxCols * (effectiveCardWidth + cardGap) - cardGap;
                return (
                  <div
                    key={`row-${rowIdx.toString()}`}
                    className="relative"
                    style={{
                      height: rowIdx < VALID_COLS.length - 1 ? cardHeight - rowOverlap : cardHeight,
                      width: rowWidth,
                    }}
                  >
                    {cols.map((colIdx) => {
                      const tc2 = state.layout[rowIdx]?.[colIdx];
                      const left = colIdx * (effectiveCardWidth + cardGap);
                      if (!tc2 || tc2.removed) {
                        return (
                          <div
                            key={`tc-${rowIdx.toString()}-${colIdx.toString()}`}
                            className="absolute"
                            style={{ left, width: effectiveCardWidth, height: cardHeight }}
                          />
                        );
                      }
                      if (!tc2.card) return null;
                      const exposed = tc2.exposed;
                      const isHinted = hint?.type === 'remove' && hint.row === rowIdx && hint.col === colIdx;
                      return (
                        <div key={`tc-${rowIdx.toString()}-${colIdx.toString()}`} className="absolute" style={{ left }}>
                          <button
                            type="button"
                            onClick={() => {
                              if (!exposed || !tc2.card) return;
                              handleSelectCard(rowIdx, colIdx);
                            }}
                            disabled={!isPlaying || loading || !exposed}
                            aria-label={cardAlt(tc2.card)}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                              isHinted ? 'ring-2 ring-ds-warning' : ''
                            } ${!exposed ? 'opacity-60' : ''}`}
                          >
                            <AnimatedCard
                              card={tc2.card}
                              width={effectiveCardWidth}
                              onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                            />
                          </button>
                        </div>
                      );
                    })}
                  </div>
                );
              })}
            </div>

            {/* Stock + Waste */}
            <div className="flex gap-4 justify-center mb-3" data-tutorial="tp-stock-waste">
              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockCount > 0 ? (
                  <AnimatedCardBack
                    width={effectiveCardWidth}
                    onClick={isPlaying ? handleDraw : undefined}
                    ariaLabel={t('draw')}
                    onFlipComplete={() => playSound('cardFlip')}
                  />
                ) : (
                  <div
                    style={{ width: effectiveCardWidth, height: cardHeight }}
                    className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>

              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                {state.waste.length > 0 ? (
                  <AnimatedCard
                    card={state.waste[state.waste.length - 1]}
                    width={effectiveCardWidth}
                    onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                  />
                ) : (
                  <div
                    style={{ width: effectiveCardWidth, height: cardHeight }}
                    className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>
            </div>

            {/* Hint display */}
            <div data-tutorial="tp-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 text-center">
                  {t('hintAvailable')}: {t(`hintType.${hint.type}`)}
                </div>
              )}
            </div>
            {frontendHintEnabled && frontendHint && (
              <div className="flex justify-center">
                <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.tripeaks.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="tp-controls">
                  <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                    {t('draw')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading}
                    />
                  )}
                  <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                    {t('giveup')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tp-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={state.phase === TriPeaksPhase.GAME_CLEAR} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
