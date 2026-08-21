import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { MrsMopMoveZone, mrsMopApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
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
import { useMrsMopGame } from '../hooks/useMrsMopGame';
import { mrsMopWinRate, useMrsMopStats } from '../hooks/useMrsMopStats';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MrsMopResponse } from '../types/card';
import { MrsMopPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MRSMOP_HELP, parseMrsMopCommand } from '../utils/cli/commands/mrsMopCommands';
import { formatMrsMopState } from '../utils/cli/formatters/mrsMopFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
// **同スート降順の並び判定は Spider と同一の規則。**共有ヘルパをそのまま使う
// (クローンが `mrsMopMovableRun` という存在しない名前に書き換えていた)。
import { isTableauAllFaceUp, spiderMovableRun } from '../utils/solitaireUtils';

/** MrsMop Solitaire tutorial step definitions. */
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

/** Renders the MrsMop Solitaire game page with 10 tableau columns and stock. */
export const MrsMopPage = withTutorial(MrsMopPageContent, 'mrsmop', SPD_TUTORIAL_STEPS);
/** Inner content of the MrsMop page, wrapped by TutorialProvider. */
function MrsMopPageContent() {
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
  } = useGamePageSetup('mrsmop');
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
  } = useMrsMopGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('mrsmop', state);
  // Live longest-column length: shrinks the per-card vertical step on mobile so the tallest
  // tableau column fits within 375×667 without scrolling (#1861). MrsMop columns grow long
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
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('mrsmop');
  const cliConfig: CliGameConfig<MrsMopResponse, Parameters<typeof mrsMopApi.exec>> = useMemo(
    () => ({
      gameName: 'mrsmop',
      parseCommand: parseMrsMopCommand,
      formatResponse: formatMrsMopState,
      helpText: MRSMOP_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === MrsMopPhase.PLAYING;

  // Movable-run hover preview: highlights the same-suit descending run that would move as a
  // unit when a tableau card is grabbed (#3061). Purely additive — never gates clicks/drags.
  const [hoveredRun, setHoveredRun] = useState<{ col: number; indices: number[] } | null>(null);

  // **配る操作が無いので、空列ガードも要らない。**クローン元 (Spider) は
  // 「空列があると配れない」ので、失敗を黙って返さないよう shake で知らせて
  // いた。Mrs. Mop に山札は無く、空列は単に「どの札でも置ける枠」でしかない。

  const dispatchMove = useCallback(
    (source: MrsMopMoveZone, target: MrsMopMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<MrsMopMoveZone>({
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
  const { getStat, recordResult } = useMrsMopStats();
  const currentStat = getStat(currentDifficulty);
  // Badge shown on the clear screen when a game beats a stored personal best.
  const [bestUpdate, setBestUpdate] = useState<{ newBestScore: boolean; newFewestMoves: boolean } | null>(null);
  // Guard so a completed game is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  const currentPhase = state?.phase;
  const currentScore = state?.score;
  const currentMoves = state?.moveCount;
  useEffect(() => {
    const ended = currentPhase === MrsMopPhase.GAME_CLEAR || currentPhase === MrsMopPhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    const won = currentPhase === MrsMopPhase.GAME_CLEAR;
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
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
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

  if (!state) return <GameSkeleton gameKey="mrsmop" layout={{ kind: 'tableau', topRow: 3, tableau: 10 }} />;

  const isPlaying = state.phase === MrsMopPhase.PLAYING;
  const isGameClear = state.phase === MrsMopPhase.GAME_CLEAR;
  const isGameOver = state.phase === MrsMopPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === 'tableau' &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // **常に配り終わっている。**山札が無いので、オートコンプリートを止めるのは
  // 伏せ札の有無だけ ── そしてそれも配り直後から常に無い。
  const autoCompleteReady = isTableauAllFaceUp(state.tableau);

  return (
    <GamePageShell
      title={tc('nav.mrsmop')}
      gameThemeBg={gameTheme.mrsmop.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/mrsMop"
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
            {/* **山札の行は無い。**Mrs. Mop は 104 枚を配り切るので、クローン元
                (Spider) の山札・配る操作・残り配り回数はどれも存在しない。
                常に空の置き場を描くと「まだ配れる」と読めてしまう。 */}

            {/* Tableau (10 columns) */}
            <div className="relative">
              <div className="flex gap-0.5 sm:gap-1 mb-3" data-tutorial="spd-tableau">
                {state.tableau.map((col, colIdx) => {
                  const tableauColZone: MrsMopMoveZone = { zone: 'tableau', col: colIdx };
                  // タップ選択でのプレビュー。onMouseEnter も onFocus も発火しない
                  // タッチ端末では、選んだ札と一緒に動く連番がどこまでかを事前に
                  // 確かめる手段が無かった (#4780, Yukon の #3152 と同じ)。
                  // 連番は列ごとに1度だけ求める (ホバー側の setHoveredRun と同じ形)。
                  const selectedRun =
                    selectedSource?.zone === 'tableau' &&
                    selectedSource.col === colIdx &&
                    selectedSource.cardIndex !== undefined
                      ? spiderMovableRun(col, selectedSource.cardIndex)
                      : null;
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
                              key={`empty-${colIdx.toString()}`}
                              type="button"
                              onClick={() => handleSelectTarget(tableauColZone)}
                              disabled={!isPlaying || loading || !selectedSource}
                              style={{ height: tableau.ch }}
                              data-testid={`spd-empty-col-${colIdx.toString()}`}
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                            >
                              {t('empty')}
                            </button>
                          ) : (
                            col.map((tc, cardIdx) => {
                              const cardZone: MrsMopMoveZone = {
                                zone: 'tableau',
                                col: colIdx,
                                cardIndex: cardIdx,
                              };
                              const inMovableRun = hoveredRun?.col === colIdx && hoveredRun.indices.includes(cardIdx);
                              const inSelectedRun = selectedRun?.includes(cardIdx) ?? false;
                              const ringClass = isSourceSelected(colIdx, cardIdx)
                                ? 'ring-2 ring-ds-warning'
                                : inMovableRun
                                  ? 'ring-2 ring-ds-success'
                                  : inSelectedRun
                                    ? 'ring-2 ring-ds-info'
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
                                      data-selected-block={inSelectedRun || undefined}
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
            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="mrsMop-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t('tableau')} {hint.fromCol} [{hint.cardIndex}] → {t('tableau')} {hint.toCol}
                </div>
              )}
            </div>
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
          <GameFooter className={`${gameTheme.mrsmop.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="spd-controls">
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
                  className="bg-ds-surface-elevated text-ds-text-primary text-sm rounded px-2 py-1"
                  aria-label={t('difficulty')}
                >
                  <option value={1}>{t('difficulty1')}</option>
                  <option value={2}>{t('difficulty2')}</option>
                  <option value={4}>{t('difficulty4')}</option>
                </select>
                {/* Per-difficulty stats: win rate + best score/fewest moves (#3062). */}
                <div data-testid="spd-stats-panel" className="text-game-text-muted text-xs mt-1">
                  {t('stats.winRate', { rate: mrsMopWinRate(currentStat) })} ({currentStat.wins}/{currentStat.plays})
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="mrsMop-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
