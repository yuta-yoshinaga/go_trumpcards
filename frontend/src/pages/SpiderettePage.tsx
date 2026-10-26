import type { TFunction } from 'i18next';
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import type { SpideretteMoveZone, spideretteApi } from '../api/gameApi';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSpideretteGame } from '../hooks/useSpideretteGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { SpideretteResponse } from '../types/card';
import { SpiderettePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSpideretteCommand, SPIDERETTE_HELP } from '../utils/cli/commands/spideretteCommands';
import { formatSpideretteState } from '../utils/cli/formatters/spideretteFormatter';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

const SPDT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="spdt-stock-pile"]',
    messageKey: 'tutorial.stockPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="spdt-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="spdt-completed-suits"]',
    messageKey: 'tutorial.completedSuits',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="spdt-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="spdt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TOTAL_FOUNDATIONS = 4;
const TABLEAU_COLS = 7;

const SPIDERETTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SpiderettePhase.PLAYING]: 'playing',
  [SpiderettePhase.GAME_CLEAR]: 'gameClear',
  [SpiderettePhase.GAME_OVER]: 'gameOver',
};

type SpideretteGame = ReturnType<typeof useSpideretteGame>;
type SpideretteView = ReturnType<typeof useSpideretteView>;
type SpideretteContext = GamePageContext<SpideretteResponse, SpideretteGame, SpideretteView>;

function useSpideretteView({ game, t }: { game: SpideretteGame; t: TFunction }) {
  const { state, loading, handleDeal } = game;
  const selectSourceHintId = useId();
  const tableau = useResponsiveTableau(TABLEAU_COLS, { padX: 32, gapPx: 2 });
  const isPlayingForKbd = state?.phase === SpiderettePhase.PLAYING;

  const [emptyDealAttemptKey, setEmptyDealAttemptKey] = useState(0);
  const hasEmptyColumn = useMemo(() => state?.tableau.some((col) => col.length === 0) ?? false, [state?.tableau]);
  const dealBlockedByEmpty = hasEmptyColumn && (state?.stockCount ?? 0) > 0;
  const handleDealGuarded = useCallback(() => {
    if (dealBlockedByEmpty) {
      setEmptyDealAttemptKey((k) => k + 1);
      return;
    }
    setEmptyDealAttemptKey(0);
    handleDeal();
  }, [dealBlockedByEmpty, handleDeal]);

  const dispatchMove = useCallback(
    (source: SpideretteMoveZone, target: SpideretteMoveZone) => {
      void game.exec('move', source, target);
    },
    [game.exec],
  );
  const dnd = useSolitaireDragDrop<SpideretteMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  // Score breakdown popover: title-tooltips are unreachable on touch (#6389).
  const [scoreOpen, setScoreOpen] = useState(false);
  const scoreRef = useRef<HTMLSpanElement>(null);

  // Dismiss the score popover on Escape or a click/tap outside its wrapper.
  useEffect(() => {
    if (!scoreOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setScoreOpen(false);
    };
    const onPointer = (e: MouseEvent) => {
      if (scoreRef.current && !scoreRef.current.contains(e.target as Node)) setScoreOpen(false);
    };
    document.addEventListener('keydown', onKey);
    document.addEventListener('mousedown', onPointer);
    return () => {
      document.removeEventListener('keydown', onKey);
      document.removeEventListener('mousedown', onPointer);
    };
  }, [scoreOpen]);

  const scoreRuleText = state
    ? t('scoreRule', {
        start: state.scoring.start,
        penalty: state.scoring.movePenalty,
        bonus: state.scoring.suitBonus,
      })
    : '';

  const phaseNames = usePhaseNames('spiderette', SPIDERETTE_PHASE_KEYS);

  // A partial final deal (1–6 cards) still counts as one deal opportunity,
  // so round up rather than down (#1676 review).
  const dealsRemaining = state ? Math.ceil(state.stockCount / TABLEAU_COLS) : 0;
  const autoCompleteReady = state ? state.stockCount === 0 && isTableauAllFaceUp(state.tableau) : false;

  return {
    selectSourceHintId,
    tableau,
    emptyDealAttemptKey,
    dealBlockedByEmpty,
    handleDealGuarded,
    dnd,
    scoreOpen,
    setScoreOpen,
    scoreRef,
    scoreRuleText,
    phaseNames,
    dealsRemaining,
    autoCompleteReady,
  };
}

