import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { slapjackApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { SlapBurst, type SlapOutcome } from '../components/SlapBurst';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useReflexShortcuts } from '../hooks/useReflexShortcuts';
import i18n from '../i18n';
import { useSound } from '../providers/SoundProvider';
import { gameTheme } from '../styles/gameTheme';
import type { SlapjackResponse } from '../types/card';
import { SlapjackEventKind, SlapjackPendingKind, SlapjackPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSlapjackCommand, slapjackHelp } from '../utils/cli/commands/slapjackCommands';
import { formatSlapjackState } from '../utils/cli/formatters/slapjackFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

type SlapjackArgs = Parameters<typeof slapjackApi.exec>;

/** CPU tick interval (ms) — drives CPU step + slap reaction polling.
 * Hard difficulty has μ=300ms σ=120ms; 100ms keeps the distribution intact. */
const SLAPJACK_TICK_INTERVAL_MS = 100;

/** Tutorial steps for the Slapjack page. */
const SJ_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sj-cpu-pile"]', messageKey: 'tutorial.cpuPile', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sj-arena"]', messageKey: 'tutorial.arena', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="sj-player-pile"]',
    messageKey: 'tutorial.playerPile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sj-step-button"]',
    messageKey: 'tutorial.stepButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sj-slap-button"]',
    messageKey: 'tutorial.slapButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Slapjack game page. */
