import { useCallback, useEffect, useMemo, useState } from 'react';
import type { teenPattiApi } from '../api/gameApi';
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
import { HintTooltip } from '../components/hint/HintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  ANTE_OPTIONS,
  CPU_DIFFICULTY_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  useTeenPattiGame,
} from '../hooks/useTeenPattiGame';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TeenPattiResponse } from '../types/card';
import { TeenPattiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTeenPattiCommand, TEEN_PATTI_HELP } from '../utils/cli/commands/teenPattiCommands';
import { formatTeenPattiState } from '../utils/cli/formatters/teenPattiFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Teen Patti tutorial step definitions. */
const TEEN_PATTI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="teenpatti-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-action-buttons"]',
    messageKey: 'tutorial.bet',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-sideshow"]',
    messageKey: 'tutorial.sideShow',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="teenpatti-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TEEN_PATTI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TeenPattiPhase.BETTING]: 'betting',
  [TeenPattiPhase.SIDE_SHOW]: 'sideShow',
  [TeenPattiPhase.SHOWDOWN]: 'showdown',
  [TeenPattiPhase.ROUND_END]: 'roundEnd',
  [TeenPattiPhase.GAME_END]: 'gameEnd',
};

/** Renders the Teen Patti game page: a 4-player Indian chips/pot vying game with Side Show. */
export const TeenPattiPage = withTutorial(TeenPattiPageContent, 'teenpatti', TEEN_PATTI_TUTORIAL_STEPS);

