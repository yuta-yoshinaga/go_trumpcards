import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { shortdeckApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuAccordion } from '../components/CpuAccordion';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CpuPlayerCard } from '../components/CpuPlayerCard';
import { EquityDisplay } from '../components/EquityDisplay';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { PokerTableLayout } from '../components/PokerTableLayout';
import { RoundResults } from '../components/RoundResults';
import { ShortDeckSkeleton } from '../components/skeleton/ShortDeckSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnOutline, btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { HoldemPhase, HoldemRebuyPhaseType } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';

/** Short Deck Hold'em tutorial step definitions. */
const SD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sd-community-cards"]',
    messageKey: 'tutorial.communityCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-learning-mode"]',
    messageKey: 'tutorial.learningMode',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sd-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Short Deck Hold'em tutorial configuration. */
const SD_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'shortdeck',
  steps: SD_TUTORIAL_STEPS,
};

const SHORTDECK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HoldemPhase.PRE_FLOP]: 'preFlop',
  [HoldemPhase.FLOP]: 'flop',
  [HoldemPhase.TURN]: 'turn',
  [HoldemPhase.RIVER]: 'river',
  [HoldemPhase.SHOWDOWN]: 'showdown',
  [HoldemPhase.END]: 'end',
  [HoldemPhase.REBUY]: 'rebuy',
};

function StatTooltip({ id, label, tooltipText }: { id: string; label: string; tooltipText: string }) {
  return (
    <button
      type="button"
      className="group relative cursor-help bg-transparent border-none p-0 font-inherit text-inherit inline"
      aria-describedby={id}
    >
      {label}
      <span
        id={id}
        className="pointer-events-none absolute bottom-full left-1/2 -translate-x-1/2 mb-1 whitespace-nowrap rounded bg-gray-900 px-2 py-1 text-xs text-white opacity-0 group-hover:opacity-100 group-focus-within:opacity-100"
        role="tooltip"
      >
        {tooltipText}
      </span>
    </button>
  );
}

function HudStats({ vpip, pfr, threeBet, af }: { vpip: number; pfr: number; threeBet: number; af: string }) {
  const { t } = useTranslation('shortdeck');
  return (
    <span className="ml-2 text-cyan-300 text-[0.8em] hidden sm:inline" data-testid="hud-stats">
      <StatTooltip id="tooltip-vpip" label={t('stats.vpip')} tooltipText={t('stats.vpipTooltip')} />:{vpip}%{' '}
      <StatTooltip id="tooltip-pfr" label={t('stats.pfr')} tooltipText={t('stats.pfrTooltip')} />:{pfr}%{' '}
      <StatTooltip id="tooltip-3bet" label={t('stats.threeBet')} tooltipText={t('stats.threeBetTooltip')} />:{threeBet}%{' '}
      <StatTooltip id="tooltip-af" label={t('stats.af')} tooltipText={t('stats.afTooltip')} />:{af}
    </span>
  );
}

