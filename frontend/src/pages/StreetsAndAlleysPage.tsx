import type { TFunction } from 'i18next';
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import type { StreetsAndAlleysMoveZone, streetsAndAlleysApi } from '../api/gameApi';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useDestinationPreview } from '../hooks/useDestinationPreview';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useStreetsAndAlleysGame } from '../hooks/useStreetsAndAlleysGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { StreetsAndAlleysResponse } from '../types/card';
import { StreetsAndAlleysPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseStreetsAndAlleysCommand, STREETSANDALLEYS_HELP } from '../utils/cli/commands/streetsandalleysCommands';
import { formatStreetsAndAlleysState } from '../utils/cli/formatters/streetsandalleysFormatter';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

const SA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sa-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sa-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sa-moves"]',
    messageKey: 'tutorial.moves',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sa-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [StreetsAndAlleysPhase.PLAYING]: 'playing',
  [StreetsAndAlleysPhase.GAME_CLEAR]: 'gameClear',
  [StreetsAndAlleysPhase.GAME_OVER]: 'gameOver',
};

type StreetsAndAlleysGame = ReturnType<typeof useStreetsAndAlleysGame>;
type StreetsAndAlleysView = ReturnType<typeof useStreetsAndAlleysView>;
type StreetsAndAlleysContext = GamePageContext<StreetsAndAlleysResponse, StreetsAndAlleysGame, StreetsAndAlleysView>;

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation');
  return t('frontendHint.tableau', { col });
}

function useStreetsAndAlleysView({ game, t }: { game: StreetsAndAlleysGame; t: TFunction; tc: TFunction }) {
  const { state, loading, selectedSource } = game;
  const previousMoveCount = useRef(state?.moveCount);
  const [announcedMoveCount, setAnnouncedMoveCount] = useState<number | null>(null);
  useEffect(() => {
    if (!state) return;
    if (previousMoveCount.current !== undefined && previousMoveCount.current !== state.moveCount) {
      setAnnouncedMoveCount(state.moveCount);
    }
    previousMoveCount.current = state.moveCount;
  }, [state]);
  const selectSourceHintId = useId();

  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    // Layout has 9 visual columns: 4 left tableau + 1 foundation strip + 4 right tableau.
    const totalCols = 9;
    const colW = Math.floor((windowWidth - padX - (totalCols - 1) * gapPx) / totalCols);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === StreetsAndAlleysPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: StreetsAndAlleysMoveZone, target: StreetsAndAlleysMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<StreetsAndAlleysMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  // Keep this hook before the early return so its hook order stays stable while
  // the game state is loading.
  const preview = useDestinationPreview<StreetsAndAlleysMoveZone>(selectedSource);

  // 選択中、またはホバー/フォーカス中の札そのもの。列の末尾しか動かせない。
  const previewSource = preview.source;
  const legalForSource =
    state && previewSource?.zone === 'tableau' && previewSource.col !== undefined
      ? state.legalTargets[`tableau-${previewSource.col}`]
      : undefined;
  // 判定は StreetsAndAlleys.go の canPlaceOn* を通してサーバーが計算する (ADR-0045)。
  const legalTargets = {
    tableau: new Set(legalForSource?.tableau ?? []),
    foundation: new Set(legalForSource?.foundation ?? []),
  };
  const selectedTargetCount = selectedSource ? legalTargets.tableau.size + legalTargets.foundation.size : null;

  const destinationStatus = useCallback(
    (isLegal: boolean) => t(isLegal ? 'legalDestinationStatus' : 'illegalDestinationStatus'),
    [t],
  );
  /** Ring for a legal destination: softer while it is only a hover preview. */
  const targetRing = preview.isPreview
    ? ' rounded ring-1 ring-ds-info/70 motion-safe:hover:ring-2 focus:ring-2'
    : ' rounded ring-1 ring-ds-info motion-safe:hover:ring-2 focus:ring-2';

  const phaseNames = usePhaseNames('streetsandalleys', SA_PHASE_KEYS);

  return {
    announcedMoveCount,
    selectSourceHintId,
    dims,
    isMobile,
    dnd,
    preview,
    legalTargets,
    selectedTargetCount,
    destinationStatus,
    targetRing,
    phaseNames,
  };
}

function LiveRegion({ t, tc, view }: StreetsAndAlleysContext) {
  return (
    <>
      <LandscapeBanner message={t('landscapeBanner')} />
      <span id={view.selectSourceHintId} className="sr-only">
        {tc('label.selectSourceFirst')}
      </span>
      <div role="status" aria-live="polite" className="sr-only" data-testid="sa-destination-announcement">
        {view.selectedTargetCount === null
          ? ''
          : t(view.selectedTargetCount === 0 ? 'noLegalDestinations' : 'legalDestinationCount', {
              count: view.selectedTargetCount,
            })}
      </div>
      <div role="status" aria-live="polite" className="sr-only" data-testid="sa-move-count-announcement">
        {view.announcedMoveCount === null ? '' : t('moveCountAnnouncement', { count: view.announcedMoveCount })}
      </div>
    </>
  );
}

