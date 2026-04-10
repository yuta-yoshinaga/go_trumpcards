import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { sevenCardStudApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CpuAccordion } from '../components/CpuAccordion';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { RoundResults } from '../components/RoundResults';
import { HoldemSkeleton } from '../components/skeleton/HoldemSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SevenCardStudResponse } from '../types/card';
import { SevenCardStudPhase, SevenCardStudRebuyPhaseType } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

/** Seven Card Stud tutorial step definitions. */
const SCS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="scs-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scs-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scs-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scs-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scs-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SCS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SevenCardStudPhase.THIRD_STREET]: 'thirdStreet',
  [SevenCardStudPhase.FOURTH_STREET]: 'fourthStreet',
  [SevenCardStudPhase.FIFTH_STREET]: 'fifthStreet',
  [SevenCardStudPhase.SIXTH_STREET]: 'sixthStreet',
  [SevenCardStudPhase.SEVENTH_STREET]: 'seventhStreet',
  [SevenCardStudPhase.SHOWDOWN]: 'showdown',
  [SevenCardStudPhase.END]: 'end',
  [SevenCardStudPhase.REBUY]: 'rebuy',
};

/** Renders the Seven Card Stud game page with door cards, betting, and showdown. */
export function SevenCardStudPage() {
  return (
    <TutorialWrapper gameName="sevencardstud" steps={SCS_TUTORIAL_STEPS}>
      <SevenCardStudPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Seven Card Stud page, wrapped by TutorialProvider. */
function SevenCardStudPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sevencardstud');
  const phaseNames = usePhaseNames('sevencardstud', SCS_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const isMobile = useIsMobile();
  const { state, loading, error, exec: execApi, retry } = useGameApi(sevenCardStudApi.exec);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sevencardstud');
  type ScsArgs = Parameters<typeof sevenCardStudApi.exec>;
  const cliConfig: CliGameConfig<SevenCardStudResponse, ScsArgs> = useMemo(
    () => ({
      gameName: 'sevencardstud',
      parseCommand: (input: string): CliParseResult<ScsArgs> => {
        const parts = input.trim().split(/\s+/);
        const cmd = parts[0]?.toLowerCase() ?? '';
        const amount = parts[1] ? Number.parseInt(parts[1], 10) : undefined;
        const validCmds = [
          'reset',
          'fold',
          'check',
          'call',
          'bet',
          'raise',
          'allin',
          'rebuy',
          'skiprebuy',
          'addon',
          'skipaddon',
          'muck',
          'show',
        ] as const;
        type Cmd = (typeof validCmds)[number];
        if (validCmds.includes(cmd as Cmd)) {
          return { args: [cmd as Cmd, amount] };
        }
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: SevenCardStudResponse) => {
        const lines: string[] = [];
        lines.push(`Phase: ${phaseNames[s.phase] ?? 'Init'} | Pot: ${s.pot}`);
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : `CPU ${p.id}`;
          const door = p.doorCards.map((c) => `${c.design[0]}${c.value}`).join(' ');
          const hole = p.holeCards.map((c) => `${c.design[0]}${c.value}`).join(' ');
          lines.push(
            `${tag}: chips=${p.chips} door=[${door}] hole=[${hole}]${p.folded ? ' FOLDED' : ''}${p.allIn ? ' ALL-IN' : ''}`,
          );
        }
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: [
        'f/fold      - Fold',
        'ck/check    - Check',
        'c/call      - Call',
        'b/bet <amt> - Bet',
        'ra/raise <amt> - Raise',
        'a/allin     - All-in',
        'rebuy       - Rebuy chips',
        'skiprebuy   - Skip rebuy',
        'addon       - Add-on chips',
        'skipaddon   - Skip add-on',
        'muck        - Muck hand',
        'show        - Show hand',
        'r/reset     - Reset game',
      ],
    }),
    [phaseNames],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const [betAmount, setBetAmount] = useState(20);
  const [cpuMetaAI, setCpuMetaAI] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('sevencardstud', state);
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

  const phase = state?.phase ?? SevenCardStudPhase.INIT;
  const isActive = phase >= SevenCardStudPhase.THIRD_STREET && phase <= SevenCardStudPhase.SEVENTH_STREET;
  const isShowdown = phase === SevenCardStudPhase.SHOWDOWN || phase === SevenCardStudPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === SevenCardStudPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase =
    phase === SevenCardStudPhase.REBUY && state?.rebuyPhaseType === SevenCardStudRebuyPhaseType.REBUY;
  const isAddonPhase =
    phase === SevenCardStudPhase.REBUY && state?.rebuyPhaseType === SevenCardStudRebuyPhaseType.ADDON;
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

  if (!state) return <HoldemSkeleton />;

  return (
    <div
      className={`flex-1 flex flex-col min-h-0 ${gameTheme.sevencardstud.bg}`}
      aria-busy={loading}
      aria-live="polite"
    >
      <GamePageHeading title={tc('nav.sevencardstud')} />
      {/* Phase indicator + info bar */}
      <PhaseIndicator phaseName={phaseNames[phase] ?? t('phase.init')} isHumanTurn={canAct}>
        <span data-tutorial="scs-pot-display">
          {tc('label.pot')} <strong>{state?.pot ?? 0}</strong>
        </span>
        <span>
          {t('anteBringIn')}:{' '}
          <strong>
            {state?.ante ?? 0}/{state?.bringIn ?? 0}
          </strong>
        </span>
        <span>
          {tc('label.dealer')}{' '}
          <strong>
            {tc('label.player')} {state?.dealerIdx ?? 0}
          </strong>
        </span>
        {state?.tournamentMode && (
          <span>{t('handNumber', { count: state.handCount, level: state.anteLevelHands })}</span>
        )}
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/sevencardstud" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable: CPU players */}
          <div className={`flex-1 overflow-y-auto pt-4 px-5 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* CPU players */}
            <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="scs-cpu-area">
              {cpuPlayers.map((p) => (
                <div key={p.id} className="mb-3 p-2 rounded bg-black/30">
                  <div className="text-white text-sm mb-1">
                    CPU {p.id}
                    <span className="ml-2 text-xs text-gray-300">{p.playStyleName}</span>
                    <span className="ml-2 text-xs">
                      {tc('betting.chips')} {p.chips}
                    </span>
                    {p.currentBet > 0 && (
                      <span className="ml-2 text-xs">
                        {tc('betting.currentBet')} {p.currentBet}
                      </span>
                    )}
                    {p.folded && <span className="ml-2 text-ds-error text-xs">[{tc('status.folded')}]</span>}
                    {p.allIn && <span className="ml-2 text-ds-warning text-xs">[{tc('status.allIn')}]</span>}
                    {isShowdown && !p.folded && p.handName && (
                      <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                        {p.handName}
                      </span>
                    )}
                  </div>
                  {/* Door cards (always visible) */}
                  <div className="text-gray-300 text-xs mb-0.5">{t('doorCards')}</div>
                  <div className="flex flex-wrap gap-1 mb-1">
                    {p.doorCards?.length
                      ? p.doorCards.map((card) => (
                          <AnimatedCard
                            key={`${card.design}-${card.value}`}
                            card={card}
                            width={cardWidth}
                            style={{ border: '3px solid transparent' }}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
                        ))
                      : !p.folded &&
                        Array.from({ length: 4 }).map((_, i) => (
                          <AnimatedCardBack key={i} width={cardWidth} onFlipComplete={() => playSound('cardFlip')} />
                        ))}
                  </div>
                  {/* Hole cards (face-down unless showdown) */}
                  <div className="text-gray-300 text-xs mb-0.5">{t('holeCards')}</div>
                  <div className="flex flex-wrap gap-1">
                    {isShowdown && !p.folded && p.holeCards?.length
                      ? p.holeCards.map((card) => (
                          <AnimatedCard
                            key={`${card.design}-${card.value}`}
                            card={card}
                            width={cardWidth}
                            style={{ border: '3px solid transparent' }}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
                        ))
                      : !p.folded &&
                        Array.from({
                          length: (state?.phase ?? 0) >= SevenCardStudPhase.SEVENTH_STREET ? 3 : 2,
                        }).map((_, i) => (
                          <AnimatedCardBack key={i} width={cardWidth} onFlipComplete={() => playSound('cardFlip')} />
                        ))}
                  </div>
                </div>
              ))}
            </CpuAccordion>

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
          <GameFooter className={`${gameTheme.sevencardstud.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="scs-player-hand">
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
                  {humanPlayer.folded && <span className="ml-2 text-ds-error text-xs">[{tc('status.folded')}]</span>}
                  {humanPlayer.allIn && <span className="ml-2 text-ds-warning text-xs">[{tc('status.allIn')}]</span>}
                  {isShowdown && !humanPlayer.folded && humanPlayer.handName && (
                    <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                      {humanPlayer.handName}
                    </span>
                  )}
                </div>
                {/* Door cards */}
                <div className="text-gray-300 text-xs mb-0.5">{t('doorCards')}</div>
                <div className="flex flex-wrap gap-1.5 mb-1">
                  {humanPlayer.doorCards?.length
                    ? humanPlayer.doorCards.map((card) => (
                        <AnimatedCard
                          key={`${card.design}-${card.value}`}
                          card={card}
                          width={cardWidth}
                          style={{ border: '3px solid transparent' }}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))
                    : !humanPlayer.folded &&
                      Array.from({ length: 4 }).map((_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth} onFlipComplete={() => playSound('cardFlip')} />
                      ))}
                </div>
                {/* Hole cards */}
                <div className="text-gray-300 text-xs mb-0.5">{t('holeCards')}</div>
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {humanPlayer.holeCards?.length
                    ? humanPlayer.holeCards.map((card) => (
                        <AnimatedCard
                          key={`${card.design}-${card.value}`}
                          card={card}
                          width={cardWidth}
                          style={{ border: '3px solid transparent' }}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))
                    : !humanPlayer.folded &&
                      Array.from({ length: 3 }).map((_, i) => (
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

            <ErrorAlert message={error} onRetry={retry} />

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
              <div data-tutorial="scs-action-buttons">
                <BettingControls
                  inputId="sevenCardStudBetAmount"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={minRaise}
                  maxBetAmount={state?.maxBetAmount}
                  potSize={state?.pot}
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
            <div className="text-center flex items-center justify-center gap-3" data-tutorial="scs-reset-button">
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
        </>
      )}
      <WinCelebration show={phase === SevenCardStudPhase.END} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
