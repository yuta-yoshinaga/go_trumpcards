import { useEffect, useMemo } from 'react';
import type { doppelkopfApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, TARGET_CHIPS_OPTIONS, useDoppelkopfGame } from '../hooks/useDoppelkopfGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DoppelkopfResponse } from '../types/card';
import { DoppelkopfPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DOPPELKOPF_HELP, parseDoppelkopfCommand } from '../utils/cli/commands/doppelkopfCommands';
import { formatDoppelkopfState } from '../utils/cli/formatters/doppelkopfFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { DOPPELKOPF_TRUMP_ORDER, isDoppelkopfTrump } from '../utils/doppelkopfTrump';
import { playerName } from '../utils/playerUtils';

/** Doppelkopf tutorial step definitions. */
const DK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dk-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="dk-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="dk-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const DOPPELKOPF_PHASE_KEYS: Readonly<Record<number, string>> = {
  [DoppelkopfPhase.PLAY]: 'play',
  [DoppelkopfPhase.TRICK_END]: 'trickEnd',
  [DoppelkopfPhase.ROUND_END]: 'roundEnd',
  [DoppelkopfPhase.GAME_END]: 'gameEnd',
};

/** Renders the Doppelkopf game page: 4-player partnership trick-taking with hidden Re/Kontra teams. */
export const DoppelkopfPage = withTutorial(DoppelkopfPageContent, 'doppelkopf', DK_TUTORIAL_STEPS);

/** Inner content of the Doppelkopf page, wrapped by TutorialProvider. */
function DoppelkopfPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('doppelkopf');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    doppelkopfConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleAnnounce,
    handleNextTrick,
    handleNextRound,
    handleHint,
    hintLoading,
  } = useDoppelkopfGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('doppelkopf');
  const doppelkopfCliConfig: CliGameConfig<DoppelkopfResponse, Parameters<typeof doppelkopfApi.exec>> = useMemo(
    () => ({
      gameName: 'doppelkopf',
      parseCommand: parseDoppelkopfCommand,
      formatResponse: formatDoppelkopfState,
      helpText: DOPPELKOPF_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, doppelkopfCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('doppelkopf', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('doppelkopf', DOPPELKOPF_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="doppelkopf" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 6 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;

  // Indices of the human's trump cards (♦ any, all Q/J, ♥10) for hand highlighting.
  const trumpIndices = humanPlayer
    ? humanPlayer.cards.reduce<number[]>((acc, card, idx) => {
        if (isDoppelkopfTrump(card)) acc.push(idx);
        return acc;
      }, [])
    : [];

  const isPlayPhase = state.phase === DoppelkopfPhase.PLAY;
  const isTrickEnd = state.phase === DoppelkopfPhase.TRICK_END;
  const isRoundEnd = state.phase === DoppelkopfPhase.ROUND_END;
  const isGameEnd = state.phase === DoppelkopfPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const teamLabel = state.youAreRe ? t('team.re') : t('team.kontra');
  // The human announces Re if on the Re team, otherwise Kontra.
  const announceLabel = state.youAreRe ? t('announceRe') : t('announceKontra');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.doppelkopf')}
      gameThemeBg={gameTheme.doppelkopf.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/doppelkopf"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
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
                    value: doppelkopfConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetChips',
                    label: t('settings.targetChips'),
                    value: doppelkopfConfig.targetChips,
                    options: TARGET_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetChips', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('trick', { n: state.trickNumber })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="dk-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="dk-info">
                {/* Trump ordering legend (collapsible) */}
                <details className="mb-2 p-2 rounded bg-black/30" data-testid="dk-trump-legend">
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                    {t('trumpLegend.title')}
                  </summary>
                  <div className="mt-1 text-ds-text-muted text-xs">
                    <div className="mb-1">{t('trumpLegend.caption')}</div>
                    <div className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
                      {DOPPELKOPF_TRUMP_ORDER.map((symbol, i) => (
                        <span key={symbol} className="inline-flex items-center gap-1">
                          <span className="font-mono text-ds-text-primary">{symbol}</span>
                          {i < DOPPELKOPF_TRUMP_ORDER.length - 1 && <span aria-hidden="true">&gt;</span>}
                        </span>
                      ))}
                    </div>
                  </div>
                </details>

                {/* Your team / announcements */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>
                    {t('yourTeam')}: {teamLabel}
                    {state.soloRe && ` (${t('soloRe')})`}
                  </div>
                  {(state.reAnnounced || state.kontraAnnounced) && (
                    <div>
                      {t('announced')}:{' '}
                      {[state.reAnnounced && t('team.re'), state.kontraAnnounced && t('team.kontra')]
                        .filter(Boolean)
                        .join(', ')}
                    </div>
                  )}
                </div>

                {/* Players: chips / cards / tricks (team shown once revealed) */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}
                          {state.teamsRevealed && ` [${p.isRe ? t('team.re') : t('team.kontra')}]`}:{' '}
                          {t('chips', { count: p.chips })} | {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}
                        {state.teamsRevealed && ` [${p.isRe ? t('team.re') : t('team.kontra')}]`}:{' '}
                        {t('chips', { count: p.chips })} | {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Round result */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.rePoints', { points: state.roundRePoints })}</div>
                    <div>{state.roundReWon ? t('roundResult.reWon') : t('roundResult.reLost')}</div>
                    <div>{t('roundResult.gamePoints', { points: state.roundGamePoints })}</div>
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
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
          <GameFooter className={`${gameTheme.doppelkopf.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="dk"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
                trumpIndices={trumpIndices}
                trumpTitle={t('trumpLegend.badgeTitle')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices.length > 0 && ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="dk-action-buttons">
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
              {canPlay && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {state.canAnnounce && (
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={handleAnnounce}
                  disabled={loading}
                  aria-label={`${announceLabel} — ${t('announceDescription')}`}
                  title={t('announceDescription')}
                >
                  {announceLabel}
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
                dataTutorial="dk-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
