import { useCallback, useMemo } from 'react';
import type { grandfathersClockApi } from '../api/gameApi';
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
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGrandfathersClockGame } from '../hooks/useGrandfathersClockGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GrandfathersClockMoveZone, GrandfathersClockResponse } from '../types/card';
import { GrandfathersClockPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GRANDFATHERSCLOCK_HELP, parseGrandfathersClockCommand } from '../utils/cli/commands/grandfathersclockCommands';
import { formatGrandfathersClockState } from '../utils/cli/formatters/grandfathersclockFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const TABLEAU_COLS = 8;
const FACE_COUNT = 12;

// Face index 0 is one o'clock, 11 is twelve. Everything the player sees stays
// 0-based so the labels, the hint text and the CUI all name the same face.
const CLOCK_HOURS = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12'] as const;

const GC_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="gc-clock"]', messageKey: 'tutorial.clock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="gc-clock"]', messageKey: 'tutorial.targets', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="gc-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="gc-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Grandfather's Clock page: twelve clock faces above eight columns. */
export const GrandfathersClockPage = withTutorial(GrandfathersClockPageContent, 'grandfathersclock', GC_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  return t('frontendHint.tableau', { col: idx });
}

function GrandfathersClockPageContent() {
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
  } = useGamePageSetup('grandfathersclock');
  const game = useGrandfathersClockGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('grandfathersclock', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } =
    useCliMode('grandfathersclock');
  const cliConfig: CliGameConfig<GrandfathersClockResponse, Parameters<typeof grandfathersClockApi.exec>> = useMemo(
    () => ({
      gameName: 'grandfathersclock',
      parseCommand: parseGrandfathersClockCommand,
      formatResponse: formatGrandfathersClockState,
      helpText: GRANDFATHERSCLOCK_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const colW = Math.floor((windowWidth - padX - (TABLEAU_COLS - 1) * gapPx) / TABLEAU_COLS);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === GrandfathersClockPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: GrandfathersClockMoveZone, target: GrandfathersClockMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<GrandfathersClockMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    game.handleReset();
  }, [game, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(game.handleGiveUp),
    [requestGiveUpConfirm, game.handleGiveUp],
  );

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: game.handleHint, label: 'hint' },
      { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: game.handleUndo, label: 'undo' },
    ],
    [game, confirmGiveUpAction],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayingForKbd && !loading });

  if (!state) {
    return (
      <GameSkeleton
        gameKey="grandfathersclock"
        layout={{ kind: 'tableau', topRow: FACE_COUNT, tableau: TABLEAU_COLS }}
      />
    );
  }

  const isPlaying = state.phase === GrandfathersClockPhase.PLAYING;
  const isGameClear = state.phase === GrandfathersClockPhase.GAME_CLEAR;
  const isGameOver = state.phase === GrandfathersClockPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const completedFaces = state.foundation.filter((f) => f.complete).length;
  // Every face is seeded at the deal, so progress means a pile grew past its
  // starter — that is when auto-complete is worth pulsing.
  const autoCompleteReady = state.foundation.some((f) => f.cards.length > 1);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx] ?? [];
    const tableauColZone: GrandfathersClockMoveZone = { zone: 'tableau', col: colIdx };
    return (
      <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
        <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{colIdx}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(tableauColZone)}
          onDragOver={dnd.handleDragOver(tableauColZone)}
          onDrop={dnd.handleDrop(tableauColZone)}
          onDragLeave={dnd.handleDragLeave}
          className="relative block"
        >
          <div className="relative" style={{ minHeight: dims.ch }}>
            {col.length === 0 ? (
              <button
                type="button"
                onClick={() => game.handleSelectTarget(tableauColZone)}
                disabled={!isPlaying || loading || !selectedSource}
                aria-label={t('emptyColumnAriaLabel', { col: colIdx })}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('empty')}
              </button>
            ) : (
              col.map((tc2, cardIdx) => {
                const isTop = cardIdx === col.length - 1;
                // Only the top card moves, so the pressed state belongs to it
                // alone even though selection is keyed on the column.
                const isSelected = isTop && isSourceSelected('tableau', colIdx);
                return (
                  <div
                    key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    {tc2.card ? (
                      <button
                        type="button"
                        onClick={() => {
                          if (selectedSource) {
                            game.handleSelectTarget(tableauColZone);
                          } else if (isTop) {
                            game.handleSelectSource(tableauColZone);
                          }
                        }}
                        disabled={!isPlaying || loading || (!isTop && !selectedSource)}
                        aria-label={cardAlt(tc2.card)}
                        aria-pressed={isSelected}
                        draggable={isPlaying && !loading && isTop}
                        onDragStart={dnd.handleDragStart(tableauColZone)}
                        onDragEnd={dnd.handleDragEnd}
                        className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${isTop ? 'cursor-pointer' : 'cursor-default'} ${isSelected ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(tableauColZone) ? 'opacity-50' : ''}`}
                      >
                        <AnimatedCard
                          card={tc2.card}
                          width={dims.cw}
                          draggable={false}
                          style={{ width: '100%' }}
                          wrapperClassName="block w-full"
                        />
                      </button>
                    ) : null}
                  </div>
                );
              })
            )}
            {col.length > 0 && <div style={{ height: (col.length - 1) * dims.co + dims.ch }} />}
          </div>
        </DropZone>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.grandfathersclock')}
      gameThemeBg={gameTheme.grandfathersclock.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/grandfathersclock"
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
          <span className="text-sm text-ds-text-muted">
            {t('moveCount')}: {state.moveCount}
          </span>
          <span className="text-sm text-ds-text-muted" data-testid="gc-face-progress">
            {t('facesComplete', { count: completedFaces, total: FACE_COUNT })}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="flex flex-wrap justify-center gap-1 sm:gap-2 mb-3" data-tutorial="gc-clock">
              {state.foundation.map((face, idx) => {
                const faceZone: GrandfathersClockMoveZone = { zone: 'foundation', col: idx };
                const top = face.cards.length > 0 ? face.cards[face.cards.length - 1] : null;
                return (
                  <div key={`f-${idx.toString()}`} className="text-center">
                    <div className="text-game-text-muted text-xs mb-1">
                      {t('hourLabel', { hour: CLOCK_HOURS[idx] })}
                    </div>
                    <DropZone
                      isDropTarget={dnd.isDropTarget(faceZone)}
                      onDragOver={dnd.handleDragOver(faceZone)}
                      onDrop={dnd.handleDrop(faceZone)}
                      onDragLeave={dnd.handleDragLeave}
                    >
                      {top ? (
                        <button
                          type="button"
                          onClick={() => game.handleSelectTarget(faceZone)}
                          disabled={!isPlaying || loading || isAutoCompleting || !selectedSource || face.complete}
                          aria-label={t('faceAriaLabel', {
                            idx,
                            hour: CLOCK_HOURS[idx],
                            target: face.targetRank,
                            count: face.cards.length,
                          })}
                          className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${face.complete ? 'opacity-60 cursor-default' : 'cursor-pointer'}`}
                        >
                          <AnimatedCard
                            card={top}
                            width={dims.cw}
                            draggable={false}
                            dealDelay={isAutoCompleting ? idx * 0.05 : 0}
                          />
                        </button>
                      ) : (
                        <div
                          role="img"
                          aria-label={t('emptyFaceAriaLabel', { idx, hour: CLOCK_HOURS[idx] })}
                          style={{ width: dims.cw, height: dims.ch }}
                          className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                        >
                          {t('empty')}
                        </div>
                      )}
                    </DropZone>
                    {/* The target rank is the whole point of the layout, so it
                        is shown rather than left implicit in the hour. */}
                    <div
                      className={`text-xs mt-0.5 ${face.complete ? 'text-ds-success' : 'text-ds-text-muted'}`}
                      aria-hidden="true"
                    >
                      →{face.targetRank}
                      {face.complete ? ' ✔' : ''}
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="flex gap-1 sm:gap-2 items-start" data-tutorial="gc-tableau">
              {Array.from({ length: TABLEAU_COLS }, (_, i) => i).map(renderTableauColumn)}
            </div>

            <div data-tutorial="gc-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
                  {t('hintAvailable')}: {formatHintZone(t, 'tableau', hint.fromCol)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toIdx)}
                </div>
              )}
            </div>
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {isGameOver && (
              <p data-testid="gc-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', { count: completedFaces, total: FACE_COUNT })}
              </p>
            )}

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
          />

          <GameFooter className={`${gameTheme.grandfathersclock.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="gc-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={game.handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
                  )}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={game.handleHint}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={game.handleAutoComplete}
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
                dataTutorial="gc-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="grandfathersclock-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
