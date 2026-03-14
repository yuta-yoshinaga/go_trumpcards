import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { baccaratApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { useActionLog } from '../hooks/useActionLog';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { BaccaratResponse } from '../types/card';
import { BACCARAT_BET_BANKER, BACCARAT_BET_PLAYER, BACCARAT_BET_TIE, BACCARAT_PHASE } from '../types/card';

const BET_TYPE_LABELS: Record<number, string> = {
  [BACCARAT_BET_PLAYER]: 'betType.player',
  [BACCARAT_BET_BANKER]: 'betType.banker',
  [BACCARAT_BET_TIE]: 'betType.tie',
};

export function BaccaratPage() {
  const { t } = useTranslation('baccarat');
  const { t: tc } = useTranslation('common');

  const [betAmount, setBetAmount] = useState(100);
  const [betType, setBetType] = useState(BACCARAT_BET_PLAYER);

  const onSuccess = useCallback((_res: BaccaratResponse) => {}, []);

  const { state, loading, error, exec } = useGameApi(baccaratApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const { actionLog, showActionLog, hideActionLog } = useActionLog('baccarat');

  if (!state) return null;

  const isBetPhase = state.phase === BACCARAT_PHASE.BET;
  const isEndPhase = state.phase === BACCARAT_PHASE.END;

  const handleBet = () => {
    exec('bet', betAmount, betType);
  };

  const handleReset = () => {
    exec('reset');
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#0d5016]" aria-busy={loading}>
      <LoadingSpinner loading={loading} />

      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <ErrorAlert message={error} />

        {/* Chips */}
        <div className="text-white text-center mb-2 font-bold">{t('label.chips', { chips: state.chips })}</div>

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Player Hand */}
        {state.playerHand.length > 0 && (
          <div className="mb-4">
            <div className="text-yellow-300 font-bold text-center mb-1">
              <span aria-hidden="true">🟡</span> {t('player')} {t('label.value', { value: state.playerHandValue })}
            </div>
            <div className="flex justify-center gap-2">
              {state.playerHand.map((card, i) => (
                <CardImage key={`p-${card.design}-${card.value}-${i}`} card={card} width={60} />
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
                <CardImage key={`b-${card.design}-${card.value}-${i}`} card={card} width={60} />
              ))}
            </div>
          </div>
        )}

        {/* Payout info */}
        {isEndPhase && (
          <div className="text-white text-center font-bold mb-2">{t('label.payout', { payout: state.payout })}</div>
        )}

        {/* Action Log */}
        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
      </div>

      {/* Footer */}
      <GameFooter className="bg-gray-800 px-4 pt-3">
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
                <option value={BACCARAT_BET_PLAYER}>{t(BET_TYPE_LABELS[BACCARAT_BET_PLAYER])}</option>
                <option value={BACCARAT_BET_BANKER}>{t(BET_TYPE_LABELS[BACCARAT_BET_BANKER])}</option>
                <option value={BACCARAT_BET_TIE}>{t(BET_TYPE_LABELS[BACCARAT_BET_TIE])}</option>
              </select>
            </div>
            <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
              {t('button.bet')}
            </button>
          </div>
        )}
        {isEndPhase && (
          <div className="flex justify-center gap-2 pb-2">
            <button type="button" className={btnPrimary} onClick={handleReset} disabled={loading}>
              {t('button.reset')}
            </button>
            <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
      </GameFooter>
    </div>
  );
}
