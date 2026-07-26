import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { razzApi } from '../api/gameApi';
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
import { formatRazzLow, razzBestLow } from '../utils/razzLow';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Razz tutorial step definitions. */
const RAZZ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="razz-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="razz-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="razz-pot-display"]',
    messageKey: 'tutorial.potDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="razz-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="razz-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const RAZZ_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SevenCardStudPhase.THIRD_STREET]: 'thirdStreet',
  [SevenCardStudPhase.FOURTH_STREET]: 'fourthStreet',
  [SevenCardStudPhase.FIFTH_STREET]: 'fifthStreet',
  [SevenCardStudPhase.SIXTH_STREET]: 'sixthStreet',
  [SevenCardStudPhase.SEVENTH_STREET]: 'seventhStreet',
  [SevenCardStudPhase.SHOWDOWN]: 'showdown',
  [SevenCardStudPhase.END]: 'end',
  [SevenCardStudPhase.REBUY]: 'rebuy',
};

/** Renders the Razz game page with door cards, betting, and showdown. */
export const RazzPage = withTutorial(RazzPageContent, 'razz', RAZZ_TUTORIAL_STEPS);
/** Inner content of the Razz page, wrapped by TutorialProvider. */
function RazzPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('razz');
  const phaseNames = usePhaseNames('razz', RAZZ_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const isMobile = useIsMobile();
  const { state, loading, error, exec: execApi, retry } = useGameApi(razzApi.exec);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('razz');
  type RazzArgs = Parameters<typeof razzApi.exec>;
  const cliConfig: CliGameConfig<SevenCardStudResponse, RazzArgs> = useMemo(
    () => ({
      gameName: 'razz',
      parseCommand: (input: string): CliParseResult<RazzArgs> => {
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
  // Tournament / ante config applied on the next reset (mirrors the CUI's tournament + ante options).
  const [ante, setAnte] = useState(1);
  const [tournamentMode, setTournamentMode] = useState(false);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('razz', state);
  const turnStartRef = useRef(0);

  useMountReset(execApi);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, { ante, tournamentMode, cpuMetaAI });
  }, [execApi, hideActionLog, ante, tournamentMode, cpuMetaAI]);

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
  // Current best Razz low from the human's known cards (door + hole), shown from 3rd street.
  const razzLow =
    isActive && humanPlayer && !humanPlayer.folded
      ? razzBestLow([...(humanPlayer.doorCards ?? []), ...(humanPlayer.holeCards ?? [])])
      : null;
  const minRaise = state?.minRaise ?? 0;
  const isMuckPhase = phase === SevenCardStudPhase.SHOWDOWN && state?.muckAvailable === true;
  const isRebuyPhase =
    phase === SevenCardStudPhase.REBUY && state?.rebuyPhaseType === SevenCardStudRebuyPhaseType.REBUY;
  const isAddonPhase =
    phase === SevenCardStudPhase.REBUY && state?.rebuyPhaseType === SevenCardStudRebuyPhaseType.ADDON;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanRebuyCount = state?.rebuyCounts?.[humanIdx] ?? 0;
  const cpuPlayers = useMemo(() => state?.players?.filter((p) => !p.isHuman) ?? [], [state?.players]);
  // Bring-in player (highest door card, posts the forced bet). Shown only during 3rd street,
  // the single round where the bring-in is relevant. `bringInPlayerIdx` indexes into state.players.
  const bringInIdx = state?.bringInPlayerIdx ?? -1;
  const bringInPlayerId =
    phase === SevenCardStudPhase.THIRD_STREET && bringInIdx >= 0 ? state?.players?.[bringInIdx]?.id : undefined;
  const bringInCardClass = 'ring-2 ring-game-status-active rounded p-0.5';
  const bringInBadgeClass =
    'inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 bg-game-status-active text-game-text-strong';

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
        gameKey="razz"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
      />
    );

  return (
    <GamePageShell
      title={tc('nav.razz')}
      gameThemeBg={gameTheme.razz.bg}
      phaseName={phaseNames[phase] ?? t('phase.init')}
      isHumanTurn={canAct}
      gamePath="/razz"
      gameEndFlag={phase === SevenCardStudPhase.END}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="razz-pot-display">
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
            <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="razz-cpu-area">
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
                    {p.id === bringInPlayerId && (
                      <span data-testid={`razz-bringin-badge-${p.id}`} className={bringInBadgeClass}>
                        {t('bringIn')}
                      </span>
                    )}
                    {isShowdown && !p.folded && p.handName && (
                      <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                        {p.handName}
                      </span>
                    )}
                  </div>
                  {/* Door cards (always visible) */}
                  <div className="text-ds-text-muted text-xs mb-0.5">{t('doorCards')}</div>
                  <div className={`flex flex-wrap gap-1 mb-1 ${p.id === bringInPlayerId ? bringInCardClass : ''}`}>
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
                  {
                    type: 'select' as const,
                    id: 'razzAnte',
                    label: t('settings.ante'),
                    tooltip: t('settings.anteHelp'),
                    value: ante,
                    options: [
                      { value: 1, label: '1' },
                      { value: 2, label: '2' },
                      { value: 5, label: '5' },
                      { value: 10, label: '10' },
                    ],
                    onSelect: (v) => setAnte(Number(v)),
                    testId: 'razz-ante-select',
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'razzTournamentMode',
                    label: t('settings.tournamentMode'),
                    tooltip: t('settings.tournamentModeHelp'),
                    checked: tournamentMode,
                    onToggle: setTournamentMode,
                  },
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
          <GameFooter className={`${gameTheme.razz.footer} px-5 py-3`}>
            {/* Human player */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="razz-player-hand">
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
                  {humanPlayer.id === bringInPlayerId && (
                    <span data-testid={`razz-bringin-badge-${humanPlayer.id}`} className={bringInBadgeClass}>
                      {t('bringIn')}
                    </span>
                  )}
                  {razzLow && (
                    <span
                      data-testid="razz-best-low"
                      className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}
                    >
                      {razzLow.complete ? t('currentLow', { low: formatRazzLow(razzLow) }) : t('currentLowIncomplete')}
                    </span>
                  )}
                  {isShowdown && !humanPlayer.folded && humanPlayer.handName && (
                    <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
                      {humanPlayer.handName}
                    </span>
                  )}
                </div>
                {/* Door cards */}
                <div className="text-ds-text-muted text-xs mb-0.5">{t('doorCards')}</div>
                <div
                  className={`flex flex-wrap gap-1.5 mb-1 ${humanPlayer.id === bringInPlayerId ? bringInCardClass : ''}`}
                >
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
              <div data-tutorial="razz-action-buttons">
                <BettingControls
                  inputId="razzBetAmount"
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
              dataTutorial="razz-reset-button"
              className="min-w-[90px]"
            />
            <ActionShortcutsPanel bindings={actionBindings} data-testid="razz-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
