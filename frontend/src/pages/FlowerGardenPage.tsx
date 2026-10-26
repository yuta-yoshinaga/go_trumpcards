import { Fragment, useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import type { FlowerGardenMoveZone, flowerGardenApi } from '../api/gameApi';
import { CardFan } from '../components/CardFan';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useBreakpoint, useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useFlowerGardenGame } from '../hooks/useFlowerGardenGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { DESKTOP_COMPACT_CARD_WIDTH } from '../styles/gameStyles';
import type { FlowerGardenResponse } from '../types/card';
import { FlowerGardenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { FLOWERGARDEN_HELP, parseFlowerGardenCommand } from '../utils/cli/commands/flowergardenCommands';
import { formatFlowerGardenState } from '../utils/cli/formatters/flowergardenFormatter';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;
// Keep in sync with the sm:gap-2 spacing on the bed row below.
const BED_GAP_PX = 8;
const FOUNDATION_SUIT_I18N_KEYS = [
  'common.suit.spade',
  'common.suit.club',
  'common.suit.heart',
  'common.suit.diamond',
] as const;

const FG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="fg-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fg-reserve"]',
    messageKey: 'tutorial.reserve',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fg-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fg-tableau"]',
    messageKey: 'tutorial.moves',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fg-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

const FLOWERGARDEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FlowerGardenPhase.PLAYING]: 'playing',
  [FlowerGardenPhase.GAME_CLEAR]: 'gameClear',
  [FlowerGardenPhase.GAME_OVER]: 'gameOver',
};

type FlowerGardenGame = ReturnType<typeof useFlowerGardenGame>;
type FlowerGardenView = ReturnType<typeof useFlowerGardenView>;
type FlowerGardenContext = GamePageContext<FlowerGardenResponse, FlowerGardenGame, FlowerGardenView>;

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation');
  if (zone === 'reserve') return t('frontendHint.reserve', { col });
  return t('frontendHint.tableau', { col });
}

function useFlowerGardenView({ game }: { game: FlowerGardenGame }) {
  const selectSourceHintId = useId();
  const moveCount = game.state?.moveCount;
  const previousMoveCount = useRef<number | null>(null);
  const [announcedMoveCount, setAnnouncedMoveCount] = useState<number | null>(null);

  useEffect(() => {
    if (previousMoveCount.current === null) {
      previousMoveCount.current = moveCount ?? null;
      return;
    }
    if (moveCount !== undefined && moveCount !== previousMoveCount.current) {
      previousMoveCount.current = moveCount;
      setAnnouncedMoveCount(moveCount);
    }
  }, [moveCount]);

  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    // Layout has 6 flower-bed fans side by side on mobile.
    const totalCols = 6;
    const colW = Math.floor((windowWidth - padX - (totalCols - 1) * gapPx) / totalCols);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);
  const reserveDims = isMobile
    ? { cw: Math.min(dims.cw, 32), ch: Math.round(Math.min(dims.cw, 32) * 1.5) }
    : { cw: dims.cw, ch: dims.ch };
  const foundationWidth = isLargeDesktop ? DESKTOP_COMPACT_CARD_WIDTH : dims.cw;
  const foundationHeight = isLargeDesktop ? Math.round(DESKTOP_COMPACT_CARD_WIDTH * 1.5) : dims.ch;

  const isPlayingForKbd = game.state?.phase === FlowerGardenPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: FlowerGardenMoveZone, target: FlowerGardenMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<FlowerGardenMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: game.loading,
  });

  const phaseNames = usePhaseNames('flowergarden', FLOWERGARDEN_PHASE_KEYS);

  return {
    selectSourceHintId,
    announcedMoveCount,
    dims,
    reserveDims,
    foundationWidth,
    foundationHeight,
    isLargeDesktop,
    dnd,
    phaseNames,
  };
}

function LiveRegion({ t, tc, view }: FlowerGardenContext) {
  return (
    <>
      <div data-testid="fg-move-count-live" className="sr-only" role="status" aria-live="polite">
        {view.announcedMoveCount === null ? '' : t('moveCountAnnouncement', { count: view.announcedMoveCount })}
      </div>
      <span id={view.selectSourceHintId} className="sr-only">
        {tc('label.selectSourceFirst')}
      </span>
    </>
  );
}

