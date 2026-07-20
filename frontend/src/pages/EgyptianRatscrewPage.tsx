import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { egyptianRatscrewApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { SlapBurst, type SlapOutcome } from '../components/SlapBurst';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useReflexShortcuts } from '../hooks/useReflexShortcuts';
import { gameTheme } from '../styles/gameTheme';
import type { EgyptianRatscrewResponse } from '../types/card';
import {
  EgyptianRatscrewEventKind,
  EgyptianRatscrewPendingKind,
  EgyptianRatscrewPhase,
  EgyptianRatscrewSlapReason,
} from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { formatEgyptianRatscrewState } from '../utils/cli/formatters/egyptianratscrewFormatter';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

type EgyptianRatscrewArgs = Parameters<typeof egyptianRatscrewApi.exec>;

/** CPU tick interval (ms) — drives CPU step + slap reaction polling.
 * Hard difficulty has μ=300ms σ=120ms; 100ms keeps the distribution intact. */
const ER_TICK_INTERVAL_MS = 100;

/** Tutorial steps for the Egyptian Ratscrew page. */
const ER_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="er-cpu-pile"]', messageKey: 'tutorial.cpuPile', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="er-arena"]', messageKey: 'tutorial.arena', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="er-player-pile"]',
    messageKey: 'tutorial.playerPile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="er-step-button"]',
    messageKey: 'tutorial.stepButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="er-slap-button"]',
    messageKey: 'tutorial.slapButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Egyptian Ratscrew game page. */
