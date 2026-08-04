import { useEffect, useMemo, useState } from 'react';
import type { musApi } from '../api/gameApi';
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
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_AMARRAKOS_OPTIONS, useMusGame } from '../hooks/useMusGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MusResponse } from '../types/card';
import { MusBetAction, MusPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MUS_HELP, parseMusCommand } from '../utils/cli/commands/musCommands';
import { formatMusState } from '../utils/cli/formatters/musFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { evalMusHand } from '../utils/musHandEval';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** i18n value keys for each Pares category (0=none, 1=par, 2=medias, 3=duples). */
const PARES_VALUE_KEYS: readonly string[] = ['paresNone', 'paresPar', 'paresMedias', 'paresDuples'];

/** Mus tutorial step definitions. */
const MUS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="mus-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="mus-amarrakos"]',
    messageKey: 'tutorial.amarrakos',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="mus-player-hand"]', messageKey: 'tutorial.mus', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mus-action-buttons"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="mus-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MUS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MusPhase.MUS]: 'mus',
  [MusPhase.DISCARD]: 'discard',
  [MusPhase.GRANDE]: 'grande',
  [MusPhase.CHICA]: 'chica',
  [MusPhase.PARES]: 'pares',
  [MusPhase.JUEGO]: 'juego',
  [MusPhase.SHOWDOWN]: 'showdown',
  [MusPhase.ROUND_END]: 'roundEnd',
  [MusPhase.GAME_END]: 'gameEnd',
};

/** Maps a betting-round phase to its i18n round-name key. */
const ROUND_NAME_KEYS: Readonly<Record<number, string>> = {
  [MusPhase.GRANDE]: 'grande',
  [MusPhase.CHICA]: 'chica',
  [MusPhase.PARES]: 'pares',
  [MusPhase.JUEGO]: 'juego',
};

/** Maps a round index (0..3) in `results` to its i18n round-name key. */
const ROUND_INDEX_KEYS: Readonly<Record<number, string>> = {
  0: 'grande',
  1: 'chica',
  2: 'pares',
  3: 'juego',
};

/** Renders the Mus game page: a 4-player Basque vying (betting) game. */
export const MusPage = withTutorial(MusPageContent, 'mus', MUS_TUTORIAL_STEPS);

