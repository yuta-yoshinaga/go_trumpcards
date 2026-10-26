import type { TFunction } from 'i18next';
import { useEffect, useMemo, useRef, useState } from 'react';
import type { tripeaksApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useChainCombo } from '../hooks/useChainCombo';
import { useTriPeaksGame } from '../hooks/useTriPeaksGame';
import { useTriPeaksStats } from '../hooks/useTriPeaksStats';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { TriPeaksResponse } from '../types/card';
import { TriPeaksPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTripeaksCommand, TRIPEAKS_HELP } from '../utils/cli/commands/tripeaksCommands';
import { formatTripeaksState } from '../utils/cli/formatters/tripeaksFormatter';
import { comboAnnouncement } from '../utils/comboAnnounce';

/** Valid column positions per row in the TriPeaks tableau. */
const VALID_COLS: readonly number[][] = [
  [0, 3, 6],
  [0, 1, 3, 4, 6, 7],
  [0, 1, 2, 3, 4, 5, 6, 7, 8],
  [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
];

/** Number of peaks in a TriPeaks tableau. */
const PEAK_COUNT = 3;

/**
 * Maps a tableau column index to its peak index (0-2). Columns are partitioned
 * into three contiguous groups — 0-2 (left peak), 3-5 (middle peak), 6-9 (right
 * peak) — so every card in the layout is attributed to exactly one peak.
 */
function peakOfColumn(col: number): number {
  if (col < 3) return 0;
  if (col < 6) return 1;
  return 2;
}

/**
 * Derives the number of remaining (present and not-yet-removed) cards in each of
 * the three peaks from the TriPeaks layout. Returns a fixed-length `[left,
 * middle, right]` tuple; a missing layout yields `[0, 0, 0]`.
 */
export function computePeakRemaining(layout: TriPeaksResponse['layout'] | undefined): number[] {
  const remaining = [0, 0, 0];
  if (!layout) return remaining;
  for (const row of layout) {
    row.forEach((cell, col) => {
      if (cell?.card && !cell.removed) {
        remaining[peakOfColumn(col)] += 1;
      }
    });
  }
  return remaining;
}

/** TriPeaks tutorial step definitions. */
const TP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tp-peaks"]',
    messageKey: 'tutorial.peaks',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type TriPeaksGame = ReturnType<typeof useTriPeaksGame>;

function useTriPeaksView({ game, t }: { game: TriPeaksGame; t: TFunction }) {
  const { playSound } = useSound();
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const phaseName = useMemo(() => {
    const p = game.state?.phase;
    if (p === TriPeaksPhase.GAME_CLEAR) return t('phase.gameClear');
    if (p === TriPeaksPhase.GAME_OVER) return t('phase.gameOver');
    return t('phase.playing');
  }, [game.state?.phase, t]);

  const combo = useChainCombo(game.state?.moveCount, game.state?.stockCount);

  // **バッジは combo >= 2 のときだけ描かれる。** 消えることは live region では
  // 伝わらないので、領域は常設して中身を差し替える。判断は Golf と共有の
  // 純関数に任せる (#5821, もとは #5520)。
  const [comboAnnounce, setComboAnnounce] = useState('');
  const prevCombo = useRef(0);
  useEffect(() => {
    const was = prevCombo.current;
    prevCombo.current = combo;
    const next = comboAnnouncement(combo, was);
    setComboAnnounce(next ? t(next.key, { count: next.count }) : '');
  }, [combo, t]);

  // Remaining-card count per peak, derived from the board layout (issue #3085).
  const peakRemaining = useMemo(() => computePeakRemaining(game.state?.layout), [game.state?.layout]);
  const isPlayingForPeaks = game.state?.phase === TriPeaksPhase.PLAYING;
  // Fire a subtle sound once when a peak's remaining count first reaches zero.
  // The ref always tracks the latest counts, so reset/undo (which refill a peak)
  // naturally re-arms the celebration without any explicit teardown.
  const prevPeakRemaining = useRef<number[]>([0, 0, 0]);
  useEffect(() => {
    const prev = prevPeakRemaining.current;
    if (isPlayingForPeaks) {
      for (let i = 0; i < PEAK_COUNT; i++) {
        if (prev[i] > 0 && peakRemaining[i] === 0) {
          playSound('cardPlace');
        }
      }
    }
    prevPeakRemaining.current = peakRemaining;
  }, [peakRemaining, isPlayingForPeaks, playSound]);

  // **得点はサーバが数える。** 以前はフロントの useTriPeaksScore が盤面の遷移から
  // 計算しており、同じ式がドメインに無いせいで CUI からは得点に手が届かなかった
  // (#5511)。式は internal/domain/TriPeaks.go に移してある。
  const score = game.state?.score ?? 0;

  // Best-record persistence in localStorage (issue #3087).
  const { stats, recordResult } = useTriPeaksStats();
  const [newBest, setNewBest] = useState(false);
  // Guard so each finished game is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  const endPhase = game.state?.phase;
  useEffect(() => {
    const ended = endPhase === TriPeaksPhase.GAME_CLEAR || endPhase === TriPeaksPhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    setNewBest(recordResult({ won: endPhase === TriPeaksPhase.GAME_CLEAR, score }));
  }, [endPhase, score, recordResult]);

  return {
    cardHeight,
    cardWidth,
    isMobile,
    windowWidth,
    phaseName,
    combo,
    comboAnnounce,
    peakRemaining,
    stats,
    newBest,
    score,
  };
}

type TriPeaksView = ReturnType<typeof useTriPeaksView>;
type TriPeaksContext = GamePageContext<TriPeaksResponse, TriPeaksGame, TriPeaksView>;

function HeaderExtra({ state, t, view }: TriPeaksContext) {
  const isPlaying = state.phase === TriPeaksPhase.PLAYING;
  // **リングと合計は同じ集合を読む。**別々に数えると、光っている札の数と
  // ヘッダーの数字がずれる。CUI の playableCount と揃える (#4783)。
  const playableCells = isPlaying ? new Set(state.playableCells) : new Set<string>();

  return (
    <>
      <span>
        {t('moveCount')}: {state.moveCount}
      </span>
      <span data-testid="tp-score">
        {t('score')}: <span className="font-bold tabular-nums">{view.score}</span>
      </span>
      {isPlaying && (
        <span data-testid="tp-playable" className="text-xs">
          {t('playableSummary', { count: playableCells.size })}
          {playableCells.size === 0 && state.stockCount > 0 && (
            <span className="ml-1 text-ds-warning">{t('drawRecommended')}</span>
          )}
        </span>
      )}
      {isPlaying && (
        <span data-testid="peak-remaining" className="flex items-center gap-1 text-xs">
          <span className="sr-only">{t('peakRemaining')}</span>
          <span aria-hidden="true">⛰</span>
          {view.peakRemaining.map((n, i) => (
            <span key={`peak-${i.toString()}`} className="flex items-center">
              {i > 0 && <span className="text-game-text-muted mx-0.5">/</span>}
              <span className="sr-only">{t(['leftPeak', 'middlePeak', 'rightPeak'][i])}</span>
              <span
                title={n === 0 ? t('peakCleared') : undefined}
                className={`font-bold ${n === 0 ? 'text-ds-success' : ''}`}
              >
                {n === 0 ? '✓' : n}
              </span>
            </span>
          ))}
        </span>
      )}
      {view.combo >= 2 && (
        <span
          data-testid="combo-badge"
          className={`px-2 py-0.5 rounded-full text-xs font-bold ${
            view.combo >= 5
              ? 'bg-ds-error text-white'
              : view.combo >= 3
                ? 'bg-ds-warning text-ds-text-on-accent'
                : 'bg-ds-info text-ds-text-on-accent'
          }`}
        >
          {t('combo', { count: view.combo })}
        </span>
      )}
      <span className="sr-only" role="status" aria-live="polite" data-testid="tripeaks-combo-announce">
        {view.comboAnnounce}
      </span>
    </>
  );
}

function Board({ game, state, t, view, frontendHint }: TriPeaksContext) {
  const isPlaying = state.phase === TriPeaksPhase.PLAYING;
  // **リングと合計は同じ集合を読む。**別々に数えると、光っている札の数と
  // ヘッダーの数字がずれる。CUI の playableCount と揃える (#4783)。
  const playableCells = isPlaying ? new Set(state.playableCells) : new Set<string>();
  const isGameClear = state.phase === TriPeaksPhase.GAME_CLEAR;
  const isGameOver = state.phase === TriPeaksPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const hintedCard = game.hint?.type === 'remove' ? state.layout[game.hint.row]?.[game.hint.col]?.card : undefined;

  // Calculate layout dimensions
  const maxCols = 10;
  const cardGap = 4;
  const ROW_OVERLAP_RATIO = view.isMobile ? 0.3 : 0.35;
  const rowOverlap = view.cardHeight * ROW_OVERLAP_RATIO;
  // px-4 on the scrollable container = 16px * 2 = 32px total horizontal padding
  const CONTAINER_PADDING = 32;
  const effectiveCardWidth = view.isMobile
    ? Math.floor((view.windowWidth - CONTAINER_PADDING - cardGap * (maxCols - 1)) / maxCols)
    : view.cardWidth;

  return (
    <>
      {/* Tableau */}
      <div data-tutorial="tp-peaks" className="flex flex-col items-center mb-3">
        {VALID_COLS.map((cols, rowIdx) => {
          const rowWidth = maxCols * (effectiveCardWidth + cardGap) - cardGap;
          return (
            <div
              key={`row-${rowIdx.toString()}`}
              className="relative"
              style={{
                height: rowIdx < VALID_COLS.length - 1 ? view.cardHeight - rowOverlap : view.cardHeight,
                width: rowWidth,
              }}
            >
              {cols.map((colIdx) => {
                const tc2 = state.layout[rowIdx]?.[colIdx];
                const left = colIdx * (effectiveCardWidth + cardGap);
                if (!tc2 || tc2.removed) {
                  return (
                    <div
                      key={`tc-${rowIdx.toString()}-${colIdx.toString()}`}
                      className="absolute"
                      style={{ left, width: effectiveCardWidth, height: view.cardHeight }}
                    />
                  );
                }
                if (!tc2.card) return null;
                const exposed = tc2.exposed;
                const isHinted = game.hint?.type === 'remove' && game.hint.row === rowIdx && game.hint.col === colIdx;
                const isPlayable = playableCells.has(`${rowIdx.toString()}-${colIdx.toString()}`);
                return (
                  <div key={`tc-${rowIdx.toString()}-${colIdx.toString()}`} className="absolute" style={{ left }}>
                    <button
                      type="button"
                      onClick={() => {
                        if (!exposed || !tc2.card) return;
                        game.handleSelectCard(rowIdx, colIdx);
                      }}
                      disabled={!isPlaying || game.loading || !exposed}
                      aria-label={isPlayable ? t('cardPlayableAria', { card: cardAlt(tc2.card) }) : cardAlt(tc2.card)}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                        isHinted ? 'ring-2 ring-ds-warning' : isPlayable ? 'ring-2 ring-ds-success/70' : ''
                      } ${!exposed ? 'opacity-60' : ''}`}
                    >
                      <AnimatedCard card={tc2.card} width={effectiveCardWidth} />
                    </button>
                  </div>
                );
              })}
            </div>
          );
        })}
      </div>

      {/* Stock + Waste */}
      <div className="flex gap-4 justify-center mb-3" data-tutorial="tp-stock-waste">
        <div className="text-center">
          <div className="text-game-text-muted text-xs mb-1">
            {t('stock')} ({state.stockCount})
          </div>
          {state.stockCount > 0 ? (
            <AnimatedCardBack
              width={effectiveCardWidth}
              onClick={isPlaying ? game.handleDraw : undefined}
              ariaLabel={t('draw')}
            />
          ) : (
            <div
              style={{ width: effectiveCardWidth, height: view.cardHeight }}
              className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
            >
              {t('empty')}
            </div>
          )}
        </div>

        <div className="text-center">
          <div className="text-game-text-muted text-xs mb-1">
            {t('waste')} <span id="tripeaks-waste-count">{t('wasteCount', { count: state.waste.length })}</span>
          </div>
          {state.waste.length > 0 ? (
            <AnimatedCard
              card={state.waste[state.waste.length - 1]}
              width={effectiveCardWidth}
              ariaDescribedBy="tripeaks-waste-count"
            />
          ) : (
            <div
              style={{ width: effectiveCardWidth, height: view.cardHeight }}
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
      <div data-tutorial="tp-hint-display" data-testid="tp-hint-live" role="status" aria-live="polite">
        {game.hint && (
          <div className="text-ds-warning text-sm mb-2 text-center">
            {game.hint.type === 'remove' && hintedCard
              ? t('removeHint', {
                  card: cardAlt(hintedCard),
                  row: game.hint.row,
                  col: game.hint.col,
                })
              : `${t('hintAvailable')}: ${t(`hintType.${game.hint.type}`)}`}
          </div>
        )}
      </div>
      <div data-testid="tp-stock-empty-guidance" role="status" aria-live="polite">
        {isPlaying && state.stockCount === 0 && (
          <div className="text-ds-warning text-sm mb-2 text-center">
            {playableCells.size > 0
              ? t('stockEmptyPlayable', { count: playableCells.size })
              : t('stockEmptyNoPlayable')}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {/* New personal-best badge on the end screen (#3087). */}
      {isEnded && view.newBest && (
        <div
          data-testid="tp-best-badge"
          role="status"
          className="text-center text-ds-success font-semibold text-sm mb-2"
        >
          {t('newBest', { score: view.score })}
        </div>
      )}

      {/* Best-record panel: highest score + clear rate (#3087). */}
      <div data-testid="tp-stats-panel" className="text-game-text-muted text-xs text-center mb-2">
        {t('bestScore')}: {view.stats.bestScore ?? '—'}
        {view.stats.plays > 0 && (
          <>
            {' · '}
            {t('clears', { wins: view.stats.wins, plays: view.stats.plays })}
          </>
        )}
      </div>
    </>
  );
}

function FooterActions({ game, state, t, busy, confirmGiveUpAction }: TriPeaksContext) {
  const isPlaying = state.phase === TriPeaksPhase.PLAYING;
  if (!isPlaying) return null;
  return (
    <div data-tutorial="tp-controls">
      <button type="button" className={btnPrimary} onClick={game.handleDraw} disabled={busy}>
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

const tripeaksDefinition: GameDefinition<
  TriPeaksResponse,
  TriPeaksGame,
  Parameters<typeof tripeaksApi.exec>,
  TriPeaksView
> = {
  key: 'tripeaks',
  useGame: useTriPeaksGame,
  useView: useTriPeaksView,
  tutorial: TP_TUTORIAL_STEPS,
  skeleton: { kind: 'tiered-rows', rows: [3, 6, 9, 10], stockWaste: true },
  cli: () => ({
    parseCommand: parseTripeaksCommand,
    formatResponse: formatTripeaksState,
    helpText: TRIPEAKS_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameClear = state.phase === TriPeaksPhase.GAME_CLEAR;
    const isGameOver = state.phase === TriPeaksPhase.GAME_OVER;
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
  shortcutsEnabled: ({ state }) => state.phase === TriPeaksPhase.PLAYING,
  shortcutsTestId: 'tri-peaks-kbd-shortcuts',
  resetTutorialId: 'tp-reset-button',
  landscapeBanner: true,
  HeaderExtra,
  Board,
  FooterActions,
};

/** Renders the TriPeaks Solitaire game page with three peaks, stock/waste, and controls. */
export const TriPeaksPage = createGamePage(tripeaksDefinition);
