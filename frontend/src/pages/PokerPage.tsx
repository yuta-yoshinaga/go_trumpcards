import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { BettingControls } from '../components/BettingControls';
import { CardImage } from '../components/CardImage';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { RoundResults } from '../components/RoundResults';
import { useActionLog } from '../hooks/useActionLog';
import { usePokerGame } from '../hooks/usePokerGame';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { PokerPhase } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';

const cardWrapBase: React.CSSProperties = {
  position: 'relative',
  cursor: 'pointer',
  transition: 'transform 0.15s',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
};

export function PokerPage() {
  const { t } = useTranslation('poker');
  const { t: tc } = useTranslation('common');
  const { state, loading, error, exec, selected, toggleCard, odds, canExchange } = usePokerGame();
  const { actionLog, showActionLog, hideActionLog } = useActionLog('poker');
  const [betAmount, setBetAmount] = useState(10);
  const [bettingLimit, setBettingLimit] = useState(0);
  const [isLowball, setIsLowball] = useState(false);

  useEffect(() => {
    if (state?.minRaise && state.minRaise > 0) {
      setBetAmount(state.minRaise);
    } else if (state) {
      setBetAmount(10);
    }
  }, [state]);

  const phase = state?.phase ?? PokerPhase.INIT;
  const isBettingPhase = phase === PokerPhase.DEAL || phase === PokerPhase.SECOND_BET;
  const isEnd = phase === PokerPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isBettingPhase && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 10;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a6b1a]" aria-busy={loading} aria-live="polite">
      <LoadingSpinner loading={loading} />
      {/* Info bar */}
      <div className="shrink-0 bg-black/40 text-white text-sm px-5 py-2 flex flex-wrap gap-x-6 gap-y-1">
        <span>
          {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
        </span>
        <span>
          {tc('label.dealer')} <strong>Player {state?.dealerIdx ?? 0}</strong>
        </span>
        {(state?.jokerCount ?? 0) > 0 && (
          <span>
            {t('joker')} <strong>{state?.jokerCount}</strong>
          </span>
        )}
        {state?.isLowball && (
          <span className="bg-yellow-600 text-white px-2 py-0.5 rounded text-xs font-bold">[{t('lowballMode')}]</span>
        )}
      </div>

      {/* Scrollable: CPU players + logs */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        {/* CPU players */}
        {state?.players
          ?.filter((p) => !p.isHuman)
          .map((p) => (
            <CpuPlayerCard
              key={p.id}
              player={p}
              showCards={isEnd}
              faceDownCount={5}
              showHandName={isEnd}
              extraInfo={
                (phase === PokerPhase.SECOND_BET || isEnd) && p.exchangeCount > 0 && !p.folded ? (
                  <span className="ml-2 text-xs">{t('exchangeCount', { count: p.exchangeCount })}</span>
                ) : undefined
              }
            />
          ))}

        {/* CPU actions log */}
        <CpuActionLog actions={state?.cpuActions} />

        {/* CPU exchanges log */}
        {state?.cpuExchanges && state.cpuExchanges.length > 0 && (
          <div className="bg-black/30 rounded p-2 mb-3 text-white text-xs">
            <div className="font-bold mb-1">{t('cpuExchange')}</div>
            {state.cpuExchanges.map((ex, i) => (
              <div key={`${i}-${ex.playerIdx}`}>
                {t('cpuExchangeEntry', { idx: ex.playerIdx, count: ex.exchangeCount })}
              </div>
            ))}
          </div>
        )}

        {/* Round results */}
        {isEnd && <RoundResults results={state?.roundResults} players={state?.players ?? []} />}
      </div>

      {/* Sticky footer: player hand + buttons */}
      <GameFooter className="bg-[#155715] border-white/20 px-5 py-3">
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
            <div className="text-white text-lg mb-1">
              {t('yourHand')}
              <span className="ml-3 text-xs">
                {tc('betting.chips')} {humanPlayer.chips}
              </span>
              {humanPlayer.currentBet > 0 && (
                <span className="ml-2 text-xs">
                  {tc('betting.currentBet')} {humanPlayer.currentBet}
                </span>
              )}
              {humanPlayer.folded && <span className="ml-2 text-red-300 text-xs">[{tc('status.folded')}]</span>}
              {humanPlayer.allIn && <span className="ml-2 text-yellow-300 text-xs">[{tc('status.allIn')}]</span>}
              {isEnd && !humanPlayer.folded && humanPlayer.handName && (
                <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                  {humanPlayer.handName}
                </span>
              )}
            </div>
            {canExchange && <div className="text-game-text-highlight text-xs mb-1">{t('exchangeInstruction')}</div>}
            <div className="flex flex-wrap gap-1.5 mb-2">
              {humanPlayer.cards?.map((card, i) => {
                const isSelected = selected.includes(i);
                return (
                  <button
                    key={`${card.design}-${card.value}`}
                    type="button"
                    aria-label={`${cardAlt(card)}${isSelected ? ` ${t('cardSelected')}` : ''}`}
                    aria-pressed={isSelected}
                    onClick={() => toggleCard(i)}
                    className="focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 focus-visible:ring-offset-2 focus-visible:ring-offset-black rounded"
                    style={{
                      background: 'none',
                      border: 'none',
                      padding: 0,
                      ...cardWrapBase,
                      cursor: canExchange ? 'pointer' : 'default',
                    }}
                  >
                    <CardImage
                      card={card}
                      width={60}
                      style={{
                        border: isSelected ? '3px solid var(--color-game-status-waiting)' : '3px solid transparent',
                        transform: isSelected ? 'translateY(-10px)' : undefined,
                        transition: 'transform 0.15s',
                      }}
                    />
                    <div
                      style={{
                        color: 'var(--color-game-status-waiting)',
                        fontSize: '0.75em',
                        fontWeight: 'bold',
                        visibility: isSelected ? 'visible' : 'hidden',
                      }}
                    >
                      {t('exchangeLabel')}
                    </div>
                  </button>
                );
              })}
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

        {/* Action log */}
        {state?.gameEndFlag && !actionLog && (
          <div className="text-center my-2">
            <button type="button" className={btnSecondary} onClick={showActionLog}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}

        <ErrorAlert message={error} />

        {/* Betting controls */}
        {canAct && (
          <BettingControls
            inputId="pokerBetAmount"
            betAmount={betAmount}
            onBetAmountChange={setBetAmount}
            minRaise={minRaise}
            maxBetAmount={state?.maxBetAmount}
            hasOutstandingBet={hasOutstandingBet}
            loading={loading}
            onCall={() => exec('call')}
            onRaise={() => exec('raise', undefined, betAmount)}
            onBet={() => exec('bet', undefined, betAmount)}
            onCheck={() => exec('check')}
            onFold={() => exec('fold')}
            onAllIn={() => exec('allin')}
          />
        )}

        {/* Draw odds panel */}
        {canExchange && odds && odds.some((o) => o.probability > 0) && (
          <div className="bg-black/40 rounded-lg px-4 py-2 mb-2 text-white text-xs" data-testid="odds-panel">
            <div className="font-bold mb-1">{t('drawOdds')}</div>
            {odds
              .filter((o) => o.probability > 0)
              .map((o) => (
                <div key={o.handRank} className="flex justify-between">
                  <span>{o.handName}</span>
                  <span>{(o.probability * 100).toFixed(1)}%</span>
                </div>
              ))}
          </div>
        )}

        {/* Exchange controls */}
        {canExchange && (
          <div className="text-center mb-2">
            <button
              type="button"
              className={`${btnWarning} min-w-[90px]`}
              disabled={loading}
              onClick={() => exec('exchange', selected)}
            >
              {t('exchangeLabel')}
            </button>
            <button
              type="button"
              className={`${btnSuccess} min-w-[90px]`}
              disabled={loading}
              onClick={() => exec('stand')}
            >
              {t('standLabel')}
            </button>
          </div>
        )}

        {/* Settings + Reset */}
        <div className="text-center flex items-center justify-center gap-3">
          <label className="text-white text-sm flex items-center gap-1">
            {tc('betting.bettingLimit')}
            <select
              value={bettingLimit}
              onChange={(e) => setBettingLimit(Number(e.target.value))}
              className="px-2 py-1 text-sm rounded bg-white/90 text-gray-900"
            >
              <option value={0}>{tc('betting.fixed')}</option>
              <option value={1}>{tc('betting.potLimit')}</option>
              <option value={2}>{tc('betting.noLimit')}</option>
            </select>
          </label>
          <label className="text-white text-sm flex items-center gap-1">
            <input type="checkbox" checked={isLowball} onChange={(e) => setIsLowball(e.target.checked)} />
            {t('lowball')}
          </label>
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => {
              hideActionLog();
              exec('reset', undefined, undefined, { bettingLimit, isLowball });
            }}
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
    </div>
  );
}
