import { useEffect, useMemo, useRef, useState } from 'react';
import type { klaverjasApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useKlaverjasGame } from '../hooks/useKlaverjasGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KlaverjasResponse } from '../types/card';
import { KlaverjasPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KLAVERJAS_HELP, parseKlaverjasCommand } from '../utils/cli/commands/klaverjasCommands';
import { formatKlaverjasState } from '../utils/cli/formatters/klaverjasFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/**
 * Card strength and points, strongest first (sync: `Klaverjas.trumpStrength` /
 * `klaverjasPlainStrength` / `cardPoints`).
 *
 * **切り札と非切り札で順序が完全に違う。**切り札は J > 9 > A > 10 …、非切り札は
 * A > 10 > K …。姉妹ゲームの Manille には早見表があるのに、より覚えにくいこちら
 * には無かった (#4757)。
 */
const KLAVERJAS_TRUMP_ROWS: ReadonlyArray<{ face: string; points: number; nameKey?: string }> = [
  { face: 'J', points: 20, nameKey: 'strengthLegend.jas' },
  { face: '9', points: 14, nameKey: 'strengthLegend.nel' },
  { face: 'A', points: 11 },
  { face: '10', points: 10 },
  { face: 'K', points: 4 },
  { face: 'Q', points: 3 },
  { face: '8', points: 0 },
  { face: '7', points: 0 },
];

/** Non-trump strength and points, strongest first. */
const KLAVERJAS_PLAIN_ROWS: ReadonlyArray<{ face: string; points: number }> = [
  { face: 'A', points: 11 },
  { face: '10', points: 10 },
  { face: 'K', points: 4 },
  { face: 'Q', points: 3 },
  { face: 'J', points: 2 },
  { face: '9', points: 0 },
  { face: '8', points: 0 },
  { face: '7', points: 0 },
];

/** Klaverjas tutorial step definitions. */
const KLAVERJAS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="klaverjas-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaverjas-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaverjas-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="klaverjas-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="klaverjas-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KLAVERJAS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KlaverjasPhase.PLAY]: 'play',
  [KlaverjasPhase.TRICK_END]: 'trickEnd',
  [KlaverjasPhase.ROUND_END]: 'roundEnd',
  [KlaverjasPhase.GAME_END]: 'gameEnd',
};

/** Renders the Klaverjas game page: a Dutch 4-player (2 vs 2) trump trick-taker with Roem bonuses. */
export const KlaverjasPage = withTutorial(KlaverjasPageContent, 'klaverjas', KLAVERJAS_TUTORIAL_STEPS);

