import { type KeyboardEvent as ReactKeyboardEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { memoryApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import {
  AUTO_NEXT_DELAY_OPTIONS,
  CPU_DIFFICULTY_OPTIONS,
  PAIR_COUNT_OPTIONS,
  useMemoryGame,
} from '../hooks/useMemoryGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, MemoryResponse } from '../types/card';
import { MemoryPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MEMORY_HELP, parseMemoryCommand } from '../utils/cli/commands/memoryCommands';
import { formatMemoryState } from '../utils/cli/formatters/memoryFormatter';
import type { CliGameConfig } from '../utils/cli/types';
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

/** Renders the Memory card matching game page with board grid and scores. */
export const MemoryPage = withTutorial(MemoryPageContent, 'memory', MEM_TUTORIAL_STEPS);
/** Inner content of the Memory page, wrapped by TutorialProvider. */
function MemoryPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('memory');
  const { cardWidth } = useCardDimensions();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    memoryConfig,
    handleConfigChange,
    autoNextDelayMs,
    setAutoNextDelayMs,
    handleFlip,
    handleNext,
  } = useMemoryGame();
  const { hintEnabled: frontendHintEnabled, setHintEnabled: setFrontendHintEnabled } = useGameHint('memory', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('memory');
  const cliConfig: CliGameConfig<MemoryResponse, Parameters<typeof memoryApi.exec>> = useMemo(
    () => ({
      gameName: 'memory',
      parseCommand: parseMemoryCommand,
      formatResponse: formatMemoryState,
      helpText: MEMORY_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isResultForKbd = state?.phase === MemoryPhase.RESULT;

  const actionBindings = useMemo(
    () => [{ key: 'n', action: handleNext, enabled: isResultForKbd, label: 'next' }],
    [handleNext, isResultForKbd],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  const phaseNames = usePhaseNames('memory', MEMORY_PHASE_KEYS);

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
    if (firstIdx >= 0) setFocusedIdx(firstIdx);
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
    });
  }, [state, t]);

  // 1枚めくった時点で、既に見て覚えている一致先があればその位置を指す。
  // レジストリの hintFactories は state しか受け取れないので、記憶を渡せる
  // ここで getMemoryHint を呼ぶ。**ヒントもハイライトも同じ位置を読む。**
  const knownMatchIdx = useMemo(() => (state ? memoryKnownMatch(state.board, seen) : null), [state, seen]);
  const frontendHint = useMemo(
    () => (frontendHintEnabled && state ? getMemoryHint(state, knownMatchIdx) : null),
    [frontendHintEnabled, state, knownMatchIdx],
  );

  const handleManualReset = useCallback(() => {
    hideActionLog();
    setSeen(new Map());
    void exec('reset', undefined, { cpuDifficulty: memoryConfig.cpuDifficulty, pairCount: memoryConfig.pairCount });
  }, [exec, hideActionLog, memoryConfig.cpuDifficulty, memoryConfig.pairCount]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="memory"
        layout={{
          kind: 'card-grid',
          count: 52,
          cols: 'grid-cols-7 sm:grid-cols-8 md:grid-cols-10 lg:grid-cols-13',
          aspectRatio: 'aspect-[2/3] max-sm:aspect-auto max-sm:h-11 lg:aspect-auto',
          gridClassName: 'lg:grid-rows-4 lg:h-full',
          topPills: 4,
        }}
      />
    );

  const isFlip1 = state.phase === MemoryPhase.FLIP1;
  const isFlip2 = state.phase === MemoryPhase.FLIP2;
  const isResult = state.phase === MemoryPhase.RESULT;
  const isGameEnd = state.phase === MemoryPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isFlip1 || isFlip2) && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.memory')}
      gameThemeBg={gameTheme.memory.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/memory"
      gameEndFlag={!!isGameEnd}
      winShow={!!state.gameEndFlag}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Settings */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: memoryConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pairCount',
                    label: t('settings.pairCount'),
                    value: memoryConfig.pairCount,
                    options: PAIR_COUNT_OPTIONS.map((pairs) => ({
                      value: pairs,
                      label: t('settings.pairCountOption', { pairs, cards: pairs * 2 }),
                    })),
                    // Re-deal at once: this changes the board's shape, so deferring it
                    // to the next reset would read as the setting doing nothing.
                    onSelect: (v) => {
                      const pairs = Number(v);
                      handleConfigChange('pairCount', v);
                      hideActionLog();
                      setSeen(new Map());
                      void exec('reset', undefined, { cpuDifficulty: memoryConfig.cpuDifficulty, pairCount: pairs });
                    },
                  },
                  {
                    type: 'select',
                    id: 'autoNextDelayMs',
                    label: t('settings.autoNextDelay'),
                    value: autoNextDelayMs,
                    options: AUTO_NEXT_DELAY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.autoNextDelayOption.${o.label}`),
                    })),
                    onSelect: (v) => setAutoNextDelayMs(Number(v)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Content area – scroll on mobile, fit-to-viewport on desktop */}
          <div className="flex-1 overflow-y-auto lg:overflow-hidden lg:flex lg:flex-col pt-3 lg:pt-1 px-4 lg:px-8">
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
                    <span className="text-ds-text-muted/80 mr-3" aria-hidden="true">
                      |
                    </span>
                  )}
                  {t('scoreLine', { name: playerName(p.id, p.isHuman), count: p.pairCount })}
                </span>
              ))}
            </div>

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
                      <span className="text-ds-text-muted/70">{t('noCapturedPairs')}</span>
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
                      disabled={loading || !isHumanTurn || bc.taken || bc.faceUp}
                      tabIndex={idx === focusedIdx ? 0 : -1}
                      onClick={() => handleFlip(idx)}
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
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.memory.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center">
              {isResult && (
                <div data-tutorial="mem-next-button">
                  <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
                    {t('nextButton')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="mem-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="memory-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
