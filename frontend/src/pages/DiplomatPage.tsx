import { useCallback, useMemo } from 'react';
import type { diplomatApi } from '../api/gameApi';
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
import { useDiplomatGame } from '../hooks/useDiplomatGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DiplomatMoveZone, DiplomatResponse } from '../types/card';
import { DiplomatPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { DIPLOMAT_HELP, parseDiplomatCommand } from '../utils/cli/commands/diplomatCommands';
import { formatDiplomatState } from '../utils/cli/formatters/diplomatFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦', '♠', '♣', '♥', '♦'] as const;
const TABLEAU_PILES = 8;
const TOTAL_CARDS = 104;

const CG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cg-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cg-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cg-tableau"]', messageKey: 'tutorial.emptyPile', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cg-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cg-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Diplomat page: eight single-card piles flanking eight foundations. */
export const DiplomatPage = withTutorial(DiplomatPageContent, 'diplomat', CG_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { pile: idx });
}

function DiplomatPageContent() {
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
  } = useGamePageSetup('diplomat');
  const game = useDiplomatGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('diplomat', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('diplomat');
  const cliConfig: CliGameConfig<DiplomatResponse, Parameters<typeof diplomatApi.exec>> = useMemo(
    () => ({
      gameName: 'diplomat',
      parseCommand: parseDiplomatCommand,
      formatResponse: formatDiplomatState,
      helpText: DIPLOMAT_HELP,
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
    const colW = Math.floor((windowWidth - padX - (TABLEAU_PILES - 1) * gapPx) / TABLEAU_PILES);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === DiplomatPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: DiplomatMoveZone, target: DiplomatMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<DiplomatMoveZone>({
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
    return <GameSkeleton gameKey="diplomat" layout={{ kind: 'tableau', topRow: 8, tableau: TABLEAU_PILES }} />;
  }

  const isPlaying = state.phase === DiplomatPhase.PLAYING;
  const isGameClear = state.phase === DiplomatPhase.GAME_CLEAR;
  const isGameOver = state.phase === DiplomatPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: DiplomatMoveZone = { zone: 'waste' };

  const renderPile = (pileIdx: number) => {
    const cards = state.tableau[pileIdx] ?? [];
    const pileZone: DiplomatMoveZone = { zone: 'tableau', col: pileIdx };
    return (
      <div key={`pile-${pileIdx.toString()}`} className="flex-1 min-w-0">
        <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{pileIdx}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(pileZone)}
          onDragOver={dnd.handleDragOver(pileZone)}
          onDrop={dnd.handleDrop(pileZone)}
          onDragLeave={dnd.handleDragLeave}
          className="relative block"
        >
          <div className="relative" style={{ minHeight: dims.ch }}>
            {cards.length === 0 ? (
              <button
                type="button"
                onClick={() => game.handleSelectTarget(pileZone)}
                // **空き列は別の列か捨て札から埋める。**Congress と違って
                // タブローも移動元にできる（Diplomat の主要な逃げ道）。
                disabled={!isPlaying || loading || !selectedSource}
                aria-label={t('emptyPileAriaLabel', { pile: pileIdx })}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('emptyColumn')}
              </button>
            ) : (
              cards.map((card, cardIdx) => {
                // Only the top card is playable; the ones under it are context.
                const isTop = cardIdx === cards.length - 1;
                return (
                  <div
                    key={`c-${pileIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    <button
                      type="button"
                      onClick={() => {
                        if (selectedSource) {
                          game.handleSelectTarget(pileZone);
                        } else if (isTop) {
                          game.handleSelectSource(pileZone);
                        }
                      }}
                      disabled={!isPlaying || loading || (!isTop && !selectedSource)}
                      aria-label={cardAlt(card)}
                      aria-pressed={isTop ? isSourceSelected('tableau', pileIdx) : undefined}
                      draggable={isTop && isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(pileZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent w-full rounded cursor-pointer ${focusRingWhite} ${isTop && isSourceSelected('tableau', pileIdx) ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <AnimatedCard
                        card={card}
                        width={dims.cw}
                        draggable={false}
                        style={{ width: '100%' }}
                        wrapperClassName="block w-full"
                      />
                    </button>
                  </div>
                );
              })
            )}
            {cards.length > 0 && <div style={{ height: (cards.length - 1) * dims.co + dims.ch }} />}
          </div>
        </DropZone>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.diplomat')}
      gameThemeBg={gameTheme.diplomat.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/diplomat"
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
              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="cg-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: DiplomatMoveZone = { zone: 'foundation', col: idx };
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
                            A
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="flex gap-2 items-start" data-tutorial="cg-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  {/* The stock only ever draws. An empty column is filled from
                      another column or the waste, never straight from the deck,
                      so the stock is not a move source in this game. */}
                  <button
                    type="button"
                    onClick={() => game.handleDraw()}
                    disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
                    aria-label={
                      state.stockCount === 0
                        ? t('emptyStockAriaLabel')
                        : t('stockAriaLabel', { count: state.stockCount })
                    }
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
                      aria-pressed={isSourceSelected('waste', undefined)}
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(wasteZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
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

            <div className="flex gap-1 sm:gap-2 items-start" data-tutorial="cg-tableau">
              {Array.from({ length: TABLEAU_PILES }, (_, i) => i).map(renderPile)}
            </div>

            <div data-tutorial="cg-hint-display">
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
              <p data-testid="cg-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
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

          <GameFooter className={`${gameTheme.diplomat.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="cg-controls">
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
                dataTutorial="cg-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="diplomat-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
