import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { baccaratApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoadmapTrendBar } from '../components/RoadmapTrendBar';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BaccaratResponse, BaccaratSideBetResult, Card } from '../types/card';
import { BaccaratBetType, BaccaratPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { computeBaccaratShoeStats } from '../utils/baccaratStats';
import { BACCARAT_HELP, parseBaccaratCommand } from '../utils/cli/commands/baccaratCommands';
import { formatBaccaratState } from '../utils/cli/formatters/baccaratFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';

const BET_TYPE_LABELS: Record<number, string> = {
  [BaccaratBetType.PLAYER]: 'betType.player',
  [BaccaratBetType.BANKER]: 'betType.banker',
  [BaccaratBetType.TIE]: 'betType.tie',
};

/** Baccarat tutorial step definitions. */
const BAC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bac-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bac-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bac-banker-hand"]',
    messageKey: 'tutorial.bankerHand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bac-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const ROAD_PLAYER = 0;
const ROAD_BANKER = 1;
const ROAD_MAX_ROWS = 6;

/**
 * Baccarat hand value for the visible card slice. Aces=1, 2-9=face, 10/J/Q/K=0,
 * total mod 10. Mirrors the server's scoring rule and lets us paint an
 * intermediate header total during the staged reveal (#1892) without spoiling
 * the third card.
 */
function baccaratHandValue(cards: readonly Card[]): number {
  let total = 0;
  for (const c of cards) total += c.value >= 10 ? 0 : c.value;
  return total % 10;
}

function BigRoadGrid({ history }: { history: number[] }) {
  const { t } = useGamePageSetup('baccarat');
  // Auto-scroll the road to the latest (right-most) column whenever the
  // history grows, so the most recent results are always in view without
  // manual scrolling. Hooks must run unconditionally, before any early return.
  const scrollRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = scrollRef.current;
    if (el && history.length > 0) el.scrollLeft = el.scrollWidth;
  }, [history.length]);

  if (history.length === 0) return null;

  // Build columns: new column on side change, ties mark previous cell
  const columns: { result: number; tie: boolean }[][] = [];
  let currentCol: { result: number; tie: boolean }[] = [];
  let lastSide: number | null = null;
  // Track the most recently placed non-tie cell so it can be highlighted.
  let latestCell: { result: number; tie: boolean } | null = null;

  for (const result of history) {
    if (result !== ROAD_PLAYER && result !== ROAD_BANKER) {
      // Tie: mark on last cell if exists
      if (currentCol.length > 0) {
        currentCol[currentCol.length - 1].tie = true;
      } else if (columns.length > 0) {
        const prevCol = columns[columns.length - 1];
        prevCol[prevCol.length - 1].tie = true;
      }
      continue;
    }
    if (lastSide !== null && result !== lastSide) {
      columns.push(currentCol);
      currentCol = [];
    }
    const cell = { result, tie: false };
    currentCol.push(cell);
    latestCell = cell;
    lastSide = result;
  }
  if (currentCol.length > 0) {
    columns.push(currentCol);
  }

  if (columns.length === 0) return null;

  // Dragon tail: if column > ROAD_MAX_ROWS, overflow goes right
  const grid: ({ result: number; tie: boolean } | null)[][] = [];
  let colOffset = 0;
  for (const col of columns) {
    for (let row = 0; row < col.length; row++) {
      const actualRow = row < ROAD_MAX_ROWS ? row : ROAD_MAX_ROWS - 1;
      const actualCol = row < ROAD_MAX_ROWS ? colOffset : colOffset + (row - ROAD_MAX_ROWS + 1);
      while (grid.length <= actualRow) grid.push([]);
      while (grid[actualRow].length <= actualCol) grid[actualRow].push(null);
      grid[actualRow][actualCol] = col[row];
    }
    const colWidth = col.length > ROAD_MAX_ROWS ? 1 + (col.length - ROAD_MAX_ROWS) : 1;
    colOffset += colWidth;
  }

  // Ensure all rows have same length
  const maxCols = Math.max(...grid.map((r) => r.length));
  for (const row of grid) {
    while (row.length < maxCols) row.push(null);
  }

  // Every cell is a coloured circle, so the run is unreadable without colour.
  // The CUI already prints the same sequence as P/B/T (baccaratHistorySymbols);
  // this is that line, spelled out for a screen reader.
  const spoken = history
    .map((r, i) =>
      t('road.spokenEntry', {
        n: i + 1,
        side: r === ROAD_PLAYER ? t('road.sidePlayer') : r === ROAD_BANKER ? t('road.sideBanker') : t('road.sideTie'),
      }),
    )
    .join('、');

  return (
    <div className="mb-4 w-full max-w-md" data-testid="big-road">
      <div className="sr-only" data-testid="big-road-summary">
        {spoken}
      </div>
      <div
        ref={scrollRef}
        data-testid="big-road-scroll"
        className="overflow-x-auto rounded border border-ds-border-subtle bg-ds-surface-elevated"
      >
        <div
          className="inline-grid gap-px bg-ds-surface-elevated"
          style={{ gridTemplateColumns: `repeat(${maxCols}, 28px)` }}
        >
          {grid.flatMap((row, ri) =>
            row.map((cell, ci) => {
              const cellKey = `r${String(ri)}c${String(ci)}`;
              const isLatest = cell !== null && cell === latestCell;
              return (
                <div key={cellKey} className="w-7 h-7 flex items-center justify-center bg-ds-surface-elevated relative">
                  {cell && (
                    <>
                      <span
                        className={`w-5 h-5 rounded-full inline-block ${cell.result === ROAD_PLAYER ? 'bg-ds-info' : 'bg-ds-error'} ${isLatest ? 'ring-2 ring-ds-accent ring-offset-1 ring-offset-ds-surface-elevated' : ''}`}
                      />
                      {cell.tie && (
                        <span className="absolute inset-0 flex items-center justify-center text-ds-success font-bold text-xs">
                          /
                        </span>
                      )}
                    </>
                  )}
                </div>
              );
            }),
          )}
        </div>
      </div>
    </div>
  );
}

