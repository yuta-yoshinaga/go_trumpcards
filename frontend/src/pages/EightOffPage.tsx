import { useCallback, useEffect, useMemo, useState } from 'react';
import type { EightOffMoveZone, eightoffApi } from '../api/gameApi';
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
import { useEightOffGame } from '../hooks/useEightOffGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, EightOffHint, EightOffResponse } from '../types/card';
import { EightOffPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { EIGHTOFF_HELP, parseEightOffCommand } from '../utils/cli/commands/eightoffCommands';
import { formatEightoffState } from '../utils/cli/formatters/eightoffFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { eightOffFoundationTarget } from '../utils/eightOffFoundationTarget';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Localized "<zone> <n>" label for a hint move endpoint (n omitted when col < 0, e.g. any-foundation). */
function eightOffZoneLabel(t: (key: string) => string, zone: string, col: number): string {
  const base = zone === 'freecell' ? t('freecell') : zone === 'foundation' ? t('foundation') : t('tableau');
  return col >= 0 ? `${base} ${col + 1}` : base;
}

/** The card a hint suggests moving: the free-cell card, or the tableau card at [fromCol][cardIndex]. */
function eightOffHintCard(state: EightOffResponse, hint: EightOffHint): Card | null {
  if (hint.fromZone === 'freecell') return state.freeCells[hint.fromCol] ?? null;
  return state.tableau[hint.fromCol]?.[hint.cardIndex] ?? null;
}

/** Eight Off tutorial step definitions. */
const EO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="eo-free-cells"]',
    messageKey: 'tutorial.freeCells',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eo-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eo-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eo-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eo-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Eight Off solitaire game page with 8 free cells, 8-column tableau, and foundation. */
