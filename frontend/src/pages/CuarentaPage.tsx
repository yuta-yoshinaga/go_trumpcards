import { useEffect, useMemo, useRef, useState } from 'react';
import { cuarentaApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CuarentaAction, CuarentaResponse } from '../types/card';
import { CuarentaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CUARENTA_HELP, parseCuarentaCommand } from '../utils/cli/commands/cuarentaCommands';
import { formatCuarentaState } from '../utils/cli/formatters/cuarentaFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** CPU difficulty options for the Cuarenta settings panel. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
];

/** Cuarenta tutorial step definitions. */
const CUARENTA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cuarenta-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuarenta-teams"]',
    messageKey: 'tutorial.teams',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuarenta-table"]',
    messageKey: 'tutorial.table',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuarenta-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cuarenta-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric Cuarenta phases to i18n phase-label keys. */
const CUARENTA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CuarentaPhase.PLAY]: 'play',
  [CuarentaPhase.ROUND_END]: 'roundEnd',
  [CuarentaPhase.GAME_END]: 'gameEnd',
};

/** Renders the Cuarenta game page: a 4-player, 2-team Ecuadorian capture game. */
export const CuarentaPage = withTutorial(CuarentaPageContent, 'cuarenta', CUARENTA_TUTORIAL_STEPS);

