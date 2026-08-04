import { useCallback, useMemo } from 'react';
import type { bisleyApi } from '../api/gameApi';
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
import { useBisleyGame } from '../hooks/useBisleyGame';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BisleyMoveZone, BisleyResponse } from '../types/card';
import { BisleyPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BISLEY_HELP, parseBisleyCommand } from '../utils/cli/commands/bisleyCommands';
import { formatBisleyState } from '../utils/cli/formatters/bisleyFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

// Suit order is fixed by domain.bisleySuitOrder so a column always belongs to the
// same suit across deals.
const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

const TABLEAU_COLS = 13;

const BISLEY_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bisley-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bisley-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bisley-tableau"]',
    messageKey: 'tutorial.moves',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bisley-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Bisley solitaire page: 13 tableau columns and two foundation rows. */
export const BisleyPage = withTutorial(BisleyPageContent, 'bisley', BISLEY_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'ace') return t('frontendHint.aceFoundation', { suit: FOUNDATION_SUITS[idx] ?? '' });
  if (zone === 'king') return t('frontendHint.kingFoundation', { suit: FOUNDATION_SUITS[idx] ?? '' });
  return t('frontendHint.tableau', { col: idx });
}

function BisleyPageContent() {
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
  } = useGamePageSetup('bisley');
  const game = useBisleyGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bisley', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bisley');
  const cliConfig: CliGameConfig<BisleyResponse, Parameters<typeof bisleyApi.exec>> = useMemo(
    () => ({
      gameName: 'bisley',
      parseCommand: parseBisleyCommand,
      formatResponse: formatBisleyState,
      helpText: BISLEY_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 2;
    // 13 tableau columns share the width; the foundations sit on their own rows above.
    const colW = Math.floor((windowWidth - padX - (TABLEAU_COLS - 1) * gapPx) / TABLEAU_COLS);
    const cw = Math.min(Math.max(colW, 24), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === BisleyPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: BisleyMoveZone, target: BisleyMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<BisleyMoveZone>({
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

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="bisley" layout={{ kind: 'tableau', topRow: 8, tableau: TABLEAU_COLS }} />;

  const isPlaying = state.phase === BisleyPhase.PLAYING;
  const isGameClear = state.phase === BisleyPhase.GAME_CLEAR;
  const isGameOver = state.phase === BisleyPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const countCards = (piles: BisleyResponse['aceFoundations']) => piles.reduce((sum, pile) => sum + pile.length, 0);
  // How many of the 52 cards reached a foundation — only needed on game over, so skip the
  // reduce during normal (frequently re-rendered) play.
  const foundationCount = isGameOver ? countCards(state.aceFoundations) + countCards(state.kingFoundations) : 0;
  // The four Aces are dealt onto the foundations, so "past the deal" means a pile
  // grew beyond its seed — that is when auto-complete starts to be worth pulsing.
  const autoCompleteReady =
    state.aceFoundations.some((pile) => pile.length > 1) || state.kingFoundations.some((pile) => pile.length > 0);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx] ?? [];
    const tableauColZone: BisleyMoveZone = { zone: 'tableau', col: colIdx };
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
              <div
                // Bisley forbids moving onto an empty column, so this is a
                // placeholder rather than a drop target; role="img" carries the
                // column number to screen readers without implying it is actionable.
                role="img"
                aria-label={t('emptyColumnAriaLabel', { col: colIdx + 1 })}
                style={{ height: dims.ch }}
                className="w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent"
              >
                {t('empty')}
              </div>
            ) : (
              col.map((tc2, cardIdx) => {
                const isTop = cardIdx === col.length - 1;
                // Only the top card is movable, so selection is per column — but
                // the pressed state has to stay on that one card, or every card
                // in a selected column would announce itself as selected.
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
                        aria-label={t('cardPosAria', { card: cardAlt(tc2.card), col: colIdx + 1, pos: cardIdx + 1 })}
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

  const renderFoundationRow = (kind: 'ace' | 'king') => {
    const piles = kind === 'ace' ? state.aceFoundations : state.kingFoundations;
    const labelKey = kind === 'ace' ? 'aceFoundationAriaLabel' : 'kingFoundationAriaLabel';
    const emptyLabelKey = kind === 'ace' ? 'emptyAceFoundationAriaLabel' : 'emptyKingFoundationAriaLabel';
    const placeholder = kind === 'ace' ? 'A' : 'K';
    return (
      <div className="flex flex-col items-center">
        <div className="text-game-text-muted text-xs mb-1">
          {t(kind === 'ace' ? 'aceFoundation' : 'kingFoundation')}
        </div>
        <div className="flex gap-1 sm:gap-2">
          {FOUNDATION_SUITS.map((suit, idx) => {
            const pile = piles[idx] ?? [];
            const foundationZone: BisleyMoveZone = { zone: kind, col: idx };
            return (
              <div key={`${kind}-${suit}`} className="text-center">
                <div className="text-game-text-muted text-xs mb-1">{suit}</div>
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
                      aria-label={t(labelKey, { suit, count: pile.length })}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                    >
                      <AnimatedCard
                        card={pile[pile.length - 1]}
                        width={dims.cw}
                        draggable={false}
                        dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                      />
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={() => game.handleSelectTarget(foundationZone)}
                      disabled={!isPlaying || loading || !selectedSource}
                      aria-label={t(emptyLabelKey, { suit })}
                      style={{ width: dims.cw, height: dims.ch }}
                      className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                    >
                      {placeholder}
                    </button>
                  )}
                </DropZone>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.bisley')}
      gameThemeBg={gameTheme.bisley.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/bisley"
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
            <div className="flex flex-wrap justify-center gap-4 sm:gap-8 mb-3" data-tutorial="bisley-foundation">
              {renderFoundationRow('ace')}
              {renderFoundationRow('king')}
            </div>

            <div className="flex gap-0.5 sm:gap-2 items-start" data-tutorial="bisley-tableau">
              {Array.from({ length: TABLEAU_COLS }, (_, i) => i).map(renderTableauColumn)}
            </div>

            <div data-tutorial="bisley-hint-display">
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
              <p data-testid="bisley-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', {
                  count: foundationCount,
                  percent: Math.round((foundationCount / 52) * 100),
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
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.bisley.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="bisley-controls">
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
                dataTutorial="bisley-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="bisley-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
