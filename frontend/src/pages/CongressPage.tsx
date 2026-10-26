import { useCallback, useId, useMemo } from 'react';
import type { congressApi } from '../api/gameApi';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCongressGame } from '../hooks/useCongressGame';
import { useElementWidth } from '../hooks/useElementWidth';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { CongressMoveZone, CongressResponse } from '../types/card';
import { CongressPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CONGRESS_HELP, parseCongressCommand } from '../utils/cli/commands/congressCommands';
import { formatCongressState } from '../utils/cli/formatters/congressFormatter';
import { tableauDimensions } from '../utils/tableauDimensions';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦', '♠', '♣', '♥', '♦'] as const;
const TABLEAU_PILES = 8;
// Matches the tableau's sm:gap-2 spacing (8px).
const TABLEAU_GAP = 8;
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

const CONGRESS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CongressPhase.PLAYING]: 'playing',
  [CongressPhase.GAME_CLEAR]: 'gameClear',
  [CongressPhase.GAME_OVER]: 'gameOver',
};

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { pile: idx });
}

/** Returns the legalTargets key for a move source, or an empty key when incomplete. */
export function congressSourceKey(source: CongressMoveZone): string {
  if (source.zone === 'tableau') return source.col === undefined ? '' : `tableau-${source.col}`;
  if (source.zone === 'waste' || source.zone === 'stock') return source.zone;
  return '';
}

type CongressGame = ReturnType<typeof useCongressGame>;
type CongressView = ReturnType<typeof useCongressView>;
type CongressContext = GamePageContext<CongressResponse, CongressGame, CongressView>;

function useCongressView({ game }: { game: CongressGame }) {
  const selectSourceHintId = useId();
  const phaseNames = usePhaseNames('congress', CONGRESS_PHASE_KEYS);
  const { cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();
  const [tableauRef, tableauRowWidth] = useElementWidth<HTMLDivElement>();

  const dims = useMemo(() => {
    if (!isMobile) {
      const measured = tableauDimensions(
        tableauRowWidth || cardWidth * TABLEAU_PILES + (TABLEAU_PILES - 1) * TABLEAU_GAP,
        TABLEAU_PILES,
        TABLEAU_GAP,
        cardOverlap / cardWidth,
      );
      return { cw: measured.cardWidth, ch: measured.cardHeight, co: measured.cardOverlap };
    }
    const padX = 16;
    const gapPx = 4;
    const colW = Math.floor((windowWidth - padX - (TABLEAU_PILES - 1) * gapPx) / TABLEAU_PILES);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardOverlap, tableauRowWidth]);

  const dispatchMove = useCallback(
    (source: CongressMoveZone, target: CongressMoveZone) => {
      // useSolitaireDragDrop only invokes onMove while isPlaying is true, which
      // requires a loaded game state.
      const currentState = game.state as CongressResponse;
      const legalTargets = currentState.legalTargets[congressSourceKey(source)];
      const isLegal =
        legalTargets !== undefined &&
        target.col !== undefined &&
        (target.zone === 'tableau'
          ? legalTargets.tableau.includes(target.col)
          : target.zone === 'foundation' && legalTargets.foundation.includes(target.col));
      if (!isLegal) return;
      void game.exec('move', source, target);
    },
    [game],
  );

  const isPlayingForKbd = game.state?.phase === CongressPhase.PLAYING;
  const dnd = useSolitaireDragDrop<CongressMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: game.loading,
  });

  return {
    selectSourceHintId,
    phaseNames,
    dims,
    tableauRef,
    dnd,
  };
}

function HeaderExtra({ state, t }: CongressContext) {
  return (
    <span className="text-sm text-ds-text-muted">
      {t('moveCount')}: {state.moveCount}
    </span>
  );
}

