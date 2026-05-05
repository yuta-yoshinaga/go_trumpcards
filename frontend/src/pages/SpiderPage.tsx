import { useCallback, useMemo, useState } from 'react';
import type { SpiderMoveZone, spiderApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DropZone } from '../components/DropZone';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { isGameRoundActive, useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSpiderGame } from '../hooks/useSpiderGame';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SpiderResponse } from '../types/card';
import { SpiderPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSpiderCommand, SPIDER_HELP } from '../utils/cli/commands/spiderCommands';
import { formatSpiderState } from '../utils/cli/formatters/spiderFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

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

/** Renders the Spider Solitaire game page with 10 tableau columns and stock. */
export const SpiderPage = withTutorial(SpiderPageContent, 'spider', SPD_TUTORIAL_STEPS);
/** Inner content of the Spider page, wrapped by TutorialProvider. */
function SpiderPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spider');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
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
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  } = useSpiderGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('spider', state);
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('spider');
  const cliConfig: CliGameConfig<SpiderResponse, Parameters<typeof spiderApi.exec>> = useMemo(
    () => ({
      gameName: 'spider',
      parseCommand: parseSpiderCommand,
      formatResponse: formatSpiderState,
      helpText: SPIDER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === SpiderPhase.PLAYING;

  // Empty-column deal guard: surfaces a shake animation + tooltip instead of failing silently.
  const [emptyDealAttemptKey, setEmptyDealAttemptKey] = useState(0);
  const hasEmptyColumn = useMemo(() => state?.tableau.some((col) => col.length === 0) ?? false, [state?.tableau]);
  const dealBlockedByEmpty = hasEmptyColumn && (state?.stockCount ?? 0) > 0;
  const handleDealGuarded = useCallback(() => {
    if (dealBlockedByEmpty) {
      setEmptyDealAttemptKey((k) => k + 1);
      return;
    }
    handleDeal();
  }, [dealBlockedByEmpty, handleDeal]);

  const dispatchMove = useCallback(
    (source: SpiderMoveZone, target: SpiderMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<SpiderMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  const currentDifficulty = state?.difficulty ?? 1;

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDealGuarded },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleDealGuarded, handleHint, handleAutoComplete, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  useGameRoundGuard(isGameRoundActive(state));

  if (!state) return <GameSkeleton gameKey="spider" layout={{ kind: 'tableau', topRow: 3, tableau: 10 }} />;

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
  const autoCompleteReady = state.stockCount === 0 && isTableauAllFaceUp(state.tableau);

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.spider.bg}`} aria-busy={loading}>
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
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/spider" />
        <span className="ml-3" data-tutorial="spd-completed-suits">
          {t('completed')}: {state.completedSuits}/8
        </span>
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
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
                  <AnimatedCardBack
                    width={cardWidth}
                    onClick={isPlaying ? handleDealGuarded : undefined}
                    ariaLabel={t('deal')}
                    onFlipComplete={() => playSound('cardFlip')}
                  />
                ) : (
                  <div
                    style={{ width: cardWidth, height: cardHeight }}
                    className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                  >
                    {t('empty')}
                  </div>
                )}
                {state.stockCount > 0 && (
                  <div className="text-game-text-muted text-xs mt-1">
                    {t('dealsRemaining', { count: dealsRemaining })}
                  </div>
                )}
              </div>
            </div>

            {/* Tableau (10 columns) */}
            <div className="relative">
              <div className="flex gap-0.5 sm:gap-1 mb-3" data-tutorial="spd-tableau">
                {state.tableau.map((col, colIdx) => {
                  const tableauColZone: SpiderMoveZone = { zone: 'tableau', col: colIdx };
                  return (
                    <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
                      <DropZone
                        isDropTarget={dnd.isDropTarget(tableauColZone)}
                        onDragOver={dnd.handleDragOver(tableauColZone)}
                        onDrop={dnd.handleDrop(tableauColZone)}
                        onDragLeave={dnd.handleDragLeave}
                        className="relative block"
                      >
                        <div className="relative" style={{ minHeight: cardHeight }}>
                          {col.length === 0 ? (
                            <button
                              key={`empty-${colIdx.toString()}-${emptyDealAttemptKey.toString()}`}
                              type="button"
                              onClick={() => handleSelectTarget(tableauColZone)}
                              disabled={!isPlaying || loading || !selectedSource}
                              style={{ height: cardHeight }}
                              data-testid={`spd-empty-col-${colIdx.toString()}`}
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}${emptyDealAttemptKey > 0 ? ' animate-shake border-ds-warning text-ds-warning' : ''}`}
                            >
                              {t('empty')}
                            </button>
                          ) : (
                            col.map((tc, cardIdx) => {
                              const cardZone: SpiderMoveZone = {
                                zone: 'tableau',
                                col: colIdx,
                                cardIndex: cardIdx,
                              };
                              return (
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
                                            handleSelectTarget(tableauColZone);
                                          } else {
                                            handleSelectSource(cardZone);
                                          }
                                        } else {
                                          handleSelectSource(cardZone);
                                        }
                                      }}
                                      disabled={!isPlaying || loading}
                                      aria-label={cardAlt(tc.card)}
                                      aria-pressed={isSourceSelected(colIdx, cardIdx)}
                                      draggable={isPlaying && !loading}
                                      onDragStart={dnd.handleDragStart(cardZone)}
                                      onDragEnd={dnd.handleDragEnd}
                                      className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected(colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                    >
                                      <AnimatedCard
                                        card={tc.card}
                                        width={cardWidth}
                                        draggable={false}
                                        style={{ width: '100%' }}
                                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                                      />
                                    </button>
                                  ) : (
                                    <AnimatedCardBack
                                      width={cardWidth}
                                      style={{ width: '100%' }}
                                      onFlipComplete={() => playSound('cardFlip')}
                                    />
                                  )}
                                </div>
                              );
                            })
                          )}
                          {col.length > 0 && <div style={{ height: (col.length - 1) * cardOverlap + cardHeight }} />}
                        </div>
                      </DropZone>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Hint display */}
            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t('tableau')} {hint.fromCol} [{hint.cardIndex}] → {t('tableau')} {hint.toCol}
              </div>
            )}
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
          <GameFooter className={`${gameTheme.spider.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="spd-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDealGuarded}
                    disabled={loading || isAutoCompleting}
                    title={dealBlockedByEmpty ? t('cannotDealEmptyColExists') : undefined}
                    aria-describedby={dealBlockedByEmpty ? 'spd-deal-blocked-reason' : undefined}
                  >
                    {t('deal')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
                  )}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleHint}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={handleGiveUp}
                    disabled={loading || isAutoCompleting}
                  >
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
                  className="bg-ds-surface-elevated text-ds-text-primary text-sm rounded px-2 py-1"
                  aria-label={t('difficulty')}
                >
                  <option value={1}>{t('difficulty1')}</option>
                  <option value={2}>{t('difficulty2')}</option>
                  <option value={4}>{t('difficulty4')}</option>
                </select>
              </div>
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="spd-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={state.phase === SpiderPhase.GAME_CLEAR} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
