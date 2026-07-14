import { useCallback, useMemo, useState } from 'react';
import { beggarmyneighbourApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { gameTheme } from '../styles/gameTheme';
import type { BeggarMyNeighbourResponse } from '../types/card';
import { BeggarMyNeighbourPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BEGGARMYNEIGHBOUR_HELP, parseBeggarMyNeighbourCommand } from '../utils/cli/commands/beggarmyneighbourCommands';
import { formatBeggarMyNeighbourState } from '../utils/cli/formatters/beggarmyneighbourFormatter';
import type { CliGameConfig } from '../utils/cli/types';

type BeggarMyNeighbourArgs = Parameters<typeof beggarmyneighbourApi.exec>;

const DEFAULT_MAX_ROUNDS = 2000;

/** Tutorial steps for the Beggar-My-Neighbour game. */
const BMN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bmn-cpu-pile"]',
    messageKey: 'tutorial.cpuPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-central-pile"]',
    messageKey: 'tutorial.centralPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-player-pile"]',
    messageKey: 'tutorial.playerPile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-step-button"]',
    messageKey: 'tutorial.stepButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-autoplay-button"]',
    messageKey: 'tutorial.autoplayButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Returns emphasis ring classes for the central pile based on the current phase. */
function centralPileEmphasis(phase: number): string {
  if (phase === BeggarMyNeighbourPhase.PAY_PENALTY) return 'ring-2 ring-ds-warning';
  if (phase === BeggarMyNeighbourPhase.COLLECT) return 'ring-2 ring-ds-success';
  return '';
}

/** Renders the Beggar-My-Neighbour game page. */
export const BeggarMyNeighbourPage = withTutorial(
  BeggarMyNeighbourPageContent,
  'beggarmyneighbour',
  BMN_TUTORIAL_STEPS,
);

/** Inner content of the Beggar-My-Neighbour page. */
function BeggarMyNeighbourPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('beggarmyneighbour');
  const { state, loading, error, exec: execApi, retry } = useGameApi(beggarmyneighbourApi.exec);
  const { cardWidth } = useCardDimensions();
  const [maxRounds, setMaxRounds] = useState(DEFAULT_MAX_ROUNDS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('beggarmyneighbour', state);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  const handleAutoPlay = useCallback(() => execApi('autoplay'), [execApi]);
  const handleReset = useCallback(() => execApi('reset', { maxRounds }), [execApi, maxRounds]);

  useMountReset(execApi);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } =
    useCliMode('beggarmyneighbour');
  const cliConfig: CliGameConfig<BeggarMyNeighbourResponse, BeggarMyNeighbourArgs> = useMemo(
    () => ({
      gameName: 'beggarmyneighbour',
      parseCommand: parseBeggarMyNeighbourCommand,
      formatResponse: formatBeggarMyNeighbourState,
      helpText: BEGGARMYNEIGHBOUR_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state || state.players.length < 2)
    return <GameSkeleton gameKey="beggarmyneighbour" layout={{ kind: 'centered', rows: [2], gap: 'wide' }} />;

  const isGameEnd = state.gameEndFlag || state.phase === BeggarMyNeighbourPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const human = state.players[0];
  const cpu = state.players[1];

  const phaseName = isGameEnd
    ? t('phase.end')
    : state.phase === BeggarMyNeighbourPhase.PAY_PENALTY
      ? t('phase.payPenalty')
      : state.phase === BeggarMyNeighbourPhase.COLLECT
        ? t('phase.collect')
        : t('phase.play');

  // Phase transitions (and the penalty countdown) are conveyed only by the
  // central-pile ring color, so mirror them into an sr-only live region.
  const phaseAnnouncement =
    state.phase === BeggarMyNeighbourPhase.PAY_PENALTY
      ? t('phaseAnnouncePenalty', { phase: phaseName, count: state.penaltyRemaining })
      : t('phaseAnnounce', { phase: phaseName });

  return (
    <GamePageShell
      title={tc('nav.beggarmyneighbour')}
      gameThemeBg={gameTheme.beggarmyneighbour.bg}
      phaseName={phaseName}
      isHumanTurn={!isGameEnd}
      gamePath="/beggarmyneighbour"
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

            {/* Announce the phase (and penalty countdown) to screen readers. */}
            <div className="sr-only" role="status" aria-live="polite" data-testid="bmn-phase-announce">
              {phaseAnnouncement}
            </div>

            {/* CPU pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="bmn-cpu-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.cpu', { id: 1 })} — {t('label.drawPile')}: {cpu.drawPileSize} / {t('label.discardPile')}:{' '}
                  {cpu.discardPileSize}
                </div>
                {/* Decorative: the count is already conveyed by the text line above,
                    so hide the card back from AT to avoid a redundant double read. */}
                <div aria-hidden="true">
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
            </div>

            {/* Central pile area */}
            <div
              className={`flex items-center justify-center gap-8 py-3 bg-black/20 rounded-lg ${centralPileEmphasis(state.phase)}`}
              data-tutorial="bmn-central-pile"
            >
              <div className="text-center">
                <div className="text-sm text-ds-text-primary font-semibold">
                  {t('label.centralPile')}: {state.centralPileSize}
                </div>
                {state.phase === BeggarMyNeighbourPhase.PAY_PENALTY && (
                  <div className="text-xs text-ds-warning mt-1">
                    {t('label.penaltyRemaining')}: {state.penaltyRemaining}
                  </div>
                )}
                {state.centralPileSize > 0 ? (
                  <div
                    className="relative mx-auto mt-1"
                    aria-hidden="true"
                    data-testid="bmn-central-pile-stack"
                    data-pile-size={state.centralPileSize}
                    style={{
                      width: cardWidth * 0.8,
                      height: cardWidth * 0.8 * 1.4 + Math.min(state.centralPileSize, 10) * 2,
                    }}
                  >
                    {Array.from({ length: Math.min(state.centralPileSize, 10) }, (_, i) => (
                      <div
                        key={i}
                        className="absolute left-0 top-0"
                        style={{ transform: `translate(${i * 1.5}px, ${i * 2}px)` }}
                      >
                        <AnimatedCardBack width={cardWidth * 0.8} />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div
                    className="rounded border border-dashed border-white/30 mx-auto mt-1"
                    style={{ width: cardWidth * 0.8, height: cardWidth * 0.8 * 1.4 }}
                  />
                )}
                {state.lastCardPlayed && (
                  <div className="mt-2">
                    <div className="text-xs text-ds-text-muted mb-1">{t('label.lastCard')}</div>
                    <AnimatedCard card={state.lastCardPlayed} width={cardWidth * 1.1} />
                  </div>
                )}
              </div>
            </div>

            {/* Player pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="bmn-player-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.you')} — {t('label.drawPile')}: {human.drawPileSize} / {t('label.discardPile')}:{' '}
                  {human.discardPileSize}
                </div>
                {/* Decorative: count is in the text line above; hide from AT. */}
                <div aria-hidden="true">
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
                    tooltip: t('settings.maxRoundsHelp'),
                    value: String(maxRounds),
                    options: [
                      { value: '500', label: '500' },
                      { value: '1000', label: '1000' },
                      { value: '2000', label: '2000' },
                      { value: '5000', label: '5000' },
                      { value: '10000', label: '10000' },
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

          <GameFooter className={`${gameTheme.beggarmyneighbour.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                onClick={handleStep}
                disabled={loading || isGameEnd}
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="bmn-step-button"
              >
                {t('button.step')}
              </button>
              <button
                type="button"
                onClick={handleAutoPlay}
                disabled={loading || isGameEnd}
                className="px-6 py-2 rounded-lg bg-ds-success hover:bg-ds-success-hover text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="autoplay-button"
                data-tutorial="bmn-autoplay-button"
              >
                {t('button.autoplay')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bmn-reset-button"
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
