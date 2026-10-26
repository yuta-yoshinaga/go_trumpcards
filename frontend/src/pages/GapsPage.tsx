import { type KeyboardEvent as ReactKeyboardEvent, useCallback, useId, useState } from 'react';
import { type GapsMoveZone, gapsApi } from '../api/gameApi';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { GapsResponse } from '../types/card';
import { GapsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, isSuitDesign, suitSymbol } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { GAPS_HELP, parseGapsCommand } from '../utils/cli/commands/gapsCommands';
import { formatGapsState } from '../utils/cli/formatters/gapsFormatter';
import { isRequestedHint } from '../utils/hintRequest';

const RED_DESIGNS = new Set(['HEART', 'DIAMOND']);

const GAPS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gaps-grid"]',
    messageKey: 'tutorial.grid',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gaps-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gaps-grid"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gaps-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GAPS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GapsPhase.PLAYING]: 'playing',
  [GapsPhase.GAME_CLEAR]: 'gameClear',
  [GapsPhase.GAME_OVER]: 'gameOver',
};

type GapsGame = ReturnType<typeof useGame>;
type GapsView = ReturnType<typeof useView>;
type GapsContext = GamePageContext<GapsResponse, GapsGame, GapsView>;

function useGame() {
  const { state, loading, error, exec, retry } = useGameApi(gapsApi.exec);

  useMountReset(exec);

  const handleReset = useCallback(() => {
    void exec('reset');
  }, [exec]);

  const handleUndo = useCallback(() => {
    void exec('undo');
  }, [exec]);

  /** Undo N moves at once to escape a stalemate (mirrors Sultan / Forty and Eight). */
  const handleUndoEscape = useCallback(
    (n: number) => {
      void exec('undo_n', undefined, undefined, n);
    },
    [exec],
  );

  const handleRedeal = useCallback(() => {
    void exec('redeal');
  }, [exec]);

  const handleGiveUp = useCallback(() => {
    void exec('giveup');
  }, [exec]);

  const handleHint = useCallback(() => {
    void exec('hint');
  }, [exec]);

  const dispatchMove = useCallback(
    (source: GapsMoveZone, target: GapsMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );

  return {
    state,
    loading,
    error,
    exec,
    retry,
    handleReset,
    handleUndo,
    handleUndoEscape,
    handleRedeal,
    handleGiveUp,
    handleHint,
    dispatchMove,
  };
}

function useView({ game }: { game: GapsGame }) {
  const { state, loading, dispatchMove, handleRedeal } = game;
  const selectSourceHintId = useId();
  const { cardWidth, cardHeight } = useCardDimensions();

  const [redealConfirmOpen, setRedealConfirmOpen] = useState(false);
  const requestRedeal = useCallback(() => setRedealConfirmOpen(true), []);
  const confirmRedeal = useCallback(() => {
    setRedealConfirmOpen(false);
    handleRedeal();
  }, [handleRedeal]);

  const isPlaying = state?.phase === GapsPhase.PLAYING;
  const dnd = useSolitaireDragDrop<GapsMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlaying,
    disabled: loading,
  });

  // Keep the existing drag path and add a click/touch/keyboard path: select a
  // movable card first, then activate the destination cell.
  const [selectedSource, setSelectedSource] = useState<GapsMoveZone | null>(null);
  const [activeCell, setActiveCell] = useState({ row: 0, col: 0 });
  const handleGridKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLButtonElement>, row: number, col: number) => {
      const moves: Record<string, [number, number]> = {
        ArrowLeft: [0, -1],
        ArrowRight: [0, 1],
        ArrowUp: [-1, 0],
        ArrowDown: [1, 0],
      };
      const move = moves[event.key];
      if (!move) return;
      event.preventDefault();
      const board = state?.grid;
      if (!board?.length) return;
      const nextRow = Math.max(0, Math.min(board.length - 1, row + move[0]));
      const nextCol = Math.max(0, Math.min((board[nextRow]?.length ?? 1) - 1, col + move[1]));
      setActiveCell({ row: nextRow, col: nextCol });
      document
        .querySelector<HTMLButtonElement>(
          `[data-testid="gaps-cell-${nextRow}-${nextCol}"], [data-testid="gaps-locked-${nextRow}-${nextCol}"]`,
        )
        ?.focus();
    },
    [state?.grid],
  );
  const isSameZone = useCallback(
    (left: GapsMoveZone | null, right: GapsMoveZone) =>
      left?.zone === right.zone && left?.row === right.row && left?.col === right.col,
    [],
  );
  const handleSelectSource = useCallback((zone: GapsMoveZone) => {
    setSelectedSource((previous) =>
      previous?.zone === zone.zone && previous.row === zone.row && previous.col === zone.col ? null : zone,
    );
  }, []);
  const handleSelectTarget = useCallback(
    (zone: GapsMoveZone) => {
      if (!selectedSource) return;
      dispatchMove(selectedSource, zone);
      setSelectedSource(null);
    },
    [dispatchMove, selectedSource],
  );

  const phaseNames = usePhaseNames('gaps', GAPS_PHASE_KEYS);

  return {
    selectSourceHintId,
    cardWidth,
    cardHeight,
    redealConfirmOpen,
    setRedealConfirmOpen,
    requestRedeal,
    confirmRedeal,
    dnd,
    selectedSource,
    activeCell,
    setActiveCell,
    handleGridKeyDown,
    isSameZone,
    handleSelectSource,
    handleSelectTarget,
    phaseNames,
  };
}