function Board({ game, state, t, tc, view, frontendHint }: CongressContext) {
  const { selectSourceHintId, dims, tableauRef, dnd } = view;
  const { loading, selectedSource, hint, isAutoCompleting } = game;
  const emptyPileAriaDisabled = !selectedSource || selectedSource.zone === 'tableau';
  const isPlaying = state.phase === CongressPhase.PLAYING;
  const isGameOver = state.phase === CongressPhase.GAME_OVER;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;

  const activeSource = dnd.dragSource ?? selectedSource;
  const isSourceSelected = (zone: string, col?: number) =>
    activeSource !== null && activeSource.zone === zone && activeSource.col === col;

  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: CongressMoveZone = { zone: 'waste' };
  const stockZone: CongressMoveZone = { zone: 'stock' };
  const legalTargets = activeSource ? state.legalTargets[congressSourceKey(activeSource)] : undefined;
  const targetRing = ' rounded ring-2 ring-ds-success';

  const renderPile = (pileIdx: number) => {
    const cards = state.tableau[pileIdx] ?? [];
    const pileZone: CongressMoveZone = { zone: 'tableau', col: pileIdx };
    return (
      <div
        key={`pile-${pileIdx.toString()}`}
        className={`flex-1 min-w-0${legalTargets?.tableau.includes(pileIdx) ? targetRing : ''}`}
        data-legal-target={legalTargets?.tableau.includes(pileIdx) ? 'true' : undefined}
      >
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
                onClick={() => {
                  if (emptyPileAriaDisabled) return;
                  game.handleSelectTarget(pileZone);
                }}
                // **空き山はタブローからは埋められない。**山札か捨て札からだけ
                // (`MoveTableauToTableau` が明示的に拒否する)。押せてしまうと
                // サーバに弾かれるまで気づけない (#4906)。
                disabled={!isPlaying || loading}
                aria-disabled={emptyPileAriaDisabled || undefined}
                aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                aria-label={t('emptyPileAriaLabel', { pile: pileIdx })}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('stockOrWaste')}
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
                      // 空の山は番号を言うのに、札が乗っている山は札名だけで
                      // どの山か分からなかった (#6358)。掴めるのは一番上だけ
                      // なので、そこも言い分ける。
                      aria-label={t(isTop ? 'pileTopCardAriaLabel' : 'pileCardAriaLabel', {
                        card: cardAlt(card),
                        pile: pileIdx,
                      })}
                      aria-pressed={isTop ? isSourceSelected('tableau', pileIdx) : undefined}
                      draggable={isTop && isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(pileZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent w-full rounded cursor-pointer ${focusRingWhite} ${isTop && isSourceSelected('tableau', pileIdx) ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <AnimatedCard
                        card={card}
                        width={dims.cw}
                        isSelected={isTop && isSourceSelected('tableau', pileIdx)}
                        draggable={false}
                        style={{ width: '100%' }}
                        wrapperClassName="block w-full"
                      />
                    </button>
                  </div>
                );
              })
            )}
            {cards.length > 0 && (
              <div data-testid="cg-tableau-column-spacer" style={{ height: (cards.length - 1) * dims.co + dims.ch }} />
            )}
          </div>
        </DropZone>
      </div>
    );
  };

  return (
    <>
      <span id={selectSourceHintId} className="sr-only">
        {tc('label.selectSourceFirst')}
      </span>
      <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
        <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="cg-foundation">
          {state.foundation.map((pile, idx) => {
            const foundationZone: CongressMoveZone = { zone: 'foundation', col: idx };
            return (
              <div
                key={`f-${idx.toString()}`}
                className={`text-center${legalTargets?.foundation.includes(idx) ? targetRing : ''}`}
                data-legal-target={legalTargets?.foundation.includes(idx) ? 'true' : undefined}
              >
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
                      disabled={!isPlaying || loading}
                      aria-disabled={!selectedSource || undefined}
                      aria-describedby={!selectedSource ? selectSourceHintId : undefined}
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
            {/* The stock is both the draw button and, when a card is
                selected against an empty pile, a move source. */}
            <button
              type="button"
              onClick={() => (selectedSource ? game.handleSelectSource(stockZone) : game.handleDraw())}
              disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
              aria-label={
                state.stockCount === 0 ? t('emptyStockAriaLabel') : t('stockAriaLabel', { count: state.stockCount })
              }
              aria-pressed={isSourceSelected('stock', undefined)}
              style={{ width: dims.cw, height: dims.ch }}
              className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${isSourceSelected('stock', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
            >
              {state.stockCount}
            </button>
          </div>
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
            {wasteTop ? (
              <>
                <div id="congress-waste-count" className="text-game-text-muted text-xs mb-1">
                  {t('wasteCount', { count: state.waste.length })}
                </div>
                <button
                  type="button"
                  onClick={() => game.handleSelectSource(wasteZone)}
                  disabled={!isPlaying || loading}
                  aria-label={t('wasteTopCardAriaLabel', { card: cardAlt(wasteTop) })}
                  aria-describedby="congress-waste-count"
                  aria-pressed={isSourceSelected('waste', undefined)}
                  draggable={isPlaying && !loading}
                  onDragStart={dnd.handleDragStart(wasteZone)}
                  onDragEnd={dnd.handleDragEnd}
                  className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                >
                  <AnimatedCard card={wasteTop} width={dims.cw} draggable={false} />
                </button>
              </>
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

      <div ref={tableauRef} className="flex gap-1 sm:gap-2 items-start pt-3" data-tutorial="cg-tableau">
        {Array.from({ length: TABLEAU_PILES }, (_, i) => i).map(renderPile)}
      </div>

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-tutorial="cg-hint-display" data-testid="cg-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2 mt-3">
            {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromIdx)} →{' '}
            {formatHintZone(t, hint.toZone, hint.toIdx)}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {isGameOver && (
        <p data-testid="cg-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
          {t('gameOverSummary', {
            count: foundationCount,
            percent: Math.round((foundationCount / TOTAL_CARDS) * 100),
          })}
        </p>
      )}
    </>
  );
}

function FooterActions({ game, state, t, confirmGiveUpAction }: CongressContext) {
  const isPlaying = state.phase === CongressPhase.PLAYING;
  const isAutoCompleting = game.isAutoCompleting;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);

  if (!isPlaying) return null;

  return (
    <div data-tutorial="cg-controls">
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleDraw}
        disabled={game.loading || isAutoCompleting || state.stockCount === 0}
      >
        {t('draw')}
      </button>
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleUndo}
        disabled={game.loading || isAutoCompleting || !state.canUndo}
      >
        {t('undo')}
      </button>
      {state.isStalemate && (
        <StalemateEscapeButton
          undoToEscape={state.undoToEscape ?? 0}
          onEscape={game.handleUndoEscape}
          disabled={game.loading || isAutoCompleting}
        />
      )}
      <button
        type="button"
        className={btnSuccess}
        onClick={game.handleHint}
        disabled={game.loading || isAutoCompleting}
      >
        {t('hint')}
      </button>
      <button
        type="button"
        className={`${btnSuccess}${autoCompleteReady && !game.loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
        onClick={game.handleAutoComplete}
        disabled={game.loading || isAutoCompleting || !autoCompleteReady}
        data-testid="autocomplete-button"
        title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
      >
        {t('autoComplete')}
      </button>
      <button
        type="button"
        className={btnDanger}
        onClick={confirmGiveUpAction}
        disabled={game.loading || isAutoCompleting}
      >
        {t('giveup')}
      </button>
    </div>
  );
}

const congressDefinition: GameDefinition<
  CongressResponse,
  CongressGame,
  Parameters<typeof congressApi.exec>,
  CongressView
> = {
  key: 'congress',
  useGame: useCongressGame,
  useView: useCongressView,
  tutorial: CG_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 8, tableau: TABLEAU_PILES },
  cli: () => ({
    parseCommand: parseCongressCommand,
    formatResponse: formatCongressState,
    helpText: CONGRESS_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === CongressPhase.GAME_CLEAR || state.phase === CongressPhase.GAME_OVER,
    won: state.phase === CongressPhase.GAME_CLEAR,
  }),
  actionLogDuringPlay: true,
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, confirmGiveUpAction }) => [
    {
      key: 'd',
      action: game.handleDraw,
      label: 'draw',
    },
    {
      key: 'h',
      action: game.handleHint,
      label: 'hint',
    },
    {
      key: 'a',
      action: game.handleAutoComplete,
      label: 'autoComplete',
    },
    {
      key: 'g',
      action: confirmGiveUpAction ?? (() => undefined),
      label: 'giveUp',
    },
    {
      key: 'z',
      action: game.handleUndo,
      label: 'undo',
    },
  ],
  shortcutsEnabled: ({ state, actionLogOpen }) => state.phase === CongressPhase.PLAYING && !actionLogOpen,
  shortcutsTestId: 'congress-kbd-shortcuts',
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  landscapeBanner: true,
  landscapeBannerInCliMode: true,
  resetTutorialId: 'cg-reset-button',
  HeaderExtra,
  Board,
  FooterActions,
};

/** Renders the Congress page: eight single-card piles flanking eight foundations. */
export const CongressPage = createGamePage(congressDefinition);
