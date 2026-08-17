import { useCallback, useMemo, useState } from 'react';
import type { SpideretteMoveZone, spideretteApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
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
import { useSpideretteGame } from '../hooks/useSpideretteGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SpideretteResponse } from '../types/card';
import { SpiderettePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSpideretteCommand, SPIDERETTE_HELP } from '../utils/cli/commands/spideretteCommands';
import { formatSpideretteState } from '../utils/cli/formatters/spideretteFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

const SPDT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="spdt-stock-pile"]',
    messageKey: 'tutorial.stockPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="spdt-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="spdt-completed-suits"]',
    messageKey: 'tutorial.completedSuits',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="spdt-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="spdt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TOTAL_FOUNDATIONS = 4;
const TABLEAU_COLS = 7;

/** Renders the Spiderette game page (Klondike setup + Spider rules with a single deck). */
export const SpiderettePage = withTutorial(SpiderettePageContent, 'spiderette', SPDT_TUTORIAL_STEPS);

function SpiderettePageContent() {
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
  } = useGamePageSetup('spiderette');
  const game = useSpideretteGame();
  const {
    state,
    loading,
    error,
    retry,
    hintError,
    selectedSource,
    hint,
    handleDeal,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  } = game;
  const runCmd = game.exec;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('spiderette', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('spiderette');
  const cliConfig: CliGameConfig<SpideretteResponse, Parameters<typeof spideretteApi.exec>> = useMemo(
    () => ({
      gameName: 'spiderette',
      parseCommand: parseSpideretteCommand,
      formatResponse: formatSpideretteState,
      helpText: SPIDERETTE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(runCmd, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const tableau = useResponsiveTableau(TABLEAU_COLS, { padX: 32, gapPx: 2 });
  const isPlayingForKbd = state?.phase === SpiderettePhase.PLAYING;

  const [emptyDealAttemptKey, setEmptyDealAttemptKey] = useState(0);
  const hasEmptyColumn = useMemo(() => state?.tableau.some((col) => col.length === 0) ?? false, [state?.tableau]);
  const dealBlockedByEmpty = hasEmptyColumn && (state?.stockCount ?? 0) > 0;
  const handleDealGuarded = useCallback(() => {
    if (dealBlockedByEmpty) {
      setEmptyDealAttemptKey((k) => k + 1);
      return;
    }
    setEmptyDealAttemptKey(0);
    handleDeal();
  }, [dealBlockedByEmpty, handleDeal]);

  const dispatchMove = useCallback(
    (source: SpideretteMoveZone, target: SpideretteMoveZone) => {
      void runCmd('move', source, target);
    },
    [runCmd],
  );
  const dnd = useSolitaireDragDrop<SpideretteMoveZone>({
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

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDealGuarded, label: 'deal' },
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleDealGuarded, handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayingForKbd && !loading });

  if (!state)
    return <GameSkeleton gameKey="spiderette" layout={{ kind: 'tableau', topRow: 1, tableau: TABLEAU_COLS }} />;

  const isPlaying = state.phase === SpiderettePhase.PLAYING;
  const isGameClear = state.phase === SpiderettePhase.GAME_CLEAR;
  const isGameOver = state.phase === SpiderettePhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === 'tableau' &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // Visual hint highlighting: ring the suggested source card and target column.
  const isHintSource = (col: number, cardIndex: number) =>
    hint != null && hint.fromCol === col && hint.cardIndex === cardIndex;
  const isHintTargetCol = (col: number) => hint != null && hint.toCol === col;

  // A partial final deal (1–6 cards) still counts as one deal opportunity,
  // so round up rather than down (#1676 review).
  const dealsRemaining = Math.ceil(state.stockCount / TABLEAU_COLS);
  const autoCompleteReady = state.stockCount === 0 && isTableauAllFaceUp(state.tableau);

  return (
    <GamePageShell
      title={tc('nav.spiderette')}
      gameThemeBg={gameTheme.spiderette.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/spiderette"
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
          {/* **数字が動く理由を推測させない** (#5593)。手数の減点とスート完成の
              加点は、どちらの画面にも書かれていなかった。数字はサーバから。 */}
          <span
            className="ml-3"
            data-testid="spiderette-score"
            title={t('scoreRule', {
              start: state.scoring.start,
              penalty: state.scoring.movePenalty,
              bonus: state.scoring.suitBonus,
            })}
          >
            {t('score')}: {state.score}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
      headerEnd={
        <span className="ml-3" data-tutorial="spdt-completed-suits">
          {t('completed')}: {state.completedSuits}/{TOTAL_FOUNDATIONS}
        </span>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            <div className="flex gap-2 mb-3 items-start">
              <div className="text-center" data-tutorial="spdt-stock-pile">
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

            <div className="relative">
              <div className="flex gap-0.5 sm:gap-1 mb-3" data-tutorial="spdt-tableau">
                {state.tableau.map((col, colIdx) => {
                  const tableauColZone: SpideretteMoveZone = { zone: 'tableau', col: colIdx };
                  return (
                    <div
                      key={`col-${colIdx.toString()}`}
                      data-testid={`spdt-col-${colIdx.toString()}`}
                      className={
                        isHintTargetCol(colIdx)
                          ? 'flex-1 min-w-0 rounded ring-1 ring-ds-success animate-pulse'
                          : 'flex-1 min-w-0'
                      }
                    >
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
                              data-testid={`spdt-empty-col-${colIdx.toString()}`}
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}${emptyDealAttemptKey > 0 ? ' animate-shake border-ds-warning text-ds-warning' : ''}`}
                            >
                              {t('empty')}
                            </button>
                          ) : (
                            col.map((tc, cardIdx) => {
                              const cardZone: SpideretteMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                              return (
                                <div
                                  key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                  className="absolute left-0 right-0"
                                  style={{ top: cardIdx * tableau.co }}
                                >
                                  {tc.faceUp && tc.card ? (
                                    <button
                                      type="button"
                                      onClick={() => {
                                        if (selectedSource) {
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
                                      data-testid={`spdt-card-${colIdx.toString()}-${cardIdx.toString()}`}
                                      className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected(colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : isHintSource(colIdx, cardIdx) ? 'ring-2 ring-ds-info animate-pulse' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
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

            {hint && (
              <div className="text-ds-warning text-sm mb-2" role="status" aria-live="polite">
                {t('hintAvailable')}: {t('tableau')} {hint.fromCol} [{hint.cardIndex}] → {t('tableau')} {hint.toCol}
              </div>
            )}
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            {/* Announce the empty-column deal guard for screen readers; the key
            forces a remount on each blocked attempt so it re-announces, mirroring
            the visual shake animation. */}
            {emptyDealAttemptKey > 0 && (
              <div
                key={`deal-warn-${emptyDealAttemptKey.toString()}`}
                className="sr-only"
                role="status"
                aria-live="assertive"
              >
                {t('cannotDealEmptyColExists')}
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

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.spiderette.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="spdt-controls">
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
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="spdt-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="spiderette-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
