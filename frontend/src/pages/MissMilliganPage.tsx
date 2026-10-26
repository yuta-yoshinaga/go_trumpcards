import { useCallback, useId, useMemo } from 'react';
import type { missMilliganApi } from '../api/gameApi';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useMissMilliganGame } from '../hooks/useMissMilliganGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { MissMilliganMoveZone, MissMilliganResponse } from '../types/card';
import { MissMilliganPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MISSMILLIGAN_HELP, parseMissMilliganCommand } from '../utils/cli/commands/missmilliganCommands';
import { formatMissMilliganState } from '../utils/cli/formatters/missmilliganFormatter';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦', '♠', '♣', '♥', '♦'] as const;
const TABLEAU_COLS = 8;
const TOTAL_CARDS = 104;

const MM_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mm-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="mm-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mm-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mm-stock"]', messageKey: 'tutorial.waive', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mm-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

const MISSMILLIGAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MissMilliganPhase.PLAYING]: 'playing',
  [MissMilliganPhase.GAME_CLEAR]: 'gameClear',
  [MissMilliganPhase.GAME_OVER]: 'gameOver',
};

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'waived') return t('frontendHint.waived');
  if (zone === 'stock') return t('frontendHint.stock');
  if (idx < 0) return t('frontendHint.deal');
  return t('frontendHint.tableau', { col: idx });
}

type MissMilliganGame = ReturnType<typeof useMissMilliganGame>;
type MissMilliganView = ReturnType<typeof useMissMilliganView>;
type MissMilliganContext = GamePageContext<MissMilliganResponse, MissMilliganGame, MissMilliganView>;

function useMissMilliganView({ game }: { game: MissMilliganGame }) {
  const selectSourceHintId = useId();
  const phaseNames = usePhaseNames('missmilligan', MISSMILLIGAN_PHASE_KEYS);
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const cw = cardWidth;
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = game.state?.phase === MissMilliganPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: MissMilliganMoveZone, target: MissMilliganMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<MissMilliganMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: game.loading,
  });

  return {
    selectSourceHintId,
    phaseNames,
    dims,
    isMobile,
    dnd,
  };
}

function HeaderExtra({ state, t }: MissMilliganContext) {
  return (
    <span className="text-sm text-ds-text-muted">
      {t('moveCount')}: {state.moveCount}
    </span>
  );
}

