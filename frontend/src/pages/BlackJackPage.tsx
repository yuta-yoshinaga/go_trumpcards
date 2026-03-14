import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { BlackJackBetOptions, BlackJackConfigInput } from '../api/gameApi';
import { blackjackApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { BjActionPhaseControls } from '../components/blackjack/BjActionPhaseControls';
import { BjBetPhaseControls } from '../components/blackjack/BjBetPhaseControls';
import { BjEarlySurrenderPhaseControls } from '../components/blackjack/BjEarlySurrenderPhaseControls';
import { BjEndPhaseControls } from '../components/blackjack/BjEndPhaseControls';
import { BjInsurancePhaseControls } from '../components/blackjack/BjInsurancePhaseControls';
import {
  BJ_COUNTING_KO,
  BJ_SIDE_BET_PERFECT_PAIRS,
  BJ_SUGGEST_DECLINE_INSURANCE,
  BJ_SUGGEST_DOUBLE,
  BJ_SUGGEST_DOUBLE_STAND,
  BJ_SUGGEST_HIT,
  BJ_SUGGEST_NONE,
  BJ_SUGGEST_SPLIT,
  BJ_SUGGEST_STAND,
  BJ_SUGGEST_SURRENDER,
} from '../components/blackjack/bjConstants';
import { HandStatusBadges } from '../components/blackjack/HandStatusBadges';
import { CardBack, CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { useActionLog } from '../hooks/useActionLog';
import { useConfirmDialog } from '../hooks/useConfirmDialog';
import { useGameApi } from '../hooks/useGameApi';
import { btnSecondary } from '../styles/buttonStyles';
import type { BlackJackResponse } from '../types/card';
import { BjPhase } from '../types/phases';

function usePhaseNames(t: (key: string) => string): Record<number, string> {
  return {
    [BjPhase.BET]: t('phase.bet'),
    [BjPhase.DEAL]: t('phase.deal'),
    [BjPhase.INSURANCE]: t('phase.insurance'),
    [BjPhase.ACTION]: t('phase.action'),
    [BjPhase.END]: t('phase.end'),
    [BjPhase.EARLY_SURRENDER]: t('phase.earlySurrender'),
  };
}

function useSuggestionLabels(t: (key: string) => string): Record<number, string> {
  return {
    [BJ_SUGGEST_HIT]: t('suggest.hit'),
    [BJ_SUGGEST_STAND]: t('suggest.stand'),
    [BJ_SUGGEST_DOUBLE]: t('suggest.double'),
    [BJ_SUGGEST_SPLIT]: t('suggest.split'),
    [BJ_SUGGEST_SURRENDER]: t('suggest.surrender'),
    [BJ_SUGGEST_DECLINE_INSURANCE]: t('suggest.decline'),
    [BJ_SUGGEST_DOUBLE_STAND]: t('suggest.double'),
  };
}

export function BlackJackPage() {
  const { t } = useTranslation('blackjack');
  const { t: tc } = useTranslation('common');
  const phaseNames = usePhaseNames(t);
  const suggestionLabels = useSuggestionLabels(t);

  const [message, setMessage] = useState('');
  const { actionLog, showActionLog, hideActionLog } = useActionLog('blackjack');
  const [betAmount, setBetAmount] = useState(10);
  const [dealerHitsSoft17, setDealerHitsSoft17] = useState(false);
  const [countingEnabled, setCountingEnabled] = useState(false);
  const [cpuPlayerCount, setCpuPlayerCount] = useState(0);
  const [perfectPairsBet, setPerfectPairsBet] = useState(0);
  const [twentyOnePlus3Bet, setTwentyOnePlus3Bet] = useState(0);
  const [handCount, setHandCount] = useState(1);
  const [doubleAfterSplit, setDoubleAfterSplit] = useState(true);
  const [countingSystem, setCountingSystem] = useState(0);
  const [deckPenetration, setDeckPenetration] = useState(75);
  const [surrenderRule, setSurrenderRule] = useState(0);
  const [autoAdvance, setAutoAdvance] = useState(0);

  const onSuccess = useCallback((res: BlackJackResponse) => {
    setMessage(res.message);
    setDealerHitsSoft17(res.dealerHitsSoft17);
    setCountingEnabled(res.countingEnabled);
    setCpuPlayerCount(res.cpuPlayerCount);
    setDoubleAfterSplit(res.doubleAfterSplit);
    setCountingSystem(res.countingSystem);
    setDeckPenetration(res.deckPenetration);
    setSurrenderRule(res.surrenderRule);
  }, []);
  const { state, loading, error, exec } = useGameApi(blackjackApi.exec, { onSuccess });
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const phase = state?.phase ?? BjPhase.BET;
  const hands = state?.hands ?? [];
  const currentHandIdx = state?.currentHandIdx ?? 0;
  const currentHand = hands[currentHandIdx];
  const playerChips = state?.player?.chips ?? 0;
  const hintEnabled = state?.hintEnabled ?? false;
  const suggestedAction = state?.suggestedAction ?? BJ_SUGGEST_NONE;
  const cpuPlayers = state?.cpuPlayers ?? [];
  const sideBetResults = state?.sideBetResults ?? [];

  const showDoubleDown =
    !!currentHand &&
    currentHand.cards.length === 2 &&
    playerChips >= currentHand.bet &&
    ((hands?.length ?? 0) <= 1 || doubleAfterSplit);
  const showSplit = !!currentHand?.canSplit && playerChips >= currentHand.bet;
  const showSurrender = !!currentHand?.canSurrender;

  const handleReset = useCallback(() => {
    hideActionLog();
    const config: BlackJackConfigInput = {
      dealerHitsSoft17,
      cpuPlayerCount,
      countingEnabled,
      doubleAfterSplit,
      countingSystem,
      deckPenetration,
      surrenderRule,
    };
    exec('reset', undefined, config);
  }, [
    exec,
    dealerHitsSoft17,
    cpuPlayerCount,
    countingEnabled,
    doubleAfterSplit,
    countingSystem,
    deckPenetration,
    surrenderRule,
    hideActionLog,
  ]);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#008000]" aria-busy={loading} aria-live="polite">
      <LoadingSpinner loading={loading} />
      {/* Phase indicator + info bar */}
      {state && (
        <PhaseIndicator
          phaseName={phaseNames[phase] ?? t('phase.bet')}
          isHumanTurn={
            phase === BjPhase.ACTION || phase === BjPhase.EARLY_SURRENDER
              ? true
              : phase === BjPhase.END
                ? false
                : undefined
          }
        >
          <span>
            {t('player')} {state.player.chips} chips
          </span>
          <span>
            {t('deck')} {state.deckCount}
            {t('deckUnit')}
          </span>
          {countingEnabled && (
            <span>
              {t(`countingSystemNames.${countingSystem}`)} RC={state.runningCount}{' '}
              {countingSystem === BJ_COUNTING_KO ? t('trueCountNA') : `TC=${state.trueCount.toFixed(1)}`}
            </span>
          )}
          <span>
            {tc('label.dealer')} {state.dealer.chips} chips
          </span>
        </PhaseIndicator>
      )}

      {/* Scrollable: dealer area + CPU players */}
      <div className="flex-1 overflow-y-auto p-4">
        {state && phase !== BjPhase.BET && (
          <div>
            <h3 className="text-white">
              {t('dealerHand')}
              {dealerHitsSoft17 ? ' (H17)' : ' (S17)'}
            </h3>
            <h3 className="text-white">
              {t('score')} {state.dealer.score ? state.dealer.score : ''}
            </h3>
            <div className="flex flex-wrap gap-2">
              {state.dealer.cards?.map((card, idx) => (
                <CardImage key={`dealer-${idx}-${card.design}-${card.value}`} card={card} width={60} />
              ))}
              {!state.dealer.score && <CardBack width={60} />}
            </div>
          </div>
        )}

        {/* CPU players */}
        {state && phase !== BjPhase.BET && cpuPlayers.length > 0 && (
          <div className="mt-4">
            {cpuPlayers.map((cpu, cpuIdx) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: CPU seats have fixed order
              <div key={cpuIdx} className="mb-3">
                <h3 className="text-yellow-200 mt-0 mb-1">
                  {tc('player.cpu', { id: cpuIdx + 1 })} ({cpu.chips} chips)
                  {cpu.insuranceBet > 0 && (
                    <span className="text-yellow-400 text-sm ml-2">
                      [{t('insurance')} {cpu.insuranceBet}]
                    </span>
                  )}
                </h3>
                {cpu.hands.map((hand, handIdx) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: CPU hands have fixed order
                  <div key={handIdx} className="mb-1">
                    <div className="text-yellow-100 text-sm">
                      {cpu.hands.length > 1 ? `${t('hand', { idx: handIdx + 1 })} ` : ''}
                      {t('score')} {hand.score} / {tc('betting.currentBet')} {hand.bet}
                      <HandStatusBadges
                        busted={hand.busted}
                        doubled={hand.doubled}
                        isBlackJack={hand.isBlackJack}
                        surrendered={hand.surrendered}
                      />
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {hand.cards.map((card, cardIdx) => (
                        <CardImage
                          key={`cpu${cpuIdx}-hand${handIdx}-${cardIdx}-${card.design}-${card.value}`}
                          card={card}
                          width={50}
                        />
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Sticky footer: player hand + result + buttons */}
      <GameFooter className="bg-[#005a00] border-white/15 px-4 py-3">
        {/* Player hands */}
        {state && phase !== BjPhase.BET && hands.length > 0 && (
          <div className="mb-2">
            {hands.map((hand, handIndex) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: player hands have fixed order per round
              <div key={`hand-${handIndex}`} className="mb-2">
                <h3 className="text-white mt-0 mb-0.5">
                  {hands.length > 1 ? t('hand', { idx: handIndex + 1 }) : t('playerHand')}
                  {handIndex === currentHandIdx &&
                    (phase === BjPhase.ACTION || phase === BjPhase.EARLY_SURRENDER) &&
                    ' (*)'}
                  <HandStatusBadges
                    busted={hand.busted}
                    doubled={hand.doubled}
                    isBlackJack={hand.isBlackJack}
                    surrendered={hand.surrendered}
                  />
                </h3>
                <h3 className="text-white mt-0 mb-0.5">
                  {t('score')} {hand.score} / {tc('betting.currentBet')} {hand.bet}
                </h3>
                <div className="flex flex-wrap gap-1.5">
                  {hand.cards.map((card, cardIdx) => (
                    <CardImage
                      key={`hand-${handIndex}-${cardIdx}-${card.design}-${card.value}`}
                      card={card}
                      width={60}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Insurance info */}
        {state && state.insuranceBet > 0 && (
          <div className="text-yellow-300 text-sm mb-1">
            {t('insurance')} {state.insuranceBet}
          </div>
        )}

        {/* Side bet results */}
        {sideBetResults.length > 0 && (
          <div className="mb-2">
            {sideBetResults.map((r) => (
              <div
                key={r.betType}
                className={`text-sm text-center px-3 py-1 rounded mb-1 ${r.payout > 0 ? 'bg-yellow-400/90 text-gray-900 font-bold' : 'bg-gray-500/70 text-white'}`}
              >
                {r.betType === BJ_SIDE_BET_PERFECT_PAIRS ? t('sideBet.perfectPairs') : t('sideBet.twentyOnePlus3')}:{' '}
                {r.payout > 0
                  ? t('sideBet.win', { name: r.resultName, payout: r.payout })
                  : t('sideBet.lose', { name: r.resultName, amount: r.betAmount })}
              </div>
            ))}
          </div>
        )}

        {/* Hint banner */}
        {hintEnabled && suggestedAction !== BJ_SUGGEST_NONE && (
          <div className="bg-yellow-300/90 text-gray-900 text-center text-sm font-bold px-3 py-1 rounded mb-2">
            {t('suggestion')} {suggestionLabels[suggestedAction]}
          </div>
        )}

        {/* Result message */}
        <GameMessageBox message={message} messageCode={state?.messageCode} messageParams={state?.messageParams} />

        {/* Action log */}
        {phase === BjPhase.END && !actionLog && (
          <div className="text-center my-2">
            <button type="button" className={btnSecondary} onClick={showActionLog}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}

        <ErrorAlert message={error} />

        {/* Phase-based buttons */}
        <div className="text-center">
          {phase === BjPhase.BET && (
            <>
              <BjBetPhaseControls
                betAmount={betAmount}
                onBetAmountChange={setBetAmount}
                deckCount={state?.deckCount ?? 1}
                onDeckCountChange={(v) => exec('setdeckcount', v)}
                cpuPlayerCount={cpuPlayerCount}
                onCpuPlayerCountChange={(v) => exec('setcpucount', v)}
                hintEnabled={hintEnabled}
                onToggleHint={() => exec('togglehint')}
                dealerHitsSoft17={dealerHitsSoft17}
                onToggleSoft17={() => exec('togglesoft17')}
                countingEnabled={countingEnabled}
                onToggleCounting={() => exec('togglecounting')}
                doubleAfterSplit={doubleAfterSplit}
                onToggleDAS={() => exec('toggledas')}
                countingSystem={countingSystem}
                onCountingSystemChange={(v) => exec('setcountingsystem', v)}
                deckPenetration={deckPenetration}
                onDeckPenetrationChange={(v) => exec('setpenetration', v)}
                surrenderRule={surrenderRule}
                onSurrenderRuleChange={(v) => exec('setsurrenderrule', v)}
                handCount={handCount}
                onHandCountChange={setHandCount}
                loading={loading}
                onBet={() => {
                  const betOptions: BlackJackBetOptions = {};
                  if (perfectPairsBet > 0) betOptions.perfectPairsBet = perfectPairsBet;
                  if (twentyOnePlus3Bet > 0) betOptions.twentyOnePlus3Bet = twentyOnePlus3Bet;
                  if (handCount > 1) betOptions.handCount = handCount;
                  exec('bet', betAmount, undefined, betOptions);
                }}
                perfectPairsBet={perfectPairsBet}
                onPerfectPairsBetChange={setPerfectPairsBet}
                twentyOnePlus3Bet={twentyOnePlus3Bet}
                onTwentyOnePlus3BetChange={setTwentyOnePlus3Bet}
              />
              <div className="flex items-center justify-center gap-2 mt-2">
                <label htmlFor="bj-auto-advance" className="text-white text-sm">
                  {t('autoAdvance')}
                </label>
                <select
                  id="bj-auto-advance"
                  value={autoAdvance}
                  onChange={(e) => setAutoAdvance(Number(e.target.value))}
                  className="px-2 py-1 rounded text-sm"
                >
                  <option value={0}>OFF</option>
                  <option value={3}>{t('autoAdvanceSec', { sec: 3 })}</option>
                  <option value={5}>{t('autoAdvanceSec', { sec: 5 })}</option>
                  <option value={10}>{t('autoAdvanceSec', { sec: 10 })}</option>
                </select>
              </div>
            </>
          )}

          {phase === BjPhase.INSURANCE && (
            <BjInsurancePhaseControls
              loading={loading}
              hintEnabled={hintEnabled}
              suggestedAction={suggestedAction}
              onInsurance={() => exec('insurance')}
              onDecline={() => exec('declineinsurance')}
            />
          )}

          {phase === BjPhase.ACTION && (
            <BjActionPhaseControls
              loading={loading}
              hintEnabled={hintEnabled}
              suggestedAction={suggestedAction}
              showDoubleDown={showDoubleDown}
              showSplit={showSplit}
              showSurrender={showSurrender}
              onHit={() => exec('hit')}
              onStand={() => exec('stand')}
              onDoubleDown={() => exec('doubledown')}
              onSplit={() => exec('split')}
              onSurrender={() => exec('surrender')}
            />
          )}

          {phase === BjPhase.EARLY_SURRENDER && (
            <BjEarlySurrenderPhaseControls
              loading={loading}
              hintEnabled={hintEnabled}
              suggestedAction={suggestedAction}
              onSurrender={() => exec('earlysurrender')}
              onContinue={() => exec('declineearlysurrender')}
            />
          )}

          {phase === BjPhase.END && (
            <BjEndPhaseControls
              loading={loading}
              onReset={handleReset}
              onManualReset={() => requestConfirm(handleReset)}
              autoAdvanceSeconds={autoAdvance > 0 ? autoAdvance : undefined}
            />
          )}
        </div>
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