export const EightOffPage = withTutorial(EightOffPageContent, 'eightoff', EO_TUTORIAL_STEPS);
/** Inner content of the Eight Off page, wrapped by TutorialProvider. */
function EightOffPageContent() {
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
  } = useGamePageSetup('eightoff');
  const {
    state,
    loading,
    error,
    exec,
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
  } = useEightOffGame();
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
    const card = eightOffHintCard(state, hint);
    setHintAnnounce(
      t('hintAnnouncement', {
        card: card ? cardAlt(card) : '',
        from: eightOffZoneLabel(t, hint.fromZone, hint.fromCol),
        to: eightOffZoneLabel(t, hint.toZone, hint.toCol),
      }),
    );
  }, [hintNonce]);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('eightoff', state);
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('eightoff');
  const cliConfig: CliGameConfig<EightOffResponse, Parameters<typeof eightoffApi.exec>> = useMemo(
    () => ({
      gameName: 'eightoff',
      parseCommand: parseEightOffCommand,
      formatResponse: formatEightoffState,
      helpText: EIGHTOFF_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === EightOffPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: EightOffMoveZone, target: EightOffMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );

  // Double-click / double-tap shortcut: auto-send an exposed card (a column's
  // top card or a free-cell card) straight to its foundation when a legal target
  // exists; otherwise do nothing (single-click selection is left untouched).
  // Mirrors the FreeCell / Easthaven foundation shortcut.
  const handleFoundationShortcut = useCallback(
    (source: EightOffMoveZone, card: Card) => {
      if (!state) return;
      const target = eightOffFoundationTarget(card, state.foundation);
      if (!target) return;
      dispatchMove(source, target);
    },
    [state, dispatchMove],
  );
  const dnd = useSolitaireDragDrop<EightOffMoveZone>({
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

  if (!state) return <GameSkeleton gameKey="eightoff" layout={{ kind: 'tableau', topRow: 8, tableau: 8 }} />;

  const isPlaying = state.phase === EightOffPhase.PLAYING;
  const isGameClear = state.phase === EightOffPhase.GAME_CLEAR;
  const isGameOver = state.phase === EightOffPhase.GAME_OVER;
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

  // Visual hint: ring the suggested source card (info) and target zone (success).
  const isHintSourceTableau = (colIdx: number, cardIdx: number) =>
    hint != null && hint.fromZone === 'tableau' && hint.fromCol === colIdx && hint.cardIndex === cardIdx;
  const isHintSourceFreecell = (cell: number) => hint != null && hint.fromZone === 'freecell' && hint.fromCol === cell;
  const isHintTarget = (zone: string, col: number) => hint != null && hint.toZone === zone && hint.toCol === col;

  return (
    <GamePageShell
      title={tc('nav.eightoff')}
      gameThemeBg={gameTheme.eightoff.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/eightoff"
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
          {/* **上限を知るには赤くなった札にマウスを乗せるしかなかった (#4801)。**
              後追いの体験になる。ほぼ同型の Penguin は同じ計算式のバッジを
              ヘッダーに常設している。 */}
          <span
            data-testid="eo-supermove-badge"
            title={t('supermoveBadgeTooltip', { cells: emptyFreeCells, cols: emptyTableauCols })}
          >
            {t('supermoveBadge', { limit: supermoveLimit })}
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
              <div className="flex gap-2" data-tutorial="eo-free-cells">
                {state.freeCells.map((card: Card | null, idx: number) => {
                  const freeCellZone: EightOffMoveZone = { zone: 'freecell', cell: idx };
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
                              // owns the foundation shortcut without a stray select.
                              if (e.detail >= 2) return;
                              handleSelectSource(freeCellZone);
                            }}
                            onDoubleClick={() => handleFoundationShortcut(freeCellZone, card)}
                            disabled={!isPlaying || loading}
                            aria-label={cardAlt(card)}
                            aria-pressed={isSourceSelected('freecell', undefined, idx)}
                            data-testid={`eo-freecell-${idx.toString()}`}
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
                            data-testid={`eo-freecell-empty-${idx.toString()}`}
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
              <div className="flex gap-2" data-tutorial="eo-foundation">
                {state.foundation.map((pile: Card[], idx: number) => {
                  const foundationZone: EightOffMoveZone = { zone: 'foundation', col: idx };
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
                            data-testid={`eo-foundation-${idx.toString()}`}
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
                            data-testid={`eo-foundation-empty-${idx.toString()}`}
                            style={{ width: cardWidth, height: cardHeight }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}${isHintTarget('foundation', idx) ? ' ring-2 ring-ds-success animate-pulse' : ''}`}
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
              <div className="flex gap-0.5 sm:gap-2 mb-3" data-tutorial="eo-tableau">
                {state.tableau.map((col: (Card | null)[], colIdx: number) => {
                  const tableauColZone: EightOffMoveZone = { zone: 'tableau', col: colIdx };
                  return (
                    <div
                      key={`col-${colIdx.toString()}`}
                      data-testid={`eo-col-${colIdx.toString()}`}
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
                              aria-label={t('emptyColumnAriaLabel', { rank: 'K' })}
                              style={{ height: cardHeight }}
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                            >
                              K
                            </button>
                          ) : (
                            col.map((card: Card | null, cardIdx: number) => {
                              const cardZone: EightOffMoveZone = {
                                zone: 'tableau',
                                col: colIdx,
                                cardIndex: cardIdx,
                              };
                              const stackSize = col.length - cardIdx;
                              const exceedsSupermove = stackSize > supermoveLimit;
                              const isTopCard = cardIdx === col.length - 1;
                              // **タッチにはホバーが無い。**タップで選んだ直後に
                              // フォーカスが外れると、どこまでが一緒に動くのか
                              // 手がかりが消えるので、選択状態にも追従させる
                              // (Easthaven #4815 と同じ形) (#5612)。
                              // `?? null` は置かない。タブローの**選択元**は必ず
                              // cardIndex を持つ (cardIndex 無しの tableau zone は
                              // 空列を指す移動先専用) ので、埋めようのない分岐が
                              // 増えるだけになる。
                              const selectedBlockStart =
                                selectedSource?.zone === 'tableau' && selectedSource.col === colIdx
                                  ? selectedSource.cardIndex
                                  : undefined;
                              // 上限判定は両方に掛ける ── 動かせない塊を「動く」と
                              // 見せてはいけない。
                              const blockStart =
                                hoveredStack !== null && hoveredStack.col === colIdx
                                  ? hoveredStack.cardIdx
                                  : selectedBlockStart;
                              const isInHoveredBlock =
                                blockStart !== undefined &&
                                cardIdx >= blockStart &&
                                col.length - blockStart <= supermoveLimit;
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
                                        // onDoubleClick owns the foundation shortcut
                                        // without a stray select/target move.
                                        if (e.detail >= 2) return;
                                        if (selectedSource) {
                                          handleSelectTarget(tableauColZone);
                                        } else {
                                          handleSelectSource(cardZone);
                                        }
                                      }}
                                      onDoubleClick={
                                        // Only a column's exposed top card can move
                                        // straight to a foundation.
                                        isTopCard ? () => handleFoundationShortcut(cardZone, card) : undefined
                                      }
                                      disabled={!isPlaying || loading}
                                      // 上限超過は title とリングだけで示していたので、
                                      // ホバーできる人にしか届かない。draggable も落として
                                      // いるのに、動かせない理由が読み上げに出ない (#5820)。
                                      aria-label={
                                        exceedsSupermove
                                          ? `${cardAlt(card)} — ${t('supermoveLimitTooltip', { limit: supermoveLimit, cells: emptyFreeCells, cols: emptyTableauCols })}`
                                          : cardAlt(card)
                                      }
                                      data-testid={`eo-tableau-${colIdx.toString()}-${cardIdx.toString()}`}
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
            <div className="sr-only" role="status" aria-live="polite" data-testid="eo-hint-announce">
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

          <GameFooter className={`${gameTheme.eightoff.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="eo-controls">
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
                dataTutorial="eo-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="eight-off-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
