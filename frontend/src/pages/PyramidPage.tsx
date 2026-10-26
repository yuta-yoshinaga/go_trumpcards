import type { TFunction } from 'i18next';
import { useEffect, useMemo, useRef, useState } from 'react';
import type { pyramidApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import {
  useBreakpoint,
  useCardDimensions,
  useDenseLargeDesktopCardSize,
  useWindowWidth,
} from '../hooks/useCardDimensions';
import { usePyramidGame } from '../hooks/usePyramidGame';
import { pyramidWinRate, usePyramidStats } from '../hooks/usePyramidStats';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { PyramidResponse } from '../types/card';
import { PyramidPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PYRAMID_HELP, parsePyramidCommand } from '../utils/cli/commands/pyramidCommands';
import { formatPyramidState } from '../utils/cli/formatters/pyramidFormatter';

/** Pyramid tutorial step definitions. */
const PY_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="py-pyramid"]',
    messageKey: 'tutorial.pyramid',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Play-area clientHeight measured at 1280x800; this keeps the row step compact on taller screens. */
const PYRAMID_PLAY_AREA_HEIGHT = 583;
const PYRAMID_VERTICAL_PADDING = 8;

/** Calculates the compact row step, bounded by the minimum tap spacing and scaled develop step. */
export function compactPyramidRowStep(rows: number, cardHeight: number, scaledDevelopStep: number): number {
  const upperBound = Math.max(56, scaledDevelopStep);
  if (rows <= 1) return upperBound;
  const availableStep = Math.floor((PYRAMID_PLAY_AREA_HEIGHT - cardHeight - PYRAMID_VERTICAL_PADDING) / (rows - 1));
  return Math.max(56, Math.min(scaledDevelopStep, availableStep));
}

type PyramidGame = ReturnType<typeof usePyramidGame>;

function usePyramidView({ game, t }: { game: PyramidGame; t: TFunction }) {
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const denseCardSize = useDenseLargeDesktopCardSize();
  const windowWidth = useWindowWidth();

  const phaseName = useMemo(() => {
    const p = game.state?.phase;
    if (p === PyramidPhase.GAME_CLEAR) return t('phase.gameClear');
    if (p === PyramidPhase.GAME_OVER) return t('phase.gameOver');
    return t('phase.playing');
  }, [game.state?.phase, t]);

  // Best-record persistence in localStorage (issue #3083).
  const { stats, recordResult } = usePyramidStats();
  const [newBestMoves, setNewBestMoves] = useState(false);
  // Guard so each finished game is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  const endPhase = game.state?.phase;
  const endMoveCount = game.state?.moveCount ?? 0;
  useEffect(() => {
    const ended = endPhase === PyramidPhase.GAME_CLEAR || endPhase === PyramidPhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    setNewBestMoves(recordResult({ won: endPhase === PyramidPhase.GAME_CLEAR, moves: endMoveCount }));
  }, [endPhase, endMoveCount, recordResult]);

  return {
    cardHeight,
    cardWidth,
    isMobile,
    isLargeDesktop,
    denseCardSize,
    windowWidth,
    phaseName,
    stats,
    newBestMoves,
  };
}

type PyramidView = ReturnType<typeof usePyramidView>;
type PyramidContext = GamePageContext<PyramidResponse, PyramidGame, PyramidView>;

function HeaderExtra({ state, t, view }: PyramidContext) {
  return (
    <>
      <span>
        {t('moveCount')}: {state.moveCount}
      </span>
      {view.stats.fewestMoves !== null && (
        <span data-testid="py-best-moves">
          {t('bestMoves')}: {t('bestMovesValue', { moves: view.stats.fewestMoves })}
        </span>
      )}
    </>
  );
}