/** Inner content of the Klaverjas page, wrapped by TutorialProvider. */
function KlaverjasPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('klaverjas');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    klaverjasConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useKlaverjasGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('klaverjas');
  const klaverjasCliConfig: CliGameConfig<KlaverjasResponse, Parameters<typeof klaverjasApi.exec>> = useMemo(
    () => ({
      gameName: 'klaverjas',
      parseCommand: parseKlaverjasCommand,
      formatResponse: formatKlaverjasState,
      helpText: KLAVERJAS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, klaverjasCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('klaverjas', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('klaverjas', KLAVERJAS_PHASE_KEYS);

  // Transient pulse whenever the combined Roem (bonus) total grows during a hand —
  // a scoring-relevant event that is otherwise easy to miss both visually and for SR users.
  const [roemPulse, setRoemPulse] = useState(false);
  const prevRoemRef = useRef<number | null>(null);
  useEffect(() => {
    const roemTotal = state?.roundRoem ? (state.roundRoem[0] ?? 0) + (state.roundRoem[1] ?? 0) : null;
    if (roemTotal == null) return;
    const prev = prevRoemRef.current;
    prevRoemRef.current = roemTotal;
    if (prev != null && roemTotal > prev) {
      setRoemPulse(true);
      const id = setTimeout(() => setRoemPulse(false), 2500);
      return () => clearTimeout(id);
    }
  }, [state?.roundRoem]);

  if (!state)
    return <GameSkeleton gameKey="klaverjas" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === KlaverjasPhase.PLAY;
  const isTrickEnd = state.phase === KlaverjasPhase.TRICK_END;
  const isRoundEnd = state.phase === KlaverjasPhase.ROUND_END;
  const isGameEnd = state.phase === KlaverjasPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const humanTeam = humanIdx % 2;
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '?';

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.klaverjas')}
      gameThemeBg={gameTheme.klaverjas.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/klaverjas"
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
                    value: klaverjasConfig.cpuDifficulty,
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
                    value: klaverjasConfig.targetPoints,
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
                  dataTutorial="klaverjas-trick-display"
                />
                {/* At TrickEnd leadPlayerIdx is the trick winner, so their team is shown before Next Trick. */}
                {isTrickEnd && (
                  <div
                    className="my-2 p-2 rounded bg-ds-accent/15 text-center text-sm font-semibold text-ds-accent"
                    role="status"
                    aria-live="polite"
                    data-testid="klaverjas-trick-winner"
                  >
                    {t('trickWinner', { team: state.leadPlayerIdx % 2 === 0 ? t('team.a') : t('team.b') })}
                  </div>
                )}
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="klaverjas-info">
                {/* Team match scores */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                </div>

                {/* Live Roem (bonus) per team, shown throughout the hand to match the CUI's
                    Roem readout; the round-result block below repeats it once the round ends. */}
                {!(isRoundEnd || isGameEnd) && (
                  <div
                    className={`mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm${
                      roemPulse ? ' motion-safe:animate-pulse ring-1 ring-ds-warning' : ''
                    }`}
                    data-testid="klaverjas-roem"
                    role="status"
                    aria-live="polite"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roem.title')}</div>
                    <div>
                      {t('roundResult.roem', { roemA: state.roundRoem[0] ?? 0, roemB: state.roundRoem[1] ?? 0 })}
                    </div>
                  </div>
                )}

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

                {/* **切り札と非切り札で強さの順序が完全に違う (#4757)。**切り札は
                    J > 9 > A > 10、非切り札は A > 10 > K。姉妹ゲームの Manille には
                    早見表があるのに、より覚えにくいこちらには無かった。 */}
                <details className="mb-2 p-2 rounded bg-black/30" data-testid="klaverjas-strength-legend">
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                    {t('strengthLegend.title')}
                  </summary>
                  <div className="mt-1 text-ds-text-muted text-xs">
                    {(
                      [
                        ['trumpCol', KLAVERJAS_TRUMP_ROWS],
                        ['plainCol', KLAVERJAS_PLAIN_ROWS],
                      ] as const
                    ).map(([headKey, rows]) => (
                      <table className="w-full mb-1" key={headKey}>
                        <caption className="text-left">{t(`strengthLegend.${headKey}`)}</caption>
                        <thead>
                          <tr>
                            <th scope="col" className="text-left font-normal">
                              {t('strengthLegend.rankCol')}
                            </th>
                            <th scope="col" className="text-left font-normal">
                              {t('strengthLegend.cardCol')}
                            </th>
                            <th scope="col" className="text-right font-normal">
                              {t('strengthLegend.pointCol')}
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {rows.map((row, i) => (
                            <tr key={row.face}>
                              <td>{i + 1}</td>
                              <td>{'nameKey' in row && row.nameKey ? `${row.face} (${t(row.nameKey)})` : row.face}</td>
                              <td className="text-right">{row.points}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ))}
                    <div className="mt-1">{t('strengthLegend.note')}</div>
                  </div>
                </details>

                {/* Round result */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] ?? 0 })}</div>
                    <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] ?? 0 })}</div>
                    <div className="mt-1">
                      {t('roundResult.roem', { roemA: state.roundRoem[0] ?? 0, roemB: state.roundRoem[1] ?? 0 })}
                    </div>
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
          <GameFooter className={`${gameTheme.klaverjas.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="klaverjas"
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
            <div data-testid="klaverjas-hint-live" role="status" aria-live="polite">
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="klaverjas-action-buttons">
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
                dataTutorial="klaverjas-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
