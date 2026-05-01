import { useCallback, useEffect, useMemo } from 'react';
import { egyptianRatscrewApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { useMountReset } from '../hooks/useMountReset';
import { gameTheme } from '../styles/gameTheme';
import type { EgyptianRatscrewResponse } from '../types/card';
import { EgyptianRatscrewEventKind, EgyptianRatscrewPendingKind, EgyptianRatscrewPhase } from '../types/phases';
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
  useGameRoundGuard(!!state && !state.gameEndFlag);
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('egyptianratscrew', state);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  const handleSlap = useCallback(() => execApi('slap'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

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
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.egyptianratscrew.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.egyptianratscrew')} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={!isGameEnd && state.isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/egyptianratscrew" />
      </PhaseIndicator>

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
              className={`flex items-center justify-center gap-8 py-3 rounded-lg transition-colors ${
                state.isSlappable ? 'bg-ds-warning/30' : 'bg-black/20'
              } ${lastEvent === EgyptianRatscrewEventKind.SLAP_WRONG ? 'ring-2 ring-ds-error' : ''}`}
              data-tutorial="er-arena"
            >
              <div className="text-center">
                <div className="text-sm text-ds-text-primary font-semibold">
                  {t('label.pileCount', { count: state.centerPileSize })}
                </div>
                {state.chanceRemaining > 0 && (
                  <div className="text-xs text-ds-warning mt-1">
                    {t('label.chanceRemaining', { count: state.chanceRemaining })}
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
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="er-step-button"
              >
                {t('button.step')}
              </button>
              <button
                type="button"
                onClick={handleSlap}
                disabled={loading || isGameEnd || state.centerPileSize === 0}
                className={`px-6 py-2 rounded-lg text-white font-bold disabled:opacity-40 disabled:cursor-not-allowed ${
                  state.isSlappable
                    ? 'bg-ds-warning hover:bg-ds-warning-hover animate-pulse'
                    : 'bg-ds-error hover:bg-ds-error'
                }`}
                data-testid="slap-button"
                data-tutorial="er-slap-button"
              >
                {t('egyptianratscrew.slap')}
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

          <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
          <WinCelebration show={isGameEnd && humanWon} />
        </>
      )}
    </div>
  );
}