function Board({ game, state, t, view, frontendHint }: PyramidContext) {
  const isPlaying = state.phase === PyramidPhase.PLAYING;
  const isGameClear = state.phase === PyramidPhase.GAME_CLEAR;

  const isSelected = (zone: 'pyramid' | 'waste', row?: number, col?: number) =>
    game.selectedCard !== null &&
    game.selectedCard.zone === zone &&
    game.selectedCard.row === row &&
    game.selectedCard.col === col;

  // When the player has selected the first card, compute the partner value (13 - value).
  // Kings (13) need no partner; ace (1) wants Q (12), etc.
  const selectedValue = (() => {
    if (!game.selectedCard) return null;
    if (game.selectedCard.zone === 'waste') {
      return state.waste.length > 0 ? state.waste[state.waste.length - 1].value : null;
    }
    if (game.selectedCard.row === undefined || game.selectedCard.col === undefined) return null;
    const pc = state.pyramid[game.selectedCard.row]?.[game.selectedCard.col];
    return pc?.card?.value ?? null;
  })();
  const partnerValue = selectedValue !== null && selectedValue < 13 ? 13 - selectedValue : null;
  const wasteTopCard = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const isWastePairCandidate =
    partnerValue !== null && !isSelected('waste') && wasteTopCard !== null && wasteTopCard.value === partnerValue;
  // The waste top is always exposed; a King there is removable alone (issue #3082).
  const isWasteExposedKing = wasteTopCard !== null && wasteTopCard.value === 13 && !isSelected('waste');

  // Calculate pyramid layout dimensions
  const maxCols = 7; // bottom row has 7 cards
  const cardGap = 4;
  /** Fraction of card height used for vertical overlap between rows (less on mobile for bigger tap targets) */
  const ROW_OVERLAP_RATIO = view.isMobile ? 0.3 : 0.35;
  const pyramidCardWidth = view.denseCardSize.width;
  const effectiveCardHeight = view.isLargeDesktop ? view.denseCardSize.height : view.cardHeight;
  const currentRowStep = view.cardHeight - view.cardHeight * ROW_OVERLAP_RATIO;
  const scaledDevelopStep = currentRowStep * (pyramidCardWidth / view.cardWidth);
  const compactRowStep = compactPyramidRowStep(state.pyramid.length, effectiveCardHeight, scaledDevelopStep);
  const rowStep = view.isLargeDesktop ? compactRowStep : view.cardHeight * (1 - ROW_OVERLAP_RATIO);
  const rowOverlap = effectiveCardHeight - rowStep;
  // px-4 on the scrollable container = 16px * 2 = 32px total horizontal padding
  const CONTAINER_PADDING = 32;
  let effectiveCardWidth: number = view.cardWidth;
  if (view.isLargeDesktop) {
    effectiveCardWidth = pyramidCardWidth;
  } else if (view.isMobile) {
    effectiveCardWidth = Math.floor((view.windowWidth - CONTAINER_PADDING - cardGap * (maxCols - 1)) / maxCols);
  }
  const pyramidWidth = maxCols * (effectiveCardWidth + cardGap) - cardGap;

  return (
    <>
      {/* Pyramid */}
      <div data-tutorial="py-pyramid" className="flex flex-col items-center mb-3">
        {state.pyramid.map((row, rowIdx) => {
          const cols = row.length;
          const rowWidth = cols * (effectiveCardWidth + cardGap) - cardGap;
          const offsetX = (pyramidWidth - rowWidth) / 2;
          return (
            <div
              key={`row-${rowIdx.toString()}`}
              className="relative"
              style={{
                height: rowIdx < state.pyramid.length - 1 ? effectiveCardHeight - rowOverlap : effectiveCardHeight,
                width: pyramidWidth,
              }}
            >
              {row.map((pc, colIdx) => {
                const left = offsetX + colIdx * (effectiveCardWidth + cardGap);
                if (pc.removed) {
                  return (
                    <div
                      key={`pc-${rowIdx.toString()}-${colIdx.toString()}`}
                      className="absolute"
                      style={{ left, width: effectiveCardWidth, height: effectiveCardHeight }}
                    />
                  );
                }
                if (!pc.card) return null;
                const exposed = pc.exposed;
                const isPairCandidate =
                  partnerValue !== null &&
                  exposed &&
                  pc.card.value === partnerValue &&
                  !isSelected('pyramid', rowIdx, colIdx);
                // A King (13) is removable alone once exposed — surface that
                // affordance so players don't hunt for a nonexistent partner
                // (issue #3082). Selected/hint states take visual precedence.
                const isExposedKing = exposed && pc.card.value === 13 && !isSelected('pyramid', rowIdx, colIdx);
                // Server hint targets (-1 sentinels for king/waste never match a cell).
                const isHintTarget =
                  !!game.hint &&
                  ((game.hint.row1 === rowIdx && game.hint.col1 === colIdx) ||
                    (game.hint.row2 === rowIdx && game.hint.col2 === colIdx));
                // Convey the card's actionability (blocked / selected / pair
                // candidate) to assistive tech, not just via disabled/color.
                const cellSelected = isSelected('pyramid', rowIdx, colIdx);
                const statusSuffix = !exposed
                  ? ` ${t('a11y.blocked')}`
                  : cellSelected
                    ? ` ${t('a11y.selected')}`
                    : isPairCandidate
                      ? ` ${t('a11y.pairCandidate')}`
                      : isExposedKing
                        ? ` ${t('a11y.kingRemovable')}`
                        : '';
                return (
                  <div key={`pc-${rowIdx.toString()}-${colIdx.toString()}`} className="absolute" style={{ left }}>
                    <button
                      type="button"
                      onClick={() => {
                        if (!exposed || !pc.card) return;
                        game.handleSelectCard({ zone: 'pyramid', row: rowIdx, col: colIdx }, pc.card.value);
                      }}
                      disabled={!isPlaying || game.loading || !exposed}
                      aria-label={`${cardAlt(pc.card)} ${t('a11y.position', { row: rowIdx, col: colIdx })}${statusSuffix}`}
                      aria-pressed={cellSelected}
                      data-pair-candidate={isPairCandidate ? 'true' : undefined}
                      data-king-removable={isExposedKing ? 'true' : undefined}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                        isSelected('pyramid', rowIdx, colIdx)
                          ? 'ring-2 ring-ds-warning'
                          : isHintTarget
                            ? 'ring-2 ring-ds-warning animate-pulse'
                            : isPairCandidate
                              ? 'ring-2 ring-ds-success animate-pulse'
                              : isExposedKing
                                ? 'ring-2 ring-ds-success'
                                : ''
                      } ${!exposed ? 'opacity-60' : ''}`}
                    >
                      <AnimatedCard card={pc.card} width={effectiveCardWidth} />
                    </button>
                  </div>
                );
              })}
            </div>
          );
        })}
      </div>

      {/* Stock + Waste */}
      <div className="flex gap-4 justify-center mb-3" data-tutorial="py-stock-waste">
        {/* Stock */}
        <div className="text-center">
          <div className="text-game-text-muted text-xs mb-1">
            {t('stock')} ({state.stockCount})
          </div>
          {state.stockCount > 0 ? (
            <AnimatedCardBack
              width={effectiveCardWidth}
              onClick={isPlaying ? game.handleDraw : undefined}
              ariaLabel={t('drawStock', { count: state.stockCount })}
            />
          ) : (
            <div
              style={{ width: effectiveCardWidth, height: effectiveCardHeight }}
              className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
            >
              {t('empty')}
            </div>
          )}
          {/* **配り直しは無い。** Draw は山札を引き切ると二度と引けないのに、
              空表示は「なし」としか言わない。標準のピラミッド (3回配り直し可) を
              知っているプレイヤーほど手詰まりの原因を誤解する (#5510)。 */}
          {state.stockCount === 0 && (
            <div className="text-game-text-muted text-[10px] mt-1 max-w-[120px]" data-testid="py-no-redeal">
              {t('noRedeal')}
            </div>
          )}
        </div>

        {/* Waste */}
        <div className="text-center">
          <div className="text-game-text-muted text-xs mb-1">
            {t('waste')} <span id="pyramid-waste-count">{t('wasteCount', { count: state.waste.length })}</span>
          </div>
          {wasteTopCard ? (
            <button
              type="button"
              onClick={() => game.handleSelectCard({ zone: 'waste' }, wasteTopCard.value)}
              disabled={!isPlaying || game.loading}
              aria-label={`${cardAlt(wasteTopCard)}${isWastePairCandidate ? ` ${t('a11y.pairCandidate')}` : ''}${isWasteExposedKing ? ` ${t('a11y.kingRemovable')}` : ''}`}
              aria-describedby="pyramid-waste-count"
              aria-pressed={isSelected('waste')}
              data-pair-candidate={isWastePairCandidate ? 'true' : undefined}
              data-king-removable={isWasteExposedKing ? 'true' : undefined}
              className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                isSelected('waste') ? 'ring-2 ring-ds-warning' : ''
              } ${isWastePairCandidate ? 'ring-2 ring-ds-success animate-pulse' : isWasteExposedKing ? 'ring-2 ring-ds-success' : ''}`}
            >
              <AnimatedCard card={wasteTopCard} width={effectiveCardWidth} />
            </button>
          ) : (
            <div
              style={{ width: effectiveCardWidth, height: effectiveCardHeight }}
              className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
            >
              {t('empty')}
            </div>
          )}
        </div>
      </div>

      {/* Hint display */}
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-tutorial="py-hint-display" data-testid="py-hint-live" role="status" aria-live="polite">
        {game.hint && (
          <div className="text-ds-warning text-sm mb-2 text-center">
            {t(`hintAnnouncement.${game.hint.type}`, {
              row1: game.hint.row1,
              col1: game.hint.col1,
              row2: game.hint.row2,
              col2: game.hint.col2,
            })}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {/* New fewest-moves record badge on the clear screen (#3083). */}
      {isGameClear && view.newBestMoves && (
        <div
          data-testid="py-best-badge"
          role="status"
          className="text-center text-ds-success font-semibold text-sm mb-2"
        >
          {t('newBestMoves', { moves: state.moveCount })}
        </div>
      )}

      {/* Best-record panel: fewest moves + clear tally (#3083). */}
      <div data-testid="py-stats-panel" className="text-game-text-muted text-xs text-center mb-2">
        {t('bestMoves')}:{' '}
        {view.stats.fewestMoves !== null ? t('bestMovesValue', { moves: view.stats.fewestMoves }) : '—'}
        {view.stats.plays > 0 && (
          <>
            {' · '}
            {t('clears', { wins: view.stats.wins, plays: view.stats.plays })}
            {' · '}
            {t('winRate', { rate: pyramidWinRate(view.stats) })}
          </>
        )}
      </div>
    </>
  );
}

function FooterActions({ game, state, t, busy, confirmGiveUpAction }: PyramidContext) {
  const isPlaying = state.phase === PyramidPhase.PLAYING;
  if (!isPlaying) return null;
  return (
    <div data-tutorial="py-controls">
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleDraw}
        disabled={busy}
        aria-label={t('drawStock', { count: state.stockCount })}
      >
        {t('draw')}
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

const pyramidDefinition: GameDefinition<
  PyramidResponse,
  PyramidGame,
  Parameters<typeof pyramidApi.exec>,
  PyramidView
> = {
  key: 'pyramid',
  useGame: usePyramidGame,
  useView: usePyramidView,
  tutorial: PY_TUTORIAL_STEPS,
  skeleton: { kind: 'tiered-rows', rows: [1, 2, 3, 4], stockWaste: true },
  cli: () => ({
    parseCommand: parsePyramidCommand,
    formatResponse: formatPyramidState,
    helpText: PYRAMID_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameClear = state.phase === PyramidPhase.GAME_CLEAR;
    const isGameOver = state.phase === PyramidPhase.GAME_OVER;
    return {
      name: view.phaseName,
      ended: isGameClear || isGameOver,
      won: isGameClear,
    };
  },
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'd', action: game.handleDraw, label: 'draw' },
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state }) => state.phase === PyramidPhase.PLAYING,
  shortcutsTestId: 'pyramid-kbd-shortcuts',
  resetTutorialId: 'py-reset-button',
  landscapeBanner: true,
  contentClassName: 'pt-3 px-4 lg:px-8 lg:flex lg:flex-wrap lg:justify-center lg:items-start lg:gap-6',
  HeaderExtra,
  Board,
  FooterActions,
};

/** Renders the Pyramid Solitaire game page with pyramid, stock/waste, and controls. */
export const PyramidPage = createGamePage(pyramidDefinition);
