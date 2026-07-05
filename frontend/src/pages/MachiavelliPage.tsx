import { useCallback, useMemo } from 'react';
import type { machiavelliApi } from '../api/gameApi';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import {
  CPU_DIFFICULTY_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useMachiavelliGame,
} from '../hooks/useMachiavelliGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MachiavelliResponse } from '../types/card';
import { MachiavelliPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MACHIAVELLI_HELP, parseMachiavelliCommand } from '../utils/cli/commands/machiavelliCommands';
import { formatMachiavelliState } from '../utils/cli/formatters/machiavelliFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

const MACHIAVELLI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MachiavelliPhase.TURN]: 'turn',
  [MachiavelliPhase.ROUND_END]: 'roundEnd',
  [MachiavelliPhase.GAME_END]: 'gameEnd',
};

/** Machiavelli tutorial step definitions. */
const MV_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="mv-table"]', messageKey: 'tutorial.table', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="mv-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-newmeld-button"]',
    messageKey: 'tutorial.newMeldButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Machiavelli game page with a shared table, melding, and layoff. */
export const MachiavelliPage = withTutorial(MachiavelliPageContent, 'machiavelli', MV_TUTORIAL_STEPS);
/** Inner content of the Machiavelli page, wrapped by TutorialProvider. */
function MachiavelliPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('machiavelli');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    machiavelliConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDraw,
    handleNewMeld,
    handleLayoff,
    handleNextRound,
  } = useMachiavelliGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('machiavelli', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('machiavelli');
  const cliConfig: CliGameConfig<MachiavelliResponse, Parameters<typeof machiavelliApi.exec>> = useMemo(
    () => ({
      gameName: 'machiavelli',
      parseCommand: parseMachiavelliCommand,
      formatResponse: formatMachiavelliState,
      helpText: MACHIAVELLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isTurnPhaseForKbd = state?.phase === MachiavelliPhase.TURN;
  const isHumanTurnForKbd = isTurnPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (selectedCardIndices.length >= 3) handleNewMeld();
  }, [selectedCardIndices, handleNewMeld]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('machiavelli', MACHIAVELLI_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      playerCount: machiavelliConfig.playerCount,
      cpuDifficulty: machiavelliConfig.cpuDifficulty,
      targetRounds: machiavelliConfig.targetRounds,
    });
  }, [
    gameExec,
    hideActionLog,
    machiavelliConfig.playerCount,
    machiavelliConfig.cpuDifficulty,
    machiavelliConfig.targetRounds,
  ]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="machiavelli"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 13 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isTurnPhase = state.phase === MachiavelliPhase.TURN;
  const isRoundEnd = state.phase === MachiavelliPhase.ROUND_END;
  const isGameEnd = state.phase === MachiavelliPhase.GAME_END || state.gameEndFlag;
  const revealCpu = isRoundEnd || isGameEnd;
  const isHumanTurn = isTurnPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.machiavelli')}
      gameThemeBg={gameTheme.machiavelli.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/machiavelli"
      gameEndFlag={isGameEnd}
      onCelebrate={() => playSound('winFanfare')}
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
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: machiavelliConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: machiavelliConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: machiavelliConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
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
              <span className="mr-4">{t('round', { n: state.roundNumber, total: state.targetRounds })}</span>
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: shared table melds */}
              <div>
                <div className="my-3 p-3 rounded bg-black/40" data-tutorial="mv-table" data-testid="machiavelli-table">
                  <div className="text-ds-text-muted text-sm mb-2">{t('table')}</div>
                  {state.table.length === 0 ? (
                    <div className="text-ds-text-muted text-sm">{t('tableEmpty')}</div>
                  ) : (
                    <div className="flex flex-col gap-2">
                      {state.table.map((meld, meldIdx) => (
                        <div
                          key={`meld-${meldIdx}-${meld.kind}-${meld.cards.map((c) => `${c.design}${c.value}`).join('')}`}
                          className="flex items-center gap-2 flex-wrap"
                        >
                          <span className="text-ds-text-muted text-xs w-14">
                            {meld.kind === 0 ? t('meldKindSet') : t('meldKindRun')}
                          </span>
                          <div className="flex flex-wrap gap-1">
                            {meld.cards.map((card, idx) => (
                              <AnimatedCard
                                key={`meld-${meldIdx}-card-${card.design}-${card.value}-${idx}`}
                                card={card}
                                width={cardWidth * 0.8}
                              />
                            ))}
                          </div>
                          {isHumanTurn && (
                            <button
                              type="button"
                              className={btnPrimary}
                              onClick={() => handleLayoff(meldIdx)}
                              disabled={loading || selectedCardIndices.length !== 1}
                            >
                              {t('layoffButton')}
                            </button>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })}
                        {revealCpu && <> | {t('deadwoodShort', { score: p.deadwood })}</>}
                      </div>
                      {revealCpu && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${p.id}-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="mv-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
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

          <GameFooter className={`${gameTheme.machiavelli.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="mv-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
                      ...selectedCardStyle(selectedCardIndices.includes(idx)),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={loading}
                    data-tutorial="mv-draw-button"
                  >
                    {t('drawButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleNewMeld}
                    disabled={loading || selectedCardIndices.length < 3}
                    data-tutorial="mv-newmeld-button"
                    data-testid="machiavelli-newmeld-button"
                  >
                    {t('newMeldButton')}
                  </button>
                </>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="mv-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