/** Inner content of the Cuarenta page, wrapped by TutorialProvider. */
function CuarentaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cuarenta');
  const { state, loading, error, exec, retry } = useGameApi(cuarentaApi.exec);

  const [cpuDifficulty, setCpuDifficulty] = useState(1);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleDifficultyChange = (value: string) => {
    const level = Number(value);
    setCpuDifficulty(level);
    exec('reset', { config: { cpuDifficulty: level } });
  };

  const phaseNames = usePhaseNames('cuarenta', CUARENTA_PHASE_KEYS);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cuarenta');
  const cliConfig: CliGameConfig<CuarentaResponse, Parameters<typeof cuarentaApi.exec>> = useMemo(
    () => ({
      gameName: 'cuarenta',
      parseCommand: parseCuarentaCommand,
      formatResponse: formatCuarentaState,
      helpText: CUARENTA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  // Pop + chime the human's caída / ronda / limpia bonus the instant it lands —
  // these Cuarenta-specific scores were buried as static text in the log (#2693).
  const [bonusCelebrationKey, setBonusCelebrationKey] = useState(0);
  const prevHumanActionRef = useRef<string | null>(null);
  const humanAction = state?.humanAction ?? null;
  const humanBonus = !!humanAction && (humanAction.isCaida || humanAction.rondaBonus > 0 || humanAction.isLimpia);
  const humanActionSig = humanAction
    ? `${humanAction.playedCard?.design ?? ''}${humanAction.playedCard?.value ?? ''}-${humanAction.isCaida}-${humanAction.rondaBonus}-${humanAction.isLimpia}`
    : '';
  useEffect(() => {
    if (prevHumanActionRef.current === null) {
      prevHumanActionRef.current = humanActionSig;
      return;
    }
    if (humanActionSig !== prevHumanActionRef.current) {
      prevHumanActionRef.current = humanActionSig;
      if (humanBonus) {
        setBonusCelebrationKey((k) => k + 1);
        playSound('chipClick', { pitchVariation: 0.1 });
      }
    }
  }, [humanActionSig, humanBonus, playSound]);

  if (!state)
    return <GameSkeleton gameKey="cuarenta" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const isGameEnd = state.phase === CuarentaPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.phase === CuarentaPhase.PLAY && state.currentTurn === 0 && !isGameEnd;
  const phaseName = phaseNames[state.phase] ?? '';

  const humanPlayer = state.players.find((p) => p.isHuman);
  // Team A = seats {0,2}; the human (seat 0) is on Team A.
  const humanTeam = humanPlayer?.team ?? 0;
  const winningTeam = state.roundWinners.length === 1 ? state.roundWinners[0] : -1;
  const humanWon = isGameEnd && state.roundWinners.includes(humanTeam);

  const teamLabel = (team: number): string => t('team', { name: team === 0 ? 'A' : 'B' });
  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  /** Renders a short capture-result line for a play action, with bonus badges.
   * `celebrate` pops the bonus badges (used for the human's freshest bonus play). */
  const renderAction = (a: CuarentaAction, celebrate = false) => {
    const badges: string[] = [];
    if (a.isCaida) badges.push(t('caida'));
    if (a.rondaBonus > 0) badges.push(t('ronda', { bonus: a.rondaBonus }));
    if (a.isLimpia) badges.push(t('limpia'));
    const captured = a.capturedCards.length;
    return (
      <div
        key={`act-${a.playerIdx}`}
        className="text-sm text-ds-text-muted flex items-center gap-2 flex-wrap py-0.5"
        // On the human's freshest bonus play, announce the row so the caída/ronda/
        // limpia badges reach SR users (they were visual/audio-only before).
        role={celebrate && badges.length > 0 ? 'status' : undefined}
        aria-live={celebrate && badges.length > 0 ? 'polite' : undefined}
        data-testid={celebrate && badges.length > 0 ? 'cuarenta-bonus-announce' : undefined}
      >
        <span className="font-semibold text-ds-text-primary">{playerLabel(a.playerIdx, a.playerIdx === 0)}</span>
        {a.playedCard && <CardImage card={a.playedCard} width={Math.round(cardWidth * 0.6)} />}
        <span>{captured > 0 ? t('captured', { count: captured }) : t('laidDown')}</span>
        {badges.map((b) => (
          <span
            key={b}
            className={`text-ds-accent font-semibold ${celebrate ? 'motion-safe:animate-bounce' : ''}`}
            data-testid={celebrate ? 'cuarenta-bonus-pop' : undefined}
          >
            {b}
          </span>
        ))}
      </div>
    );
  };

  const handleManualReset = () => {
    hideActionLog();
    exec('reset', { config: { cpuDifficulty } });
  };

  return (
    <GamePageShell
      title={tc('nav.cuarenta')}
      gameThemeBg={gameTheme.cuarenta.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/cuarenta"
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label}`),
                    })),
                    onSelect: handleDifficultyChange,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="cuarenta-info">
              <span>{t('deck', { count: state.remainingDeck })}</span>
            </div>

            {/* Team scores */}
            <div className="mb-2 p-2 rounded bg-black/30 flex justify-center gap-6" data-tutorial="cuarenta-teams">
              {state.teamScores.map((score, team) => (
                <span
                  key={`team-${team}`}
                  className={`text-sm font-semibold ${team === humanTeam ? 'text-ds-warning' : 'text-ds-text-primary'}`}
                >
                  {t('teamScore', { name: teamLabel(team), score, target: state.config.targetScore })}
                </span>
              ))}
            </div>

            {/* Players (team / hand / captured / current turn) */}
            <div className="mb-2 p-2 rounded bg-black/20">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  className={`text-sm py-0.5 flex items-center gap-3 ${
                    p.id === state.currentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>{teamLabel(p.team)}</span>
                  <span>{t('captured', { count: p.capturedCount })}</span>
                </div>
              ))}
            </div>

            {/* Central table */}
            <div className="mb-2 p-3 rounded bg-black/20 text-center" data-tutorial="cuarenta-table">
              <div className="text-ds-text-muted text-xs mb-1">
                {t('table')} — {t('tableCount', { count: state.tableCards.length })}
              </div>
              {state.tableCards.length > 0 ? (
                <div className="flex flex-wrap justify-center gap-2">
                  {state.tableCards.map((c, i) => (
                    <CardImage key={`table-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              ) : (
                <div className="text-ds-text-muted text-sm">{t('tableEmpty')}</div>
              )}
            </div>

            {/* Last plays + capture results */}
            {(state.humanAction || state.cpuActions.length > 0) && (
              <div className="mb-2 p-2 rounded bg-black/20">
                <div className="mb-1 text-ds-text-muted text-xs">{t('lastPlays')}</div>
                {state.humanAction && (
                  <div key={bonusCelebrationKey}>{renderAction(state.humanAction, humanBonus)}</div>
                )}
                {state.cpuActions.map((a) => renderAction(a))}
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
          <GameFooter className={`${gameTheme.cuarenta.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="cuarenta-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-2">
                {humanPlayer?.cards.map((c, i) => (
                  <button
                    key={`hand-${i}`}
                    type="button"
                    onClick={() => isHumanTurn && exec('play', { handIndex: i })}
                    disabled={!isHumanTurn || loading}
                    className={`rounded transition-all ${
                      isHumanTurn ? 'cursor-pointer hover:opacity-90 hover:-translate-y-1' : 'cursor-default'
                    }`}
                    data-testid={`hand-card-${i}`}
                    aria-label={t('playCardAria', { card: cardAlt(c) })}
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanTurn && <div className="text-ds-text-muted text-xs mb-2">{t('turnNotice')}</div>}

            <div className="flex flex-wrap gap-2 items-center">
              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose', { name: winningTeam === 0 ? 'A' : 'B' })}
                </span>
              )}
              {state.phase === CuarentaPhase.ROUND_END && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cuarenta-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
