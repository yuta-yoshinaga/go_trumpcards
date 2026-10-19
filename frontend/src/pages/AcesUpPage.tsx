import { useEffect, useRef, useState } from 'react';
import type { acesupApi } from '../api/gameApi';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useAcesUpGame } from '../hooks/useAcesUpGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { AcesUpResponse } from '../types/card';
import { AcesUpPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { ACESUP_HELP, parseAcesUpCommand } from '../utils/cli/commands/acesupCommands';
import { formatAcesUpState } from '../utils/cli/formatters/acesupFormatter';

type AcesUpGame = ReturnType<typeof useAcesUpGame>;
type AcesUpContext = GamePageContext<AcesUpResponse, AcesUpGame>;

function cardActionKey(card: { removable: boolean; movable: boolean }): string {
  if (card.removable && card.movable) return 'cardAction.removableAndMovable';
  if (card.removable) return 'cardAction.removable';
  if (card.movable) return 'cardAction.movable';
  return 'cardAction.unavailable';
}

/** Aces Up tutorial step definitions. */
const ACESUP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="acesup-columns"]',
    messageKey: 'tutorial.columns',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="acesup-columns"]',
    messageKey: 'tutorial.rules',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="acesup-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="acesup-stock"]',
    messageKey: 'tutorial.stock',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="acesup-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const COL_COUNT = 4;

/** Number of non-ace cards that must be discarded to win (52 - 4 aces). */
const DISCARD_GOAL = 48;

function HeaderExtra({ state, t }: AcesUpContext) {
  return (
    <>
      <span>
        {t('moveCount')}: {state.moveCount}
      </span>
      <span>
        {t('discard')}: {state.discardCount}
      </span>
    </>
  );
}

function LiveRegion({ state, t }: AcesUpContext) {
  const previousProgress = useRef<{ moveCount: number; discardCount: number; stockCount: number } | null>(null);
  const [progressAnnouncement, setProgressAnnouncement] = useState('');

  useEffect(() => {
    const currentProgress = {
      moveCount: state.moveCount,
      discardCount: state.discardCount,
      stockCount: state.stockCount,
    };
    const previous = previousProgress.current;
    previousProgress.current = currentProgress;
    if (
      previous &&
      (previous.moveCount !== currentProgress.moveCount ||
        previous.discardCount !== currentProgress.discardCount ||
        previous.stockCount !== currentProgress.stockCount)
    ) {
      setProgressAnnouncement(
        currentProgress.stockCount === 0 ? t('stockDepleted', currentProgress) : t('progressUpdate', currentProgress),
      );
    }
  }, [state, t]);

  return (
    <div data-testid="acesup-progress-live" className="sr-only" role="status" aria-live="polite" aria-atomic="true">
      {progressAnnouncement}
    </div>
  );
}