function Board({ game, state, t, tc, frontendHint, view }: MissMilliganContext) {
  const { dims, isMobile, dnd, selectSourceHintId } = view;
  const { loading, selectedSource, hint, isAutoCompleting } = game;
  const { hint: frontendHintData, hintEnabled: frontendHintEnabled } = frontendHint;

  const isPlaying = state.phase === MissMilliganPhase.PLAYING;
  const isGameOver = state.phase === MissMilliganPhase.GAME_OVER;
  const foundationCount = state.foundation.reduce((sum, pile) => sum + pile.length, 0);
  const dealsRemaining = Math.ceil(state.stockCount / TABLEAU_COLS);
  const foundationPercent = Math.round((foundationCount / TOTAL_CARDS) * 100);
  const isHolding = state.waived.length > 0;
  const selectedTableauRunIsLegal =
    selectedSource?.zone === 'tableau' && selectedSource.col !== undefined && selectedSource.cardIndex !== undefined
      ? state.runStarts[selectedSource.col].includes(selectedSource.cardIndex)
      : selectedSource?.zone === 'waived';
  const targetKey =
    selectedSource?.zone === 'waived'
      ? 'waived'
      : selectedSource?.zone === 'tableau' && selectedSource.col !== undefined && selectedSource.cardIndex !== undefined
        ? `tableau-${selectedSource.col}-${selectedSource.cardIndex}`
        : '';
  const serverTargets = selectedTableauRunIsLegal ? state.legalTargets[targetKey] : undefined;
  const legalTargets = {
    tableau: new Set(serverTargets?.tableau ?? []),
    foundation: new Set(serverTargets?.foundation ?? []),
  };

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // Waive はドメイン側では cardIndex から列末尾までの塊を退避できるのに、
  // ボタンが cardIndex を渡していなかったので常に最上段 1 枚だけだった (#6352)。
  // 選択が別の列や別のゾーンにあるときは渡さない — 渡すと関係のない列の
  // 途中から退避しようとする。
  const waiveCardIndex = (colIdx: number): number | undefined => {
    if (selectedSource === null || selectedSource.zone !== 'tableau' || selectedSource.col !== colIdx) return undefined;
    const index = selectedSource.cardIndex;
    return index !== undefined && state.runStarts[colIdx].includes(index) ? index : undefined;
  };

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx] ?? [];
    const tableauColZone: MissMilliganMoveZone = { zone: 'tableau', col: colIdx };
    return (
      <div
        key={`col-${colIdx.toString()}`}
        className={`${isMobile ? 'shrink-0' : 'flex-1 min-w-0'} ${legalTargets.tableau.has(colIdx) ? 'rounded ring-2 ring-ds-success' : ''}`}
        style={isMobile ? { width: dims.cw } : undefined}
        data-testid={`mm-tableau-col-${colIdx}`}
      >
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
              <button
                type="button"
                onClick={() => game.handleSelectTarget(tableauColZone)}
                disabled={!isPlaying || loading}
                aria-disabled={!selectedSource || undefined}
                aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                aria-label={t('emptyColumnAriaLabel', { col: colIdx })}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('kingsOnly')}
              </button>
            ) : (
              col.map((tc2, cardIdx) => {
                // Any card can head a run, so every card is a potential source.
                const cardZone: MissMilliganMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                const selectedRun =
                  selectedSource?.zone === 'tableau' &&
                  selectedSource.col === colIdx &&
                  selectedSource.cardIndex !== undefined;
                const isInSelectedRun =
                  selectedRun && selectedSource.cardIndex !== undefined && cardIdx >= selectedSource.cardIndex;
                const selectedRunRing = selectedTableauRunIsLegal ? 'ring-ds-warning' : 'ring-ds-error';
                return (
                  <div
                    key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                    className={`absolute left-0 right-0 ${isInSelectedRun ? `ring-2 ${selectedRunRing}` : ''}`}
                    style={{ top: cardIdx * dims.co }}
                  >
                    {tc2.card ? (
                      <button
                        type="button"
                        onClick={() => {
                          if (selectedSource) {
                            game.handleSelectTarget(tableauColZone);
                          } else {
                            game.handleSelectSource(cardZone);
                          }
                        }}
                        disabled={!isPlaying || loading}
                        aria-label={t('cardPosAria', { card: cardAlt(tc2.card), col: colIdx, pos: cardIdx + 1 })}
                        aria-pressed={isSelected}
                        draggable={isPlaying && !loading}
                        onDragStart={dnd.handleDragStart(cardZone)}
                        onDragEnd={dnd.handleDragEnd}
                        className={`p-0 border-0 bg-transparent w-full rounded cursor-pointer ${focusRingWhite} ${isSelected ? `ring-2 ${selectedRunRing}` : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
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
        {/* Waiving is per column and only reachable once the stock is gone, so
            the control lives with the column rather than in the footer. */}
        {state.canWaive && col.length > 0 && (
          <button
            type="button"
            onClick={() => game.handleWaive(colIdx, waiveCardIndex(colIdx))}
            disabled={!isPlaying || loading}
            aria-label={t('waiveColumnAriaLabel', { col: colIdx })}
            className={`${btnPrimary} w-full mt-1 text-xs`}
          >
            {t('waive')}
          </button>
        )}
      </div>
    );
  };

  const waivedZone: MissMilliganMoveZone = { zone: 'waived' };

  return (
    <>
      <span id={selectSourceHintId} className="sr-only">
        {tc('label.selectSourceFirst')}
      </span>
      <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
        <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="mm-foundation">
          {state.foundation.map((pile, idx) => {
            const foundationZone: MissMilliganMoveZone = { zone: 'foundation', col: idx };
            return (
              <div
                key={`f-${idx.toString()}`}
                className={`text-center ${legalTargets.foundation.has(idx) ? 'rounded ring-2 ring-ds-success' : ''}`}
                data-testid={`mm-foundation-${idx}`}
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

        <div className="flex gap-2 items-start" data-tutorial="mm-stock">
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
            <button
              type="button"
              onClick={game.handleDeal}
              disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0 || isHolding}
              aria-label={t('stockAriaLabel', { count: state.stockCount })}
              title={isHolding ? t('dealBlockedByWaive') : undefined}
              style={{ width: dims.cw, height: dims.ch }}
              className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
            >
              {state.stockCount}
            </button>
            <div className="text-game-text-muted text-xs mt-1">{t('dealsRemaining', { count: dealsRemaining })}</div>
          </div>
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('waived')}</div>
            {isHolding ? (
              <button
                type="button"
                onClick={() => game.handleSelectSource(waivedZone)}
                disabled={!isPlaying || loading}
                aria-label={t('waivedAriaLabel', { count: state.waived.length })}
                aria-pressed={isSourceSelected('waived', undefined, undefined)}
                draggable={isPlaying && !loading}
                onDragStart={dnd.handleDragStart(waivedZone)}
                onDragEnd={dnd.handleDragEnd}
                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waived', undefined, undefined) ? 'ring-2 ring-ds-warning' : ''}`}
              >
                <AnimatedCard card={state.waived[0]} width={dims.cw} draggable={false} />
                <span className="mt-1 block text-ds-text-primary text-xs" data-testid="mm-waived-count">
                  {t('waivedCount', { count: state.waived.length })}
                </span>
              </button>
            ) : (
              <div
                role="img"
                aria-label={state.canWaive ? t('waiveReadyAriaLabel') : t('emptyWaivedAriaLabel')}
                style={{ width: dims.cw, height: dims.ch }}
                className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
              >
                {state.canWaive ? t('waiveReady') : t('empty')}
              </div>
            )}
          </div>
        </div>
      </div>

      <div className={isMobile ? 'overflow-x-auto' : ''} data-tutorial="mm-tableau">
        <div className={`flex gap-1 sm:gap-2 items-start ${isMobile ? 'flex-nowrap w-max' : ''}`}>
          {Array.from({ length: TABLEAU_COLS }, (_, i) => i).map(renderTableauColumn)}
        </div>
      </div>

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-tutorial="mm-hint-display" data-testid="mm-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2 mt-3">
            {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
            {formatHintZone(t, hint.toZone, hint.toIdx)}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHintData} enabled={frontendHintEnabled} t={t} />
      </div>

      {isPlaying && (
        <p data-testid="mm-foundation-progress" className="text-ds-text-muted text-sm text-center mt-1 tabular-nums">
          {t('foundationProgress', { count: foundationCount, percent: foundationPercent })}
        </p>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {isGameOver && (
        <p data-testid="mm-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
          {t('gameOverSummary', {
            count: foundationCount,
            percent: foundationPercent,
          })}
        </p>
      )}
    </>
  );
}

function FooterActions({ game, state, t, confirmGiveUpAction, busy }: MissMilliganContext) {
  const isPlaying = state.phase === MissMilliganPhase.PLAYING;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);
  const isHolding = state.waived.length > 0;
  if (!isPlaying) return null;
  return (
    <div data-tutorial="mm-controls">
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleDeal}
        disabled={busy || game.isAutoCompleting || state.stockCount === 0 || isHolding}
        title={isHolding ? t('dealBlockedByWaive') : undefined}
      >
        {t('deal')}
      </button>
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleUndo}
        disabled={busy || game.isAutoCompleting || !state.canUndo}
      >
        {t('undo')}
      </button>
      {state.isStalemate && (
        <StalemateEscapeButton
          undoToEscape={state.undoToEscape ?? 0}
          onEscape={game.handleUndoEscape}
          disabled={busy || game.isAutoCompleting}
        />
      )}
      <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.isAutoCompleting}>
        {t('hint')}
      </button>
      <button
        type="button"
        className={`${btnSuccess}${autoCompleteReady && !busy && !game.isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
        onClick={game.handleAutoComplete}
        disabled={busy || game.isAutoCompleting || !autoCompleteReady}
        data-testid="autocomplete-button"
        title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
      >
        {t('autoComplete')}
      </button>
      <button
        type="button"
        className={btnDanger}
        onClick={confirmGiveUpAction}
        disabled={busy || game.isAutoCompleting}
      >
        {t('giveup')}
      </button>
    </div>
  );
}

const missMilliganDefinition: GameDefinition<
  MissMilliganResponse,
  MissMilliganGame,
  Parameters<typeof missMilliganApi.exec>,
  MissMilliganView
> = {
  key: 'missmilligan',
  useGame: useMissMilliganGame,
  useView: useMissMilliganView,
  tutorial: MM_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 8, tableau: TABLEAU_COLS },
  landscapeBanner: true,
  landscapeBannerInCliMode: true,
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  cli: () => ({
    parseCommand: parseMissMilliganCommand,
    formatResponse: formatMissMilliganState,
    helpText: MISSMILLIGAN_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === MissMilliganPhase.GAME_CLEAR || state.phase === MissMilliganPhase.GAME_OVER,
    won: state.phase === MissMilliganPhase.GAME_CLEAR,
  }),
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'd', action: game.handleDeal, label: 'deal' },
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state }) => state.phase === MissMilliganPhase.PLAYING,
  shortcutsTestId: 'missmilligan-kbd-shortcuts',
  resetTutorialId: 'mm-reset-button',
  HeaderExtra,
  Board,
  FooterActions,
};

/** Renders the Miss Milligan page: eight columns, eight foundations, and the waived slot. */
export const MissMilliganPage = createGamePage(missMilliganDefinition);
