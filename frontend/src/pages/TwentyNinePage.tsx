import { useEffect, useMemo } from 'react';
import type { twentyNineApi } from '../api/gameApi';
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
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTwentyNineGame } from '../hooks/useTwentyNineGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TwentyNineResponse } from '../types/card';
import { TwentyNinePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTwentyNineCommand, TWENTY_NINE_HELP } from '../utils/cli/commands/twentyNineCommands';
import { formatTwentyNineState } from '../utils/cli/formatters/twentyNineFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = no trump). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Bid button options (Pass / 16 / 20 / 24 / 28). */
const BIDS: { value: number; key: string }[] = [
  { value: 0, key: 'bid.pass' },
  { value: 16, key: 'bid.sixteen' },
  { value: 20, key: 'bid.twenty' },
  { value: 24, key: 'bid.twentyfour' },
  { value: 28, key: 'bid.twentyeight' },
];

/** Twenty-Nine (29) tutorial step definitions. */
const TWENTY_NINE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="twentynine-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="twentynine-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="twentynine-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="twentynine-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TWENTY_NINE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TwentyNinePhase.BID]: 'bid',
  [TwentyNinePhase.PLAY]: 'play',
  [TwentyNinePhase.TRICK_END]: 'trickEnd',
  [TwentyNinePhase.ROUND_END]: 'roundEnd',
  [TwentyNinePhase.GAME_END]: 'gameEnd',
};

/** Renders the Twenty-Nine (29) game page: an Indian/Bangladeshi 4-player (2 vs 2) hidden-trump bidding trick-taker. */
export const TwentyNinePage = withTutorial(TwentyNinePageContent, 'twentynine', TWENTY_NINE_TUTORIAL_STEPS);

/** Inner content of the Twenty-Nine (29) page, wrapped by TutorialProvider. */
function TwentyNinePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('twentynine');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    twentyNineConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useTwentyNineGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('twentynine');
  const cliConfig: CliGameConfig<TwentyNineResponse, Parameters<typeof twentyNineApi.exec>> = useMemo(
    () => ({
      gameName: 'twentynine',
      parseCommand: parseTwentyNineCommand,
      formatResponse: formatTwentyNineState,
      helpText: TWENTY_NINE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('twentynine', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('twentynine', TWENTY_NINE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="twentynine" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanTeam = humanIdx % 2;
  const isHumanTurn = state.isHumanTurn;
  const isHumanBidTurn = state.isHumanBidTurn;

  const isBidPhase = state.phase === TwentyNinePhase.BID;
  const isPlayPhase = state.phase === TwentyNinePhase.PLAY;
  const isTrickEnd = state.phase === TwentyNinePhase.TRICK_END;
  const isRoundEnd = state.phase === TwentyNinePhase.ROUND_END;
  const isGameEnd = state.phase === TwentyNinePhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  // The trump suit is hidden until trumpRevealed flips true mid-play.
  const trumpSymbol = !state.trumpRevealed
    ? t('hiddenTrump')
    : state.trumpSuit === 0
      ? t('noTrump')
      : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');

  // The current highest (non-pass) bid; a new non-pass bid must beat it.
  const highestBid = Math.max(0, ...state.bids);

  const contractLabel = state.contract === 0 ? t('contractUndecided') : String(state.contract);

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.twentynine')}
      gameThemeBg={gameTheme.twentynine.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || isHumanBidTurn) && !isGameEnd}
      gamePath="/twentynine"
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
                    value: twentyNineConfig.cpuDifficulty,
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
                    value: twentyNineConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
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
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="twentynine-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            <div className="text-ds-text-muted text-center mb-2 text-sm">
              {state.declarerIdx >= 0
                ? t('contractLine', {
                    name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
                    team: state.declarerIdx % 2 === 0 ? t('team.a') : t('team.b'),
                    contract: contractLabel,
                  })
                : t('contractUndecided')}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="twentynine-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* Team match scores */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                </div>

                {/* Players grouped by team, with the declarer badge */}
                {([0, 1] as const).map((team) => (
                  <div key={team} className="mb-2 p-2 rounded bg-black/30">
                    <div className="text-ds-text-primary text-xs mb-1">{team === 0 ? t('team.a') : t('team.b')}</div>
                    {isMobile ? (
                      <details>
                        <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                          {t('players')}
                        </summary>
                        <div className="mt-1">{renderTeamPlayers(team)}</div>
                      </details>
                    ) : (
                      renderTeamPlayers(team)
                    )}
                  </div>
                ))}

                {/* Live round card points during bidding/play (matches the CUI's roundPoints
                    readout); the round-result block below takes over once the round ends. */}
                {!(isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="tn29-round-points"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundPointsTitle')}</div>
                    <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] })}</div>
                    <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] })}</div>
                  </div>
                )}

                {/* Round result: card points per team */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] ?? 0 })}</div>
                    <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] ?? 0 })}</div>
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
          <GameFooter className={`${gameTheme.twentynine.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="twentynine"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="twentynine-action-buttons">
              {isBidPhase && isHumanBidTurn && (
                <>
                  <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
                  {BIDS.map((b) => {
                    // Pass (0) is always allowed; a non-pass bid must beat the current highest.
                    const tooLow = b.value !== 0 && b.value <= highestBid;
                    const disabled = loading || tooLow;
                    const reason = tooLow ? t('bidDisabledReason', { currentBid: highestBid }) : undefined;
                    // The title lives on the wrapping span: browsers suppress native tooltips on
                    // disabled buttons, so hovering the span still surfaces the reason.
                    return (
                      <span key={b.value} title={reason} data-testid={`bid-wrap-${b.value}`}>
                        <button
                          type="button"
                          className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                          onClick={() => handleBid(b.value)}
                          disabled={disabled}
                          aria-disabled={disabled}
                          aria-label={reason ? `${t(b.key)} — ${reason}` : undefined}
                          data-testid={`bid-${b.value}`}
                        >
                          {t(b.key)}
                        </button>
                      </span>
                    );
                  })}
                </>
              )}
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
                dataTutorial="twentynine-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );

  /** Renders the per-player rows for one team (seats with seat%2 === team). */
  function renderTeamPlayers(team: number) {
    if (!state) return null;
    return state.players
      .filter((p) => p.id % 2 === team)
      .map((p) => (
        <div key={p.id} className="py-0.5 flex items-center gap-2 text-ds-text-muted text-sm">
          <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
          </span>
          {p.isDeclarer && (
            <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">{t('declarerBadge')}</span>
          )}
        </div>
      ));
  }
}
