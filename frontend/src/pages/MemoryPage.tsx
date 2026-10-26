import type { TFunction } from 'i18next';
import { type KeyboardEvent as ReactKeyboardEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { memoryApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  AUTO_NEXT_DELAY_OPTIONS,
  CPU_DIFFICULTY_OPTIONS,
  PAIR_COUNT_OPTIONS,
  useMemoryGame,
} from '../hooks/useMemoryGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { Card, MemoryResponse } from '../types/card';
import { MemoryPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MEMORY_HELP, parseMemoryCommand } from '../utils/cli/commands/memoryCommands';
import { formatMemoryState } from '../utils/cli/formatters/memoryFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { type GridDir, moveFocus } from '../utils/gridNav';
import { getMemoryHint } from '../utils/hints/memoryHint';
import { memoryKnownMatch } from '../utils/memoryKnownMatch';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Memory tutorial step definitions. */
const MEM_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mem-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mem-board"]',
    messageKey: 'tutorial.board',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mem-next-button"]',
    messageKey: 'tutorial.nextButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mem-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MEMORY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MemoryPhase.FLIP1]: 'flip1',
  [MemoryPhase.FLIP2]: 'flip2',
  [MemoryPhase.RESULT]: 'result',
  [MemoryPhase.GAME_END]: 'gameEnd',
};

/** Maps arrow keys to a grid navigation direction. */
const ARROW_DIRS: Readonly<Record<string, GridDir>> = {
  ArrowLeft: 'left',
  ArrowRight: 'right',
  ArrowUp: 'up',
  ArrowDown: 'down',
};

/**
 * Column count of the board grid at the current breakpoint, mirroring the
 * Tailwind `grid-cols-*` classes (7 / 8 / 10 / 13). Falls back to 13 when
 * `matchMedia` is unavailable (SSR).
 */
function boardColumns(): number {
  if (typeof window === 'undefined' || !window.matchMedia) return 13;
  if (window.matchMedia('(min-width: 1024px)').matches) return 13;
  if (window.matchMedia('(min-width: 768px)').matches) return 10;
  if (window.matchMedia('(min-width: 640px)').matches) return 8;
  return 7;
}

type MemoryGame = ReturnType<typeof useMemoryGame>;

