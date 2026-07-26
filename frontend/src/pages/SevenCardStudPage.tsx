import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { sevenCardStudApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { BettingControls } from '../components/BettingControls';
import { CpuAccordion } from '../components/CpuAccordion';
import { CpuActionLog } from '../components/CpuActionLog';
import { CpuActionToast } from '../components/CpuActionToast';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { RoundResults } from '../components/RoundResults';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { placeholderCardStyle } from '../styles/cardStyles';
import { handNameBadgeClass } from '../styles/gameConstants';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SevenCardStudResponse } from '../types/card';
import { SevenCardStudPhase, SevenCardStudRebuyPhaseType } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';
import { evaluateBestHand, pokerHandKey } from '../utils/pokerSquaresUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

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
export const SevenCardStudPage = withTutorial(SevenCardStudPageContent, 'sevencardstud', SCS_TUTORIAL_STEPS);
/** Inner content of the Seven Card Stud page, wrapped by TutorialProvider. */
function SevenCardStudPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sevencardstud');
  const phaseNames = usePhaseNames('sevencardstud', SCS_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
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

  useMountReset(execApi);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, { cpuMetaAI });
  }, [execApi, hideActionLog, cpuMetaAI]);

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

  // Live best-hand strength for the human, computed from their visible door +
  // hole cards (best 5 of up to 7). Shown during the streets to aid betting
  // decisions; the server-supplied `handName` takes over at showdown. Only the
  // human's cards are used, so no opponent information leaks.
  const currentHandKey = useMemo(() => {
    if (!humanPlayer) return null;
    const cards = [...(humanPlayer.doorCards ?? []), ...(humanPlayer.holeCards ?? [])];
    const rank = evaluateBestHand(cards);
    return rank == null ? null : pokerHandKey(rank);
  }, [humanPlayer]);

  const actionBindings = useMemo(
    () => [
      {
        key: 'c',
        action: () => execApi('call', undefined, undefined, getElapsed()),
        enabled: hasOutstandingBet,
        labelKey: 'kbd.action.call',
      },
      {
        key: 'r',
        action: () =>
          hasOutstandingBet
            ? execApi('raise', betAmount, undefined, getElapsed())
            : execApi('bet', betAmount, undefined, getElapsed()),
        labelKey: 'kbd.action.raiseOrBet',
      },
      {
        key: 'k',
        action: () => execApi('check', undefined, undefined, getElapsed()),
        enabled: !hasOutstandingBet,
        labelKey: 'kbd.action.check',
      },
      { key: 'f', action: () => execApi('fold', undefined, undefined, getElapsed()), labelKey: 'kbd.action.fold' },
      { key: 'a', action: () => execApi('allin', undefined, undefined, getElapsed()), labelKey: 'kbd.action.allin' },
    ],
    [execApi, hasOutstandingBet, betAmount, getElapsed],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: canAct && !loading,
  });

  if (!state)
    return (
      <GameSkeleton
        gameKey="sevencardstud"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.sevencardstud')}
      gameThemeBg={gameTheme.sevencardstud.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath="/sevencardstud"
      gameEndFlag={phase === SevenCardStudPhase.SHOWDOWN || phase === SevenCardStudPhase.END}
      winShow={phase === SevenCardStudPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
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
            {tc('label.dealer')} <strong>{findPlayerName(state.players, state.dealerIdx)}</strong>
          </span>
          {state?.tournamentMode && (
            <span>{t('handNumber', { count: state.handCount, level: state.anteLevelHands })}</span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
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
                  <div className="text-ds-text-primary text-sm mb-1">
                    CPU {p.id}
                    <span className="ml-2 text-xs text-ds-text-muted">{p.playStyleName}</span>
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
                  <div className="text-ds-text-muted text-xs mb-0.5">{t('doorCards')}</div>
                  <div className="flex flex-wrap gap-1 mb-1">
                    {p.doorCards?.length
                      ? p.doorCards.map((card) => (
                          <AnimatedCard
                            key={`${card.design}-${card.value}`}
                            card={card}
                            width={cardWidth}
                            style={placeholderCardStyle}
                          />
                        ))
                      : !p.folded &&
                        Array.from({ length: 4 }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
                  </div>
                  {/* Hole cards (face-down unless showdown) */}
                  <div className="text-ds-text-muted text-xs mb-0.5">{t('holeCards')}</div>
                  <div className="flex flex-wrap gap-1">
                    {isShowdown && !p.folded && p.holeCards?.length
                      ? p.holeCards.map((card) => (
                          <AnimatedCard
                            key={`${card.design}-${card.value}`}
                            card={card}
                            width={cardWidth}
                            style={placeholderCardStyle}
                          />
                        ))
                      : !p.folded &&
                        Array.from({
                          length: (state?.phase ?? 0) >= SevenCardStudPhase.SEVENTH_STREET ? 3 : 2,
                        }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
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

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                  {
                    type: 'checkbox' as const,
                    id: 'cpuMetaAI',
                    label: t('settings.cpuMetaAI'),
                    checked: cpuMetaAI,
                    onToggle: setCpuMetaAI,
                  },
                ],
              },
            ]}
          />

          {/* Sticky footer: player hand + buttons */}
          <GameFooter className={`${gameTheme.sevencardstud.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="scs-player-hand">
                <div className="text-ds-text-primary text-lg mb-1">
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
                  {isActive && !humanPlayer.folded && currentHandKey && (
                    <span
                      data-testid="scs-current-hand"
                      className="inline-block ml-2 text-xs rounded px-2 py-0.5 bg-black/25 text-ds-text-muted"
                    >
                      {t('currentHand')}: {t(`hand.${currentHandKey}`)}
                    </span>
                  )}
                </div>
                {/* Door cards */}
                <div className="text-ds-text-muted text-xs mb-0.5">{t('doorCards')}</div>
                <div className="flex flex-wrap gap-1.5 mb-1">
                  {humanPlayer.doorCards?.length
                    ? humanPlayer.doorCards.map((card) => (
                        <AnimatedCard
                          key={`${card.design}-${card.value}`}
                          card={card}
                          width={cardWidth}
                          style={placeholderCardStyle}
                        />
                      ))
                    : !humanPlayer.folded &&
                      Array.from({ length: 4 }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
                </div>
                {/* Hole cards */}
                <div className="text-ds-text-muted text-xs mb-0.5">{t('holeCards')}</div>
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {humanPlayer.holeCards?.length
                    ? humanPlayer.holeCards.map((card) => (
                        <AnimatedCard
                          key={`${card.design}-${card.value}`}
                          card={card}
                          width={cardWidth}
                          style={placeholderCardStyle}
                        />
                      ))
                    : !humanPlayer.folded &&
                      Array.from({ length: 3 }).map((_, i) => <AnimatedCardBack key={i} width={cardWidth} />)}
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
                <p className="text-ds-text-primary mb-2">
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
                <p className="text-ds-text-primary mb-2">{t('addon.prompt', { chips: state?.addonChips })}</p>
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

            <GameResetButton
              isGameEnd={phase === SevenCardStudPhase.SHOWDOWN || phase === SevenCardStudPhase.END}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
              dataTutorial="scs-reset-button"
              className="min-w-[90px]"
            />
            <ActionShortcutsPanel bindings={actionBindings} data-testid="seven-card-stud-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
