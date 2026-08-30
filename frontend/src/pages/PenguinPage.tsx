import { useCallback, useEffect, useMemo, useState } from 'react';
import type { PenguinMoveZone, penguinApi } from '../api/gameApi';
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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { usePenguinGame } from '../hooks/usePenguinGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, PenguinHint, PenguinResponse } from '../types/card';
import { PenguinPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PENGUIN_HELP, parsePenguinCommand } from '../utils/cli/commands/penguinCommands';
import { formatPenguinState } from '../utils/cli/formatters/penguinFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Localized "<zone> <n>" label for a hint move endpoint (n omitted when col < 0, e.g. any-foundation). */
function penguinZoneLabel(t: (key: string) => string, zone: string, col: number): string {
  const base = zone === 'freecell' ? t('freecell') : zone === 'foundation' ? t('foundation') : t('tableau');
  return col >= 0 ? `${base} ${col + 1}` : base;
}

/** The card a hint suggests moving: the free-cell card, or the tableau card at [fromCol][cardIndex]. */
function penguinHintCard(state: PenguinResponse, hint: PenguinHint): Card | null {
  if (hint.fromZone === 'freecell') return state.freeCells[hint.fromCol] ?? null;
  return state.tableau[hint.fromCol]?.[hint.cardIndex] ?? null;
}

/** Convert a card value (1-13) to display label. */
function baseRankLabel(rank: number): string {
  switch (rank) {
    case 1:
      return 'A';
    case 11:
      return 'J';
    case 12:
      return 'Q';
    case 13:
      return 'K';
    default:
      return String(rank);
  }
}

/** Compute the rank label that precedes baseRank with wraparound (1→K, 2→A). */
function prevRankLabel(baseRank: number): string {
  const prev = baseRank === 1 ? 13 : baseRank - 1;
  return baseRankLabel(prev);
}

/** Penguin tutorial step definitions. */
const PG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pg-free-cells"]',
    messageKey: 'tutorial.freeCells',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pg-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pg-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pg-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pg-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Penguin solitaire game page with 7 free cells, 7-column tableau, and foundation. */
