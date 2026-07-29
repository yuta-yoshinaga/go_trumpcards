import { useCallback, useMemo } from 'react';
import type { duchessApi } from '../api/gameApi';
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
import { useDuchessGame } from '../hooks/useDuchessGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DuchessMoveZone, DuchessResponse } from '../types/card';
import { DuchessPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { DUCHESS_HELP, parseDuchessCommand } from '../utils/cli/commands/duchessCommands';
import { formatDuchessState } from '../utils/cli/formatters/duchessFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;
const TABLEAU_COLS = 4;
const RESERVE_FANS = 4;
const TOTAL_CARDS = 52;

const DU_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="du-reserve"]', messageKey: 'tutorial.reserve', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="du-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="du-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="du-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="du-tableau"]', messageKey: 'tutorial.reserveOnly', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="du-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Duchess page: four reserve fans, four columns, four foundations, and the stock. */
export const DuchessPage = withTutorial(DuchessPageContent, 'duchess', DU_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'reserve') return t('frontendHint.reserve', { idx });
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { col: idx });
}

function DuchessPageContent() {
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
  } = useGamePageSetup('duchess');
  const game = useDuchessGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('duchess', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('duchess');
  const cliConfig: CliGameConfig<DuchessResponse, Parameters<typeof duchessApi.exec>> = useMemo(
    () => ({
      gameName: 'duchess',
      parseCommand: parseDuchessCommand,
      formatResponse: formatDuchessState,
      helpText: DUCHESS_HELP,
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

  const isPlayingForKbd = state?.phase === DuchessPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: DuchessMoveZone, target: DuchessMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<DuchessMoveZone>({
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
      { key: 'd', action: game.handleDraw, label: 'draw' },
      { key: 'h', action: game.handleHint, label: 'hint' },
      { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: game.handleUndo, label: 'undo' },
    ],
    [game, confirmGiveUpAction],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayingForKbd && !loading });

  if (!state) {
    return <GameSkeleton gameKey="duchess" layout={{ kind: 'tableau', topRow: 4, tableau: TABLEAU_COLS }} />;
  }

  const isPlaying = state.phase === DuchessPhase.PLAYING;
  const isGameClear = state.phase === DuchessPhase.GAME_CLEAR;
  const isGameOver = state.phase === DuchessPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);
  // While any reserve card remains, empty columns are the reserve's exit only.
  const reserveRemaining = state.reserve.reduce((sum, fan) => sum + fan.length, 0);

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx] ?? [];
    const tableauColZone: DuchessMoveZone = { zone: 'tableau', col: colIdx };
    const reserveOnly = col.length === 0 && reserveRemaining > 0;
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
                aria-label={
                  reserveOnly
                    ? t('reserveOnlyColumnAriaLabel', { col: colIdx })
                    : t('emptyColumnAriaLabel', { col: colIdx })
                }
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {reserveOnly ? t('reserveOnly') : ''}
              </button>
            ) : (
              col.map((tc2, cardIdx) => {
                // Any card can head a run, so every card is a potential source.
                const cardZone: DuchessMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
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
                          } else {
                            game.handleSelectSource(cardZone);
                          }
                        }}
                        disabled={!isPlaying || loading}
                        aria-label={cardAlt(tc2.card)}
                        aria-pressed={isSelected}
                        draggable={isPlaying && !loading}
                        onDragStart={dnd.handleDragStart(cardZone)}
                        onDragEnd={dnd.handleDragEnd}
                        className={`p-0 border-0 bg-transparent w-full rounded cursor-pointer ${focusRingWhite} ${isSelected ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
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

  const renderReserveFan = (fanIdx: number) => {
    const fan = state.reserve[fanIdx] ?? [];
    const top = fan.length > 0 ? fan[fan.length - 1] : null;
    const fanZone: DuchessMoveZone = { zone: 'reserve', col: fanIdx };
    if (!top) {
      return (
        <div key={`r-${fanIdx.toString()}`} className="text-center">
          <div className="text-game-text-muted text-xs mb-1">{fanIdx}</div>
          <div
            role="img"
            aria-label={t('emptyFanAriaLabel', { idx: fanIdx })}
            style={{ width: dims.cw, height: dims.ch }}
            className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
          >
            {t('empty')}
          </div>
        </div>
      );
    }
    // Before the base rank is set, a fan top is not a move source — clicking it
    // sets the rank instead, so the two states get different handlers and labels.
    const selecting = state.awaitingBaseRank;
    return (
      <div key={`r-${fanIdx.toString()}`} className="text-center">
        <div className="text-game-text-muted text-xs mb-1">
          {fanIdx} ({fan.length})
        </div>
        <button
          type="button"
          onClick={() => (selecting ? game.handleChooseBase(fanIdx) : game.handleSelectSource(fanZone))}
          disabled={!isPlaying || loading}
          aria-label={selecting ? t('chooseBaseAriaLabel', { idx: fanIdx }) : cardAlt(top)}
          aria-pressed={selecting ? undefined : isSourceSelected('reserve', fanIdx, undefined)}
          draggable={isPlaying && !loading && !selecting}
          onDragStart={dnd.handleDragStart(fanZone)}
          onDragEnd={dnd.handleDragEnd}
          className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('reserve', fanIdx, undefined) ? 'ring-2 ring-ds-warning' : ''} ${selecting ? 'ring-2 ring-ds-success animate-pulse' : ''}`}
        >
          <AnimatedCard card={top} width={dims.cw} draggable={false} />
        </button>
      </div>
    );
  };

  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: DuchessMoveZone = { zone: 'waste' };

  return (
    <GamePageShell
      title={tc('nav.duchess')}
      gameThemeBg={gameTheme.duchess.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/duchess"
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
            {t('baseRankLabel')}: {state.awaitingBaseRank ? t('baseRankUnset') : state.baseRank}
          </span>
          <span className="text-sm text-ds-text-muted">
            {t('moveCount')}: {state.moveCount}
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
            {state.awaitingBaseRank && isPlaying && (
              <p className="text-ds-warning text-sm text-center mb-2">{t('chooseBaseBanner')}</p>
            )}

            <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="du-reserve">
                {Array.from({ length: RESERVE_FANS }, (_, i) => i).map(renderReserveFan)}
              </div>

              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="du-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: DuchessMoveZone = { zone: 'foundation', col: idx };
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
                      <div className="text-game-text-muted text-xs mb-1">{FOUNDATION_SUITS[idx]}</div>
                      <DropZone
                        isDropTarget={dnd.isDropTarget(foundationZone)}
                        onDragOver={dnd.handleDragOver(foundationZone)}
                        onDrop={dnd.handleDrop(foundationZone)}
                        onDragLeave={dnd.handleDragLeave}
                      >
                        {pile.length > 0 ? (
                          <button
                            type="button"
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                            aria-label={t('foundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              idx,
                              count: pile.length,
                            })}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={dims.cw}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.1 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx], idx })}
                            style={{ width: dims.cw, height: dims.ch }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            {state.awaitingBaseRank ? '?' : state.baseRank}
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="flex gap-2 items-start" data-tutorial="du-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  <button
                    type="button"
                    onClick={game.handleDraw}
                    disabled={
                      !isPlaying || loading || isAutoCompleting || state.stockCount === 0 || state.awaitingBaseRank
                    }
                    aria-label={t('stockAriaLabel', { count: state.stockCount })}
                    style={{ width: dims.cw, height: dims.ch }}
                    className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    {state.stockCount}
                  </button>
                </div>
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                  {wasteTop ? (
                    <button
                      type="button"
                      onClick={() => game.handleSelectSource(wasteZone)}
                      disabled={!isPlaying || loading}
                      aria-label={cardAlt(wasteTop)}
                      aria-pressed={isSourceSelected('waste', undefined, undefined)}
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(wasteZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined, undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <AnimatedCard card={wasteTop} width={dims.cw} draggable={false} />
                    </button>
                  ) : (
                    <div
                      role="img"
                      aria-label={t('emptyWasteAriaLabel')}
                      style={{ width: dims.cw, height: dims.ch }}
                      className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="flex gap-1 sm:gap-2 items-start" data-tutorial="du-tableau">
              {Array.from({ length: TABLEAU_COLS }, (_, i) => i).map(renderTableauColumn)}
            </div>

            <div data-tutorial="du-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromIdx)} →{' '}
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
              <p data-testid="du-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', {
                  count: foundationCount,
                  percent: Math.round((foundationCount / TOTAL_CARDS) * 100),
                })}
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

          <GameFooter className={`${gameTheme.duchess.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="du-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleDraw}
                    disabled={loading || isAutoCompleting || state.stockCount === 0 || state.awaitingBaseRank}
                  >
                    {t('draw')}
                  </button>
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
                dataTutorial="du-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="duchess-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
