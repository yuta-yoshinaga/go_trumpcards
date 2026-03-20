import { useEffect, useMemo, useState } from 'react';
import { baccaratApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { BaccaratSkeleton } from '../components/skeleton/BaccaratSkeleton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { BaccaratSideBetResult } from '../types/card';
import { BaccaratBetType, BaccaratPhase } from '../types/phases';

const BET_TYPE_LABELS: Record<number, string> = {
  [BaccaratBetType.PLAYER]: 'betType.player',
  [BaccaratBetType.BANKER]: 'betType.banker',
  [BaccaratBetType.TIE]: 'betType.tie',
};

const ROAD_PLAYER = 0;
const ROAD_BANKER = 1;
const ROAD_MAX_ROWS = 6;

function BigRoadGrid({ history }: { history: number[] }) {
  if (history.length === 0) return null;

  // Build columns: new column on side change, ties mark previous cell
  const columns: { result: number; tie: boolean }[][] = [];
  let currentCol: { result: number; tie: boolean }[] = [];
  let lastSide: number | null = null;

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
    currentCol.push({ result, tie: false });
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

  return (
    <div className="mb-4" data-testid="big-road">
      <div
        className="inline-grid gap-px bg-gray-700 border border-gray-600 rounded overflow-x-auto"
        style={{ gridTemplateColumns: `repeat(${maxCols}, 28px)` }}
      >
        {grid.flatMap((row, ri) =>
          row.map((cell, ci) => {
            const cellKey = `r${String(ri)}c${String(ci)}`;
            return (
              <div key={cellKey} className="w-7 h-7 flex items-center justify-center bg-gray-900 relative">
                {cell && (
                  <>
                    <span
                      className={`w-5 h-5 rounded-full inline-block ${cell.result === ROAD_PLAYER ? 'bg-blue-500' : 'bg-red-500'}`}
                    />
                    {cell.tie && (
                      <span className="absolute inset-0 flex items-center justify-center text-green-400 font-bold text-xs">
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
        <div key={r.betType} className="text-white text-sm">
          {r.betType === 1 ? t('sideBet.playerPair') : t('sideBet.bankerPair')}:{' '}
          {r.resultType === 1 ? t('sideBet.pair') : t('sideBet.noPair')} ({t('label.payout', { payout: r.payout })})
        </div>
      ))}
    </div>
  );
}

/** Renders the Baccarat game page with betting and result display. */
export function BaccaratPage() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('baccarat');

  const [betAmount, setBetAmount] = useState(100);
  const [betType, setBetType] = useState<number>(BaccaratBetType.PLAYER);
  const [playerPairBet, setPlayerPairBet] = useState(0);
  const [bankerPairBet, setBankerPairBet] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi } = useGameApi(baccaratApi.exec);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const isBetPhase = state?.phase === BaccaratPhase.BET;
  const isEndPhase = state?.phase === BaccaratPhase.END;

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', betAmount, betType, playerPairBet, bankerPairBet),
        enabled: isBetPhase,
      },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, betAmount, betType, playerPairBet, bankerPairBet, isBetPhase, isEndPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <BaccaratSkeleton />;

  const handleBet = () => {
    execApi('bet', betAmount, betType, playerPairBet, bankerPairBet);
  };

  const handleReset = () => {
    execApi('reset');
  };

  const handleClearHistory = () => {
    execApi('clearhistory');
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy={loading}>
      {/* Phase indicator */}
      <PhaseIndicator phaseName={isBetPhase ? t('phase.bet') : t('phase.end')}>
        <span>{t('label.chips', { chips: state.chips })}</span>
      </PhaseIndicator>

      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Player Hand */}
        {state.playerHand.length > 0 && (
          <div className="mb-4">
            <div className="text-yellow-300 font-bold text-center mb-1">
              <span aria-hidden="true">🟡</span> {t('player')} {t('label.value', { value: state.playerHandValue })}
            </div>
            <div className="flex justify-center gap-2">
              {state.playerHand.map((card, i) => (
                <CardImage key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
              ))}
            </div>
          </div>
        )}

        {/* Banker Hand */}
        {state.bankerHand.length > 0 && (
          <div className="mb-4">
            <div className="text-red-300 font-bold text-center mb-1">
              <span aria-hidden="true">🔴</span> {t('banker')} {t('label.value', { value: state.bankerHandValue })}
            </div>
            <div className="flex justify-center gap-2">
              {state.bankerHand.map((card, i) => (
                <CardImage key={`b-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
              ))}
            </div>
          </div>
        )}

        {/* Payout info */}
        {isEndPhase && (
          <div className="text-white text-center font-bold mb-2">{t('label.payout', { payout: state.payout })}</div>
        )}

        {/* Side bet results */}
        {isEndPhase && state.sideBetResults.length > 0 && (
          <SideBetResultsDisplay results={state.sideBetResults} t={t} />
        )}

        {/* Big Road */}
        <div className="flex flex-col items-center">
          {state.history.length > 0 && <div className="text-white text-sm font-bold mb-1">{t('road.title')}</div>}
          <BigRoadGrid history={state.history} />
          {state.history.length > 0 && (
            <button
              type="button"
              className="text-xs text-gray-400 underline mb-2"
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
      <GameFooter className="bg-gray-800 px-4 pt-3">
        <ErrorAlert message={error} />
        {isBetPhase && (
          <div className="flex flex-col items-center gap-2 pb-2">
            <div className="flex items-center gap-2">
              <label htmlFor="baccarat-bet-amount" className="text-white text-sm">
                {t('label.betAmount')}
              </label>
              <input
                id="baccarat-bet-amount"
                type="number"
                min={10}
                max={state.chips}
                step={10}
                value={betAmount}
                onChange={(e) => setBetAmount(Number(e.target.value))}
                className="w-24 px-2 py-1 rounded text-sm"
              />
            </div>
            <div className="flex items-center gap-2">
              <label htmlFor="baccarat-bet-type" className="text-white text-sm">
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
            {/* Side bet inputs */}
            <div className="flex items-center gap-2">
              <label htmlFor="baccarat-pp-bet" className="text-white text-sm">
                {t('sideBet.playerPair')}
              </label>
              <input
                id="baccarat-pp-bet"
                type="number"
                min={0}
                max={state.chips}
                step={10}
                value={playerPairBet}
                onChange={(e) => setPlayerPairBet(Number(e.target.value))}
                className="w-20 px-2 py-1 rounded text-sm"
              />
            </div>
            <div className="flex items-center gap-2">
              <label htmlFor="baccarat-bp-bet" className="text-white text-sm">
                {t('sideBet.bankerPair')}
              </label>
              <input
                id="baccarat-bp-bet"
                type="number"
                min={0}
                max={state.chips}
                step={10}
                value={bankerPairBet}
                onChange={(e) => setBankerPairBet(Number(e.target.value))}
                className="w-20 px-2 py-1 rounded text-sm"
              />
            </div>
            <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
              {t('button.bet')}
            </button>
          </div>
        )}
        {isEndPhase && (
          <div className="flex justify-center gap-2 pb-2">
            <button type="button" className={btnPrimary} onClick={() => requestConfirm(handleReset)} disabled={loading}>
              {t('button.reset')}
            </button>
            <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
      </GameFooter>
      <ConfirmDialog
        open={confirmOpen}
        title={tc('button.confirmReset')}
        message={tc('button.confirmResetMessage')}
        confirmLabel={tc('button.confirm')}
        cancelLabel={tc('button.cancel')}
        onConfirm={confirmReset}
        onCancel={cancelReset}
      />
    </div>
  );
}