function HeaderExtra({ state, t }: GapsContext) {
  return (
    <span>
      {t('moveCount')}: {state.moveCount}
    </span>
  );
}

function AboveContent({ state, t, tc, view }: GapsContext) {
  return (
    <ConfirmDialog
      open={view.redealConfirmOpen}
      title={t('redealConfirmTitle')}
      message={t('redealConfirmMessage', { remaining: state.redealsRemaining })}
      confirmLabel={tc('button.confirm')}
      cancelLabel={tc('button.cancel')}
      onConfirm={view.confirmRedeal}
      onCancel={() => view.setRedealConfirmOpen(false)}
    />
  );
}

function Board({ state, t, tc, busy, view, frontendHint }: GapsContext) {
  const {
    selectSourceHintId,
    cardWidth,
    cardHeight,
    dnd,
    selectedSource,
    activeCell,
    setActiveCell,
    handleGridKeyDown,
    isSameZone,
    handleSelectSource,
    handleSelectTarget,
  } = view;
  const isPlaying = state.phase === GapsPhase.PLAYING;
  const lockedPrefixLengths = state.lockedPrefixLengths;

  return (
    <>
      <span id={selectSourceHintId} className="sr-only">
        {tc('label.selectSourceFirst')}
      </span>
      <div
        data-tutorial="gaps-grid"
        className="mx-auto mb-3 flex w-fit max-w-full flex-col items-start gap-1 overflow-x-auto p-1"
      >
        {state.grid.map((row, rIdx) => {
          const lockedCount = lockedPrefixLengths[rIdx];
          return (
            <div key={`row-${rIdx.toString()}`} className="flex gap-1">
              {row.map((cell, cIdx) => {
                const zone: GapsMoveZone = { zone: 'grid', row: rIdx, col: cIdx };
                // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
                // ヒントを載せるので、`state.hint` を直接読むと常時ハイライトになる (#4605)。
                const requestedHint = isRequestedHint(state) ? state.hint : undefined;
                const isHintFrom = requestedHint && requestedHint.fromRow === rIdx && requestedHint.fromCol === cIdx;
                const isHintTo = requestedHint && requestedHint.toRow === rIdx && requestedHint.toCol === cIdx;
                if (cell === null) {
                  const ghost = state.gapNeeds?.[rIdx]?.[cIdx];
                  // Mirror the (aria-hidden) ghost hint into the cell's label so
                  // SR users learn which card each gap accepts, or that it's blocked.
                  const gapAria =
                    ghost?.kind === 'needed'
                      ? t('gapAriaNeeded', { card: cardAlt({ design: ghost.design, value: ghost.value }) })
                      : ghost?.kind === 'anySuit'
                        ? t('gapAriaAnySuit', { value: valueName(ghost.value) })
                        : ghost?.kind === 'blocked'
                          ? t('gapAriaBlocked')
                          : t('gap');
                  return (
                    <DropZone
                      key={`cell-${rIdx.toString()}-${cIdx.toString()}`}
                      isDropTarget={dnd.isDropTarget(zone)}
                      onDragOver={dnd.handleDragOver(zone)}
                      onDragLeave={dnd.handleDragLeave}
                      onDrop={dnd.handleDrop(zone)}
                      className="relative"
                    >
                      {/* Keep drag handlers on an enabled div: browsers do not dispatch dragover/drop to disabled form controls. */}
                      <button
                        type="button"
                        tabIndex={activeCell.row === rIdx && activeCell.col === cIdx ? 0 : -1}
                        onFocus={() => setActiveCell({ row: rIdx, col: cIdx })}
                        onKeyDown={(event) => handleGridKeyDown(event, rIdx, cIdx)}
                        onClick={() => handleSelectTarget(zone)}
                        aria-label={gapAria}
                        data-testid={`gaps-cell-${rIdx.toString()}-${cIdx.toString()}`}
                        className={`relative flex items-center justify-center rounded border-2 ${
                          dnd.isDropTarget(zone)
                            ? 'border-ds-warning bg-ds-warning/20'
                            : 'border-dashed border-white/30'
                        } ${isHintTo ? 'ring-2 ring-ds-warning' : ''} ${focusRingWhite}`}
                        style={{ width: cardWidth, height: cardHeight }}
                        disabled={!isPlaying || busy}
                        aria-disabled={!selectedSource || undefined}
                        aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                      >
                        {ghost?.kind === 'needed' && (
                          <span
                            aria-hidden="true"
                            data-testid={`gaps-ghost-${rIdx}-${cIdx}`}
                            className={`text-base font-semibold opacity-30 ${RED_DESIGNS.has(ghost.design) ? 'text-ds-error-text' : 'text-ds-text-primary'}`}
                          >
                            {isSuitDesign(ghost.design) ? suitSymbol(ghost.design) : undefined}
                            {valueName(ghost.value)}
                          </span>
                        )}
                        {ghost?.kind === 'anySuit' && (
                          <span
                            aria-hidden="true"
                            data-testid={`gaps-ghost-${rIdx}-${cIdx}`}
                            className="text-base font-semibold text-white opacity-30"
                          >
                            {valueName(ghost.value)}
                          </span>
                        )}
                        {ghost?.kind === 'blocked' && (
                          <span
                            aria-hidden="true"
                            data-testid={`gaps-ghost-${rIdx}-${cIdx}-blocked`}
                            className="text-xl opacity-40"
                          >
                            🚫
                          </span>
                        )}
                      </button>
                    </DropZone>
                  );
                }
                const isLocked = cIdx < lockedCount;
                // Locked cards survive redeal — make them visually fixed by
                // dropping the draggable affordance, so users don't try to
                // move them and end up with a no-op drop.
                const ringClass = isHintFrom ? 'ring-2 ring-ds-warning' : isLocked ? 'ring-2 ring-ds-success' : '';
                return (
                  <button
                    type="button"
                    key={`cell-${rIdx.toString()}-${cIdx.toString()}`}
                    tabIndex={activeCell.row === rIdx && activeCell.col === cIdx ? 0 : -1}
                    onFocus={() => setActiveCell({ row: rIdx, col: cIdx })}
                    onKeyDown={(event) => handleGridKeyDown(event, rIdx, cIdx)}
                    draggable={isPlaying && !busy && !isLocked}
                    onDragStart={dnd.handleDragStart(zone)}
                    onDragEnd={dnd.handleDragEnd}
                    onClick={() => (selectedSource ? handleSelectTarget(zone) : !isLocked && handleSelectSource(zone))}
                    aria-label={t(isLocked ? 'lockedCardAria' : 'cardAria', {
                      card: cardAlt(cell),
                      row: rIdx,
                      col: cIdx,
                      locked: t('lockedAria'),
                    })}
                    disabled={!isPlaying || busy}
                    aria-pressed={isSameZone(selectedSource, zone)}
                    data-testid={
                      isLocked
                        ? `gaps-locked-${rIdx.toString()}-${cIdx.toString()}`
                        : `gaps-cell-${rIdx.toString()}-${cIdx.toString()}`
                    }
                    className={`relative p-0 border-0 bg-transparent rounded ${focusRingWhite} ${ringClass} ${dnd.isDragSource(zone) ? 'opacity-50' : ''}`}
                  >
                    <AnimatedCard card={cell} width={cardWidth} />
                    {isLocked && (
                      <span
                        aria-hidden="true"
                        title={t('lockedTooltip')}
                        className="absolute top-0.5 right-0.5 text-[10px] leading-none bg-ds-success/80 text-ds-text-on-accent rounded-sm px-1 py-0.5 shadow"
                      >
                        🔒
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          );
        })}
      </div>

      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterActions({ game, state, t, busy, view, confirmGiveUpAction }: GapsContext) {
  const isPlaying = state.phase === GapsPhase.PLAYING;
  if (!isPlaying) return null;
  // Derive the total from the already-computed per-row lengths to avoid a second board pass.
  const lockedTotal = state.lockedPrefixLengths.reduce((sum, n) => sum + n, 0);

  return (
    <div data-tutorial="gaps-controls" data-testid="gaps-controls" className="flex gap-2 flex-wrap">
      <span className="self-center text-ds-text-primary text-sm">
        {t('redealsRemaining')}: {state.redealsRemaining}
      </span>
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleUndo}
        disabled={busy || !state.canUndo}
        aria-keyshortcuts="z"
      >
        {t('undo')}
        <KbdBadge label={t('kbd.undo')} />
      </button>
      {/* 手詰まりはドメインが判定して返している (#5609)。CUI は #4800 で
          警告を出しているのに、Web には脱出手段の案内が無かった。 */}
      {state.isStalemate && (
        <StalemateEscapeButton
          undoToEscape={state.undoToEscape ?? 0}
          onEscape={game.handleUndoEscape}
          disabled={busy}
        />
      )}
      <button
        type="button"
        className={btnPrimary}
        onClick={view.requestRedeal}
        disabled={busy || state.redealsRemaining <= 0}
        data-testid="gaps-redeal-button"
        aria-keyshortcuts="d"
      >
        {t('redealKeepLabel', {
          used: state.redealsUsed,
          total: state.redealsUsed + state.redealsRemaining,
          locked: lockedTotal,
        })}
        <KbdBadge label={t('kbd.redeal')} />
      </button>
      <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy} aria-keyshortcuts="h">
        {t('hint')}
        <KbdBadge label={t('kbd.hint')} />
      </button>
      <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={busy} aria-keyshortcuts="g">
        {t('giveup')}
        <KbdBadge label={t('kbd.giveUp')} />
      </button>
    </div>
  );
}

const gapsDefinition: GameDefinition<GapsResponse, GapsGame, Parameters<typeof gapsApi.exec>, GapsView> = {
  key: 'gaps',
  useGame,
  useView,
  tutorial: GAPS_TUTORIAL_STEPS,
  skeleton: { kind: 'tiered-rows', rows: [13, 13, 13, 13] },
  cli: ({ state }) => ({
    parseCommand: (input: string) => parseGapsCommand(input, state),
    formatResponse: formatGapsState,
    helpText: GAPS_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase] ?? '',
    ended: state.phase === GapsPhase.GAME_CLEAR || state.phase === GapsPhase.GAME_OVER,
    won: state.phase === GapsPhase.GAME_CLEAR,
  }),
  reset: ({ game }) => {
    void game.exec('reset');
  },
  // Keyboard shortcuts mirror the sibling solitaire pages (EightOff / Baker's
  // Dozen). Bind letter keys only — Enter/Space would double-fire on a focused
  // button. Undo/redeal expose per-binding `enabled` so a disabled action's key
  // is inert, matching the button's own disabled state.
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, state, view, confirmGiveUpAction }) => [
    { key: 'z', action: game.handleUndo, enabled: !!state.canUndo },
    { key: 'd', action: view.requestRedeal, enabled: state.redealsRemaining > 0 },
    { key: 'h', action: game.handleHint },
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined) },
  ],
  shortcutsEnabled: ({ state, view }) => state.phase === GapsPhase.PLAYING && !view.redealConfirmOpen,
  landscapeBanner: true,
  resetTutorialId: 'gaps-reset-button',
  HeaderExtra,
  AboveContent,
  Board,
  FooterActions,
};

/** Renders the Gaps / Montana solitaire game page. */
export const GapsPage = createGamePage(gapsDefinition);
