import { useEffect, useMemo } from 'react';
import type { aluetteApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useAluetteGame } from '../hooks/useAluetteGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { ALUETTE_HAND_SIZE, type AluetteResponse, aluetteLuetteName } from '../types/card';
import { AluettePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { ALUETTE_HELP, parseAluetteCommand } from '../utils/cli/commands/aluetteCommands';
import { formatAluetteState } from '../utils/cli/formatters/aluetteFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Aluette tutorial step definitions. */
const ALUETTE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="aluette-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="aluette-luettes"]',
    messageKey: 'tutorial.luettes',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="aluette-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="aluette-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="aluette-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const ALUETTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [AluettePhase.PLAY]: 'play',
  [AluettePhase.TRICK_END]: 'trickEnd',
  [AluettePhase.ROUND_END]: 'roundEnd',
  [AluettePhase.GAME_END]: 'gameEnd',
};

/** Renders the Aluette game page: a Breton 4-player 2v2 trick-taker with no trump suit. */
export const AluettePage = withTutorial(AluettePageContent, 'aluette', ALUETTE_TUTORIAL_STEPS);

/** Inner content of the Aluette page, wrapped by TutorialProvider. */
function AluettePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('aluette');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    aluetteConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useAluetteGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('aluette');
  const cliConfig: CliGameConfig<AluetteResponse, Parameters<typeof aluetteApi.exec>> = useMemo(
    () => ({
      gameName: 'aluette',
      parseCommand: parseAluetteCommand,
      formatResponse: formatAluetteState,
      helpText: ALUETTE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('aluette', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('aluette', ALUETTE_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton
        gameKey="aluette"
        layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: ALUETTE_HAND_SIZE }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const isHumanTurn = state.isHumanTurn;
  const luettes = state.luettes ?? [];

  const isPlayPhase = state.phase === AluettePhase.PLAY;
  const isTrickEnd = state.phase === AluettePhase.TRICK_END;
  const isRoundEnd = state.phase === AluettePhase.ROUND_END;
  const isGameEnd = state.phase === AluettePhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.aluette')}
      gameThemeBg={gameTheme.aluette.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/aluette"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerTeam === humanTeam}
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: aluetteConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetPoints',
                    label: t('settings.targetPoints'),
                    value: aluetteConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="aluette-info">
              <span className="mr-4">{t('mene', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            {/*
              **Strength is per card, not per value.** The 3 of coins is the best
              card in the deck while the 3 of swords is ordinary, so the ranking
              cannot be read off the hand — the table is permanent, not a tooltip.
            */}
            <div className="text-center mb-2 text-sm" data-tutorial="aluette-luettes" data-testid="aluette-luettes">
              <span className="font-semibold text-ds-warning">{t('luetteLegend')}: </span>
              <span className="text-ds-text-muted">
                {luettes.map((l) => `${t(`luette.${l.name}`)} (${t(`suit.${l.design}`)}${l.value})`).join(' > ')}
              </span>
            </div>

            <div className={lgTwoColGrid}>
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="aluette-trick-display"
                />
              </div>

              <div>
                {/* Team scores — this is a team game, not an individual one. */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="aluette-team-scores">
                  {state.teamScores.map((score, team) => (
                    <div
                      key={team}
                      className={`py-0.5 ${team === humanTeam ? 'text-ds-text-primary font-semibold' : ''}`}
                    >
                      {t('teamScore', { n: team, score })}
                    </div>
                  ))}
                </div>

                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}
                          {p.isDealer ? ` / ${t('dealerBadge')}` : ''}): {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}
                        {p.isDealer ? ` / ${t('dealerBadge')}` : ''}): {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('roundResult.tricks', {
                          name: playerName(p.id, p.isHuman),
                          count: state.roundTricks[p.id] ?? 0,
                        })}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.aluette.footer} px-4 py-2.5`}>
            {/*
              A luette in hand must announce itself. "♦3" alone does not tell the
              player they are holding the strongest card in the deck.
            */}
            {humanPlayer && humanPlayer.cards.length > 0 && (
              <div className="mb-1 text-center text-xs text-ds-text-muted" data-testid="aluette-hand-luettes">
                {humanPlayer.cards
                  .map((c, i) => ({ i, name: aluetteLuetteName(luettes, c) }))
                  .filter((e) => e.name)
                  .map((e) => `[${e.i}] ${t(`luette.${e.name}`)}`)
                  .join('  ') || t('noLuetteInHand')}
              </div>
            )}

            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="aluette"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="aluette-action-buttons">
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('playButton')}
                </button>
              )}
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="aluette-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