/** Inner content of the Mus page, wrapped by TutorialProvider. */
function MusPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('mus');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    musConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleMus,
    handleDiscard,
    handleBet,
    handleNextRound,
  } = useMusGame();

  // Envido stepper amount (local UI state).
  const [envidoAmount, setEnvidoAmount] = useState(2);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('mus');
  const musCliConfig: CliGameConfig<MusResponse, Parameters<typeof musApi.exec>> = useMemo(
    () => ({
      gameName: 'mus',
      parseCommand: parseMusCommand,
      formatResponse: formatMusState,
      helpText: MUS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, musCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('mus', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('mus', MUS_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="mus" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isMusPhase = state.phase === MusPhase.MUS;
  const isDiscardPhase = state.phase === MusPhase.DISCARD;
  const isBetPhase =
    state.phase === MusPhase.GRANDE ||
    state.phase === MusPhase.CHICA ||
    state.phase === MusPhase.PARES ||
    state.phase === MusPhase.JUEGO;
  const isRoundEnd = state.phase === MusPhase.ROUND_END || state.phase === MusPhase.SHOWDOWN;
  const isGameEnd = state.phase === MusPhase.GAME_END || state.gameEndFlag;

  const roundNameKey = ROUND_NAME_KEYS[state.phase];

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.mus')}
      gameThemeBg={gameTheme.mus.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/mus"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerTeam >= 0 && state.winnerTeam === state.humanTeam}
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
                    value: musConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetAmarrakos',
                    label: t('settings.targetAmarrakos'),
                    value: musConfig.targetAmarrakos,
                    options: TARGET_AMARRAKOS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetAmarrakos', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="mus-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              {roundNameKey && <span>{t(`phase.${roundNameKey}`)}</span>}
            </div>

            {/* Amarrakos (team scores) */}
            <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-tutorial="mus-amarrakos">
              <div className="mb-1 text-ds-text-primary">{t('amarrakos')}</div>
              {state.amarrakos.map((score, team) => (
                <div key={`team-${team}`} className={team === state.humanTeam ? 'text-ds-text-primary' : ''}>
                  {t('amarrakosTeam', { team, score })}
                  {team === state.humanTeam ? ` (${t('yourTeam')})` : ''}
                </div>
              ))}
              {state.pendingStake !== 0 && (
                <div className="mt-1 text-ds-warning">
                  {t('pendingStake', { amount: state.pendingStake === -1 ? t('ordagoLabel') : state.pendingStake })}
                </div>
              )}
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })}
                </div>
              ))}
            </div>

            {/* Round (betting) results */}
            {(isRoundEnd || isGameEnd) && state.results.some((r) => r.team >= 0) && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                {state.results.map((r, i) =>
                  r.team >= 0 ? (
                    <div key={`result-${i}`}>
                      {t('roundResult.line', {
                        round: t(`roundResult.${ROUND_INDEX_KEYS[i] ?? i}`),
                        team: r.team,
                        stake: r.stake,
                      })}
                    </div>
                  ) : null,
                )}
              </div>
            )}

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
          <GameFooter className={`${gameTheme.mus.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="mus"
              />
            )}

            {humanPlayer &&
              humanPlayer.cards.length === 4 &&
              (() => {
                const evalResult = evalMusHand(humanPlayer.cards);
                const paresValue = t(`handSummary.${PARES_VALUE_KEYS[evalResult.paresCategory]}`);
                const juegoValue =
                  evalResult.points === 31
                    ? t('handSummary.juegoBest')
                    : evalResult.hasJuego
                      ? t('handSummary.juegoYes', { points: evalResult.points })
                      : t('handSummary.juegoPunto', { points: evalResult.points });
                const paresActive = state.phase === MusPhase.PARES;
                const juegoActive = state.phase === MusPhase.JUEGO;
                return (
                  <div
                    className="mt-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="mus-hand-summary"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('handSummary.title')}</div>
                    {/* Grande and Chica are two of the four betting rounds, and the
                        panel said nothing during either of them (#4721). */}
                    <div
                      className={state.phase === MusPhase.GRANDE ? 'text-ds-warning font-semibold' : ''}
                      data-testid="mus-summary-grande"
                    >
                      {t('handSummary.grande')}:{' '}
                      {evalResult.highestRank === null
                        ? t('handSummary.paresNone')
                        : t('handSummary.rankValue', { rank: evalResult.highestRank })}
                    </div>
                    <div
                      className={state.phase === MusPhase.CHICA ? 'text-ds-warning font-semibold' : ''}
                      data-testid="mus-summary-chica"
                    >
                      {t('handSummary.chica')}:{' '}
                      {evalResult.lowestRank === null
                        ? t('handSummary.paresNone')
                        : t('handSummary.rankValue', { rank: evalResult.lowestRank })}
                    </div>
                    <div className={paresActive ? 'text-ds-warning font-semibold' : ''} data-testid="mus-summary-pares">
                      {t('handSummary.pares')}: {paresValue}
                    </div>
                    <div className={juegoActive ? 'text-ds-warning font-semibold' : ''} data-testid="mus-summary-juego">
                      {t('handSummary.juego')}: {juegoValue}
                    </div>
                  </div>
                );
              })()}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="mus-action-buttons">
              {isMusPhase && isHumanTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => handleMus(true)} disabled={loading}>
                    {t('musButton')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={() => handleMus(false)} disabled={loading}>
                    {t('corteButton')}
                  </button>
                </>
              )}

              {isDiscardPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length === 0}
                  >
                    {t('discardButton')} ({t('discardSelected', { count: selectedCardIndices.length })})
                  </button>
                  <span className="text-ds-text-muted text-sm" data-testid="mus-discard-guide">
                    {selectedCardIndices.length === 0
                      ? t('discardGuide')
                      : t('discardCount', { count: selectedCardIndices.length })}
                  </span>
                </>
              )}

              {isBetPhase && isHumanTurn && (
                <>
                  {state.canPaso && (
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => handleBet(MusBetAction.PASO)}
                      disabled={loading}
                    >
                      {t('bet.paso')}
                    </button>
                  )}
                  {state.canEnvido && (
                    <fieldset className="flex items-center gap-1 border-0 p-0 m-0">
                      <legend className="sr-only">{t('envidoStepperLabel')}</legend>
                      <button
                        type="button"
                        className={btnSecondary}
                        onClick={() => setEnvidoAmount((a) => Math.max(2, a - 1))}
                        disabled={loading}
                        aria-label={t('envidoDecrease')}
                      >
                        −
                      </button>
                      <span
                        className="text-ds-text-primary text-sm min-w-[3rem] text-center"
                        aria-live="polite"
                        aria-atomic="true"
                      >
                        {t('envidoAmount', { amount: envidoAmount })}
                      </span>
                      <button
                        type="button"
                        className={btnSecondary}
                        onClick={() => setEnvidoAmount((a) => a + 1)}
                        disabled={loading}
                        aria-label={t('envidoIncrease')}
                      >
                        ＋
                      </button>
                      <button
                        type="button"
                        className={btnPrimary}
                        onClick={() => handleBet(MusBetAction.ENVIDO, envidoAmount)}
                        disabled={loading}
                      >
                        {t('bet.envido')}
                      </button>
                    </fieldset>
                  )}
                  {state.canOrdago && (
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleBet(MusBetAction.ORDAGO)}
                      disabled={loading}
                    >
                      {t('bet.ordago')}
                    </button>
                  )}
                  {state.canQuiero && (
                    <button
                      type="button"
                      className={btnSuccess}
                      onClick={() => handleBet(MusBetAction.QUIERO)}
                      disabled={loading}
                    >
                      {t('bet.quiero')}
                    </button>
                  )}
                  {state.canNoQuiero && (
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => handleBet(MusBetAction.NO_QUIERO)}
                      disabled={loading}
                    >
                      {t('bet.noquiero')}
                    </button>
                  )}
                </>
              )}

              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="mus-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