/** Inner content of the Teen Patti page, wrapped by TutorialProvider. */
function TeenPattiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('teenpatti');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    teenPattiConfig,
    handleConfigChange,
    reset,
    handleSee,
    handleBet,
    handleRaise,
    handleFold,
    handleShow,
    handleSideShow,
    handleRespondSideShow,
    handleNextRound,
  } = useTeenPattiGame();

  // Raise stake amount (local UI state).
  const [raiseStake, setRaiseStake] = useState(2);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('teenpatti');
  const cliConfig: CliGameConfig<TeenPattiResponse, Parameters<typeof teenPattiApi.exec>> = useMemo(
    () => ({
      gameName: 'teenpatti',
      parseCommand: parseTeenPattiCommand,
      formatResponse: formatTeenPattiState,
      helpText: TEEN_PATTI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('teenpatti', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('teenpatti', TEEN_PATTI_PHASE_KEYS);

  // No card selection in Teen Patti — bets are placed, not cards. Provide a stable no-op.
  const noopToggle = useCallback(() => {}, []);

  if (!state)
    return <GameSkeleton gameKey="teenpatti" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 3 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isBettingPhase = state.phase === TeenPattiPhase.BETTING;
  const isSideShowPhase = state.phase === TeenPattiPhase.SIDE_SHOW;
  const isRoundEnd = state.phase === TeenPattiPhase.ROUND_END;
  const isGameEnd = state.phase === TeenPattiPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const isHumanBetTurn = isBettingPhase && isHumanTurn;
  // The human is being asked to accept/decline a pending Side Show.
  const isHumanSideShowTarget = isSideShowPhase && state.sideShowTarget === humanIdx;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.teenpatti')}
      gameThemeBg={gameTheme.teenpatti.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/teenpatti"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.matchWinnerIdx === humanIdx}
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
                    value: teenPattiConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: teenPattiConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: teenPattiConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
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
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="teenpatti-info">
              <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { amount: state.pot })}</span>
              <span>{t('stake', { amount: state.stake })}</span>
            </div>

            {isHumanBetTurn && (
              <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">{t('betNotice')}</div>
            )}

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="teenpatti-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => {
                const badge = p.out
                  ? t('badge.out')
                  : p.folded
                    ? t('badge.folded')
                    : p.seen
                      ? t('badge.seen')
                      : t('badge.blind');
                return (
                  <div
                    key={p.id}
                    className={`text-sm py-0.5 ${p.id === state.currentPlayerIdx ? 'text-ds-warning' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
                  >
                    {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
                    {t('roundBet', { amount: p.roundBet })} · [{badge}]{p.handName ? ` · ${handName(p.handName)}` : ''}
                  </div>
                );
              })}
            </div>

            {/* Pending Side Show request + response — unified, emphasized panel */}
            {isSideShowPhase && state.sideShowRequester >= 0 && state.sideShowTarget >= 0 && (
              <div
                className="mb-2 p-3 rounded bg-black/30 ring-2 ring-ds-warning motion-safe:animate-pulse"
                data-tutorial="teenpatti-sideshow"
                data-testid="teenpatti-sideshow-panel"
              >
                <div className="text-ds-warning text-sm font-semibold mb-1">{t('sideShowTitle')}</div>
                <div className="text-ds-text-primary text-sm">
                  {t('sideShowPending', {
                    requester: playerLabel(state.sideShowRequester, state.sideShowRequester === humanIdx),
                    target: playerLabel(state.sideShowTarget, state.sideShowTarget === humanIdx),
                  })}
                </div>
                {isHumanSideShowTarget && (
                  <div className="flex flex-wrap gap-2 mt-2">
                    <button
                      type="button"
                      className={btnSuccess}
                      onClick={() => handleRespondSideShow(true)}
                      disabled={loading}
                    >
                      {t('acceptButton')}
                    </button>
                    <button
                      type="button"
                      className={btnDanger}
                      onClick={() => handleRespondSideShow(false)}
                      disabled={loading}
                    >
                      {t('declineButton')}
                    </button>
                  </div>
                )}
              </div>
            )}

            {/* Revealed hands at showdown */}
            {state.isShowdown && (
              <div className="mb-2 p-2 rounded bg-black/30">
                {state.players
                  .filter((p) => !p.isHuman && p.cards.length > 0)
                  .map((p) => (
                    <div key={p.id} className="mb-1">
                      <div className="text-ds-text-muted text-xs mb-0.5">
                        {playerLabel(p.id, p.isHuman)}
                        {p.handName ? ` — ${handName(p.handName)}` : ''}
                      </div>
                      <div className="flex gap-1">
                        {p.cards.map((c, i) => (
                          <CardImage key={`${p.id}-${i}`} card={c} width={cardWidth} />
                        ))}
                      </div>
                    </div>
                  ))}
              </div>
            )}

            {/* Deal result */}
            {(isRoundEnd || isGameEnd) && state.roundWinnerIdx >= 0 && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div>
                  {t('roundResult.winner', {
                    name: playerLabel(state.roundWinnerIdx, state.roundWinnerIdx === humanIdx),
                    pot: state.pot,
                  })}
                </div>
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
          <GameFooter className={`${gameTheme.teenpatti.footer} px-4 py-2.5`}>
            {humanPlayer && (humanPlayer.seen || state.isShowdown) && humanPlayer.cards.length > 0 ? (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={[]}
                toggleCard={noopToggle}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="teenpatti"
              />
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="teenpatti-player-hand">
                {t('handLabel')} — {t('badge.blind')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="teenpatti-action-buttons">
              {isHumanBetTurn && (
                <>
                  {!humanPlayer?.seen && (
                    <button type="button" className={btnSecondary} onClick={handleSee} disabled={loading}>
                      {t('seeButton')}
                    </button>
                  )}
                  <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                    {t('betButton', { amount: state.stake })}
                  </button>
                  <div className="flex items-center gap-1">
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => setRaiseStake((a) => Math.max(state.stake + 1, a - 1))}
                      disabled={loading}
                      aria-label="-"
                    >
                      −
                    </button>
                    <span className="text-ds-text-primary text-sm min-w-[4rem] text-center">
                      {t('raisePrompt')} {raiseStake}
                    </span>
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => setRaiseStake((a) => a + 1)}
                      disabled={loading}
                      aria-label="+"
                    >
                      ＋
                    </button>
                    <button
                      type="button"
                      className={btnWarning}
                      onClick={() => handleRaise(raiseStake)}
                      disabled={loading}
                    >
                      {t('raiseButton', { amount: raiseStake })}
                    </button>
                  </div>
                  <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                    {t('foldButton')}
                  </button>
                  {state.canShow && (
                    <button type="button" className={btnPrimary} onClick={handleShow} disabled={loading}>
                      {t('showButton')}
                    </button>
                  )}
                  {state.canRequestSideShow && (
                    <button type="button" className={btnSecondary} onClick={handleSideShow} disabled={loading}>
                      {t('sideShowButton')}
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
                dataTutorial="teenpatti-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