/**
 * Renders a compact statistics row for the shoe history: Player / Banker / Tie
 * counts with appearance rates, plus the current side streak. Returns null when
 * the history is empty. Aggregation is delegated to `computeBaccaratShoeStats`.
 */
function ShoeStatsPanel({
  history,
  t,
}: {
  history: number[];
  t: (key: string, params?: Record<string, unknown>) => string;
}) {
  if (history.length === 0) return null;
  const stats = computeBaccaratShoeStats(history);
  const streakLabel =
    stats.streakSide === ROAD_PLAYER
      ? t('betType.player')
      : stats.streakSide === ROAD_BANKER
        ? t('betType.banker')
        : null;

  return (
    <div className="mb-2 text-xs text-ds-text-primary" data-testid="baccarat-shoe-stats">
      <div className="flex flex-wrap items-center justify-center gap-x-3 gap-y-1">
        <span className="text-ds-info">{t('stats.player', { n: stats.playerCount, pct: stats.playerPct })}</span>
        <span className="text-ds-error">{t('stats.banker', { n: stats.bankerCount, pct: stats.bankerPct })}</span>
        <span className="text-ds-success">{t('stats.tie', { n: stats.tieCount, pct: stats.tiePct })}</span>
        {streakLabel && (
          <span className="font-bold text-ds-warning">
            {t('stats.streak', { side: streakLabel, k: stats.streakCount })}
          </span>
        )}
      </div>
    </div>
  );
}

function SideBetResultsDisplay({
  results,
  t,
}: {
  results: BaccaratSideBetResult[];
  t: (key: string, params?: Record<string, unknown>) => string;
}) {
  if (results.length === 0) return null;
  return (
    <div className="mb-2 text-center" data-testid="side-bet-results">
      {results.map((r) => (
        <div key={r.betType} className="text-ds-text-primary text-sm">
          {r.betType === 1 ? t('sideBet.playerPair') : t('sideBet.bankerPair')}:{' '}
          {r.resultType === 1 ? t('sideBet.pair') : t('sideBet.noPair')} ({t('label.payout', { payout: r.payout })})
        </div>
      ))}
    </div>
  );
}