/** Renders the Short Deck Hold'em game page with community cards, betting, and showdown. */
export function ShortDeckPage() {
  const { t: tShortDeck } = useTranslation('shortdeck');
  return (
    <TutorialProvider config={SD_TUTORIAL_CONFIG} translateMessage={tShortDeck}>
      <ShortDeckPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Short Deck Hold'em page, wrapped by TutorialProvider. */
function ShortDeckPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('shortdeck');
  const phaseNames = usePhaseNames('shortdeck', SHORTDECK_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const { state, loading, error, exec: execApi } = useGameApi(shortdeckApi.exec);
  const [betAmount, setBetAmount] = useState(20);
  const [learningMode, setLearningMode] = useState(false);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('shortdeck', state);
  const turnStartRef = useRef(0);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  useEffect(() => {
    if (state?.minRaise && state.minRaise > 0) {
      setBetAmount(state.minRaise);
    } else if (state) {
      setBetAmount(20);
    }
  }, [state]);

  useEffect(() => {
    if (state && state.currentTurn === state.players?.find((p) => p.isHuman)?.id) {
      turnStartRef.current = Date.now();
    }
  }, [state]);

  const getElapsed = useCallback(() => {
    if (!cpuMetaAI || turnStartRef.current === 0) return 0;
    const elapsed = Date.now() - turnStartRef.current;
    turnStartRef.current = 0;
    return elapsed;
  }, [cpuMetaAI]);

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
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);

  const actionBindings = useMemo(
    () => [
      { key: 'c', action: () => execApi('call', undefined, undefined, getElapsed()), enabled: hasOutstandingBet },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? execApi('raise', betAmount, undefined, getElapsed())
            : execApi('bet', betAmount, undefined, getElapsed()),
      },
      { key: 'k', action: () => execApi('check', undefined, undefined, getElapsed()), enabled: !hasOutstandingBet },
      { key: 'f', action: () => execApi('fold', undefined, undefined, getElapsed()) },
      { key: 'a', action: () => execApi('allin', undefined, undefined, getElapsed()) },
    ],
    [execApi, hasOutstandingBet, betAmount, getElapsed],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state) return <ShortDeckSkeleton />;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.shortdeck.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.shortdeck')} />
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseNames[phase] ?? t('phase.init')} isHumanTurn={canAct}>
        <span data-tutorial="sd-pot-display">
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
        <TutorialButton />
        <ManualButton gamePath="/shortdeck" />
      </PhaseIndicator>

      {/* Scrollable: community cards + CPU players */}
      <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
        {/* Community cards + CPU players (poker table layout on desktop, accordion on mobile) */}
        {(() => {
          const communityCardsContent = (
            <>
              <div className="text-white text-lg mb-1.5">{t('communityCards')}</div>
              <div className="flex flex-wrap gap-2">
                {state?.communityCards?.length
                  ? state.communityCards.map((card) => (
                      <AnimatedCard
                        key={`${card.design}-${card.value}`}
                        card={card}
                        width={cardWidth}
                        style={{ border: '3px solid transparent' }}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    ))
                  : Array.from({ length: 5 }).map((_, i) => (
                      <AnimatedCardBack key={i} width={cardWidth} onFlipComplete={() => playSound('cardFlip')} />
                    ))}
              </div>
            </>
          );
          const cpuPlayerCards = cpuPlayers.map((p) => (
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
          ));

          if (!isMobile) {
            return (
              <PokerTableLayout
                communityCardsTutorial="sd-community-cards"
                cpuAreaTutorial="sd-cpu-area"
                communityCards={communityCardsContent}
                cpuPlayers={cpuPlayerCards}
              />
            );
          }

          return (
            <>
              <div
                className="sticky top-0 z-10 bg-game-bg-green-poker pb-1 shadow-sm"
                data-tutorial="sd-community-cards"
              >
                {communityCardsContent}
              </div>
              <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="sd-cpu-area">
                {cpuPlayerCards}
              </CpuAccordion>
            </>
          );
        })()}

        {/* CPU actions: toast on mobile, inline log on desktop */}
        {isMobile ? <CpuActionToast actions={state?.cpuActions} /> : <CpuActionLog actions={state?.cpuActions} />}

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
      <GameFooter className={`${gameTheme.shortdeck.footer} px-5 py-3`}>
        {/* Learning mode toggle */}
        <div
          className="flex items-center gap-2 mb-2"
          data-testid="learning-mode-toggle"
          data-tutorial="sd-learning-mode"
        >
          <label htmlFor="learningModeCheckbox" className="text-white text-sm cursor-pointer">
            {t('learning.toggle')}
          </label>
          <input
            id="learningModeCheckbox"
            type="checkbox"
            checked={learningMode}
            onChange={(e) => setLearningMode(e.target.checked)}
          />
        </div>

        {/* Equity display */}
        {learningMode && state?.equity && state.potOdds != null && (
          <EquityDisplay equity={state.equity} potOdds={state.potOdds} />
        )}

        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2" data-tutorial="sd-player-hand">
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
                    <AnimatedCard
                      key={`${card.design}-${card.value}`}
                      card={card}
                      width={cardWidth}
                      style={{ border: '3px solid transparent' }}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  ))
                : !humanPlayer.folded &&
                  Array.from({ length: 2 }).map((_, i) => (
                    <AnimatedCardBack key={i} width={cardWidth} onFlipComplete={() => playSound('cardFlip')} />
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
                onClick={() => execApi('muck')}
              >
                {t('muck.muck')}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading}
                onClick={() => execApi('show')}
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
                onClick={() => execApi('rebuy')}
              >
                {t('rebuy.accept')}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading}
                onClick={() => execApi('skiprebuy')}
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
                onClick={() => execApi('addon')}
              >
                {t('addon.accept')}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading}
                onClick={() => execApi('skipaddon')}
              >
                {t('addon.skip')}
              </button>
            </div>
          </div>
        )}

        {/* Hint */}
        {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

        {/* Betting controls */}
        {canAct && (
          <div data-tutorial="sd-action-buttons">
            <BettingControls
              inputId="shortdeckBetAmount"
              betAmount={betAmount}
              onBetAmountChange={setBetAmount}
              minRaise={minRaise}
              maxBetAmount={state?.maxBetAmount}
              hasOutstandingBet={hasOutstandingBet}
              loading={loading}
              onCall={() => execApi('call', undefined, undefined, getElapsed())}
              onRaise={() => execApi('raise', betAmount, undefined, getElapsed())}
              onBet={() => execApi('bet', betAmount, undefined, getElapsed())}
              onCheck={() => execApi('check', undefined, undefined, getElapsed())}
              onFold={() => execApi('fold', undefined, undefined, getElapsed())}
              onAllIn={() => execApi('allin', undefined, undefined, getElapsed())}
            />
          </div>
        )}

        {/* Settings + Reset */}
        <div className="text-center flex items-center justify-center gap-3" data-tutorial="sd-reset-button">
          <label className="text-white text-sm flex items-center gap-1">
            <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
            {tc('hint.toggle', { ns: 'tutorial' })}
          </label>
          <label className="text-white text-sm flex items-center gap-1">
            <input type="checkbox" checked={cpuMetaAI} onChange={(e) => setCpuMetaAI(e.target.checked)} />
            {t('settings.cpuMetaAI')}
          </label>
          <button
            type="button"
            className={`${btnOutline} min-w-[90px]`}
            disabled={loading}
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                execApi('reset', undefined, { cpuMetaAI });
              })
            }
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
      <WinCelebration show={phase === HoldemPhase.END} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
