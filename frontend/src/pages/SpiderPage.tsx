import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { SpiderMoveZone, spiderApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { AutoCompleteReadyBadge } from '../components/AutoCompleteReadyBadge';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DropZone } from '../components/DropZone';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSpiderGame } from '../hooks/useSpiderGame';
import { spiderWinRate, useSpiderStats } from '../hooks/useSpiderStats';
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
import { hintCheckboxItem } from '../utils/settingsItems';
import { isTableauAllFaceUp, spiderMovableRun } from '../utils/solitaireUtils';

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
  const {
    t,
    tc,
    actionLog,
    showActionLog,
    hideActionLog,
    confirmOpen,
    requestConfirm,
    confirmReset,
    cancelReset,
    giveUpConfirmOpen,
    requestGiveUpConfirm,
    confirmGiveUp,
    cancelGiveUp,
  } = useGamePageSetup('spider');
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
  // Live longest-column length: shrinks the per-card vertical step on mobile so the tallest
  // tableau column fits within 375×667 without scrolling (#1861). Spider columns grow long
  // (initial deal up to 6, +5 stock deals = ~11 minimum, plus accumulated sequences).
  const maxColCards = useMemo(
    () => state?.tableau.reduce((m, col) => (col.length > m ? col.length : m), 0) ?? 0,
    [state?.tableau],
  );
  // Responsive 10-column dimensions matching this page's `px-4` scroll container and `gap-0.5`
  // tableau so a 375 px viewport doesn't crush each card below 28 px (#1648). Stock uses the
  // same dimensions so cards don't visibly pop when the deal animation moves them to the tableau.
  const tableau = useResponsiveTableau(10, { padX: 32, gapPx: 2, maxColCards });
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

  // Movable-run hover preview: highlights the same-suit descending run that would move as a
  // unit when a tableau card is grabbed (#3061). Purely additive — never gates clicks/drags.
  const [hoveredRun, setHoveredRun] = useState<{ col: number; indices: number[] } | null>(null);

  // Empty-column deal guard: surfaces a shake animation + tooltip instead of failing silently.
  const [emptyDealAttemptKey, setEmptyDealAttemptKey] = useState(0);
  const hasEmptyColumn = useMemo(() => state?.tableau.some((col) => col.length === 0) ?? false, [state?.tableau]);
  const dealBlockedByEmpty = hasEmptyColumn && (state?.stockCount ?? 0) > 0;
  const handleDealGuarded = useCallback(() => {
    if (dealBlockedByEmpty) {
      setEmptyDealAttemptKey((k) => k + 1);
      return;
    }
    // Reset on a successful deal so a future empty-column attempt can re-trigger the shake.
    setEmptyDealAttemptKey(0);
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

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const currentDifficulty = state?.difficulty ?? 1;

  // Per-difficulty play statistics persisted in localStorage (#3062).
  const { getStat, recordResult } = useSpiderStats();
  const currentStat = getStat(currentDifficulty);
  // Badge shown on the clear screen when a game beats a stored personal best.
  const [bestUpdate, setBestUpdate] = useState<{ newBestScore: boolean; newFewestMoves: boolean } | null>(null);
  // Guard so a completed game is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  const currentPhase = state?.phase;
  const currentScore = state?.score;
  const currentMoves = state?.moveCount;
  useEffect(() => {
    const ended = currentPhase === SpiderPhase.GAME_CLEAR || currentPhase === SpiderPhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    const won = currentPhase === SpiderPhase.GAME_CLEAR;
    const update = recordResult({
      difficulty: currentDifficulty,
      won,
      score: currentScore ?? 0,
      moves: currentMoves ?? 0,
    });
    setBestUpdate(won ? update : null);
  }, [currentPhase, currentDifficulty, currentScore, currentMoves, recordResult]);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDealGuarded },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
    ],
    [handleDealGuarded, handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  // Announce + sound a fanfare cue whenever a new K→A suit is completed, so the
  // mid-game progress (not just the final win) gives feedback. The count text is
  // an aria-live region, so screen readers hear it too.
  const [suitJustCompleted, setSuitJustCompleted] = useState(false);
  const prevCompletedRef = useRef<number | null>(null);
  useEffect(() => {
    const completed = state?.completedSuits ?? 0;
    const prev = prevCompletedRef.current;
    prevCompletedRef.current = completed;
    if (prev != null && completed > prev) {
      setSuitJustCompleted(true);
      playSound('cardPlace');
      const id = setTimeout(() => setSuitJustCompleted(false), 2500);
      return () => clearTimeout(id);
    }
  }, [state?.completedSuits, playSound]);

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
    <GamePageShell
      title={tc('nav.spider')}
      gameThemeBg={gameTheme.spider.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/spider"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      giveUpConfirmOpen={giveUpConfirmOpen}
      confirmGiveUp={confirmGiveUp}
      cancelGiveUp={cancelGiveUp}
      headerExtra={
        <>
          <span>
            {t('moveCount')}: {state.moveCount}
          </span>
          <span className="ml-3">
            {t('score')}: {state.score}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
      headerEnd={
        <>
          <span className="ml-3" role="status" data-tutorial="spd-completed-suits">
            {t('completed')}: {state.completedSuits}/8
          </span>
          {/* Dedicated live region for the completion flash: keeping it separate
              from the counter means its removal never re-reads the counter text. */}
          <span className="ml-1 text-ds-success font-semibold" role="status">
            {suitJustCompleted ? <span data-testid="spd-suit-complete">{t('suitCompleted')}</span> : null}
          </span>
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
            {/* Stock row */}
            <div className="flex gap-2 mb-3 items-start">
              {/* Stock */}
              <div className="text-center" data-tutorial="spd-stock-pile">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockCount > 0 ? (
                  <AnimatedCardBack
                    width={tableau.cw}
                    onClick={isPlaying ? handleDealGuarded : undefined}
                    ariaLabel={t('deal')}
                  />
                ) : (
                  <div
                    style={{ width: tableau.cw, height: tableau.ch }}
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
                        <div className="relative" style={{ minHeight: tableau.ch }}>
                          {col.length === 0 ? (
                            <button
                              key={`empty-${colIdx.toString()}-${emptyDealAttemptKey.toString()}`}
                              type="button"
                              onClick={() => handleSelectTarget(tableauColZone)}
                              disabled={!isPlaying || loading || !selectedSource}
                              style={{ height: tableau.ch }}
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
                              const inMovableRun = hoveredRun?.col === colIdx && hoveredRun.indices.includes(cardIdx);
                              const ringClass = isSourceSelected(colIdx, cardIdx)
                                ? 'ring-2 ring-ds-warning'
                                : inMovableRun
                                  ? 'ring-2 ring-ds-success'
                                  : '';
                              return (
                                <div
                                  key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                  className="absolute left-0 right-0"
                                  style={{ top: cardIdx * tableau.co }}
                                >
                                  {tc.faceUp && tc.card ? (
                                    <button
                                      type="button"
                                      onMouseEnter={() =>
                                        setHoveredRun({ col: colIdx, indices: spiderMovableRun(col, cardIdx) })
                                      }
                                      onMouseLeave={() => setHoveredRun(null)}
                                      onFocus={() =>
                                        setHoveredRun({ col: colIdx, indices: spiderMovableRun(col, cardIdx) })
                                      }
                                      onBlur={() => setHoveredRun(null)}
                                      data-movable-run={inMovableRun ? 'true' : undefined}
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
                                      className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${ringClass} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                    >
                                      <AnimatedCard
                                        card={tc.card}
                                        width={tableau.cw}
                                        draggable={false}
                                        style={{ width: '100%' }}
                                      />
                                    </button>
                                  ) : (
                                    <AnimatedCardBack width={tableau.cw} style={{ width: '100%' }} />
                                  )}
                                </div>
                              );
                            })
                          )}
                          {col.length > 0 && <div style={{ height: (col.length - 1) * tableau.co + tableau.ch }} />}
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
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Personal-best badge on the clear screen (#3062). */}
            {isGameClear && bestUpdate && (bestUpdate.newBestScore || bestUpdate.newFewestMoves) && (
              <div
                data-testid="spd-best-badge"
                role="status"
                className="text-center text-ds-success font-semibold text-sm mb-2"
              >
                {t('stats.newBest')}
              </div>
            )}

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
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
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
                  <AutoCompleteReadyBadge ready={autoCompleteReady} />
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={confirmGiveUpAction}
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
                    const difficulty = Number(e.target.value);
                    // Mid-game the change discards progress, so confirm first (#2188);
                    // after the game ends there is nothing to lose.
                    if (isEnded) {
                      handleResetWithConfig({ difficulty });
                    } else {
                      requestConfirm(() => handleResetWithConfig({ difficulty }));
                    }
                  }}
                  className="bg-ds-surface-elevated text-ds-text-primary text-sm rounded px-2 py-1 min-h-[44px]"
                  aria-label={t('difficulty')}
                >
                  <option value={1}>{t('difficulty1')}</option>
                  <option value={2}>{t('difficulty2')}</option>
                  <option value={4}>{t('difficulty4')}</option>
                </select>
                {/* Per-difficulty stats: win rate + best score/fewest moves (#3062). */}
                <div data-testid="spd-stats-panel" className="text-game-text-muted text-xs mt-1">
                  {t('stats.winRate', { rate: spiderWinRate(currentStat) })} ({currentStat.wins}/{currentStat.plays})
                  {currentStat.bestScore !== null && <> · {t('stats.best', { score: currentStat.bestScore })}</>}
                  {currentStat.fewestMoves !== null && (
                    <> · {t('stats.fewestMoves', { moves: currentStat.fewestMoves })}</>
                  )}
                </div>
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
    </GamePageShell>
  );
}