/** Renders the Baccarat game page with betting and result display. */
export const BaccaratPage = withTutorial(BaccaratPageContent, 'baccarat', BAC_TUTORIAL_STEPS);
/** Inner content of the Baccarat page, wrapped by TutorialProvider. */
function BaccaratPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('baccarat');

  const [betAmount, setBetAmount] = useState(100);
  const [betType, setBetType] = useState<number>(BaccaratBetType.PLAYER);
  const [playerPairBet, setPlayerPairBet] = useState(0);
  const [bankerPairBet, setBankerPairBet] = useState(0);
  // Snapshot of the last accepted bet, used to power the one-click Rebet button at end-phase.
  const [lastBet, setLastBet] = useState<{
    amount: number;
    type: number;
    pp: number;
    bp: number;
  } | null>(null);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(baccaratApi.exec);
  const hintState = useMemo(() => (state ? { ...state, betType } : null), [state, betType]);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('baccarat', hintState);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('baccarat');
  const cliConfig: CliGameConfig<BaccaratResponse, Parameters<typeof baccaratApi.exec>> = useMemo(
    () => ({
      gameName: 'baccarat',
      parseCommand: parseBaccaratCommand,
      formatResponse: formatBaccaratState,
      helpText: BACCARAT_HELP,
      localCommand: hintLocalCommand(hint),
    }),
    [hint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === BaccaratPhase.BET;
  const isEndPhase = state?.phase === BaccaratPhase.END;

  // Staged reveal for the showdown so the third-card rule isn't a black box (#1892).
  // Steps: 1 = initial 2+2 cards, 2 = player's 3rd, 3 = banker's 3rd, 4 = payout/result.
  // The hand "signature" string keys the effect so each new round restarts the animation
  // without flicker mid-render.
  const playerHand = state?.playerHand ?? [];
  const bankerHand = state?.bankerHand ?? [];
  const handSignature = isEndPhase ? `${playerHand.length}:${bankerHand.length}:${state?.payout ?? 0}` : 'bet';
  const [revealStep, setRevealStep] = useState(0);
  // biome-ignore lint/correctness/useExhaustiveDependencies: handSignature already encodes playerHand.length and bankerHand.length — listing them would re-fire the effect for the same hand.
  useEffect(() => {
    if (!isEndPhase) {
      setRevealStep(0);
      return;
    }
    setRevealStep(1);
    const timers: ReturnType<typeof setTimeout>[] = [];
    let delay = 600;
    if (playerHand.length === 3) {
      const d = delay;
      timers.push(setTimeout(() => setRevealStep((s) => Math.max(s, 2)), d));
      delay += 600;
    }
    if (bankerHand.length === 3) {
      const d = delay;
      timers.push(setTimeout(() => setRevealStep((s) => Math.max(s, 3)), d));
      delay += 600;
    }
    timers.push(setTimeout(() => setRevealStep((s) => Math.max(s, 4)), delay));
    return () => {
      for (const timer of timers) clearTimeout(timer);
    };
  }, [isEndPhase, handSignature]);
  const playerCardsShown = revealStep >= 2 ? playerHand.length : Math.min(playerHand.length, 2);
  const bankerCardsShown = revealStep >= 3 ? bankerHand.length : Math.min(bankerHand.length, 2);
  // showResultDetails is only consumed under `isEndPhase && ...`, so the !isEndPhase branch
  // was unreachable; collapse to the meaningful condition.
  const showResultDetails = revealStep >= 4;

  // Visible baccarat hand value for the currently revealed slice, so the header
  // total doesn't spoil the third card before it animates in. A=1, 2-9=face,
  // 10/J/Q/K=0, total mod 10.
  const visiblePlayerValue = isEndPhase ? baccaratHandValue(playerHand.slice(0, playerCardsShown)) : 0;
  const visibleBankerValue = isEndPhase ? baccaratHandValue(bankerHand.slice(0, bankerCardsShown)) : 0;

  const handleBet = useCallback(() => {
    setLastBet({ amount: betAmount, type: betType, pp: playerPairBet, bp: bankerPairBet });
    execApi('bet', betAmount, betType, playerPairBet, bankerPairBet);
  }, [execApi, betAmount, betType, playerPairBet, bankerPairBet]);

  const handleReset = useCallback(() => {
    execApi('reset');
  }, [execApi]);

  const totalLastBet = lastBet ? lastBet.amount + lastBet.pp + lastBet.bp : 0;
  const canRebet = lastBet !== null && totalLastBet > 0 && state !== null && totalLastBet <= state.chips;

  const handleRebet = useCallback(async () => {
    if (!lastBet) return;
    await execApi('reset');
    await execApi('bet', lastBet.amount, lastBet.type, lastBet.pp, lastBet.bp);
  }, [execApi, lastBet]);

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: handleBet, enabled: isBetPhase, label: 'bet' },
      { key: 'r', action: handleReset, enabled: isEndPhase, label: 'reset' },
      // Power-user shortcut: 'e' replays the last bet at end phase (consistent with the 'r' reset binding).
      { key: 'e', action: handleRebet, enabled: isEndPhase && canRebet, label: 'rebet' },
    ],
    [handleBet, handleReset, handleRebet, isBetPhase, isEndPhase, canRebet],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="baccarat" layout={{ kind: 'casino-table', sections: [2, 2] }} />;

  const handleClearHistory = () => {
    execApi('clearhistory');
  };

  return (
    <GamePageShell
      title={tc('nav.baccarat')}
      gameThemeBg={gameTheme.baccarat.bg}
      phaseName={isBetPhase ? t('phase.bet') : t('phase.end')}
      gamePath="/baccarat"
      gameEndFlag={isEndPhase}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>{t('label.chips', { chips: state.chips })}</span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div
            data-testid="card-area"
            className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
              .filter(Boolean)
              .join(' ')}
          >
            {isBetPhase && state.playerHand.length === 0 && (
              <div className="flex flex-col items-center justify-center py-6 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
                <details className="bg-black/30 rounded-lg w-full max-w-sm">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <ul className="text-ds-text-muted text-sm space-y-1 px-4 pb-3">
                    {(['playerWin', 'bankerWin', 'tie', 'playerPair', 'bankerPair'] as const).map((key) => (
                      <li key={key}>{t(`payoutRef.${key}`)}</li>
                    ))}
                  </ul>
                </details>
              </div>
            )}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Player Hand — staged reveal hides the 3rd card until step 2 */}
            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="bac-player-hand">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}{' '}
                  {t('label.value', { value: isEndPhase ? visiblePlayerValue : state.playerHandValue })}
                </div>
                <div className="flex justify-center gap-2" data-testid="bac-player-cards">
                  {state.playerHand.slice(0, playerCardsShown).map((card, i) => (
                    <AnimatedCard key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {/* Banker Hand — staged reveal hides the 3rd card until step 3 */}
            {state.bankerHand.length > 0 && (
              <div className="mb-4" data-tutorial="bac-banker-hand">
                <div className="text-ds-error font-bold text-center mb-1">
                  <span aria-hidden="true">🔴</span> {t('banker')}{' '}
                  {t('label.value', { value: isEndPhase ? visibleBankerValue : state.bankerHandValue })}
                </div>
                <div className="flex justify-center gap-2" data-testid="bac-banker-cards">
                  {state.bankerHand.slice(0, bankerCardsShown).map((card, i) => (
                    <AnimatedCard key={`b-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {/* Payout info — gated behind the final reveal step */}
            {isEndPhase && showResultDetails && (
              <div className="text-ds-text-primary text-center font-bold mb-2" data-testid="bac-payout">
                {t('label.payout', { payout: state.payout })}
              </div>
            )}

            {/* Side bet results — gated behind the final reveal step */}
            {isEndPhase && showResultDetails && state.sideBetResults.length > 0 && (
              <SideBetResultsDisplay results={state.sideBetResults} t={t} />
            )}

            {/* Big Road */}
            <div className="flex flex-col items-center">
              {state.history.length > 0 && (
                <div className="text-ds-text-primary text-sm font-bold mb-1">{t('road.title')}</div>
              )}
              <ShoeStatsPanel history={state.history} t={t} />
              <RoadmapTrendBar
                history={state.history}
                leftCode={ROAD_PLAYER}
                rightCode={ROAD_BANKER}
                leftLabel={t('betType.player')}
                rightLabel={t('betType.banker')}
                testId="baccarat-trend-bar"
              />
              <BigRoadGrid history={state.history} />
              {state.history.length > 0 && (
                <button
                  type="button"
                  className="text-xs text-ds-text-muted underline mb-2"
                  onClick={handleClearHistory}
                  disabled={loading}
                >
                  {t('road.clear')}
                </button>
              )}
            </div>

            {/* Action Log */}
            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.baccarat.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'baccarat-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="bac-bet-controls">
                <ChipBetInput
                  id="baccarat-bet-amount"
                  label={t('label.betAmount')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={state.chips}
                />
                <div className="flex items-center gap-2">
                  <label htmlFor="baccarat-bet-type" className="text-ds-text-primary text-sm">
                    {t('label.betTarget')}
                  </label>
                  <select
                    id="baccarat-bet-type"
                    value={betType}
                    onChange={(e) => setBetType(Number(e.target.value))}
                    className="px-2 py-1 rounded text-sm"
                  >
                    <option value={BaccaratBetType.PLAYER}>{t(BET_TYPE_LABELS[BaccaratBetType.PLAYER])}</option>
                    <option value={BaccaratBetType.BANKER}>{t(BET_TYPE_LABELS[BaccaratBetType.BANKER])}</option>
                    <option value={BaccaratBetType.TIE}>{t(BET_TYPE_LABELS[BaccaratBetType.TIE])}</option>
                  </select>
                </div>
                {/* Side bet inputs — collapsed by default to reduce clutter for beginners, but
                    auto-expanded when a side bet carries over from a prior round so a non-zero
                    wager is never hidden behind the summary (and re-bet unintentionally). */}
                <details
                  className="bg-black/30 rounded-lg w-full max-w-sm"
                  data-testid="baccarat-sidebet-details"
                  open={playerPairBet > 0 || bankerPairBet > 0 || undefined}
                >
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('sideBet.title')}
                  </summary>
                  <div className="flex flex-col items-center gap-2 px-4 pb-3">
                    <ChipBetInput
                      id="baccarat-pp-bet"
                      label={t('sideBet.playerPair')}
                      value={playerPairBet}
                      onChange={setPlayerPairBet}
                      max={state.chips}
                      min={0}
                      widthClass="w-20"
                    />
                    <ChipBetInput
                      id="baccarat-bp-bet"
                      label={t('sideBet.bankerPair')}
                      value={bankerPairBet}
                      onChange={setBankerPairBet}
                      max={state.chips}
                      min={0}
                      widthClass="w-20"
                    />
                  </div>
                </details>
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isEndPhase && (
              <div className="flex justify-center gap-2 pb-2">
                {canRebet && (
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleRebet}
                    disabled={loading}
                    data-testid="bac-rebet-button"
                  >
                    {t('button.rebet', { amount: totalLastBet })}
                  </button>
                )}
                <GameResetButton
                  isGameEnd={isEndPhase}
                  onReset={handleReset}
                  requestConfirm={requestConfirm}
                  loading={loading}
                  dataTutorial="bac-reset-button"
                />
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('actionLog.view')}
                </button>
              </div>
            )}
            <ActionShortcutsPanel bindings={actionBindings} data-testid="baccarat-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
