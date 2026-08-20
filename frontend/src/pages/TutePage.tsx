import { useEffect, useMemo } from 'react';
import type { tuteApi } from '../api/gameApi';
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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTuteGame } from '../hooks/useTuteGame';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TuteResponse } from '../types/card';
import { TutePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTuteCommand, TUTE_HELP } from '../utils/cli/commands/tuteCommands';
import { formatTuteState } from '../utils/cli/formatters/tuteFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;
/** Suit id → `suitName.*` i18n key (1=♠ .. 4=♦). */
const SUIT_KEYS = ['', 'spade', 'club', 'heart', 'diamond'] as const;

/** Tute tutorial step definitions. */
const TUTE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tute-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tute-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tute-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="tute-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="tute-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TUTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TutePhase.PLAY]: 'play',
  [TutePhase.TRICK_END]: 'trickEnd',
  [TutePhase.ROUND_END]: 'roundEnd',
  [TutePhase.GAME_END]: 'gameEnd',
};

/** Renders the Tute game page: a Spanish 4-player (2 vs 2) trump trick-taker with marriage and Tute declarations. */
export const TutePage = withTutorial(TutePageContent, 'tute', TUTE_TUTORIAL_STEPS);

/** Inner content of the Tute page, wrapped by TutorialProvider. */
function TutePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tute');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    tuteConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleDeclareMarriage,
    handleDeclareTute,
    handleNextTrick,
    handleNextRound,
  } = useTuteGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tute');
  const tuteCliConfig: CliGameConfig<TuteResponse, Parameters<typeof tuteApi.exec>> = useMemo(
    () => ({
      gameName: 'tute',
      parseCommand: parseTuteCommand,
      formatResponse: formatTuteState,
      helpText: TUTE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, tuteCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tute', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tute', TUTE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="tute" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 6 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;

  const isPlayPhase = state.phase === TutePhase.PLAY;
  const isTrickEnd = state.phase === TutePhase.TRICK_END;
  const isRoundEnd = state.phase === TutePhase.ROUND_END;
  const isGameEnd = state.phase === TutePhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  // Rendered both during play (so a marriage is seen to score) and at round end.
  const teamPointLines = (
    <>
      <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] })}</div>
      <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] })}</div>
    </>
  );
  const humanTeam = humanIdx % 2;
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '?';

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.tute')}
      gameThemeBg={gameTheme.tute.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/tute"
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
                    value: tuteConfig.cpuDifficulty,
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
                    value: tuteConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{t('trump', { suit: trumpSymbol })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="tute-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="tute-info">
                {/* Team scores */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                </div>

                {/* Declared marriage suits (persistent readout of state.declaredSuits) */}
                <div
                  className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                  data-testid="tute-declared-marriages"
                >
                  <div className="mb-1 text-ds-text-primary">{t('declaredMarriages.title')}</div>
                  {([1, 2, 3, 4] as const).map((suit) => {
                    const declared = state.declaredSuits[suit] ?? false;
                    return (
                      <div key={suit} className="py-0.5">
                        <span className="mr-1">{SUIT_SYMBOLS[suit]}</span>
                        {suit === state.trumpSuit && <span className="mr-1 text-ds-warning">★</span>}
                        <span className={declared ? 'text-ds-text-primary' : ''}>
                          {declared ? t('declaredMarriages.declared') : t('declaredMarriages.undeclared')}
                        </span>
                      </div>
                    );
                  })}
                  <div className="mt-1 text-xs">{t('declaredMarriages.trumpNote')}</div>
                </div>

                {/* Players: cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* A marriage scores the moment it is declared, and the button even
                    announces how much — but the total only appeared at round end,
                    so the player could not see it land (#4722). */}
                {(isPlayPhase || isTrickEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="tute-running-points"
                    role="status"
                    aria-live="polite"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.runningTitle')}</div>
                    {teamPointLines}
                  </div>
                )}

                {/* Round result */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {teamPointLines}
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
          <GameFooter className={`${gameTheme.tute.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="tute"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="tute-hint-live" role="status" aria-live="polite">
              {state.hint && isRequestedHint(state) && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                  {state.hint.cardIndices &&
                    state.hint.cardIndices.length > 0 &&
                    ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="tute-action-buttons">
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
              {canPlay && (state.canDeclareMarriage || state.canDeclareTute) && (
                <fieldset
                  className="flex flex-wrap items-center gap-2 border-0 p-0 m-0"
                  data-testid="tute-declarations"
                >
                  <legend className="text-xs text-ds-text-muted mr-1">{t('declarationsLabel')}</legend>
                  {state.canDeclareMarriage &&
                    humanPlayer &&
                    ([1, 2, 3, 4] as const)
                      .filter((suit) => {
                        // Only offer suits the human actually holds K+Q of and has
                        // not already declared, so every button is a legal move.
                        const design = (['', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const)[suit];
                        const hasK = humanPlayer.cards.some((c) => c.design === design && c.value === 13);
                        const hasQ = humanPlayer.cards.some((c) => c.design === design && c.value === 12);
                        return hasK && hasQ && !state.declaredSuits[suit];
                      })
                      .map((suit) => (
                        <button
                          key={suit}
                          type="button"
                          className={btnSecondary}
                          onClick={() => handleDeclareMarriage(suit)}
                          disabled={loading}
                          aria-label={t('marriageAria', {
                            suit: t(`suitName.${SUIT_KEYS[suit]}`),
                            points: suit === state.trumpSuit ? 40 : 20,
                          })}
                        >
                          {t('declareMarriage', { suit: SUIT_SYMBOLS[suit] })}
                        </button>
                      ))}
                  {state.canDeclareTute && (
                    <>
                      <button
                        type="button"
                        className={btnSecondary}
                        onClick={handleDeclareTute}
                        disabled={loading}
                        title={t('tuteHelp')}
                        aria-describedby="tute-help-desc"
                        data-testid="tute-declare-button"
                      >
                        {t('declareTute')}
                      </button>
                      <span id="tute-help-desc" className="sr-only">
                        {t('tuteHelp')}
                      </span>
                    </>
                  )}
                </fieldset>
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
                dataTutorial="tute-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