function Board({ game, state, t, busy, frontendHint }: AcesUpContext) {
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const isPlaying = state.phase === AcesUpPhase.PLAYING;
  const hint = game.hint;

  // Drag a movable top card onto an empty column to move it (the "Move [n]" buttons
  // remain as keyboard/tap-friendly fallbacks; D&D is additive, never required).
  const dnd = useSolitaireDragDrop<{ zone: string; col: number }>({
    onMove: (source) => game.handleMove(source.col),
    isPlaying,
    disabled: busy,
  });

  const cardGap = 6;
  const ROW_OVERLAP_RATIO = isMobile ? 0.45 : 0.4;
  const rowOverlap = cardHeight * ROW_OVERLAP_RATIO;

  return (
    <>
      {/* Tableau: 4 columns */}
      <div data-tutorial="acesup-columns" className="flex justify-center gap-3 mb-3">
        {Array.from({ length: COL_COUNT }, (_, colIdx) => {
          const col = state.columns[colIdx] ?? [];
          const topIdx = col.length - 1;
          const topCard = col[topIdx];
          const columnZone = { zone: 'column', col: colIdx };
          const isDropping = dnd.isDropTarget(columnZone);
          const isHinted = (hint?.type === 'remove' || hint?.type === 'move') && hint.col === colIdx;
          const stackHeight = col.length === 0 ? cardHeight : cardHeight + (col.length - 1) * (cardHeight - rowOverlap);
          return (
            <div key={`col-${colIdx.toString()}`} className="flex flex-col items-center">
              <div className="relative" style={{ width: cardWidth + cardGap, height: stackHeight }}>
                {col.length === 0 ? (
                  <DropZone
                    isDropTarget={isDropping}
                    onDragOver={dnd.handleDragOver(columnZone)}
                    onDrop={dnd.handleDrop(columnZone)}
                    onDragLeave={dnd.handleDragLeave}
                  >
                    <div
                      data-testid={`acesup-empty-${colIdx.toString()}`}
                      style={{ width: cardWidth, height: cardHeight }}
                      className={`rounded border-2 border-dashed ${
                        isDropping || isHinted ? 'border-ds-warning' : 'border-white/30'
                      } text-game-text-muted text-xs flex items-center justify-center`}
                    >
                      {t('empty')}
                    </div>
                  </DropZone>
                ) : (
                  col.map((c, rowIdx) => {
                    const top = rowIdx * (cardHeight - rowOverlap);
                    const isTop = rowIdx === topIdx;
                    if (isTop) {
                      const isCardDisabled = !isPlaying || busy || !c.removable;
                      const isDragSource = dnd.isDragSource(columnZone);
                      return (
                        <div key={`c-${colIdx.toString()}-${rowIdx.toString()}`} className="absolute" style={{ top }}>
                          <button
                            type="button"
                            onClick={() => game.handleRemove(colIdx)}
                            disabled={isCardDisabled}
                            aria-label={t('cardLabel', {
                              column: colIdx + 1,
                              card: cardAlt(c.card),
                              action: t(cardActionKey(c)),
                            })}
                            draggable={isPlaying && !busy && c.movable === true}
                            onDragStart={dnd.handleDragStart(columnZone)}
                            onDragEnd={dnd.handleDragEnd}
                            className={`relative p-0 ${c.movable ? 'border border-ds-info' : 'border-0'} bg-transparent rounded ${focusRingWhite} ${
                              isHinted ? 'ring-2 ring-ds-warning' : ''
                            } ${isCardDisabled ? 'cursor-not-allowed' : 'cursor-pointer'} ${
                              (!c.removable && !c.movable) || isDragSource ? 'opacity-50' : ''
                            } ${!c.removable && !c.movable ? 'grayscale' : ''}`}
                          >
                            {c.movable && (
                              <span
                                aria-hidden="true"
                                className="absolute top-1 right-1 z-10 rounded border border-ds-info bg-ds-surface px-1 text-xs leading-tight text-ds-info"
                              >
                                ↗
                              </span>
                            )}
                            <span aria-hidden="true">
                              <AnimatedCard card={c.card} width={cardWidth} />
                            </span>
                          </button>
                        </div>
                      );
                    }
                    return (
                      <div key={`c-${colIdx.toString()}-${rowIdx.toString()}`} className="absolute" style={{ top }}>
                        <AnimatedCard card={c.card} width={cardWidth} />
                      </div>
                    );
                  })
                )}
              </div>
              {/* Per-column move control */}
              <button
                type="button"
                className={`${btnSecondary} mt-1 text-xs`}
                onClick={() => game.handleMove(colIdx)}
                disabled={!isPlaying || busy || !topCard?.movable}
              >
                {t('move')} [{colIdx}]
              </button>
            </div>
          );
        })}
      </div>

      {/* Stock + discard pile */}
      <div className="flex gap-4 justify-center mb-3" data-tutorial="acesup-stock">
        <div className="text-center">
          <div className="text-game-text-muted text-xs mb-1">
            {t('stock')} ({state.stockCount})
          </div>
          {state.stockCount > 0 ? (
            <AnimatedCardBack
              width={cardWidth}
              onClick={isPlaying && !busy ? game.handleDraw : undefined}
              ariaLabel={t('draw')}
            />
          ) : (
            <div
              style={{ width: cardWidth, height: cardHeight }}
              className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
            >
              {t('empty')}
            </div>
          )}
        </div>
        <div className="text-center" data-testid="acesup-discard-pile">
          <div className="text-game-text-muted text-xs mb-1">
            {t('discardPile')} ({state.discardCount}/{DISCARD_GOAL})
          </div>
          {state.discardTop ? (
            <div data-testid="acesup-discard-top">
              <AnimatedCard
                key={`discard-${state.discardCount.toString()}`}
                card={state.discardTop}
                width={cardWidth}
              />
            </div>
          ) : (
            <div
              data-testid="acesup-discard-empty"
              style={{ width: cardWidth, height: cardHeight }}
              className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
            >
              {t('empty')}
            </div>
          )}
        </div>
      </div>

      {/* Hint display */}
      <div className="mb-2 flex justify-center" data-tutorial="acesup-hint-display">
        {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
            付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
            読み上げられないことがある (#5955, #6663)。 */}
        <div data-testid="acesup-hint-live" role="status" aria-live="polite">
          {hint && (
            <HintTooltip
              reason={hint.type === 'draw' ? t('hintReason.draw') : t(`hintReason.${hint.type}`, { col: hint.col })}
              confidence="strong"
            />
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy, confirmGiveUpAction }: AcesUpContext) {
  const isPlaying = state.phase === AcesUpPhase.PLAYING;
  if (!isPlaying) return null;

  const hasRemovable = state.columns.some((col) => col.length > 0 && col[col.length - 1]?.removable === true);

  return (
    <div data-tutorial="acesup-controls">
      <button type="button" className={btnPrimary} onClick={game.handleDraw} disabled={busy || state.stockCount === 0}>
        {t('draw')}
      </button>
      <button
        type="button"
        className={btnSecondary}
        onClick={game.handleRemoveAll}
        disabled={busy || !hasRemovable}
        data-testid="acesup-remove-all"
      >
        {t('removeAll')}
      </button>
      <button type="button" className={btnPrimary} onClick={game.handleUndo} disabled={busy || !state.canUndo}>
        {t('undo')}
      </button>
      {state.isStalemate && (
        <StalemateEscapeButton
          undoToEscape={state.undoToEscape ?? 0}
          onEscape={game.handleUndoEscape}
          disabled={busy}
        />
      )}
      <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy}>
        {t('hint')}
      </button>
      <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={busy}>
        {t('giveup')}
      </button>
    </div>
  );
}

const acesUpDefinition: GameDefinition<AcesUpResponse, AcesUpGame, Parameters<typeof acesupApi.exec>> = {
  key: 'acesup',
  useGame: useAcesUpGame,
  tutorial: ACESUP_TUTORIAL_STEPS,
  skeleton: { kind: 'tiered-rows', rows: [4, 4, 4, 4], columns: true },
  cli: () => ({
    parseCommand: parseAcesUpCommand,
    formatResponse: formatAcesUpState,
    helpText: ACESUP_HELP,
  }),
  busy: (game) => game.loading || game.isRemovingAll,
  phase: ({ state, t }) => {
    const isGameClear = state.phase === AcesUpPhase.GAME_CLEAR;
    const isGameOver = state.phase === AcesUpPhase.GAME_OVER;
    return {
      name: isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing'),
      ended: isGameClear || isGameOver,
      won: isGameClear,
    };
  },
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'd', action: game.handleDraw, label: 'draw' },
    { key: 'h', action: game.handleHint, label: 'hint' },
    // AcesUp always has handleGiveUp, so the fallback only satisfies the optional type.
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state }) => state.phase === AcesUpPhase.PLAYING,
  shortcutsTestId: 'aces-up-kbd-shortcuts',
  resetTutorialId: 'acesup-reset-button',
  HeaderExtra,
  LiveRegion,
  Board,
  FooterActions,
};

/** Renders the Aces Up game page with 4 columns, deal/remove/move controls. */
export const AcesUpPage = createGamePage(acesUpDefinition);