function HeaderExtra({ state, t }: FlowerGardenContext) {
  return (
    <span className="text-sm text-ds-text-muted">
      {t('moveCount')}: {state.moveCount}
    </span>
  );
}

function Board({ game, state, t, tc, view, frontendHint }: FlowerGardenContext) {
  const { selectSourceHintId, dims, reserveDims, foundationWidth, foundationHeight, isLargeDesktop, dnd } = view;
  const { loading, selectedSource, hint, isAutoCompleting } = game;

  const isPlaying = state.phase === FlowerGardenPhase.PLAYING;
  const isGameOver = state.phase === FlowerGardenPhase.GAME_OVER;
  // How many of the 52 cards reached a foundation — only needed on game over, so skip the
  // reduce during normal (frequently re-rendered) play.
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // リングは**置ける先だけ**に付ける。
  const legalTargetKey =
    selectedSource?.col === undefined
      ? undefined
      : `${selectedSource.zone === 'reserve' ? 'reserve' : 'tableau'}-${selectedSource.col}`;
  const selectedLegalTargets = legalTargetKey ? state.legalTargets[legalTargetKey] : undefined;
  const legalTargets = {
    tableau: new Set(selectedLegalTargets?.tableau ?? []),
    foundation: new Set(selectedLegalTargets?.foundation ?? []),
  };

  const BedsLayout = isLargeDesktop ? 'div' : Fragment;
  const bedsLayoutProps = isLargeDesktop ? { className: 'flex flex-wrap justify-center items-start gap-4' } : {};

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx];
    const tableauColZone: FlowerGardenMoveZone = { zone: 'tableau', col: colIdx };
    return (
      <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
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
                disabled={!isPlaying || loading}
                aria-disabled={!selectedSource || undefined}
                aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                style={{ height: dims.ch }}
                data-target-candidate={legalTargets.tableau.has(colIdx) || undefined}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite} ${
                  legalTargets.tableau.has(colIdx) ? 'ring-2 ring-ds-success' : ''
                }`}
              >
                {t('empty')}
              </button>
            ) : (
              col.map((tc, cardIdx) => {
                const cardZone: FlowerGardenMoveZone = {
                  zone: 'tableau',
                  col: colIdx,
                  cardIndex: cardIdx,
                };
                const isTop = cardIdx === col.length - 1;
                const isTargetCandidate =
                  isTop && legalTargets.tableau.has(colIdx) && !isSourceSelected('tableau', colIdx, cardIdx);
                return (
                  <div
                    key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    {tc.card ? (
                      <button
                        type="button"
                        onClick={() => {
                          if (selectedSource) {
                            game.handleSelectTarget(tableauColZone);
                          } else if (isTop) {
                            game.handleSelectSource(cardZone);
                          }
                        }}
                        disabled={!isPlaying || loading || (!isTop && !selectedSource)}
                        aria-label={t('tableauCardAriaLabel', {
                          card: cardAlt(tc.card),
                          col: colIdx,
                          pos: cardIdx,
                        })}
                        aria-pressed={isSourceSelected('tableau', colIdx, cardIdx)}
                        draggable={isPlaying && !loading && isTop}
                        onDragStart={dnd.handleDragStart(cardZone)}
                        onDragEnd={dnd.handleDragEnd}
                        data-target-candidate={isTargetCandidate || undefined}
                        className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${isTop ? 'cursor-pointer' : 'cursor-default'} ${isSourceSelected('tableau', colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : ''} ${isTargetCandidate ? 'ring-2 ring-ds-success' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                      >
                        <AnimatedCard
                          card={tc.card}
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

  const renderReserveCell = (cellIdx: number) => {
    const reserveCard = state.reserve[cellIdx];
    const reserveZone: FlowerGardenMoveZone = { zone: 'reserve', col: cellIdx };
    // Number each bouquet slot 0..15 to match formatHintZone's raw reserve col
    // (and the CUI's [r0]..[r15]), so players can map hint text to a card.
    return (
      <div key={`r-${cellIdx.toString()}`} className="flex flex-col items-center">
        {reserveCard ? (
          <button
            type="button"
            onClick={() => game.handleSelectSource(reserveZone)}
            disabled={!isPlaying || loading}
            aria-label={t('reserveCardAriaLabel', { card: cardAlt(reserveCard), idx: cellIdx })}
            aria-pressed={isSourceSelected('reserve', cellIdx)}
            draggable={isPlaying && !loading}
            onDragStart={dnd.handleDragStart(reserveZone)}
            onDragEnd={dnd.handleDragEnd}
            className={`p-0 border-0 bg-transparent rounded cursor-pointer ${focusRingWhite} ${isSourceSelected('reserve', cellIdx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(reserveZone) ? 'opacity-50' : ''}`}
          >
            <AnimatedCard card={reserveCard} width={reserveDims.cw} draggable={false} />
          </button>
        ) : (
          <div
            role="img"
            aria-label={t('emptyReserveSlot', { idx: cellIdx })}
            className="rounded border border-dashed border-white/10"
            style={{ width: reserveDims.cw, height: reserveDims.ch }}
          />
        )}
        <span className="text-xs text-ds-text-muted mt-0.5" aria-hidden="true">
          #{cellIdx}
        </span>
      </div>
    );
  };

  const renderFoundations = () =>
    state.foundation.map((pile, idx) => {
      const foundationZone: FlowerGardenMoveZone = { zone: 'foundation', col: idx };
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
                disabled={!isPlaying || loading || isAutoCompleting}
                aria-disabled={!selectedSource || undefined}
                aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                aria-label={t('foundationAriaLabel', {
                  suit: tc(FOUNDATION_SUIT_I18N_KEYS[idx]),
                  card: cardAlt(pile[pile.length - 1]),
                  count: pile.length,
                })}
                data-target-candidate={legalTargets.foundation.has(idx) || undefined}
                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${legalTargets.foundation.has(idx) ? 'ring-2 ring-ds-success' : ''}`}
              >
                <AnimatedCard
                  card={pile[pile.length - 1]}
                  width={foundationWidth}
                  draggable={false}
                  dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                />
              </button>
            ) : (
              <button
                type="button"
                onClick={() => game.handleSelectTarget(foundationZone)}
                disabled={!isPlaying || loading}
                aria-disabled={!selectedSource || undefined}
                aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                aria-label={t('emptyFoundationAriaLabel', { suit: tc(FOUNDATION_SUIT_I18N_KEYS[idx]) })}
                style={{ width: foundationWidth, height: foundationHeight }}
                data-target-candidate={legalTargets.foundation.has(idx) || undefined}
                className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${legalTargets.foundation.has(idx) ? 'ring-2 ring-ds-success' : ''}`}
              >
                A
              </button>
            )}
          </DropZone>
          <div
            data-testid={`fg-foundation-progress-${idx.toString()}`}
            className={`text-xs mt-1 tabular-nums ${pile.length === 13 ? 'text-ds-success' : 'text-game-text-muted'}`}
          >
            {pile.length === 13 && <span aria-hidden="true">✓ </span>}
            {t('foundationProgress', { count: pile.length })}
          </div>
        </div>
      );
    });

  return (
    <>
      <div className="flex flex-wrap gap-2 sm:gap-3 items-start justify-center mb-3">
        <div
          className={isLargeDesktop ? 'flex flex-col gap-1 w-full' : 'flex flex-col gap-1'}
          data-tutorial="fg-reserve"
        >
          <span className="text-game-text-muted text-xs">{t('reserve')}</span>
          <span className="text-game-text-muted text-xs">{t('reserveSelectionHint')}</span>
          {/* 16 bouquet cards laid out as a grid (4 cols on mobile, 8 on sm+) so every
              slot stays clearly visible instead of cramming into one wrapping row (#3283). */}
          {/* Add the card button's 6px border to each card's fan width. */}
          {isLargeDesktop ? (
            <CardFan itemWidth={dims.cw + 6} className="w-full" data-testid="fg-bouquet-fan">
              {state.reserve.map((_, idx) => renderReserveCell(idx))}
            </CardFan>
          ) : (
            <div className="grid grid-cols-4 sm:grid-cols-8 gap-1 sm:gap-2 justify-items-center">
              {state.reserve.map((_, idx) => renderReserveCell(idx))}
            </div>
          )}
        </div>

        {!isLargeDesktop && (
          <div className="flex items-start gap-1 sm:gap-2" data-tutorial="fg-foundation">
            {renderFoundations()}
          </div>
        )}
      </div>

      <details className="mb-1 text-center text-ds-text-muted text-xs">
        <summary className="cursor-pointer">{t('rules')}</summary>
        <p data-testid="fg-rules-note">{t('rulesNote')}</p>
      </details>

      <BedsLayout {...bedsLayoutProps}>
        <div
          className="flex gap-1 sm:gap-2"
          data-tutorial="fg-tableau"
          style={isLargeDesktop ? { width: 6 * dims.cw + 5 * BED_GAP_PX } : undefined}
        >
          {[0, 1, 2, 3, 4, 5].map(renderTableauColumn)}
        </div>
        {isLargeDesktop && (
          <div className="grid grid-cols-2 gap-1" data-tutorial="fg-foundation">
            {renderFoundations()}
          </div>
        )}
      </BedsLayout>

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-tutorial="fg-hint-display" data-testid="fg-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2 mt-3">
            {t('hintAvailable', { card: cardAlt(hint.card) })}: {formatHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
            {formatHintZone(t, hint.toZone, hint.toCol)}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {isGameOver && (
        <p data-testid="fg-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
          {t('gameOverSummary', {
            count: foundationCount,
            percent: Math.round((foundationCount / 52) * 100),
          })}
        </p>
      )}
    </>
  );
}

function FooterActions({ game, state, t, confirmGiveUpAction }: FlowerGardenContext) {
  const isPlaying = state.phase === FlowerGardenPhase.PLAYING;
  if (!isPlaying) return null;
  // Auto-complete becomes useful once a foundation has built past its ace, so
  // pulse the button only then (mirrors Crescent / Spiderette).
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);

  return (
    <div data-tutorial="fg-controls">
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleUndo}
        disabled={game.loading || game.isAutoCompleting || !state.canUndo}
      >
        {t('undo')}
      </button>
      {state.isStalemate && (
        <StalemateEscapeButton
          undoToEscape={state.undoToEscape ?? 0}
          onEscape={game.handleUndoEscape}
          disabled={game.loading || game.isAutoCompleting}
        />
      )}
      <button
        type="button"
        className={btnSuccess}
        onClick={game.handleHint}
        disabled={game.loading || game.isAutoCompleting}
      >
        {t('hint')}
      </button>
      <button
        type="button"
        className={`${btnSuccess}${autoCompleteReady && !game.loading && !game.isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
        onClick={game.handleAutoComplete}
        disabled={game.loading || game.isAutoCompleting || !autoCompleteReady}
        data-testid="autocomplete-button"
        title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
      >
        {t('autoComplete')}
      </button>
      <button
        type="button"
        className={btnDanger}
        onClick={confirmGiveUpAction}
        disabled={game.loading || game.isAutoCompleting}
      >
        {t('giveup')}
      </button>
    </div>
  );
}