export const PenguinPage = withTutorial(PenguinPageContent, 'penguin', PG_TUTORIAL_STEPS);
/** Inner content of the Penguin page, wrapped by TutorialProvider. */
function PenguinPageContent() {
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
  } = useGamePageSetup('penguin');
  const {
    state,
    loading,
    error,
    exec: apiExec,
    retry,
    hintError,
    selectedSource,
    hint,
    hintNonce,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  } = usePenguinGame();
  // Screen-reader announcement for the hint (visually it is only ring highlights).
  // Driven off hintNonce so it fires once per hint request, reading the current
  // hint/state snapshot; a null hint after a request means no legal move exists.
  const [hintAnnounce, setHintAnnounce] = useState('');
  // biome-ignore lint/correctness/useExhaustiveDependencies: intentionally react only to a new hint request (hintNonce); adding hint/state/t would re-run on unrelated updates and re-announce.
  useEffect(() => {
    // Skip on a failed hint fetch (hintError set, hint left null) so we don't
    // wrongly announce "no moves" alongside the network-error banner.
    if (hintNonce === 0 || !state || hintError) return;
    if (!hint) {
      setHintAnnounce(t('hintNoMoves'));
      return;
    }
    const card = penguinHintCard(state, hint);
    setHintAnnounce(
      t('hintAnnouncement', {
        card: card ? cardAlt(card) : '',
        from: penguinZoneLabel(t, hint.fromZone, hint.fromCol),
        to: penguinZoneLabel(t, hint.toZone, hint.toCol),
      }),
    );
  }, [hintNonce]);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('penguin', state);
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('penguin');
  const cliConfig: CliGameConfig<PenguinResponse, Parameters<typeof penguinApi.exec>> = useMemo(
    () => ({
      gameName: 'penguin',
      parseCommand: parsePenguinCommand,
      formatResponse: formatPenguinState,
      helpText: PENGUIN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === PenguinPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: PenguinMoveZone, target: PenguinMoveZone) => {
      void apiExec('move', source, target);
    },
    [apiExec],
  );
  const dnd = useSolitaireDragDrop<PenguinMoveZone>({
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

  const [hoveredStack, setHoveredStack] = useState<{ col: number; cardIdx: number } | null>(null);

  if (!state) return <GameSkeleton gameKey="penguin" layout={{ kind: 'tableau', topRow: 7, tableau: 7 }} />;

  const isPlaying = state.phase === PenguinPhase.PLAYING;
  const isGameClear = state.phase === PenguinPhase.GAME_CLEAR;
  const isGameOver = state.phase === PenguinPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const emptyFreeCells = state.freeCells.filter((c) => c === null).length;
  const emptyTableauCols = state.tableau.filter((col: (Card | null)[]) => col.length === 0).length;
  // **上限はドメインが決める。**ここで一般式を持つと、空き列自身を経由地に
  // 使えないぶんの差 (maxMovableCardsToEmptyColumn) が抜け、空き列宛ての束を
  // 「動かせる」と見せてサーバーに弾かれる (#5614)。
  const supermoveLimit = state.maxMovableCards;
  const emptyColLimit = state.maxMovableCardsToEmptyColumn;

  // 選択中の束の枚数。空き列が受け取れるかはこれと emptyColLimit で決まる。
  const selectedStackSize =
    selectedSource?.zone === 'tableau' && selectedSource.col !== undefined && selectedSource.cardIndex !== undefined
      ? (state.tableau[selectedSource.col]?.length ?? 0) - selectedSource.cardIndex
      : 0;

  const emptyColPlaceholder = prevRankLabel(state.baseRank);

  const isSourceSelected = (zone: string, col?: number, cell?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cell === cell &&
    selectedSource.cardIndex === cardIndex;

  const isHintSourceTableau = (colIdx: number, cardIdx: number) =>
    hint != null && hint.fromZone === 'tableau' && hint.fromCol === colIdx && hint.cardIndex === cardIdx;
  const isHintSourceFreecell = (cell: number) => hint != null && hint.fromZone === 'freecell' && hint.fromCol === cell;
  const isHintTarget = (zone: string, col: number) => hint != null && hint.toZone === zone && hint.toCol === col;

  return (
    <GamePageShell
      title={tc('nav.penguin')}
      gameThemeBg={gameTheme.penguin.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/penguin"
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
          <span>
            {t('baseRank')}: {baseRankLabel(state.baseRank)}
          </span>
          <span data-testid="pg-supermove-badge" title={t('supermoveBadgeTooltip')}>
            {t('supermoveBadge', {
              limit: supermoveLimit,
              cells: emptyFreeCells,
              cols: emptyTableauCols,
            })}
            {emptyColLimit > 0 && <> {t('supermoveToEmpty', { limit: emptyColLimit })}</>}
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
              <div className="flex gap-2" data-tutorial="pg-free-cells">
                {state.freeCells.map((card: Card | null, idx: number) => {
                  const freeCellZone: PenguinMoveZone = { zone: 'freecell', cell: idx };
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
                            onClick={() => handleSelectSource(freeCellZone)}
                            disabled={!isPlaying || loading}
                            aria-label={cardAlt(card)}
                            aria-pressed={isSourceSelected('freecell', undefined, idx)}
                            data-testid={`pg-freecell-${idx.toString()}`}
                            draggable={isPlaying && !loading}
                            onDragStart={dnd.handleDragStart(freeCellZone)}
                            onDragEnd={dnd.handleDragEnd}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${isSourceSelected('freecell', undefined, idx) ? 'ring-2 ring-ds-warning' : isHintSourceFreecell(idx) ? 'ring-2 ring-ds-info animate-pulse' : focusRingWhite} ${dnd.isDragSource(freeCellZone) ? 'opacity-50' : ''}`}
                          >
                            <AnimatedCard card={card} width={cardWidth} draggable={false} />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(freeCellZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFreecellAriaLabel', { idx: String(idx) })}
                            data-testid={`pg-freecell-empty-${idx.toString()}`}
                            style={{ width: cardWidth, height: cardHeight }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}${isHintTarget('freecell', idx) ? ' ring-2 ring-ds-success animate-pulse' : ''}`}
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
              <div className="flex flex-col items-center">
                <div className="flex gap-2" data-tutorial="pg-foundation">
                  {state.foundation.map((pile: Card[], idx: number) => {
                    const foundationZone: PenguinMoveZone = { zone: 'foundation', col: idx };
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
                              data-testid={`pg-foundation-${idx.toString()}`}
                              className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}${isHintTarget('foundation', idx) ? ' ring-2 ring-ds-success animate-pulse' : ''}`}
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
                              data-testid={`pg-foundation-empty-${idx.toString()}`}
                              style={{ width: cardWidth, height: cardHeight }}
                              className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}${isHintTarget('foundation', idx) ? ' ring-2 ring-ds-success animate-pulse' : ''}`}
                            >
                              {baseRankLabel(state.baseRank)}
                            </button>
                          )}
                        </DropZone>
                      </div>
                    );
                  })}
                </div>
                <div
                  className="text-game-text-muted text-[10px] mt-1 max-w-[12rem] text-center"
                  data-testid="pg-base-rank-legend"
                >
                  {t('baseRankLegend', {
                    rank: baseRankLabel(state.baseRank),
                    prev: prevRankLabel(state.baseRank),
                  })}
                </div>
              </div>
            </div>

            {/* Tableau */}
            <div className="relative">
              <div className="flex gap-0.5 sm:gap-2 mb-3" data-tutorial="pg-tableau">
                {state.tableau.map((col: (Card | null)[], colIdx: number) => {
                  const tableauColZone: PenguinMoveZone = { zone: 'tableau', col: colIdx };
                  return (
                    <div
                      key={`col-${colIdx.toString()}`}
                      data-testid={`pg-col-${colIdx.toString()}`}
                      className={`flex-1 min-w-0${isHintTarget('tableau', colIdx) ? ' rounded ring-2 ring-ds-success animate-pulse' : ''}`}
                    >
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
                              aria-label={t('emptyColumnAriaLabel', { rank: emptyColPlaceholder })}
                              style={{ height: cardHeight }}
                              data-testid={`pg-empty-col-${colIdx.toString()}`}
                              // 空き列だけ上限が低い。選んだ束が超えているなら、
                              // クリックする前に分かるようにする (#5614)。
                              data-empty-col-blocked={selectedStackSize > emptyColLimit ? 'true' : undefined}
                              title={
                                selectedStackSize > emptyColLimit
                                  ? t('emptyColLimitTooltip', { limit: emptyColLimit, size: selectedStackSize })
                                  : undefined
                              }
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${
                                selectedStackSize > emptyColLimit ? 'opacity-50' : ''
                              }`}
                            >
                              {emptyColPlaceholder}
                            </button>
                          ) : (
                            col.map((card: Card | null, cardIdx: number) => {
                              const cardZone: PenguinMoveZone = {
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
                                      onClick={() => {
                                        if (selectedSource) {
                                          handleSelectTarget(tableauColZone);
                                        } else {
                                          handleSelectSource(cardZone);
                                        }
                                      }}
                                      disabled={!isPlaying || loading}
                                      // 上限超過は title とリングだけで示していたので、
                                      // ホバーできる人にしか届かない。draggable も落として
                                      // いるのに、動かせない理由が読み上げに出ない (#5820)。
                                      aria-label={
                                        exceedsSupermove
                                          ? `${cardAlt(card)} — ${t('supermoveLimitTooltip', { limit: supermoveLimit, cells: emptyFreeCells, cols: emptyTableauCols })}`
                                          : cardAlt(card)
                                      }
                                      aria-pressed={isSourceSelected('tableau', colIdx, undefined, cardIdx)}
                                      data-testid={`pg-tableau-${colIdx.toString()}-${cardIdx.toString()}`}
                                      draggable={isPlaying && !loading && !exceedsSupermove}
                                      onDragStart={dnd.handleDragStart(cardZone)}
                                      onDragEnd={dnd.handleDragEnd}
                                      onMouseEnter={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onMouseLeave={() => setHoveredStack(null)}
                                      onFocus={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onBlur={() => setHoveredStack(null)}
                                      title={
                                        exceedsSupermove
                                          ? t('supermoveLimitTooltip', {
                                              limit: supermoveLimit,
                                              cells: emptyFreeCells,
                                              cols: emptyTableauCols,
                                            })
                                          : undefined
                                      }
                                      data-supermove-blocked={exceedsSupermove ? 'true' : undefined}
                                      data-supermove-block={isInHoveredBlock ? 'true' : undefined}
                                      className={[
                                        'p-0 border-0 bg-transparent cursor-pointer w-full rounded',
                                        isSourceSelected('tableau', colIdx, undefined, cardIdx)
                                          ? 'ring-2 ring-ds-warning'
                                          : isHintSourceTableau(colIdx, cardIdx)
                                            ? 'ring-2 ring-ds-info animate-pulse'
                                            : exceedsSupermove
                                              ? 'opacity-60 ring-1 ring-ds-error'
                                              : isInHoveredBlock
                                                ? 'ring-2 ring-ds-success'
                                                : focusRingWhite,
                                        dnd.isDragSource(cardZone) && 'opacity-50',
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

            {/* Hint is shown visually via ring highlights on the suggested source card
                and target zone; this sr-only live region conveys the same move (or the
                no-move result) to screen readers. */}
            <div key={hintNonce} className="sr-only" role="status" aria-live="polite" data-testid="pg-hint-announce">
              {hintAnnounce}
            </div>
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

          <GameFooter className={`${gameTheme.penguin.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="pg-controls">
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
                dataTutorial="pg-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="penguin-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
