import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { holdemApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CardBack, CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { RoundResults } from '../components/RoundResults';
import { HoldemSkeleton } from '../components/skeleton/HoldemSkeleton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useActionLog } from '../hooks/useActionLog';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useConfirmDialog } from '../hooks/useConfirmDialog';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { HoldemPhase, HoldemRebuyPhaseType } from '../types/phases';

function usePhaseNames(t: (key: string) => string): Record<number, string> {
  return {
    [HoldemPhase.PRE_FLOP]: t('phase.preFlop'),
    [HoldemPhase.FLOP]: t('phase.flop'),
    [HoldemPhase.TURN]: t('phase.turn'),
    [HoldemPhase.RIVER]: t('phase.river'),
    [HoldemPhase.SHOWDOWN]: t('phase.showdown'),
    [HoldemPhase.END]: t('phase.end'),
    [HoldemPhase.REBUY]: t('phase.rebuy'),
  };
}

function StatTooltip({ id, label, tooltipText }: { id: string; label: string; tooltipText: string }) {
  return (
    // biome-ignore lint/a11y/noNoninteractiveTabindex: tabIndex needed for keyboard tooltip access
    <span className="group relative cursor-help" tabIndex={0} aria-describedby={id}>
      {label}
      <span
        id={id}
        className="pointer-events-none absolute bottom-full left-1/2 -translate-x-1/2 mb-1 whitespace-nowrap rounded bg-gray-900 px-2 py-1 text-xs text-white opacity-0 group-hover:opacity-100 group-focus-within:opacity-100"
        role="tooltip"
      >
        {tooltipText}
      </span>
    </span>
  );
}

function HudStats({ vpip, pfr, threeBet, af }: { vpip: number; pfr: number; threeBet: number; af: string }) {
  const { t } = useTranslation('holdem');
  return (
    <span className="ml-2 text-cyan-300 text-[0.8em]" data-testid="hud-stats">
      <StatTooltip id="tooltip-vpip" label={t('stats.vpip')} tooltipText={t('stats.vpipTooltip')} />:{vpip}%{' '}
      <StatTooltip id="tooltip-pfr" label={t('stats.pfr')} tooltipText={t('stats.pfrTooltip')} />:{pfr}%{' '}
      <StatTooltip id="tooltip-3bet" label={t('stats.threeBet')} tooltipText={t('stats.threeBetTooltip')} />:{threeBet}%{' '}
      <StatTooltip id="tooltip-af" label={t('stats.af')} tooltipText={t('stats.afTooltip')} />:{af}
    </span>
  );
}

export function HoldemPage() {
  const { t } = useTranslation('holdem');
  const { t: tc } = useTranslation('common');
  const phaseNames = usePhaseNames(t);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec } = useGameApi(holdemApi.exec);
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();
  const [betAmount, setBetAmount] = useState(20);
  const { actionLog, showActionLog, hideActionLog } = useActionLog('holdem');

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
  const isMuckPhase = phase === HoldemPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase = phase === HoldemPhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.REBUY;
  const isAddonPhase = phase === HoldemPhase.REBUY && state?.rebuyPhaseType === HoldemRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;

  const actionBindings = useMemo(
    () => [
      { key: 'c', action: () => exec('call'), enabled: hasOutstandingBet },
      { key: 'r', action: () => (hasOutstandingBet ? exec('raise', betAmount) : exec('bet', betAmount)) },
      { key: 'k', action: () => exec('check'), enabled: !hasOutstandingBet },
      { key: 'f', action: () => exec('fold') },
      { key: 'a', action: () => exec('allin') },
    ],
    [exec, hasOutstandingBet, betAmount],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state) return <HoldemSkeleton />;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green-poker" aria-busy={loading} aria-live="polite">
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseNames[phase] ?? t('phase.init')} isHumanTurn={canAct}>
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
      </PhaseIndicator>

      {/* Scrollable: community cards + CPU players */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        {/* Community cards */}
        <div className="mb-4">
          <div className="text-white text-lg mb-1.5">{t('communityCards')}</div>
          <div className="flex flex-wrap gap-2">
            {state?.communityCards?.length
              ? state.communityCards.map((card) => (
                  <CardImage
                    key={`${card.design}-${card.value}`}
                    card={card}
                    width={cardWidth}
                    style={{ border: '3px solid transparent' }}
                  />
                ))
              : Array.from({ length: 5 }).map((_, i) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
                  <CardBack key={i} width={cardWidth} />
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
                p.totalHands > 0 ? <HudStats vpip={p.vpip} pfr={p.pfr} threeBet={p.threeBet} af={p.af} /> : undefined
              }
            />
          ))}

        {/* CPU actions log */}
        <CpuActionLog actions={state?.cpuActions} />

        {/* Round results */}
        {isShowdown && <RoundResults results={state?.roundResults} players={state?.players ?? []} />}

        {/* Action log */}
        <ActionLogSection
          isEndPhase={!!state?.gameEndFlag}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
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
              {humanPlayer.totalHands > 0 && (
                <HudStats
                  vpip={humanPlayer.vpip}
                  pfr={humanPlayer.pfr}
                  threeBet={humanPlayer.threeBet}
                  af={humanPlayer.af}
                />
              )}
              {humanPlayer.currentBet > 0 && (
                <span className="ml-2 text-xs">
                  {tc('betting.currentBet')} {humanPlayer.currentBet}
                </span>
              )}
              {humanPlayer.folded && <span className="ml-2 text-red-300 text-xs">[{tc('status.folded')}]</span>}
              {humanPlayer.allIn && <span className="ml-2 text-yellow-300 text-xs">[{tc('status.allIn')}]</span>}
              {isShowdown && !humanPlayer.folded && humanPlayer.handName && (
                <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
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
                      width={cardWidth}
                      style={{ border: '3px solid transparent' }}
                    />
                  ))
                : !humanPlayer.folded &&
                  Array.from({ length: 2 }).map((_, i) => (
                    // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
                    <CardBack key={i} width={cardWidth} />
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

        {/* Muck/Show controls */}
        {isMuckPhase && (
          <div className="mb-2 text-center" data-testid="muck-controls">
            <div className="flex justify-center gap-2">
              <button
                type="button"
                className={`${btnPrimary} min-w-[90px]`}
                disabled={loading}
                onClick={() => exec('muck')}
              >
                {t('muck.muck')}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading}
                onClick={() => exec('show')}
              >
                {t('muck.show')}
              </button>
            </div>
          </div>
        )}

        {/* Rebuy/Addon controls */}
        {isRebuyPhase && (
          <div className="mb-2 text-center" data-testid="rebuy-controls">
            <p className="text-white mb-2">
              {t('rebuy.prompt', { chips: state?.rebuyChips, used: humanRebuyCount, max: state?.rebuyMaxCount })}
            </p>
            <div className="flex justify-center gap-2">
              <button
                type="button"
                className={`${btnPrimary} min-w-[90px]`}
                disabled={loading}
                onClick={() => exec('rebuy')}
              >
                {t('rebuy.accept')}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading}
                onClick={() => exec('skiprebuy')}
              >
                {t('rebuy.skip')}
              </button>
            </div>
          </div>
        )}
        {isAddonPhase && (
          <div className="mb-2 text-center" data-testid="addon-controls">
            <p className="text-white mb-2">{t('addon.prompt', { chips: state?.addonChips })}</p>
            <div className="flex justify-center gap-2">
              <button
                type="button"
                className={`${btnPrimary} min-w-[90px]`}
                disabled={loading}
                onClick={() => exec('addon')}
              >
                {t('addon.accept')}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading}
                onClick={() => exec('skipaddon')}
              >
                {t('addon.skip')}
              </button>
            </div>
          </div>
        )}

        {/* Betting controls */}
        {canAct && (
          <BettingControls
            inputId="holdemBetAmount"
            betAmount={betAmount}
            onBetAmountChange={setBetAmount}
            minRaise={minRaise}
            maxBetAmount={state?.maxBetAmount}
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

        {/* Reset */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                exec('reset');
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
