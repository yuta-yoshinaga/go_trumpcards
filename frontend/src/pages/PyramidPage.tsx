import { useCallback, useMemo } from 'react';
import type { pyramidApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePyramidGame } from '../hooks/usePyramidGame';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PyramidResponse } from '../types/card';
import { PyramidPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PYRAMID_HELP, parsePyramidCommand } from '../utils/cli/commands/pyramidCommands';
import { formatPyramidState } from '../utils/cli/formatters/pyramidFormatter';
import type { CliGameConfig } from '../utils/cli/types';

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

/** Renders the Pyramid Solitaire game page with pyramid, stock/waste, and controls. */
export const PyramidPage = withTutorial(PyramidPageContent, 'pyramid', PY_TUTORIAL_STEPS);
/** Inner content of the Pyramid page, wrapped by TutorialProvider. */
function PyramidPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pyramid');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    selectedCard,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleSelectCard,
  } = usePyramidGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pyramid', state);
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pyramid');
  const cliConfig: CliGameConfig<PyramidResponse, Parameters<typeof pyramidApi.exec>> = useMemo(
    () => ({
      gameName: 'pyramid',
      parseCommand: parsePyramidCommand,
      formatResponse: formatPyramidState,
      helpText: PYRAMID_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

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

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  if (!state)
    return <GameSkeleton gameKey="pyramid" layout={{ kind: 'tiered-rows', rows: [1, 2, 3, 4], stockWaste: true }} />;

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
  // px-4 on the scrollable container = 16px * 2 = 32px total horizontal padding
  const CONTAINER_PADDING = 32;
  const effectiveCardWidth = isMobile
    ? Math.floor((windowWidth - CONTAINER_PADDING - cardGap * (maxCols - 1)) / maxCols)
    : cardWidth;
  const pyramidWidth = maxCols * (effectiveCardWidth + cardGap) - cardGap;

  return (
    <GamePageShell
      title={tc('nav.pyramid')}
      gameThemeBg={gameTheme.pyramid.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/pyramid"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {t('moveCount')}: {state.moveCount}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          {/* Scrollable area */}
          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Pyramid */}
            <div data-tutorial="py-pyramid" className="flex flex-col items-center mb-3">
              {state.pyramid.map((row, rowIdx) => {
                const cols = row.length;
                const rowWidth = cols * (effectiveCardWidth + cardGap) - cardGap;
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
                      const left = offsetX + colIdx * (effectiveCardWidth + cardGap);
                      if (pc.removed) {
                        return (
                          <div
                            key={`pc-${rowIdx.toString()}-${colIdx.toString()}`}
                            className="absolute"
                            style={{ left, width: effectiveCardWidth, height: cardHeight }}
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
                              isSelected('pyramid', rowIdx, colIdx) ? 'ring-2 ring-ds-warning' : ''
                            } ${!exposed ? 'opacity-60' : ''}`}
                          >
                            <AnimatedCard
                              card={pc.card}
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
            <div className="flex gap-4 justify-center mb-3" data-tutorial="py-stock-waste">
              {/* Stock */}
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
                      isSelected('waste') ? 'ring-2 ring-ds-warning' : ''
                    }`}
                  >
                    <AnimatedCard
                      card={state.waste[state.waste.length - 1]}
                      width={effectiveCardWidth}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  </button>
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
            <div data-tutorial="py-hint-display">
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

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Action log */}
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

          {/* Footer */}
          <GameFooter className={`${gameTheme.pyramid.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="py-controls">
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
                dataTutorial="py-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
