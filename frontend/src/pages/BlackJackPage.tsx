import { useCallback, useEffect, useState } from 'react';
import type { BlackJackConfigInput, BlackJackSideBetInput } from '../api/gameApi';
import { blackjackApi } from '../api/gameApi';
import { BjActionPhaseControls } from '../components/blackjack/BjActionPhaseControls';
import { BjBetPhaseControls } from '../components/blackjack/BjBetPhaseControls';
import { BjEndPhaseControls } from '../components/blackjack/BjEndPhaseControls';
import { BjInsurancePhaseControls } from '../components/blackjack/BjInsurancePhaseControls';
import {
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
import { CardBack, CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { useGameApi } from '../hooks/useGameApi';
import type { BlackJackResponse } from '../types/card';
import { BjPhase } from '../types/phases';

const SUGGESTION_LABELS: Record<number, string> = {
  [BJ_SUGGEST_HIT]: 'ヒット',
  [BJ_SUGGEST_STAND]: 'スタンド',
  [BJ_SUGGEST_DOUBLE]: 'ダブルダウン',
  [BJ_SUGGEST_SPLIT]: 'スプリット',
  [BJ_SUGGEST_SURRENDER]: 'サレンダー',
  [BJ_SUGGEST_DECLINE_INSURANCE]: '辞退',
  [BJ_SUGGEST_DOUBLE_STAND]: 'ダブルダウン',
};

export function BlackJackPage() {
  const [message, setMessage] = useState('');
  const [betAmount, setBetAmount] = useState(10);
  const [dealerHitsSoft17, setDealerHitsSoft17] = useState(false);
  const [countingEnabled, setCountingEnabled] = useState(false);
  const [cpuPlayerCount, setCpuPlayerCount] = useState(0);
  const [perfectPairsBet, setPerfectPairsBet] = useState(0);
  const [twentyOnePlus3Bet, setTwentyOnePlus3Bet] = useState(0);
  const [autoAdvance, setAutoAdvance] = useState(0);

  const onSuccess = useCallback((res: BlackJackResponse) => {
    setMessage(res.message);
    setDealerHitsSoft17(res.dealerHitsSoft17);
    setCountingEnabled(res.countingEnabled);
    setCpuPlayerCount(res.cpuPlayerCount);
  }, []);
  const { state, loading, error, exec } = useGameApi(blackjackApi.exec, { onSuccess });

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

  const showDoubleDown = !!currentHand && currentHand.cards.length === 2 && playerChips >= currentHand.bet;
  const showSplit = !!currentHand?.canSplit && playerChips >= currentHand.bet;
  const showSurrender = !!currentHand?.canSurrender;

  const handleReset = useCallback(() => {
    const config: BlackJackConfigInput = {
      dealerHitsSoft17,
      cpuPlayerCount,
      countingEnabled,
    };
    exec('reset', undefined, config);
  }, [exec, dealerHitsSoft17, cpuPlayerCount, countingEnabled]);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#008000]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">処理中...</span>}
      {/* Chip info bar */}
      {state && (
        <div className="shrink-0 bg-black/40 text-white text-sm px-4 py-1.5 flex justify-between flex-wrap gap-1">
          <span>プレイヤー: {state.player.chips} chips</span>
          <span>デッキ: {state.deckCount}デッキ</span>
          {countingEnabled && (
            <span>
              RC={state.runningCount} TC={state.trueCount.toFixed(1)}
            </span>
          )}
          <span>ディーラー: {state.dealer.chips} chips</span>
        </div>
      )}

      {/* Scrollable: dealer area + CPU players */}
      <div className="flex-1 overflow-y-auto p-4">
        {state && phase !== BjPhase.BET && (
          <div>
            <h3 className="text-white">ディーラー手札{dealerHitsSoft17 ? ' (H17)' : ' (S17)'}</h3>
            <h3 className="text-white">スコア {state.dealer.score ? state.dealer.score : ''}</h3>
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
                  CPU {cpuIdx + 1} ({cpu.chips} chips)
                </h3>
                {cpu.hands.map((hand, handIdx) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: CPU hands have fixed order
                  <div key={handIdx} className="mb-1">
                    <div className="text-yellow-100 text-sm">
                      {cpu.hands.length > 1 ? `ハンド ${handIdx + 1} ` : ''}
                      スコア {hand.score} / ベット {hand.bet}
                      {hand.busted && ' [BUST]'}
                      {hand.doubled && ' [DD]'}
                      {hand.isBlackJack && ' [BJ]'}
                      {hand.surrendered && ' [SUR]'}
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
      <div
        className="shrink-0 bg-[#005a00] border-t border-white/15 px-4 py-3"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      >
        {/* Player hands */}
        {state && phase !== BjPhase.BET && hands.length > 0 && (
          <div className="mb-2">
            {hands.map((hand, handIndex) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: player hands have fixed order per round
              <div key={`hand-${handIndex}`} className="mb-2">
                <h3 className="text-white mt-0 mb-0.5">
                  {hands.length > 1 ? `ハンド ${handIndex + 1}` : 'プレイヤー手札'}
                  {handIndex === currentHandIdx && phase === BjPhase.ACTION && ' (*)'}
                  {hand.busted && ' [BUST]'}
                  {hand.doubled && ' [DD]'}
                  {hand.isBlackJack && ' [BJ]'}
                  {hand.surrendered && (
                    <span className="ml-1 text-xs bg-gray-500 text-white px-1 rounded">SURRENDER</span>
                  )}
                </h3>
                <h3 className="text-white mt-0 mb-0.5">
                  スコア {hand.score} / ベット {hand.bet}
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
          <div className="text-yellow-300 text-sm mb-1">インシュランス: {state.insuranceBet}</div>
        )}

        {/* Side bet results */}
        {sideBetResults.length > 0 && (
          <div className="mb-2">
            {sideBetResults.map((r) => (
              <div
                key={r.betType}
                className={`text-sm text-center px-3 py-1 rounded mb-1 ${r.payout > 0 ? 'bg-yellow-400/90 text-gray-900 font-bold' : 'bg-gray-500/70 text-white'}`}
              >
                {r.betType === BJ_SIDE_BET_PERFECT_PAIRS ? 'Perfect Pairs' : '21+3'}:{' '}
                {r.payout > 0 ? `${r.resultName} WIN +${r.payout}` : `LOSE -${r.betAmount}`}
              </div>
            ))}
          </div>
        )}

        {/* Hint banner */}
        {hintEnabled && suggestedAction !== BJ_SUGGEST_NONE && (
          <div className="bg-yellow-300/90 text-gray-900 text-center text-sm font-bold px-3 py-1 rounded mb-2">
            推奨: {SUGGESTION_LABELS[suggestedAction]}
          </div>
        )}

        {/* Result message */}
        {message && (
          <div className="bg-black/70 text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2 rounded-lg">
            {message}
          </div>
        )}

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
                onCpuPlayerCountChange={setCpuPlayerCount}
                hintEnabled={hintEnabled}
                onToggleHint={() => exec('togglehint')}
                dealerHitsSoft17={dealerHitsSoft17}
                onToggleSoft17={() => exec('togglesoft17')}
                countingEnabled={countingEnabled}
                onToggleCounting={() => exec('togglecounting')}
                loading={loading}
                onBet={() => {
                  const sideBets: BlackJackSideBetInput = {};
                  if (perfectPairsBet > 0) sideBets.perfectPairsBet = perfectPairsBet;
                  if (twentyOnePlus3Bet > 0) sideBets.twentyOnePlus3Bet = twentyOnePlus3Bet;
                  exec('bet', betAmount, undefined, sideBets);
                }}
                perfectPairsBet={perfectPairsBet}
                onPerfectPairsBetChange={setPerfectPairsBet}
                twentyOnePlus3Bet={twentyOnePlus3Bet}
                onTwentyOnePlus3BetChange={setTwentyOnePlus3Bet}
              />
              <div className="flex items-center justify-center gap-2 mt-2">
                <label htmlFor="bj-auto-advance" className="text-white text-sm">
                  自動進行:
                </label>
                <select
                  id="bj-auto-advance"
                  value={autoAdvance}
                  onChange={(e) => setAutoAdvance(Number(e.target.value))}
                  className="px-2 py-1 rounded text-sm"
                >
                  <option value={0}>OFF</option>
                  <option value={3}>3秒</option>
                  <option value={5}>5秒</option>
                  <option value={10}>10秒</option>
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

          {phase === BjPhase.END && (
            <BjEndPhaseControls
              loading={loading}
              onReset={handleReset}
              autoAdvanceSeconds={autoAdvance > 0 ? autoAdvance : undefined}
            />
          )}
        </div>
      </div>
    </div>
  );
}
