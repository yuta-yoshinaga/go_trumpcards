import { useCallback, useMemo, useState } from 'react';
import type { bakersgameApi, FreeCellMoveZone } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
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
import { useBakersGameGame } from '../hooks/useBakersGameGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, FreeCellResponse } from '../types/card';
import { FreeCellPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { bakersGameAutoMoveTarget } from '../utils/bakersGameAutoMoveTarget';
import { cardAlt } from '../utils/cardAlt';
import { FREECELL_HELP, parseFreecellCommand } from '../utils/cli/commands/freecellCommands';
import { formatFreecellState } from '../utils/cli/formatters/freecellFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Baker's Game tutorial step definitions. */
const BG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="fc-free-cells"]',
    messageKey: 'tutorial.freeCells',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Baker's Game solitaire page (FreeCell with same-suit stacking). */
export const BakersGamePage = withTutorial(BakersGamePageContent, 'bakersgame', BG_TUTORIAL_STEPS);
/** Inner content of the Baker's Game page, wrapped by TutorialProvider. */
function BakersGamePageContent() {
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
  } = useGamePageSetup('bakersgame');
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
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    handleAutoMove,
    isAutoCompleting,
  } = useBakersGameGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bakersgame', state);
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bakersgame');
  const cliConfig: CliGameConfig<FreeCellResponse, Parameters<typeof bakersgameApi.exec>> = useMemo(
    () => ({
      gameName: 'bakersgame',
      parseCommand: parseFreecellCommand,
      formatResponse: formatFreecellState,
      helpText: FREECELL_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === FreeCellPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: FreeCellMoveZone, target: FreeCellMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<FreeCellMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  // Double-click / double-tap shortcut: auto-send an exposed card (a tableau
  // top card or a free-cell card) to its foundation, falling back to an empty
  // free cell, when a legal target exists; otherwise do nothing (no error,
  // selection untouched). Disabled while auto-completing.
  const handleAutoMoveShortcut = useCallback(
    (source: FreeCellMoveZone, card: Card) => {
      if (!state || isAutoCompleting) return;
      const target = bakersGameAutoMoveTarget(card, state.foundation, state.freeCells);
      if (!target) return;
      handleAutoMove(source, target);
    },
    [state, isAutoCompleting, handleAutoMove],
  );

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  const [hoveredStack, setHoveredStack] = useState<{ col: number; cardIdx: number } | null>(null);

  if (!state) return <GameSkeleton gameKey="bakersgame" layout={{ kind: 'tableau', topRow: 8, tableau: 8 }} />;

  const isPlaying = state.phase === FreeCellPhase.PLAYING;
  const isGameClear = state.phase === FreeCellPhase.GAME_CLEAR;
  const isGameOver = state.phase === FreeCellPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  // Supermove limit: a tableau stack of N cards can only move when
  // (1 + freeCells) * 2^emptyCols >= N. We compute the upper-bound limit (the
  // optimistic case where the destination is NOT one of the empty columns) and
  // mark anything deeper than that as undraggable, with a red ring + tooltip
  // so the player sees the cap before the engine rejects the move.
  const emptyFreeCells = state.freeCells.filter((c) => c === null).length;
  const emptyTableauCols = state.tableau.filter((col: (Card | null)[]) => col.length === 0).length;
  const supermoveLimit = (1 + emptyFreeCells) * 2 ** emptyTableauCols;

  const isSourceSelected = (zone: string, col?: number, cell?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cell === cell &&
    selectedSource.cardIndex === cardIndex;

  return (
    <GamePageShell
      title={tc('nav.bakersgame')}
      gameThemeBg={gameTheme.bakersgame.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/bakersgame"
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
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Free cells + Foundation row */}
            <div className="flex gap-2 mb-3 items-start flex-wrap">
              {/* Free cells */}
              <div className="flex gap-2" data-tutorial="fc-free-cells">
                {state.freeCells.map((card: Card | null, idx: number) => {
                  const freeCellZone: FreeCellMoveZone = { zone: 'freecell', cell: idx };
                  return (
                    <div key={`fc-${idx.toString()}`} className="text-center">
                      <div className="text-game-text-muted text-xs mb-1">
                        <span className="hidden sm:inline">
                          {t('freecell')} {idx}
                        </span>
                        <span className="sm:hidden">
                          {t('freecellShort')}
                          {idx}
                        </span>
                      </div>
                      <DropZone
                        isDropTarget={dnd.isDropTarget(freeCellZone)}
                        onDragOver={dnd.handleDragOver(freeCellZone)}
                        onDrop={dnd.handleDrop(freeCellZone)}
                        onDragLeave={dnd.handleDragLeave}
                      >
                        {card ? (
                          <button
                            type="button"
                            onClick={(e) => {
                              // The second click of a double-click also fires
                              // onClick (detail === 2); ignore it so onDoubleClick
                              // owns the auto-move and selection stays put.
                              if (e.detail >= 2) return;
                              handleSelectSource(freeCellZone);
                            }}
                            onDoubleClick={() => handleAutoMoveShortcut(freeCellZone, card)}
                            disabled={!isPlaying || loading}
                            aria-label={cardAlt(card)}
                            aria-pressed={isSourceSelected('freecell', undefined, idx)}
                            draggable={isPlaying && !loading}
                            onDragStart={dnd.handleDragStart(freeCellZone)}
                            onDragEnd={dnd.handleDragEnd}
                            className={[
                              'p-0 border-0 bg-transparent cursor-pointer rounded',
                              focusRingWhite,
                              isSourceSelected('freecell', undefined, idx) && 'ring-2 ring-ds-warning',
                              dnd.isDragSource(freeCellZone) && 'opacity-50',
                            ]
                              .filter(Boolean)
                              .join(' ')}
                          >
                            <AnimatedCard card={card} width={cardWidth} draggable={false} />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(freeCellZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFreecellAriaLabel', { idx: String(idx) })}
                            style={{ width: cardWidth, height: cardHeight }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            {t('empty')}
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="w-4" />

              {/* Foundation piles */}
              <div className="flex gap-2" data-tutorial="fc-foundation">
                {state.foundation.map((pile: Card[], idx: number) => {
                  const foundationZone: FreeCellMoveZone = { zone: 'foundation', col: idx };
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
                            onClick={() => handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                            aria-label={t('foundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              cardCount: String(pile.length),
                            })}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={cardWidth}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                            style={{ width: cardWidth, height: cardHeight }}
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
            </div>

            {/* Tableau */}
            <div className="relative">
              <div className="flex gap-0.5 sm:gap-2 mb-3" data-tutorial="fc-tableau">
                {state.tableau.map((col: (Card | null)[], colIdx: number) => {
                  const tableauColZone: FreeCellMoveZone = { zone: 'tableau', col: colIdx };
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
                              type="button"
                              onClick={() => handleSelectTarget(tableauColZone)}
                              disabled={!isPlaying || loading || !selectedSource}
                              style={{ height: cardHeight }}
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                              aria-label={t('emptyTableauColumnAriaLabel', { idx: String(colIdx) })}
                            >
                              {t('emptyTableauColumnLabel')}
                            </button>
                          ) : (
                            col.map((card: Card | null, cardIdx: number) => {
                              const cardZone: FreeCellMoveZone = {
                                zone: 'tableau',
                                col: colIdx,
                                cardIndex: cardIdx,
                              };
                              const stackSize = col.length - cardIdx;
                              const exceedsSupermove = stackSize > supermoveLimit;
                              const isInHoveredBlock =
                                hoveredStack !== null &&
                                hoveredStack.col === colIdx &&
                                cardIdx >= hoveredStack.cardIdx &&
                                col.length - hoveredStack.cardIdx <= supermoveLimit;
                              return (
                                <div
                                  key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                  className="absolute left-0 right-0"
                                  style={{ top: cardIdx * cardOverlap }}
                                >
                                  {card ? (
                                    <button
                                      type="button"
                                      onClick={(e) => {
                                        // The second click of a double-click also
                                        // fires onClick (detail === 2); ignore it so
                                        // onDoubleClick owns the auto-move without
                                        // issuing a stray self-target move.
                                        if (e.detail >= 2) return;
                                        if (selectedSource) {
                                          handleSelectTarget(tableauColZone);
                                        } else {
                                          handleSelectSource(cardZone);
                                        }
                                      }}
                                      onDoubleClick={
                                        // Only the exposed top card of a column can auto-move.
                                        cardIdx === col.length - 1
                                          ? () => handleAutoMoveShortcut(cardZone, card)
                                          : undefined
                                      }
                                      disabled={!isPlaying || loading}
                                      aria-label={cardAlt(card)}
                                      aria-pressed={isSourceSelected('tableau', colIdx, undefined, cardIdx)}
                                      draggable={isPlaying && !loading && !exceedsSupermove}
                                      onDragStart={dnd.handleDragStart(cardZone)}
                                      onDragEnd={dnd.handleDragEnd}
                                      onMouseEnter={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onMouseLeave={() => setHoveredStack(null)}
                                      onFocus={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onBlur={() => setHoveredStack(null)}
                                      title={
                                        exceedsSupermove
                                          ? t('supermoveLimitTooltip', { limit: supermoveLimit })
                                          : undefined
                                      }
                                      data-supermove-blocked={exceedsSupermove ? 'true' : undefined}
                                      data-supermove-block={isInHoveredBlock ? 'true' : undefined}
                                      className={[
                                        'p-0 border-0 bg-transparent cursor-pointer w-full rounded',
                                        focusRingWhite,
                                        isSourceSelected('tableau', colIdx, undefined, cardIdx) &&
                                          'ring-2 ring-ds-warning',
                                        dnd.isDragSource(cardZone) && 'opacity-50',
                                        exceedsSupermove && 'opacity-60 ring-1 ring-ds-error',
                                        isInHoveredBlock && 'ring-2 ring-ds-success',
                                      ]
                                        .filter(Boolean)
                                        .join(' ')}
                                    >
                                      <AnimatedCard
                                        card={card}
                                        width={cardWidth}
                                        draggable={false}
                                        style={{ width: '100%' }}
                                      />
                                    </button>
                                  ) : (
                                    <div style={{ width: cardWidth, height: cardHeight }} />
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
                {/* Zone identifiers (tableau/freecell/foundation) double as i18n
                    keys, so they localize instead of showing raw English. */}
                {t('hintAvailable')}: {t(hint.fromZone)}
                {hint.fromCol >= 0 ? ` ${hint.fromCol}` : ''} → {t(hint.toZone)}
                {hint.toCol >= 0 ? ` ${hint.toCol}` : ''}
              </div>
            )}
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

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

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.bakersgame.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <span
                  data-testid="bg-movable-count"
                  aria-live="polite"
                  className="text-ds-text-primary text-xs font-medium bg-ds-surface-elevated rounded px-2 py-1"
                >
                  {t('movableCount', { count: supermoveLimit })}
                </span>
              )}
              {isPlaying && (
                <div data-tutorial="fc-controls">
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
                    className={btnSuccess}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting}
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
                dataTutorial="fc-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