function useMemoryView({ game, t }: { game: MemoryGame; t: TFunction }) {
  const { state } = game;
  const phaseNames = usePhaseNames('memory', MEMORY_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();

  // **見た「位置」ではなく見た「札」を覚える。**目のマークは「前に見た」としか
  // 言わず、そこに何があったかを思い出す助けにはならなかった (#4775)。
  // サーバは裏返った札の値を送らない (不正防止として正しい) ので、これは
  // プレイヤー自身が画面で見た情報の記録であって、追加の情報ではない。
  const [seen, setSeen] = useState<ReadonlyMap<number, Card>>(() => new Map());

  // Captured-pairs panel is expanded on desktop and collapsed on mobile so the
  // board grid keeps its full height on small screens (#3028).
  const [pairsOpen, setPairsOpen] = useState<boolean>(
    () => typeof window !== 'undefined' && window.matchMedia('(min-width: 1024px)').matches,
  );

  useEffect(() => {
    if (!state) return;
    setSeen((prev) => {
      let changed = false;
      const next = new Map(prev);
      state.board.forEach((bc, idx) => {
        if (bc.faceUp && bc.card && !next.has(idx)) {
          next.set(idx, bc.card);
          changed = true;
        }
      });
      return changed ? next : prev;
    });
  }, [state]);

  // Roving-focus board navigation (#3029): arrow keys move a single focus index
  // across the card grid; Enter/Space still natively activate the focused button.
  const boardRef = useRef<HTMLDivElement>(null);
  const [focusedIdx, setFocusedIdx] = useState(0);
  const [cols, setCols] = useState<number>(() => boardColumns());

  useEffect(() => {
    const onResize = () => setCols(boardColumns());
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

  // Keep the roving tab-stop on a flippable cell: taken (removed) and face-up
  // cards are disabled/hidden and cannot hold focus.
  useEffect(() => {
    const board = state?.board;
    if (!board) return;
    const cur = board[focusedIdx];
    if (cur && !cur.taken && !cur.faceUp) return;
    const firstIdx = board.findIndex((c) => !c.taken && !c.faceUp);
    const focusedCardTaken = cur === undefined ? false : cur.taken;
    if (firstIdx >= 0) {
      setFocusedIdx(firstIdx);
      if (focusedCardTaken) {
        boardRef.current?.querySelector<HTMLButtonElement>(`[data-testid="board-${firstIdx.toString()}"]`)?.focus();
      }
    } else if (focusedCardTaken) {
      boardRef.current?.focus();
    }
  }, [state?.board, focusedIdx]);

  const focusCell = useCallback((idx: number) => {
    setFocusedIdx(idx);
    boardRef.current?.querySelector<HTMLButtonElement>(`[data-testid="board-${idx.toString()}"]`)?.focus();
  }, []);

  const handleBoardKeyDown = useCallback(
    (e: ReactKeyboardEvent<HTMLDivElement>) => {
      const dir = ARROW_DIRS[e.key];
      const board = state?.board;
      if (!dir || !board) return;
      e.preventDefault();
      const skip = (i: number) => {
        const bc = board[i];
        return !bc || bc.taken || bc.faceUp;
      };
      const target = moveFocus(focusedIdx, dir, cols, board.length, skip);
      if (target !== focusedIdx) focusCell(target);
    },
    [state?.board, cols, focusedIdx, focusCell],
  );

  // Announce the flip outcome (card faces + match/mismatch) to screen readers.
  const flipAnnounce = useMemo(() => {
    if (!state || state.phase !== MemoryPhase.RESULT) return '';
    const c1 = state.board[state.firstFlipPos]?.card;
    const c2 = state.board[state.secondFlipPos]?.card;
    if (!c1 || !c2) return '';
    return t(state.lastMatchResult ? 'announce.match' : 'announce.mismatch', {
      first: cardAlt(c1),
      second: cardAlt(c2),
      firstPosition: state.firstFlipPos + 1,
      secondPosition: state.secondFlipPos + 1,
    });
  }, [state, t]);

  // 1枚めくった時点で、既に見て覚えている一致先があればその位置を指す。
  // レジストリの hintFactories は state しか受け取れないので、記憶を渡せる
  // ここで getMemoryHint を呼ぶ。**ヒントもハイライトも同じ位置を読む。**
  const knownMatchIdx = useMemo(() => (state ? memoryKnownMatch(state.board, seen) : null), [state, seen]);

  return {
    phaseNames,
    cardWidth,
    seen,
    setSeen,
    pairsOpen,
    setPairsOpen,
    boardRef,
    focusedIdx,
    handleBoardKeyDown,
    flipAnnounce,
    knownMatchIdx,
  };
}

type MemoryView = ReturnType<typeof useMemoryView>;
type MemoryContext = GamePageContext<MemoryResponse, MemoryGame, MemoryView>;

function Board({ game, state, t, view, frontendHint, busy }: MemoryContext) {
  const {
    seen,
    pairsOpen,
    setPairsOpen,
    boardRef,
    focusedIdx,
    handleBoardKeyDown,
    flipAnnounce,
    knownMatchIdx,
    cardWidth,
  } = view;
  const { hintEnabled: frontendHintEnabled } = frontendHint;
  const isFlip1 = state.phase === MemoryPhase.FLIP1;
  const isFlip2 = state.phase === MemoryPhase.FLIP2;
  const isGameEnd = state.phase === MemoryPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isFlip1 || isFlip2) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const highestPairCount = Math.max(...state.players.map((player) => player.pairCount));
  const leaders = state.players.filter((player) => player.pairCount === highestPairCount);
  const resultNames = leaders.map((player) => playerName(player.id, player.isHuman)).join(t('listSeparator'));

  const frontendHintResult = useMemo(
    () => (frontendHintEnabled && state ? getMemoryHint(state, knownMatchIdx) : null),
    [frontendHintEnabled, state, knownMatchIdx],
  );

  return (
    <>
      {/* Player scores – compact inline layout to maximise board visibility */}
      <div
        className="my-1 px-2 py-1 rounded bg-black/30 text-ds-text-primary text-sm flex flex-wrap items-center gap-y-0.5 lg:shrink-0"
        data-tutorial="mem-score-table"
        role="status"
        aria-label={t('scores')}
      >
        {state.players.map((p, idx) => (
          <span key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
            {idx > 0 && (
              <span className="text-ds-text-muted mr-3" aria-hidden="true">
                |
              </span>
            )}
            {t('scoreLine', { name: playerName(p.id, p.isHuman), count: p.pairCount })}
          </span>
        ))}
        <span className="ml-3 inline-block rounded bg-ds-surface px-2 py-1 text-ds-text-primary tabular-nums">
          {t('remainingPairs', {
            remaining: state.board.filter((card) => !card.taken).length / 2,
            total: state.board.length / 2,
          })}
        </span>
      </div>

      {isGameEnd && (
        <p className="my-2 text-center font-bold text-ds-text-primary" data-testid="memory-game-result">
          {leaders.length === 1
            ? t('winnerResult', { name: resultNames, count: highestPairCount })
            : t('tieResult', { names: resultNames, count: highestPairCount })}
        </p>
      )}

      {/* Captured pairs – mini cards per player. Collapsible on mobile so
          the board grid keeps its full height (#3028). */}
      <details
        className="my-1 px-2 py-1 rounded bg-black/20 text-ds-text-primary text-sm lg:shrink-0"
        data-testid="mem-captured"
        open={pairsOpen}
        onToggle={(e) => setPairsOpen(e.currentTarget.open)}
      >
        <summary className="cursor-pointer select-none font-bold py-0.5">{t('capturedPairs')}</summary>
        <div className="mt-1 flex flex-col gap-1">
          {state.players.map((p) => (
            <div
              key={p.id}
              data-testid={`mem-captured-${p.id.toString()}`}
              className="flex items-center gap-1 flex-wrap"
            >
              <span className={`shrink-0 ${p.isHuman ? 'text-ds-accent' : 'text-ds-text-muted'}`}>
                {playerName(p.id, p.isHuman)}
              </span>
              {p.pairs.length === 0 ? (
                <span className="text-ds-text-muted">{t('noCapturedPairs')}</span>
              ) : (
                p.pairs.map((c, i) => (
                  <AnimatedCard
                    key={`p${p.id.toString()}-pair-${i.toString()}-${c.design}${c.value.toString()}`}
                    card={c}
                    width={Math.max(20, Math.round(cardWidth * 0.5))}
                    silent
                  />
                ))
              )}
            </div>
          ))}
        </div>
      </details>

      {/* Board: responsive grid (7/8/10/13 columns). Narrow screens drop the 2/3 aspect
          so all 8 rows fit the play area; on lg it fills the remaining height. */}
      <div
        className="my-3 max-sm:my-1 lg:my-1 p-1 rounded bg-black/40 lg:flex-1 lg:min-h-0 lg:overflow-hidden"
        data-tutorial="mem-board"
      >
        {/* biome-ignore lint/a11y/noStaticElementInteractions: keydown only routes arrow keys for roving focus; the real controls are the child <button>s */}
        <div
          ref={boardRef}
          tabIndex={-1}
          onKeyDown={handleBoardKeyDown}
          className="grid grid-cols-7 gap-0.5 sm:gap-1 sm:grid-cols-8 md:grid-cols-10 lg:grid-cols-13 lg:grid-rows-4 lg:h-full"
        >
          {state.board.map((bc, idx) => {
            const wasVisited = !bc.faceUp && !bc.taken && seen.has(idx);
            const isKnownMatch = frontendHintEnabled && idx === knownMatchIdx;
            return (
              <button
                type="button"
                key={`board-${idx.toString()}`}
                data-testid={`board-${idx.toString()}`}
                aria-label={
                  bc.faceUp && bc.card
                    ? cardAlt(bc.card)
                    : // 目のマークを読み上げるなら、より強い「一致がある」も読み上げる。
                      // 見ただけの印だけ伝えて一致を伏せると、スクリーンリーダー利用者だけが
                      // 情報の少ない側に置かれる。
                      isKnownMatch
                      ? `${t('cardFaceDown', { position: idx + 1 })} (${t('knownMatchMark')})`
                      : wasVisited
                        ? `${t('cardFaceDown', { position: idx + 1 })} (${t('visitedMark')})`
                        : t('cardFaceDown', { position: idx + 1 })
                }
                disabled={busy || !isHumanTurn || bc.taken || bc.faceUp}
                tabIndex={idx === focusedIdx ? 0 : -1}
                onClick={() => game.handleFlip(idx)}
                className={`memory-card relative aspect-[2/3] max-sm:aspect-auto max-sm:h-11 min-h-[44px] min-w-[44px] lg:aspect-auto rounded ${focusRingWhite} ${
                  bc.taken
                    ? 'hidden'
                    : bc.faceUp
                      ? 'bg-white ring-2 ring-ds-warning shadow-lg shadow-ds-warning/30'
                      : isKnownMatch
                        ? 'bg-ds-info border border-white/10 ring-2 ring-ds-success motion-safe:animate-pulse'
                        : wasVisited
                          ? 'bg-ds-info border border-white/10 ring-1 ring-ds-accent hover:ring-ds-warning'
                          : 'bg-ds-info border border-white/10 hover:ring-1 hover:ring-ds-warning'
                } transition-all`}
                data-known-match={isKnownMatch || undefined}
              >
                <div className={`memory-card-inner${bc.faceUp ? ' flipped' : ''}`}>
                  <div className="memory-card-back">
                    <img src="/images/z01.png" alt="" className="w-full h-full object-contain rounded" />
                    {wasVisited && (
                      <span
                        data-testid={`board-visited-${idx.toString()}`}
                        aria-hidden="true"
                        className="absolute inset-0 rounded bg-black/25 pointer-events-none flex items-start justify-end p-0.5"
                      >
                        <span className="text-sm leading-none" title={t('visitedMark')}>
                          {'\u{1F441}'}
                        </span>
                      </span>
                    )}
                  </div>
                  <div className="memory-card-front">
                    {bc.card && <AnimatedCard card={bc.card} width={cardWidth} />}
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Polite live region announcing flip results to screen readers (#3029) */}
      <span className="sr-only" role="status" aria-live="polite" data-testid="mem-flip-announce">
        {flipAnnounce}
      </span>

      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHintResult} enabled={frontendHintEnabled} t={t} />
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: MemoryContext) {
  const isResult = state.phase === MemoryPhase.RESULT;
  return (
    <>
      {isResult && (
        <div data-tutorial="mem-next-button">
          <button type="button" className={btnSuccess} onClick={game.handleNext} disabled={busy}>
            {t('nextButton')}
          </button>
        </div>
      )}
    </>
  );
}

const memoryDefinition: GameDefinition<MemoryResponse, MemoryGame, Parameters<typeof memoryApi.exec>, MemoryView> = {
  key: 'memory',
  useGame: useMemoryGame,
  useView: useMemoryView,
  tutorial: MEM_TUTORIAL_STEPS,
  skeleton: {
    kind: 'card-grid',
    count: 52,
    cols: 'grid-cols-7 sm:grid-cols-8 md:grid-cols-10 lg:grid-cols-13',
    aspectRatio: 'aspect-[2/3] max-sm:aspect-auto max-sm:h-11 lg:aspect-auto',
    gridClassName: 'lg:grid-rows-4 lg:h-full',
    topPills: 4,
  },
  cli: ({ state, view, frontendHint }) => ({
    parseCommand: parseMemoryCommand,
    formatResponse: formatMemoryState,
    helpText: MEMORY_HELP,
    localCommand: hintLocalCommand(frontendHint.hintEnabled && state ? getMemoryHint(state, view.knownMatchIdx) : null),
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === MemoryPhase.GAME_END || state.gameEndFlag;
    const highestPairCount = Math.max(...state.players.map((player) => player.pairCount));
    const leaders = state.players.filter((player) => player.pairCount === highestPairCount);
    const humanIsSoleWinner = leaders.length === 1 && leaders[0]?.isHuman === true;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && humanIsSoleWinner,
    };
  },
  shell: ({ state }) => {
    const isFlip1 = state.phase === MemoryPhase.FLIP1;
    const isFlip2 = state.phase === MemoryPhase.FLIP2;
    const isGameEnd = state.phase === MemoryPhase.GAME_END || state.gameEndFlag;
    const isHumanTurn = (isFlip1 || isFlip2) && state.players[state.currentPlayerIdx]?.isHuman === true;
    const highestPairCount = Math.max(...state.players.map((player) => player.pairCount));
    const leaders = state.players.filter((player) => player.pairCount === highestPairCount);
    const humanIsNotAmongLeaders = !leaders.some((player) => player.isHuman);
    return {
      isHumanTurn,
      lossShow: !!isGameEnd && leaders.length === 1 && humanIsNotAmongLeaders,
    };
  },
  reset: ({ game, view }) => {
    view.setSeen(new Map());
    void game.exec('reset', undefined, {
      cpuDifficulty: game.memoryConfig.cpuDifficulty,
      pairCount: game.memoryConfig.pairCount,
    });
  },
  settings: ({ game, t, tc, frontendHint, view, hideActionLog }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.memoryConfig.cpuDifficulty,
          // **強さの差は「CPU がどれだけ覚えているか」の差。** domain の
          // retentionChance / decayRate が難易度ごとに 30%/60%/95% 保持、
          // 15%/5%/1% 減衰と切り替わるのに、選択肢は Easy/Normal/Hard と
          // しか出ておらず、何が変わるのか分からなかった (#5492)。
          // 数値は出さない -- 実装が変われば嘘になる。順序だけを言う。
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t('settings.difficultyOption', {
              level: t(`settings.${o.label.toLowerCase()}`),
              memory: t(`settings.memory${o.label}`),
            }),
          })),
          tooltip: t('settings.difficultyTooltip'),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'pairCount',
          label: t('settings.pairCount'),
          value: game.memoryConfig.pairCount,
          options: PAIR_COUNT_OPTIONS.map((pairs) => ({
            value: pairs,
            label: t('settings.pairCountOption', { pairs, cards: pairs * 2 }),
          })),
          // Re-deal at once: this changes the board's shape, so deferring it
          // to the next reset would read as the setting doing nothing.
          onSelect: (v) => {
            const pairs = Number(v);
            game.handleConfigChange('pairCount', v);
            hideActionLog();
            view.setSeen(new Map());
            void game.exec('reset', undefined, { cpuDifficulty: game.memoryConfig.cpuDifficulty, pairCount: pairs });
          },
        },
        {
          type: 'select',
          id: 'autoNextDelayMs',
          label: t('settings.autoNextDelay'),
          value: game.autoNextDelayMs,
          options: AUTO_NEXT_DELAY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.autoNextDelayOption.${o.label}`),
          })),
          onSelect: (v) => game.setAutoNextDelayMs(Number(v)),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: 'lg:overflow-hidden lg:flex lg:flex-col pt-3 lg:pt-1 px-4 lg:px-8',
  shortcuts: ({ game, state }) => [
    { key: 'n', action: game.handleNext, enabled: state.phase === MemoryPhase.RESULT, label: 'next' },
  ],
  shortcutsTestId: 'memory-kbd-shortcuts',
  resetTutorialId: 'mem-reset-button',
  actionsClassName: 'flex gap-2 items-center',
  Board,
  FooterActions,
};

/** Renders the Memory card matching game page with board grid and scores. */
export const MemoryPage = createGamePage(memoryDefinition);