function HeaderExtra({ state, t }: StreetsAndAlleysContext) {
  return (
    <span className="text-sm text-ds-text-muted">
      {t('moveCount')}: {state.moveCount}
    </span>
  );
}

function Board({ game, state, t, frontendHint, view }: StreetsAndAlleysContext) {
  const { loading, selectedSource, hint, isAutoCompleting } = game;
  const { dims, isMobile, dnd, preview, legalTargets, destinationStatus, targetRing, selectSourceHintId } = view;

  const isPlaying = state.phase === StreetsAndAlleysPhase.PLAYING;
  const isGameOver = state.phase === StreetsAndAlleysPhase.GAME_OVER;
  // How many of the 52 cards reached a foundation — only needed on game over, so skip the
  // reduce during normal (frequently re-rendered) play.
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  const hasSelectedSource = selectedSource !== null;

  const attachScrollTestId = useCallback((el: HTMLDivElement | null) => {
    if (el?.parentElement) {
      el.parentElement.setAttribute('data-testid', 'sa-board-scroll');
    }
  }, []);

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx];
    const tableauColZone: StreetsAndAlleysMoveZone = { zone: 'tableau', col: colIdx };
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
                aria-label={hasSelectedSource ? t('emptyTableauAriaLabel', { col: colIdx }) : undefined}
                aria-description={hasSelectedSource ? destinationStatus(legalTargets.tableau.has(colIdx)) : undefined}
                style={{ height: dims.ch }}
                data-target-candidate={legalTargets.tableau.has(colIdx) || undefined}
                data-preview-target={legalTargets.tableau.has(colIdx) && preview.isPreview ? 'true' : undefined}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}${
                  legalTargets.tableau.has(colIdx) ? targetRing : ''
                }`}
              >
                {t('empty')}
              </button>
            ) : (
              col.map((tc, cardIdx) => {
                const cardZone: StreetsAndAlleysMoveZone = {
                  zone: 'tableau',
                  col: colIdx,
                  cardIndex: cardIdx,
                };
                const isTop = cardIdx === col.length - 1;
                const isSelfSource = isSourceSelected('tableau', colIdx, cardIdx);
                // Ring only the columns that can actually accept the selection, so the
                // cue means "this move is legal" rather than "something is selected".
                const isTargetCandidate = isTop && !isSelfSource && legalTargets.tableau.has(colIdx);
                return (
                  <div
                    key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    {tc.card ? (
                      <button
                        type="button"
                        {...(isTop ? preview.previewProps(cardZone) : {})}
                        onClick={() => {
                          if (selectedSource) {
                            game.handleSelectTarget(tableauColZone);
                          } else if (isTop) {
                            game.handleSelectSource(cardZone);
                          }
                        }}
                        disabled={!isPlaying || loading || (!isTop && !selectedSource)}
                        aria-label={t('tableauCardAriaLabel', { card: cardAlt(tc.card), col: colIdx })}
                        aria-description={
                          hasSelectedSource ? destinationStatus(legalTargets.tableau.has(colIdx)) : undefined
                        }
                        aria-pressed={isSelfSource}
                        data-target-candidate={isTargetCandidate || undefined}
                        data-preview-target={isTargetCandidate && preview.isPreview ? 'true' : undefined}
                        draggable={isPlaying && !loading && isTop}
                        onDragStart={dnd.handleDragStart(cardZone)}
                        onDragEnd={dnd.handleDragEnd}
                        className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${isTop ? 'cursor-pointer' : 'cursor-default'} ${isSelfSource ? 'ring-2 ring-ds-warning' : ''}${isTargetCandidate ? targetRing : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
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

  return (
    <>
      <div
        ref={attachScrollTestId}
        className="flex gap-2 sm:gap-3 items-start"
        data-testid="sa-board-content"
        // 9 card columns + gaps: 2 outer gap-2 (8px) + 6 inner gap-1 (4px) = 40px.
        style={{ minWidth: isMobile ? `${9 * dims.cw + 40}px` : undefined }}
      >
        <div className="flex-1 flex gap-1 sm:gap-2" data-tutorial="sa-tableau">
          {[0, 1, 2, 3].map(renderTableauColumn)}
        </div>

        <div
          className="flex flex-col items-center gap-1 sm:gap-2 mb-3 lg:grid lg:grid-cols-2 lg:self-start"
          data-tutorial="sa-foundation"
        >
          {state.foundation.map((pile, idx) => {
            const foundationZone: StreetsAndAlleysMoveZone = { zone: 'foundation', col: idx };
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
                        suit: FOUNDATION_SUITS[idx],
                        count: pile.length,
                      })}
                      aria-description={
                        hasSelectedSource ? destinationStatus(legalTargets.foundation.has(idx)) : undefined
                      }
                      data-target-candidate={legalTargets.foundation.has(idx) || undefined}
                      data-preview-target={legalTargets.foundation.has(idx) && preview.isPreview ? 'true' : undefined}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}${
                        legalTargets.foundation.has(idx) ? targetRing : ''
                      }`}
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
                      disabled={!isPlaying || loading}
                      aria-disabled={!selectedSource || undefined}
                      aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                      aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                      aria-description={
                        hasSelectedSource ? destinationStatus(legalTargets.foundation.has(idx)) : undefined
                      }
                      style={{ width: dims.cw, height: dims.ch }}
                      data-target-candidate={legalTargets.foundation.has(idx) || undefined}
                      data-preview-target={legalTargets.foundation.has(idx) && preview.isPreview ? 'true' : undefined}
                      className={`border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}${
                        legalTargets.foundation.has(idx) ? targetRing : ' rounded'
                      }`}
                    >
                      A
                    </button>
                  )}
                </DropZone>
                <div
                  data-testid={`sa-foundation-progress-${idx.toString()}`}
                  className={`text-xs mt-1 tabular-nums ${pile.length === 13 ? 'text-ds-success' : 'text-game-text-muted'}`}
                >
                  {pile.length === 13 && <span aria-hidden="true">✓ </span>}
                  {t('foundationProgress', { count: pile.length })}
                </div>
              </div>
            );
          })}
        </div>

        <div className="flex-1 flex gap-1 sm:gap-2">{[4, 5, 6, 7].map(renderTableauColumn)}</div>
      </div>

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで現れるので変化として扱われず、読み上げられない
        ことがある (#5597)。
      */}
      <div data-tutorial="sa-hint-display" data-testid="sa-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2 mt-3">
            {t('hintAvailable')}: {formatHintZone(t, 'tableau', hint.fromCol)} →{' '}
            {formatHintZone(t, hint.toZone, hint.toCol)}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {isGameOver && (
        <p data-testid="sa-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
          {t('gameOverSummary', {
            count: foundationCount,
            percent: Math.round((foundationCount / 52) * 100),
          })}
        </p>
      )}
    </>
  );
}

