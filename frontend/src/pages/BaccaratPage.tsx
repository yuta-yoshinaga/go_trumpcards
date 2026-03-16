import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
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
import { useActionLog } from '../hooks/useActionLog';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useConfirmDialog } from '../hooks/useConfirmDialog';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import type { BaccaratResponse } from '../types/card';
import { BaccaratBetType, BaccaratPhase } from '../types/phases';

const BET_TYPE_LABELS: Record<number, string> = {
  [BaccaratBetType.PLAYER]: 'betType.player',
  [BaccaratBetType.BANKER]: 'betType.banker',
  [BaccaratBetType.TIE]: 'betType.tie',
};

export function BaccaratPage() {
  const { t } = useTranslation('baccarat');
  const { t: tc } = useTranslation('common');

  const [betAmount, setBetAmount] = useState(100);
  const [betType, setBetType] = useState<number>(BaccaratBetType.PLAYER);

  const onSuccess = useCallback((_res: BaccaratResponse) => {}, []);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec } = useGameApi(baccaratApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const { actionLog, showActionLog, hideActionLog } = useActionLog('baccarat');
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();

  const isBetPhase = state?.phase === BaccaratPhase.BET;
  const isEndPhase = state?.phase === BaccaratPhase.END;

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: () => exec('bet', betAmount, betType), enabled: isBetPhase },
      { key: 'r', action: () => exec('reset'), enabled: isEndPhase },
    ],
    [exec, betAmount, betType, isBetPhase, isEndPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <BaccaratSkeleton />;

  const handleBet = () => {
    exec('bet', betAmount, betType);
  };

  const handleReset = () => {
    exec('reset');
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