export const SlapjackPage = withTutorial(SlapjackPageContent, 'slapjack', SJ_TUTORIAL_STEPS);
function SlapjackPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('slapjack');
  const { state, loading, error, exec: execApi, retry } = useGameApi(slapjackApi.exec);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('slapjack', state);

  // Flipping the next card onto the arena is the human's tap — give it a card
  // sound (respects the global mute via SoundProvider).
  const handleStep = useCallback(() => {
    return execApi('step');
  }, [execApi]);
  const handleSlap = useCallback(() => execApi('slap'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

  // SlapBurst trigger: refire whenever a new SLAP_CORRECT or SLAP_WRONG event arrives.
  const [slapBurst, setSlapBurst] = useState<{ key: number; outcome: SlapOutcome; label: string }>({
    key: 0,
    outcome: 'correct',
    label: '',
  });
  // Screen-reader announcement for the slap outcome (who slapped + correct/miss).
  // The visual SlapBurst alone is invisible to assistive tech (#2607).
  const [slapAnnounce, setSlapAnnounce] = useState('');
  const prevSlapEventRef = useRef<{ kind: number; player: number }>({ kind: -1, player: -1 });
  useEffect(() => {
    if (!state) {
      // A reset blanks `state` momentarily; clear the live region so a fresh
      // game never surfaces the previous round's stale announcement.
      setSlapAnnounce('');
      return;
    }
    const kind = state.lastEventKind;
    const player = state.lastEventPlayerIdx;
    const prev = prevSlapEventRef.current;
    if (
      (kind === SlapjackEventKind.SLAP_CORRECT || kind === SlapjackEventKind.SLAP_WRONG) &&
      (kind !== prev.kind || player !== prev.player)
    ) {
      const outcome: SlapOutcome = kind === SlapjackEventKind.SLAP_CORRECT ? 'correct' : 'wrong';
      const label = outcome === 'wrong' ? t('slapjack.burst.miss') : t('slapjack.burst.jack');
      // Counter (not Date.now()) keeps repeated slap events distinct even
      // when they happen within the same millisecond.
      setSlapBurst((prevBurst) => ({ key: prevBurst.key + 1, outcome, label }));
      const slapper = player === 0 ? tc('player.you') : tc('player.cpu', { id: player });
      setSlapAnnounce(t(`slapjack.slapAnnounce.${outcome}`, { player: slapper }));
      // Sound only for the human's own slap so a fanfare never celebrates the
      // CPU (and a buzz never blames the player for the CPU's miss). Mute is
      // handled globally by SoundProvider.
      if (player === 0) {
        playSound(outcome === 'correct' ? 'winFanfare' : 'errorBuzz');
      }
      prevSlapEventRef.current = { kind, player };
    }
  }, [state, t, tc, playSound]);

  useMountReset(execApi);

  // CPU tick driver: poll only while a CPU action is pending (#4748).
  //
  // **手番ではなく予約でゲートする。**CPU のスラップは人間の手番中にも予約される
  // (SlapjackPendingSlap) ので、「CPU の手番中だけ」に絞ると CPU が J を叩かなくなる。
  const isCpuPending = state?.pendingKind !== undefined && state.pendingKind !== SlapjackPendingKind.NONE;
  const isGameRunning = !!state && !state.gameEndFlag;
  useEffect(() => {
    if (!isGameRunning || !isCpuPending) return;
    const id = window.setInterval(() => {
      void execApi('tick');
    }, SLAPJACK_TICK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [isGameRunning, isCpuPending, execApi]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('slapjack');
  // slapjackHelp() reads i18n internally, so depend on i18n.language to
  // re-localize the CLI help after a runtime language switch.
  // biome-ignore lint/correctness/useExhaustiveDependencies: i18n.language drives help re-localization
  const cliConfig: CliGameConfig<SlapjackResponse, SlapjackArgs> = useMemo(
    () => ({
      gameName: 'slapjack',
      parseCommand: parseSlapjackCommand,
      formatResponse: formatSlapjackState,
      helpText: slapjackHelp(),
    }),
    [i18n.language],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const reflexShortcutsEnabled =
    !!state && !state.gameEndFlag && state.phase !== SlapjackPhase.GAME_END && !cliEnabled && !loading;
  useReflexShortcuts({
    onStep: handleStep,
    onSlap: handleSlap,
    enabled: reflexShortcutsEnabled,
    // Mirror the step / slap button `disabled` predicates so the keyboard
    // shortcut never fires when the visible button is greyed out.
    stepEnabled: !!state?.isHumanTurn,
    slapEnabled: (state?.centerPileSize ?? 0) > 0,
  });

  if (!state || state.players.length < 2) {
    return <GameSkeleton gameKey="slapjack" layout={{ kind: 'centered', rows: [1, 1, 1] }} />;
  }

  const isGameEnd = state.gameEndFlag || state.phase === SlapjackPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const human = state.players[0];
  const cpu = state.players[1];

  const phaseName = isGameEnd ? t('phase.end') : state.isTopJack ? t('phase.slap') : t('phase.play');
  const lastEvent = state.lastEventKind;

  return (
    <GamePageShell
      title={tc('nav.slapjack')}
      gameThemeBg={gameTheme.slapjack.bg}
      phaseName={phaseName}
      isHumanTurn={isGameEnd ? undefined : state.isHumanTurn}
      gamePath="/slapjack"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            <ErrorAlert message={error} onRetry={retry} />

            {/* CPU pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="sj-cpu-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.cpu', { id: 1 })} — {t('label.stock')}: {cpu.stockSize}
                </div>
                {cpu.stockSize > 0 ? (
                  <AnimatedCardBack width={cardWidth * 0.9} />
                ) : (
                  <div
                    className="rounded border border-white/20"
                    style={{ width: cardWidth * 0.9, height: cardWidth * 0.9 * 1.4 }}
                  />
                )}
              </div>
            </div>

            {/* Center pile / arena */}
            <div
              className={`relative flex items-center justify-center gap-8 py-3 rounded-lg transition-colors ${
                state.isTopJack ? 'bg-ds-warning/30' : 'bg-black/20'
              } ${lastEvent === SlapjackEventKind.SLAP_WRONG ? 'ring-2 ring-ds-error' : ''}`}
              data-tutorial="sj-arena"
            >
              <SlapBurst triggerKey={slapBurst.key} outcome={slapBurst.outcome} label={slapBurst.label} />
              <div className="text-center">
                <div className="text-sm text-ds-text-primary font-semibold">
                  {t('label.pileCount', { count: state.centerPileSize })}
                </div>
                <div className="mt-2">
                  {state.topCard ? (
                    <AnimatedCard card={state.topCard} width={cardWidth * 1.2} />
                  ) : (
                    <div
                      className="rounded border border-dashed border-white/30 mx-auto"
                      style={{ width: cardWidth * 1.2, height: cardWidth * 1.2 * 1.4 }}
                    />
                  )}
                </div>
                {state.isTopJack && (
                  <div className="text-base text-ds-warning font-bold mt-2 animate-pulse">
                    {t('slapjack.jackOnTop')}
                  </div>
                )}
                {/* Screen-reader announcement for the flash slap chance. */}
                <div className="sr-only" aria-live="assertive" aria-atomic="true" data-testid="sj-jack-announce">
                  {state.isTopJack ? t('slapjack.jackAnnounce') : ''}
                </div>
                {/* Screen-reader announcement for the slap outcome (#2607). */}
                <div
                  className="sr-only"
                  role="status"
                  aria-live="polite"
                  aria-atomic="true"
                  data-testid="sj-slap-announce"
                >
                  {slapAnnounce}
                </div>
              </div>
            </div>

            {/* Human pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="sj-player-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.you')} — {t('label.stock')}: {human.stockSize}
                </div>
                {human.stockSize > 0 ? (
                  <AnimatedCardBack width={cardWidth * 0.9} />
                ) : (
                  <div
                    className="rounded border border-white/20"
                    style={{ width: cardWidth * 0.9, height: cardWidth * 0.9 * 1.4 }}
                  />
                )}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(state.cpuDifficulty),
                    options: [
                      { value: '0', label: t('settings.difficulty.easy') },
                      { value: '1', label: t('settings.difficulty.normal') },
                      { value: '2', label: t('settings.difficulty.hard') },
                    ],
                    onSelect: (v: string) => execApi('reset', { config: { cpuDifficulty: Number.parseInt(v, 10) } }),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.slapjack.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                onClick={handleStep}
                disabled={loading || isGameEnd || !state.isHumanTurn}
                className="min-h-[44px] min-w-[44px] px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="sj-step-button"
              >
                {t('button.step')}
                <KbdBadge label={t('kbd.step')} />
              </button>
              <button
                type="button"
                onClick={handleSlap}
                disabled={loading || isGameEnd || state.centerPileSize === 0}
                className={`min-h-[44px] min-w-[44px] px-6 py-2 rounded-lg text-white font-bold disabled:opacity-40 disabled:cursor-not-allowed ${
                  state.isTopJack
                    ? 'bg-ds-warning hover:bg-ds-warning-hover animate-pulse'
                    : 'bg-ds-error hover:bg-ds-error'
                }`}
                data-testid="slap-button"
                data-tutorial="sj-slap-button"
              >
                {t('slapjack.slap')}
                <KbdBadge label={t('kbd.slap')} />
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sj-reset-button"
              />
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