function HeaderExtra({ state, t, view }: SpideretteContext) {
  const { scoreRef, scoreOpen, setScoreOpen, scoreRuleText } = view;
  return (
    <>
      <span>
        {t('moveCount')}: {state.moveCount}
      </span>
      {/* **数字が動く理由を推測させない** (#5593, #6389)。内訳ポップオーバーは
          タッチ端末でも到達可能。数字はサーバから。 */}
      <span className="relative inline-flex ml-3" ref={scoreRef}>
        <button
          type="button"
          data-testid="spiderette-score"
          className="normal-case inline-flex items-center justify-center min-h-[44px] min-w-[44px] rounded px-2 py-0.5 text-xs text-ds-text-primary hover:text-ds-accent focus:outline-none focus:ring-1 focus:ring-ds-accent"
          aria-expanded={scoreOpen}
          aria-haspopup="true"
          onClick={() => setScoreOpen((o) => !o)}
          title={scoreRuleText}
        >
          {t('score')}: {state.score}
        </button>
        {scoreOpen && (
          /* aria-label is the region's *name*, so it must not repeat the body —
             a screen reader would read the same sentence twice. */
          <section
            data-testid="spiderette-score-popover"
            aria-label={t('score')}
            className="absolute top-full left-0 z-10 mt-1 w-max max-w-[16rem] rounded-lg bg-ds-surface border border-ds-accent/40 p-2 text-[11px] text-ds-text-primary shadow-lg"
          >
            {scoreRuleText}
          </section>
        )}
      </span>
    </>
  );
}

function AboveContent({ tc, view }: SpideretteContext) {
  return (
    <span id={view.selectSourceHintId} className="sr-only">
      {tc('label.selectSourceFirst')}
    </span>
  );
}