function FooterActions({ game, state, t, confirmGiveUpAction }: StreetsAndAlleysContext) {
  const isPlaying = state.phase === StreetsAndAlleysPhase.PLAYING;
  if (!isPlaying) return null;

  // Auto-complete becomes useful once a foundation has built past its ace, so
  // pulse the button only then (mirrors Crescent / Spiderette).
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);

  return (
    <div data-tutorial="sa-controls">
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleUndo}
        disabled={game.loading || game.isAutoCompleting || !state.canUndo}
        aria-keyshortcuts="z"
      >
        {t('undo')}
        <KbdBadge label={t('kbd.undo')} />
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
        aria-keyshortcuts="h"
      >
        {t('hint')}
        <KbdBadge label={t('kbd.hint')} />
      </button>
      <button
        type="button"
        className={`${btnSuccess}${autoCompleteReady && !game.loading && !game.isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
        onClick={game.handleAutoComplete}
        disabled={game.loading || game.isAutoCompleting || !autoCompleteReady}
        data-testid="autocomplete-button"
        title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
        aria-keyshortcuts="a"
      >
        {t('autoComplete')}
        <KbdBadge label={t('kbd.autoComplete')} />
      </button>
      <button
        type="button"
        className={btnDanger}
        onClick={confirmGiveUpAction ?? (() => undefined)}
        disabled={game.loading || game.isAutoCompleting}
        aria-keyshortcuts="g"
      >
        {t('giveup')}
        <KbdBadge label={t('kbd.giveUp')} />
      </button>
    </div>
  );
}

const streetsAndAlleysDefinition: GameDefinition<
  StreetsAndAlleysResponse,
  StreetsAndAlleysGame,
  Parameters<typeof streetsAndAlleysApi.exec>,
  StreetsAndAlleysView
> = {
  key: 'streetsandalleys',
  useGame: useStreetsAndAlleysGame,
  useView: useStreetsAndAlleysView,
  tutorial: SA_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 4, tableau: 8 },
  cli: () => ({
    parseCommand: parseStreetsAndAlleysCommand,
    formatResponse: formatStreetsAndAlleysState,
    helpText: STREETSANDALLEYS_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameClear = state.phase === StreetsAndAlleysPhase.GAME_CLEAR;
    const isGameOver = state.phase === StreetsAndAlleysPhase.GAME_OVER;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameClear || isGameOver,
      won: isGameClear,
    };
  },
  contentClassName: 'overflow-x-auto pt-3 px-2 sm:px-4 lg:px-8',
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state, game }) => state.phase === StreetsAndAlleysPhase.PLAYING && !game.isAutoCompleting,
  shortcutsTestId: 'streetsandalleys-kbd-shortcuts',
  resetTutorialId: 'sa-reset-button',
  HeaderExtra,
  LiveRegion,
  Board,
  FooterActions,
};

/** Renders the Streets and Alleys solitaire game page with 8 tableau columns and 4 foundations. */
export const StreetsAndAlleysPage = createGamePage(streetsAndAlleysDefinition);
