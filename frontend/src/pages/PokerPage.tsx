import { useCallback, useEffect, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { RoundResults } from '../components/RoundResults';
import { PokerSkeleton } from '../components/skeleton/PokerSkeleton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { usePokerGame } from '../hooks/usePokerGame';
import { btnPrimary, btnSuccess, btnWarning, focusRingBlue } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { PokerPhase } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';

const POKER_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PokerPhase.INIT]: 'init',
  [PokerPhase.DEAL]: 'deal',
  [PokerPhase.EXCHANGE]: 'exchange',
  [PokerPhase.SECOND_BET]: 'secondBet',
  [PokerPhase.END]: 'end',
};

/** Renders the 5-card Draw Poker game page with betting and card exchange. */
export function PokerPage() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('poker');
  const phaseNames = usePhaseNames('poker', POKER_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec, selected, toggleCard, clearSelection, odds, canExchange } = usePokerGame();
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
  const cardCount = humanPlayer?.cards?.length ?? 0;

  useCardKeyboardNav({
    cardCount,
    onToggle: toggleCard,
    onConfirm: useCallback(() => {
      if (canExchange && !loading) exec('exchange', selected);
    }, [canExchange, loading, exec, selected]),
    onClear: clearSelection,
    enabled: canExchange,
  });

  if (!state) return <PokerSkeleton />;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green-poker" aria-busy={loading} aria-live="polite">
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseNames[phase] ?? t('phase.init')} isHumanTurn={canAct || canExchange}>
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
      </PhaseIndicator>

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
      <GameFooter className="bg-game-bg-green-poker-dark border-white/20 px-5 py-3">
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
                    className={`${focusRingBlue} rounded`}
                    style={{
                      background: 'none',
                      padding: 0,
                      cursor: canExchange ? 'pointer' : 'default',
                      borderRadius: 8,
                      ...selectedCardStyle(isSelected),
                      boxSizing: 'border-box',
                    }}
                  >
                    <CardImage card={card} width={cardWidth} />
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
        <ActionLogSection
          isEndPhase={!!state?.gameEndFlag}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />

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
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                exec('reset', undefined, undefined, { bettingLimit, isLowball });
              })
            }
          >
            {tc('button.reset')}
          </button>
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