export const EgyptianRatscrewPage = withTutorial(EgyptianRatscrewPageContent, 'egyptianratscrew', ER_TUTORIAL_STEPS);
function EgyptianRatscrewPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('egyptianratscrew');
  const { state, loading, error, exec: execApi, retry } = useGameApi(egyptianRatscrewApi.exec);
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('egyptianratscrew', state);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  const handleSlap = useCallback(() => execApi('slap'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

  // SlapBurst trigger: refire whenever a new SLAP_CORRECT or SLAP_WRONG event arrives.
  const [slapBurst, setSlapBurst] = useState<{ key: number; outcome: SlapOutcome; label: string }>({
    key: 0,
    outcome: 'correct',
    label: '',
  });
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
      (kind === EgyptianRatscrewEventKind.SLAP_CORRECT || kind === EgyptianRatscrewEventKind.SLAP_WRONG) &&
      (kind !== prev.kind || player !== prev.player)
    ) {
      const outcome: SlapOutcome = kind === EgyptianRatscrewEventKind.SLAP_CORRECT ? 'correct' : 'wrong';
      const label =
        outcome === 'wrong'
          ? t('egyptianratscrew.burst.miss')
          : state.lastSlapReason === EgyptianRatscrewSlapReason.SANDWICH
            ? t('egyptianratscrew.burst.sandwich')
            : t('egyptianratscrew.burst.pair');
      // Incrementing counter is more robust than Date.now() for trigger keys:
      // back-to-back events within the same ms still register as distinct.
      setSlapBurst((prevBurst) => ({ key: prevBurst.key + 1, outcome, label }));
      const slapper = player === 0 ? tc('player.you') : tc('player.cpu', { id: player });
      if (outcome === 'correct') {
        const reason =
          state.lastSlapReason === EgyptianRatscrewSlapReason.SANDWICH
            ? t('egyptianratscrew.slapReason.sandwich')
            : t('egyptianratscrew.slapReason.pair');
        setSlapAnnounce(t('egyptianratscrew.slapAnnounce.correct', { player: slapper, reason }));
      } else {
        setSlapAnnounce(t('egyptianratscrew.slapAnnounce.wrong', { player: slapper }));
      }
      prevSlapEventRef.current = { kind, player };
    }
  }, [state, t, tc]);

  useMountReset(execApi);

  // CPU tick driver: poll only while a CPU action is pending. Narrow deps so
  // the interval is not torn down on every state change.
  const isCpuPending = state?.pendingKind !== undefined && state.pendingKind !== EgyptianRatscrewPendingKind.NONE;
  const isGameRunning = !!state && !state.gameEndFlag;
  useEffect(() => {
    if (!isGameRunning || !isCpuPending) return;
    const id = window.setInterval(() => {
      void execApi('tick');
    }, ER_TICK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [isGameRunning, isCpuPending, execApi]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('egyptianratscrew');
  const cliConfig: CliGameConfig<EgyptianRatscrewResponse, EgyptianRatscrewArgs> = useMemo(
    () => ({
      gameName: 'egyptianratscrew',
      parseCommand: (input: string): CliParseResult<EgyptianRatscrewArgs> => {
        const cmd = input.trim().toLowerCase();
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'step' || cmd === 's') return { args: ['step'] };
        if (cmd === 'slap' || cmd === 'j') return { args: ['slap'] };
        if (cmd === 'tick') return { args: ['tick'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: formatEgyptianRatscrewState,
      helpText: [
        's/step  - Flip top of stock onto pile',
        'j/slap  - Slap the pile (when pair or sandwich is on top)',
        'tick    - Advance CPU by one tick',
        'r/reset - Reset game',
        'l/log   - Show action log',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const reflexShortcutsEnabled =
    !!state && !state.gameEndFlag && state.phase !== EgyptianRatscrewPhase.GAME_END && !cliEnabled && !loading;
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
    return (
      <div
        className={`flex-1 flex flex-col min-h-0 ${gameTheme.egyptianratscrew.bg} items-center justify-center text-ds-text-muted`}
      >
        Loading…
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag || state.phase === EgyptianRatscrewPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const human = state.players[0];
  const cpu = state.players[1];

  const phaseName = isGameEnd
    ? t('phase.end')
    : state.isSlappable
      ? t('phase.slap')
      : state.chanceRemaining > 0
        ? t('phase.chance')
        : t('phase.play');
  const lastEvent = state.lastEventKind;

  return (
    <GamePageShell
      title={tc('nav.egyptianratscrew')}
      gameThemeBg={gameTheme.egyptianratscrew.bg}
      phaseName={phaseName}
      isHumanTurn={isGameEnd ? undefined : state.isHumanTurn}
      gamePath="/egyptianratscrew"
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
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {/* CPU pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="er-cpu-pile">
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
                state.isSlappable ? 'bg-ds-warning/30' : 'bg-black/20'
              } ${lastEvent === EgyptianRatscrewEventKind.SLAP_WRONG ? 'ring-2 ring-ds-error' : ''}`}
              data-tutorial="er-arena"
            >
              <SlapBurst triggerKey={slapBurst.key} outcome={slapBurst.outcome} label={slapBurst.label} />
              {/* Screen-reader announcement for the slap outcome (parity with Slapjack #2607). */}
              <div
                className="sr-only"
                role="status"
                aria-live="polite"
                aria-atomic="true"
                data-testid="er-slap-announce"
              >
                {slapAnnounce}
              </div>
              <div className="text-center">
                <div className="text-sm text-ds-text-primary font-semibold">
                  {t('label.pileCount', { count: state.centerPileSize })}
                </div>
                {state.chanceRemaining > 0 && (
                  <div className="mt-1 flex flex-col items-center gap-0.5" data-testid="er-chance-row">
                    {/* Dot row: one filled pip per remaining flip. Keyed on the count so it
                        re-mounts and re-pulses each time a chance is consumed (#2608). */}
                    <div
                      key={`chance-${state.chanceRemaining}`}
                      className="flex items-center justify-center gap-1 animate-pulse"
                      aria-hidden="true"
                    >
                      {Array.from({ length: state.chanceRemaining }, (_, i) => (
                        <span key={`dot${i}`} className="inline-block h-2 w-2 rounded-full bg-ds-warning" />
                      ))}
                    </div>
                    <div className="text-xs text-ds-warning" role="status">
                      {t('label.chanceRemaining', { count: state.chanceRemaining })}
                      {' — '}
                      {t('chanceResponder', {
                        player: state.isHumanTurn ? tc('player.you') : tc('player.cpu', { id: state.currentTurnIdx }),
                      })}
                    </div>
                  </div>
                )}
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
                {state.isSlappable && (
                  <div className="text-base text-ds-warning font-bold mt-2 animate-pulse">
                    {t('egyptianratscrew.slappable')}
                  </div>
                )}
                {state.isSlappable && state.lastSlapReason !== EgyptianRatscrewSlapReason.NONE && (
                  <div
                    data-testid="er-slap-reason"
                    className="mt-1 inline-flex items-center gap-1 rounded-full bg-ds-warning/20 px-2 py-0.5 text-xs font-medium text-ds-warning"
                  >
                    {state.lastSlapReason === EgyptianRatscrewSlapReason.PAIR
                      ? `👯 ${t('egyptianratscrew.slapReason.pair')}`
                      : `🥪 ${t('egyptianratscrew.slapReason.sandwich')}`}
                  </div>
                )}
              </div>
            </div>

            {/* Human pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="er-player-pile">
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
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}
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
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.egyptianratscrew.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                onClick={handleStep}
                disabled={loading || isGameEnd || !state.isHumanTurn}
                className="min-h-[44px] min-w-[44px] px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="er-step-button"
              >
                {t('button.step')}
                <KbdBadge label={t('kbd.step')} />
              </button>
              <button
                type="button"
                onClick={handleSlap}
                disabled={loading || isGameEnd || state.centerPileSize === 0}
                className={`min-h-[44px] min-w-[44px] px-6 py-2 rounded-lg text-white font-bold disabled:opacity-40 disabled:cursor-not-allowed ${
                  state.isSlappable
                    ? 'bg-ds-warning hover:bg-ds-warning-hover animate-pulse'
                    : 'bg-ds-error hover:bg-ds-error'
                }`}
                data-testid="slap-button"
                data-tutorial="er-slap-button"
              >
                {t('egyptianratscrew.slap')}
                <KbdBadge label={t('kbd.slap')} />
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="er-reset-button"
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
