import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { holdemApi } from '../api/gameApi';
import { BettingControls } from '../components/BettingControls';
import { CardBack, CardImage } from '../components/CardImage';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { RoundResults } from '../components/RoundResults';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary } from '../styles/buttonStyles';

import { handNameBadgeStyle } from '../styles/gameConstants';
import { HoldemPhase } from '../types/phases';

function usePhaseNames(t: (key: string) => string): Record<number, string> {
  return {
    [HoldemPhase.PRE_FLOP]: t('phase.preFlop'),
    [HoldemPhase.FLOP]: t('phase.flop'),
    [HoldemPhase.TURN]: t('phase.turn'),
    [HoldemPhase.RIVER]: t('phase.river'),
    [HoldemPhase.SHOWDOWN]: t('phase.showdown'),
    [HoldemPhase.END]: t('phase.end'),
  };
}

export function HoldemPage() {
  const { t } = useTranslation('holdem');
  const { t: tc } = useTranslation('common');
  const phaseNames = usePhaseNames(t);
  const { state, loading, error, exec } = useGameApi(holdemApi.exec);
  const [betAmount, setBetAmount] = useState(20);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  useEffect(() => {
    if (state?.minRaise && state.minRaise > 0) {
      setBetAmount(state.minRaise);
    } else if (state) {
      setBetAmount(20);
    }
  }, [state]);

  const phase = state?.phase ?? HoldemPhase.INIT;
  const isActive = phase >= HoldemPhase.PRE_FLOP && phase <= HoldemPhase.RIVER;
  const isShowdown = phase === HoldemPhase.SHOWDOWN || phase === HoldemPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a6b1a]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">{tc('status.loading')}</span>}
      {/* Info bar */}
      <div className="shrink-0 bg-black/40 text-white text-sm px-5 py-2 flex flex-wrap gap-x-6 gap-y-1">
        <span>
          <strong>{phaseNames[phase] ?? t('phase.init')}</strong>
        </span>
        <span>
          {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
        </span>
        <span>
          SB/BB:{' '}
          <strong>
            {state?.smallBlind ?? 0}/{state?.bigBlind ?? 0}
          </strong>
        </span>
        <span>
          {tc('label.dealer')} <strong>Player {state?.dealerIdx ?? 0}</strong>
        </span>
        {state?.tournamentMode && (
          <span>{t('handNumber', { count: state.handCount, level: state.blindLevelHands })}</span>
        )}
      </div>

      {/* Scrollable: community cards + CPU players */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        {/* Community cards */}
        <div className="mb-4">
          <div className="text-white text-[1.1em] mb-1.5">{t('communityCards')}</div>
          <div className="flex flex-wrap gap-2">
            {state?.communityCards?.length
              ? state.communityCards.map((card) => (
                  <CardImage
                    key={`${card.design}-${card.value}`}
                    card={card}
                    width={60}
                    style={{ border: '3px solid transparent' }}
                  />
                ))
              : Array.from({ length: 5 }).map((_, i) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
                  <CardBack key={i} width={60} />
                ))}
          </div>
        </div>

        {/* CPU players */}
        {state?.players
          ?.filter((p) => !p.isHuman)
          .map((p) => (
            <CpuPlayerCard
              key={p.id}
              player={p}
              showCards={isShowdown}
              faceDownCount={2}
              showHandName={isShowdown}
              extraInfo={
                p.totalHands > 0 ? (
                  <span className="ml-2 text-cyan-300 text-[0.8em]">
                    VPIP:{p.vpip}% PFR:{p.pfr}%
                  </span>
                ) : undefined
              }
            />
          ))}

        {/* CPU actions log */}
        <CpuActionLog actions={state?.cpuActions} />

        {/* Round results */}
        {isShowdown && <RoundResults results={state?.roundResults} players={state?.players ?? []} />}
      </div>

      {/* Sticky footer: player hand + buttons */}
      <GameFooter className="bg-[#155715] border-white/20 px-5 py-3">
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
            <div className="text-white text-[1.1em] mb-1">
              {t('yourHand')}
              <span className="ml-3 text-[0.85em]">
                {tc('betting.chips')} {humanPlayer.chips}
              </span>
              {humanPlayer.totalHands > 0 && (
                <span className="ml-2 text-cyan-300 text-[0.8em]">
                  VPIP:{humanPlayer.vpip}% PFR:{humanPlayer.pfr}%
                </span>
              )}
              {humanPlayer.currentBet > 0 && (
                <span className="ml-2 text-[0.85em]">
                  {tc('betting.currentBet')} {humanPlayer.currentBet}
                </span>
              )}
              {humanPlayer.folded && <span className="ml-2 text-red-300 text-[0.85em]">[{tc('status.folded')}]</span>}
              {humanPlayer.allIn && <span className="ml-2 text-yellow-300 text-[0.85em]">[{tc('status.allIn')}]</span>}
              {isShowdown && !humanPlayer.folded && humanPlayer.handName && (
                <span
                  className="inline-block ml-2 text-[0.85em] font-bold rounded px-2 py-0.5"
                  style={handNameBadgeStyle}
                >
                  {humanPlayer.handName}
                </span>
              )}
            </div>
            <div className="flex flex-wrap gap-1.5 mb-2">
              {humanPlayer.cards?.length
                ? humanPlayer.cards.map((card) => (
                    <CardImage
                      key={`${card.design}-${card.value}`}
                      card={card}
                      width={60}
                      style={{ border: '3px solid transparent' }}
                    />
                  ))
                : !humanPlayer.folded &&
                  Array.from({ length: 2 }).map((_, i) => (
                    // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
                    <CardBack key={i} width={60} />
                  ))}
            </div>
          </div>
        )}

        {/* Message */}
        <GameMessageBox
          message={state?.message}
          messageCode={state?.messageCode}
          messageParams={state?.messageParams}
          alwaysVisible
        />

        <ErrorAlert message={error} />

        {/* Betting controls */}
        {canAct && (
          <BettingControls
            inputId="holdemBetAmount"
            betAmount={betAmount}
            onBetAmountChange={setBetAmount}
            minRaise={minRaise}
            hasOutstandingBet={hasOutstandingBet}
            loading={loading}
            onCall={() => exec('call')}
            onRaise={() => exec('raise', betAmount)}
            onBet={() => exec('bet', betAmount)}
            onCheck={() => exec('check')}
            onFold={() => exec('fold')}
            onAllIn={() => exec('allin')}
          />
        )}

        {/* Reset button */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => exec('reset')}
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
    </div>
  );
}