const flowerGardenDefinition: GameDefinition<
  FlowerGardenResponse,
  FlowerGardenGame,
  Parameters<typeof flowerGardenApi.exec>,
  FlowerGardenView
> = {
  key: 'flowergarden',
  useGame: useFlowerGardenGame,
  useView: useFlowerGardenView,
  tutorial: FG_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 4, tableau: 6 },
  cli: () => ({
    parseCommand: parseFlowerGardenCommand,
    formatResponse: formatFlowerGardenState,
    helpText: FLOWERGARDEN_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === FlowerGardenPhase.GAME_CLEAR || state.phase === FlowerGardenPhase.GAME_OVER,
    won: state.phase === FlowerGardenPhase.GAME_CLEAR,
  }),
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state }) => state.phase === FlowerGardenPhase.PLAYING,
  shortcutsTestId: 'flower-garden-kbd-shortcuts',
  resetTutorialId: 'fg-reset-button',
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  landscapeBanner: true,
  landscapeBannerInCliMode: true,
  HeaderExtra,
  LiveRegion,
  Board,
  FooterActions,
};

/** Renders the Flower Garden solitaire game page with 6 flower-bed fans, a 16-card bouquet reserve, and 4 foundations. */
export const FlowerGardenPage = createGamePage(flowerGardenDefinition);