function Board({ game, state, t, frontendHint, view, busy }: SpideretteContext) {
  const { tableau, emptyDealAttemptKey, handleDealGuarded, dnd, dealsRemaining, selectSourceHintId } = view;
  const { selectedSource, hint, handleSelectSource, handleSelectTarget } = game;

  const isPlaying = state.phase === SpiderettePhase.PLAYING;

  const isSourceSelected = (col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === 'tableau' &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // Visual hint highlighting: ring the suggested source card and target column.
  const hintedCard = hint ? (state.tableau[hint.fromCol]?.[hint.cardIndex]?.card ?? null) : null;
  const isHintSource = (col: number, cardIndex: number) =>
    hint != null && hint.fromCol === col && hint.cardIndex === cardIndex;
  const isHintTargetCol = (col: number) => hint != null && hint.toCol === col;

  return (
    <>
      <div className="flex gap-2 mb-3 items-start">
        <div className="text-center" data-tutorial="spdt-stock-pile">
          <div className="text-game-text-muted text-xs mb-1">
            {t('stock')} ({state.stockCount})
          </div>
          {state.stockCount > 0 ? (
            <AnimatedCardBack
              width={tableau.cw}
              onClick={isPlaying ? handleDealGuarded : undefined}
              ariaLabel={t('deal')}
            />
          ) : (
            <div
              style={{ width: tableau.cw, height: tableau.ch }}
              className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
            >
              {t('empty')}
            </div>
          )}
          {state.stockCount > 0 && (
            <div className="text-game-text-muted text-xs mt-1">{t('dealsRemaining', { count: dealsRemaining })}</div>
          )}
        </div>
      </div>

      <div className="relative">
        <div className="flex gap-0.5 sm:gap-1 mb-3" data-tutorial="spdt-tableau">
          {state.tableau.map((col, colIdx) => {
            const tableauColZone: SpideretteMoveZone = { zone: 'tableau', col: colIdx };
            return (
              <div
                key={`col-${colIdx.toString()}`}
                data-testid={`spdt-col-${colIdx.toString()}`}
                className={
                  isHintTargetCol(colIdx)
                    ? 'flex-1 min-w-0 rounded ring-1 ring-ds-success animate-pulse'
                    : 'flex-1 min-w-0'
                }
              >
                <DropZone
                  isDropTarget={dnd.isDropTarget(tableauColZone)}
                  onDragOver={dnd.handleDragOver(tableauColZone)}
                  onDrop={dnd.handleDrop(tableauColZone)}
                  onDragLeave={dnd.handleDragLeave}
                  className="relative block"
                >
                  <div className="relative" style={{ minHeight: tableau.ch }}>
                    {col.length === 0 ? (
                      <button
                        key={`empty-${colIdx.toString()}-${emptyDealAttemptKey.toString()}`}
                        type="button"
                        onClick={() => handleSelectTarget(tableauColZone)}
                        disabled={!isPlaying || busy}
                        aria-disabled={!selectedSource || undefined}
                        aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                        aria-label={t('emptyColumnAriaLabel', { col: colIdx })}
                        style={{ height: tableau.ch }}
                        data-testid={`spdt-empty-col-${colIdx.toString()}`}
                        className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}${emptyDealAttemptKey > 0 ? ' animate-shake border-ds-warning text-ds-warning' : ''}`}
                      >
                        {t('empty')}
                      </button>
                    ) : (
                      col.map((tc, cardIdx) => {
                        const cardZone: SpideretteMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                        return (
                          <div
                            key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                            className="absolute left-0 right-0"
                            style={{ top: cardIdx * tableau.co }}
                          >
                            {tc.faceUp && tc.card ? (
                              <button
                                type="button"
                                onClick={() => {
                                  if (selectedSource) {
                                    if (selectedSource.col !== colIdx) {
                                      handleSelectTarget(tableauColZone);
                                    } else {
                                      handleSelectSource(cardZone);
                                    }
                                  } else {
                                    handleSelectSource(cardZone);
                                  }
                                }}
                                disabled={
                                  !isPlaying ||
                                  busy ||
                                  ((!selectedSource || selectedSource.col === colIdx) &&
                                    !state.runStarts[colIdx].includes(cardIdx))
                                }
                                aria-label={t('tableauCardAriaLabel', {
                                  card: cardAlt(tc.card),
                                  col: colIdx,
                                  pos: cardIdx,
                                })}
                                aria-pressed={isSourceSelected(colIdx, cardIdx)}
                                draggable={isPlaying && !busy}
                                onDragStart={dnd.handleDragStart(cardZone)}
                                onDragEnd={dnd.handleDragEnd}
                                data-testid={`spdt-card-${colIdx.toString()}-${cardIdx.toString()}`}
                                className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected(colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : isHintSource(colIdx, cardIdx) ? 'ring-2 ring-ds-info animate-pulse' : ''} ${!state.runStarts[colIdx].includes(cardIdx) ? 'opacity-40 cursor-default' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                              >
                                <AnimatedCard
                                  card={tc.card}
                                  width={tableau.cw}
                                  draggable={false}
                                  style={{ width: '100%' }}
                                />
                              </button>
                            ) : (
                              <AnimatedCardBack width={tableau.cw} style={{ width: '100%' }} />
                            )}
                          </div>
                        );
                      })
                    )}
                    {col.length > 0 && <div style={{ height: (col.length - 1) * tableau.co + tableau.ch }} />}
                  </div>
                </DropZone>
              </div>
            );
          })}
        </div>
      </div>

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="spiderette-hint-live" role="status" aria-live="polite">
        {hint && hintedCard && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintMoveAnnouncement', {
              card: cardAlt(hintedCard),
              tableau: t('tableau'),
              fromCol: hint.fromCol,
              toCol: hint.toCol,
            })}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      {/* Announce the empty-column deal guard for screen readers; the key
      forces a remount on each blocked attempt so it re-announces, mirroring
      the visual shake animation. */}
      {emptyDealAttemptKey > 0 && (
        <div
          key={`deal-warn-${emptyDealAttemptKey.toString()}`}
          className="sr-only"
          role="status"
          aria-live="assertive"
        >
          {t('cannotDealEmptyColExists')}
        </div>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ state, t }: SpideretteContext) {
  return (
    <div className="sr-only" role="status" aria-live="polite" data-testid="spiderette-stalemate-status">
      {state.isStalemate && (state.undoToEscape ?? 0) > 0
        ? t('stalemateWithEscape', { count: state.undoToEscape })
        : ''}
    </div>
  );
}

function FooterActions({ game, state, t, view, busy, confirmGiveUpAction }: SpideretteContext) {
  const isPlaying = state.phase === SpiderettePhase.PLAYING;
  const { handleDealGuarded, dealBlockedByEmpty, autoCompleteReady } = view;
  if (!isPlaying) return null;
  return (
    <div data-tutorial="spdt-controls">
      <button
        type="button"
        className={btnPrimary}
        onClick={handleDealGuarded}
        disabled={busy || game.isAutoCompleting}
        title={dealBlockedByEmpty ? t('cannotDealEmptyColExists') : undefined}
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

const spideretteDefinition: GameDefinition<
  SpideretteResponse,
  SpideretteGame,
  Parameters<typeof spideretteApi.exec>,
  SpideretteView
> = {
  key: 'spiderette',
  useGame: useSpideretteGame,
  useView: useSpideretteView,
  tutorial: SPDT_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 1, tableau: TABLEAU_COLS },
  cli: () => ({
    parseCommand: parseSpideretteCommand,
    formatResponse: formatSpideretteState,
    helpText: SPIDERETTE_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameClear = state.phase === SpiderettePhase.GAME_CLEAR;
    const isGameOver = state.phase === SpiderettePhase.GAME_OVER;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameClear || isGameOver,
      won: isGameClear,
    };
  },
  shell: ({ state, t }) => ({
    headerEnd: (
      <span className="ml-3" data-tutorial="spdt-completed-suits">
        {t('completed')}: {state.completedSuits}/{TOTAL_FOUNDATIONS}
      </span>
    ),
  }),
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, view, confirmGiveUpAction }) => [
    { key: 'd', action: view.handleDealGuarded, label: 'deal' },
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state }) => state.phase === SpiderettePhase.PLAYING,
  shortcutsTestId: 'spiderette-kbd-shortcuts',
  landscapeBanner: true,
  resetTutorialId: 'spdt-reset-button',
  HeaderExtra,
  AboveContent,
  Board,
  FooterStatus,
  FooterActions,
};

/** Renders the Spiderette game page (Klondike setup + Spider rules with a single deck). */
export const SpiderettePage = createGamePage(spideretteDefinition);
