import { useCallback, useMemo } from 'react';
import type { napoleonsSquareApi } from '../api/gameApi';
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
import { useNapoleonsSquareGame } from '../hooks/useNapoleonsSquareGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { NapoleonsSquareMoveZone, NapoleonsSquareResponse } from '../types/card';
import { NapoleonsSquarePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { NAPOLEONSSQUARE_HELP, parseNapoleonsSquareCommand } from '../utils/cli/commands/napoleonssquareCommands';
import { formatNapoleonsSquareState } from '../utils/cli/formatters/napoleonssquareFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

// Two foundations per suit; the order is pinned by domain.napoleonsSquareSuitOrder
// so a pile keeps its position across deals.
const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦', '♠', '♣', '♥', '♦'] as const;

const TABLEAU_COLS = 12;
const TOTAL_CARDS = 104;

const NS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ns-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ns-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ns-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ns-tableau"]', messageKey: 'tutorial.runs', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ns-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Napoleon's Square page: 12 tableau columns around 8 foundations. */
export const NapoleonsSquarePage = withTutorial(NapoleonsSquarePageContent, 'napoleonssquare', NS_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { col: idx });
}

function NapoleonsSquarePageContent() {
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
  } = useGamePageSetup('napoleonssquare');
  const game = useNapoleonsSquareGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('napoleonssquare', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('napoleonssquare');
  const cliConfig: CliGameConfig<NapoleonsSquareResponse, Parameters<typeof napoleonsSquareApi.exec>> = useMemo(
    () => ({
      gameName: 'napoleonssquare',
      parseCommand: parseNapoleonsSquareCommand,
      formatResponse: formatNapoleonsSquareState,
      helpText: NAPOLEONSSQUARE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 2;
    const colW = Math.floor((windowWidth - padX - (TABLEAU_COLS - 1) * gapPx) / TABLEAU_COLS);
    const cw = Math.min(Math.max(colW, 24), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === NapoleonsSquarePhase.PLAYING;

  const dispatchMove = useCallback(
    (source: NapoleonsSquareMoveZone, target: NapoleonsSquareMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<NapoleonsSquareMoveZone>({
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
    return <GameSkeleton gameKey="napoleonssquare" layout={{ kind: 'tableau', topRow: 8, tableau: TABLEAU_COLS }} />;
  }

  const isPlaying = state.phase === NapoleonsSquarePhase.PLAYING;
  const isGameClear = state.phase === NapoleonsSquarePhase.GAME_CLEAR;
  const isGameOver = state.phase === NapoleonsSquarePhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // Only needed on game over, so skip the reduce during normal play.
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  // Every foundation starts seeded with its Ace, so "progress" means a pile grew
  // past that — that is when auto-complete is worth pulsing.
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx] ?? [];
    const tableauColZone: NapoleonsSquareMoveZone = { zone: 'tableau', col: colIdx };
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
                aria-label={t('emptyColumnAriaLabel', { col: colIdx + 1 })}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('empty')}
              </button>
            ) : (
              col.map((tc2, cardIdx) => {
                // Any card can head a run, so every card is a potential source —
                // unlike the one-card-at-a-time solitaires where only the top is.
                const cardZone: NapoleonsSquareMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
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

  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: NapoleonsSquareMoveZone = { zone: 'waste' };

  return (
    <GamePageShell
      title={tc('nav.napoleonssquare')}
      gameThemeBg={gameTheme.napoleonssquare.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/napoleonssquare"
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
            <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="ns-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: NapoleonsSquareMoveZone = { zone: 'foundation', col: idx };
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
                              idx: idx + 1,
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
                            aria-label={t('emptyFoundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              idx: idx + 1,
                            })}
                            style={{ width: dims.cw, height: dims.ch }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            A
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="flex gap-2 items-start" data-tutorial="ns-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  <button
                    type="button"
                    onClick={game.handleDraw}
                    disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
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

            <div className="flex gap-0.5 sm:gap-2 items-start" data-tutorial="ns-tableau">
              {Array.from({ length: TABLEAU_COLS }, (_, i) => i).map(renderTableauColumn)}
            </div>

            <div data-tutorial="ns-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toCol)}
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
              <p data-testid="ns-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
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

          <GameFooter className={`${gameTheme.napoleonssquare.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="ns-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleDraw}
                    disabled={loading || isAutoCompleting || state.stockCount === 0}
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
                dataTutorial="ns-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="napoleonssquare-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
