import { useCallback, useEffect, useMemo, useState } from 'react';
import { warApi } from '../api/gameApi';
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
import { WarSkeleton } from '../components/skeleton/WarSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import type { WarResponse } from '../types/card';
import { WarPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

type WarArgs = Parameters<typeof warApi.exec>;

const DEFAULT_MAX_ROUNDS = 500;

/** Tutorial steps for the War game. */
const WR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="wr-cpu-pile"]', messageKey: 'tutorial.cpuPile', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wr-arena"]', messageKey: 'tutorial.arena', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="wr-player-pile"]',
    messageKey: 'tutorial.playerPile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wr-step-button"]',
    messageKey: 'tutorial.stepButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the War (戦争) game page. */
export function WarPage() {
  return (
    <TutorialWrapper gameName="war" steps={WR_TUTORIAL_STEPS}>
      <WarPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the War page. */
function WarPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('war');
  const { state, loading, error, exec: execApi, retry } = useGameApi(warApi.exec);
  const { cardWidth } = useCardDimensions();
  const [maxRounds, setMaxRounds] = useState(DEFAULT_MAX_ROUNDS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('war', state);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  const handleReset = useCallback(() => execApi('reset', { maxRounds }), [execApi, maxRounds]);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('war');
  const cliConfig: CliGameConfig<WarResponse, WarArgs> = useMemo(
    () => ({
      gameName: 'war',
      parseCommand: (input: string): CliParseResult<WarArgs> => {
        const cmd = input.trim().toLowerCase();
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'step' || cmd === 's') return { args: ['step'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: WarResponse) => {
        const lines: string[] = [];
        const phaseName = s.phase === WarPhase.WAR_BURY ? 'WAR' : s.phase === WarPhase.RESOLVED ? 'Resolved' : 'Reveal';
        lines.push(`Phase: ${phaseName} | Pot: ${s.warPotSize} | Round: ${s.roundsPlayed}`);
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : 'CPU';
          lines.push(`${tag}: draw=${p.drawPileSize} discard=${p.discardPileSize} total=${p.totalCards}`);
        }
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: ['s/step  - Flip next card', 'r/reset - Reset game', 'l/log   - Show action log'],
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state || state.players.length < 2) return <WarSkeleton />;

  const isGameEnd = state.gameEndFlag || state.phase === WarPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const human = state.players[0];
  const cpu = state.players[1];

  const phaseName = isGameEnd
    ? t('phase.end')
    : state.phase === WarPhase.WAR_BURY
      ? t('phase.war')
      : state.phase === WarPhase.RESOLVED
        ? t('phase.resolved')
        : t('phase.reveal');

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green" aria-busy={loading}>
      <GamePageHeading title={tc('nav.war')} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={!isGameEnd}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/war" />
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

            {/* CPU area */}
            <div className="flex items-center justify-center gap-4" data-tutorial="wr-cpu-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.cpu', { id: 1 })} — {t('label.drawPile')}: {cpu.drawPileSize} / {t('label.discardPile')}:{' '}
                  {cpu.discardPileSize}
                </div>
                {cpu.drawPileSize > 0 ? (
                  <AnimatedCardBack width={cardWidth * 0.9} />
                ) : (
                  <div
                    className="rounded border border-white/20"
                    style={{ width: cardWidth * 0.9, height: cardWidth * 0.9 * 1.4 }}
                  />
                )}
              </div>
            </div>

            {/* Revealed cards + pot */}
            <div
              className="flex items-center justify-center gap-8 py-3 bg-black/20 rounded-lg"
              data-tutorial="wr-arena"
            >
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">CPU</div>
                {state.cpuRevealed ? (
                  <AnimatedCard card={state.cpuRevealed} width={cardWidth * 1.1} />
                ) : (
                  <div
                    className="rounded border border-dashed border-white/30"
                    style={{ width: cardWidth * 1.1, height: cardWidth * 1.1 * 1.4 }}
                  />
                )}
              </div>
              <div className="text-center">
                <div className="text-sm text-ds-text-primary font-semibold">
                  {t('label.potCount', { count: state.warPotSize })}
                </div>
                <div className="text-xs text-ds-text-muted mt-1">
                  {t('label.rounds')}: {state.roundsPlayed} / {state.config.maxRounds}
                </div>
              </div>
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">{tc('player.you')}</div>
                {state.playerRevealed ? (
                  <AnimatedCard card={state.playerRevealed} width={cardWidth * 1.1} />
                ) : (
                  <div
                    className="rounded border border-dashed border-white/30"
                    style={{ width: cardWidth * 1.1, height: cardWidth * 1.1 * 1.4 }}
                  />
                )}
              </div>
            </div>

            {/* Player area */}
            <div className="flex items-center justify-center gap-4" data-tutorial="wr-player-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.you')} — {t('label.drawPile')}: {human.drawPileSize} / {t('label.discardPile')}:{' '}
                  {human.discardPileSize}
                </div>
                {human.drawPileSize > 0 ? (
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
                    id: 'maxRounds',
                    label: t('settings.maxRounds'),
                    value: String(maxRounds),
                    options: [
                      { value: '100', label: '100' },
                      { value: '250', label: '250' },
                      { value: '500', label: '500' },
                      { value: '1000', label: '1000' },
                      { value: '2000', label: '2000' },
                    ],
                    onSelect: (v: string) => setMaxRounds(Number.parseInt(v, 10)),
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

          <GameFooter className="bg-game-bg-green-dark border-white/20 px-4 py-2.5">
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                onClick={handleStep}
                disabled={loading || isGameEnd}
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="wr-step-button"
              >
                {t('button.step')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="wr-reset-button"
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
