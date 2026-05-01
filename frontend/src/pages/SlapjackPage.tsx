import { useCallback, useEffect, useMemo } from 'react';
import { slapjackApi } from '../api/gameApi';
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
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { gameTheme } from '../styles/gameTheme';
import type { SlapjackResponse } from '../types/card';
import { SlapjackEventKind, SlapjackPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

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
export function SlapjackPage() {
  return (
    <TutorialWrapper gameName="slapjack" steps={SJ_TUTORIAL_STEPS}>
      <SlapjackPageContent />
    </TutorialWrapper>
  );
}

function SlapjackPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('slapjack');
  const { state, loading, error, exec: execApi, retry } = useGameApi(slapjackApi.exec);
  // Issue #1609: warn before tab close / reload while a round is in progress.
  useGameRoundGuard(!!state && !state.gameEndFlag);
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('slapjack', state);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  const handleSlap = useCallback(() => execApi('slap'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  // CPU tick driver while the game is active.
  useEffect(() => {
    if (!state) return;
    if (state.gameEndFlag) return;
    const id = window.setInterval(() => {
      void execApi('tick');
    }, SLAPJACK_TICK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [state, execApi]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('slapjack');
  const cliConfig: CliGameConfig<SlapjackResponse, SlapjackArgs> = useMemo(
    () => ({
      gameName: 'slapjack',
      parseCommand: (input: string): CliParseResult<SlapjackArgs> => {
        const cmd = input.trim().toLowerCase();
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'step' || cmd === 's') return { args: ['step'] };
        if (cmd === 'slap' || cmd === 'j') return { args: ['slap'] };
        if (cmd === 'tick') return { args: ['tick'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: SlapjackResponse) => {
        const lines: string[] = [];
        const phase = s.phase === SlapjackPhase.GAME_END ? 'End' : 'Play';
        const top = s.topCard ? `${s.topCard.value}` : '--';
        lines.push(`Phase: ${phase} | Pile: ${s.centerPileSize} | Top: ${top} | Turn: P${s.currentTurnIdx}`);
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : 'CPU';
          lines.push(`${tag}: stock=${p.stockSize}`);
        }
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: [
        's/step  - Flip top of stock onto pile',
        'j/slap  - Slap the pile (when J is on top)',
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
        className={`flex-1 flex flex-col min-h-0 ${gameTheme.slapjack.bg} items-center justify-center text-ds-text-muted`}
      >
        Loading…
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag || state.phase === SlapjackPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const human = state.players[0];
  const cpu = state.players[1];

  const phaseName = isGameEnd ? t('phase.end') : state.isTopJack ? t('phase.slap') : t('phase.play');
  const lastEvent = state.lastEventKind;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.slapjack.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.slapjack')} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={!isGameEnd && state.isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/slapjack" />
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
              className={`flex items-center justify-center gap-8 py-3 rounded-lg transition-colors ${
                state.isTopJack ? 'bg-ds-warning/30' : 'bg-black/20'
              } ${lastEvent === SlapjackEventKind.SLAP_WRONG ? 'ring-2 ring-ds-error' : ''}`}
              data-tutorial="sj-arena"
            >
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

          <GameFooter className={`${gameTheme.slapjack.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                onClick={handleStep}
                disabled={loading || isGameEnd || !state.isHumanTurn}
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="sj-step-button"
              >
                {t('button.step')}
              </button>
              <button
                type="button"
                onClick={handleSlap}
                disabled={loading || isGameEnd || state.centerPileSize === 0}
                className={`px-6 py-2 rounded-lg text-white font-bold disabled:opacity-40 disabled:cursor-not-allowed ${
                  state.isTopJack
                    ? 'bg-ds-warning hover:bg-ds-warning-hover animate-pulse'
                    : 'bg-ds-error hover:bg-ds-error'
                }`}
                data-testid="slap-button"
                data-tutorial="sj-slap-button"
              >
                {t('slapjack.slap')}
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

          <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
          <WinCelebration show={isGameEnd && humanWon} />
        </>
      )}
    </div>
  );
}
