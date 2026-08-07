import { useCallback, useEffect, useMemo, useState } from 'react';
import type { threeCardBragApi } from '../api/gameApi';
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
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
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
  useThreeCardBragGame,
} from '../hooks/useThreeCardBragGame';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ThreeCardBragResponse } from '../types/card';
import { ThreeCardBragPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseThreeCardBragCommand, THREE_CARD_BRAG_HELP } from '../utils/cli/commands/threeCardBragCommands';
import { formatThreeCardBragState } from '../utils/cli/formatters/threeCardBragFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';
import {
  clampThreeCardBragRaise,
  threeCardBragActualCost,
  threeCardBragRaiseBounds,
} from '../utils/threeCardBragRaise';

/** Three Card Brag tutorial step definitions. */
const THREE_CARD_BRAG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="threecardbrag-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="threecardbrag-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="threecardbrag-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="threecardbrag-action-buttons"]',
    messageKey: 'tutorial.bet',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="threecardbrag-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const THREE_CARD_BRAG_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ThreeCardBragPhase.BETTING]: 'betting',
  [ThreeCardBragPhase.SHOWDOWN]: 'showdown',
  [ThreeCardBragPhase.ROUND_END]: 'roundEnd',
  [ThreeCardBragPhase.GAME_END]: 'gameEnd',
};

/** Renders the Three Card Brag game page: a 4-player British chips/pot vying game. */
export const ThreeCardBragPage = withTutorial(
  ThreeCardBragPageContent,
  'threecardbrag',
  THREE_CARD_BRAG_TUTORIAL_STEPS,
);

/** Inner content of the Three Card Brag page, wrapped by TutorialProvider. */
function ThreeCardBragPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('threecardbrag');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    threeCardBragConfig,
    handleConfigChange,
    reset,
    handleSee,
    handleBet,
    handleRaise,
    handleFold,
    handleShow,
    handleNextRound,
  } = useThreeCardBragGame();

  // Raise stake amount (local UI state).
  const [raiseStake, setRaiseStake] = useState(2);

  // Raise bounds mirror the CUI (ThreeCardBragCuiPresenter.threeCardBragRaiseRangeStr):
  // min = stake + 1; max = affordable ceiling (Seen players pay double, halving it).
  const humanForBounds = state?.players.find((p) => p.isHuman);
  const {
    min: raiseMin,
    max: raiseMax,
    canRaise,
  } = threeCardBragRaiseBounds(state?.stake ?? 0, humanForBounds?.chips ?? 0, humanForBounds?.seen ?? false);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // Keep the raise amount synced to the current stake/chip bounds: when the
  // stake rises (or chips/seen change) re-clamp so it never sits below min or
  // above the affordable max.
  useEffect(() => {
    setRaiseStake((a) => clampThreeCardBragRaise(a, raiseMin, raiseMax));
  }, [raiseMin, raiseMax]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('threecardbrag');
  const cliConfig: CliGameConfig<ThreeCardBragResponse, Parameters<typeof threeCardBragApi.exec>> = useMemo(
    () => ({
      gameName: 'threecardbrag',
      parseCommand: parseThreeCardBragCommand,
      formatResponse: formatThreeCardBragState,
      helpText: THREE_CARD_BRAG_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('threecardbrag', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('threecardbrag', THREE_CARD_BRAG_PHASE_KEYS);

  // No card selection in Brag — bets are placed, not cards. Provide a stable no-op.
  const noopToggle = useCallback(() => {}, []);

  if (!state)
    return (
      <GameSkeleton gameKey="threecardbrag" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 3 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isBettingPhase = state.phase === ThreeCardBragPhase.BETTING;
  const isRoundEnd = state.phase === ThreeCardBragPhase.ROUND_END;
  const isGameEnd = state.phase === ThreeCardBragPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const isHumanBetTurn = isBettingPhase && isHumanTurn;

  // Actual chips the human pays: a Seen player pays double the nominal stake
  // (matches the domain's callCost rule in ThreeCardBrag.go). Surface the real
  // cost on the Bet/Raise buttons so a Seen player isn't surprised by the 2x.
  const humanSeen = humanPlayer?.seen ?? false;
  const betCost = threeCardBragActualCost(state.stake, humanSeen);
  const raiseCost = threeCardBragActualCost(raiseStake, humanSeen);

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.threecardbrag')}
      gameThemeBg={gameTheme.threecardbrag.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/threecardbrag"
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
                    value: threeCardBragConfig.cpuDifficulty,
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
                    value: threeCardBragConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: threeCardBragConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="threecardbrag-info">
              <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { amount: state.pot })}</span>
              <span>{t('stake', { amount: state.stake })}</span>
            </div>

            {isHumanBetTurn && (
              <>
                <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">{t('betNotice')}</div>
                <div className="text-ds-text-muted text-center mb-2 text-xs" data-testid="tcb-cost-notice">
                  {humanSeen ? t('costNotice.seen', { cost: betCost }) : t('costNotice.blind', { cost: betCost })}
                </div>
              </>
            )}

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="threecardbrag-players">
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
          <GameFooter className={`${gameTheme.threecardbrag.footer} px-4 py-2.5`}>
            {humanPlayer && (humanPlayer.seen || state.isShowdown) && humanPlayer.cards.length > 0 ? (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={[]}
                toggleCard={noopToggle}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="threecardbrag"
              />
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="threecardbrag-player-hand">
                {t('handLabel')} — {t('badge.blind')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* **バックエンドは Output() で常に hint を詰めているのに、ページが
                一度も読んでいなかった (#4728)。**設定にヒントのチェックボックスは
                あるのに、サーバーの提案自体は画面に出ていなかった。他の全ゲームと
                同じく、明示的に要求されたときだけ出す。 */}
            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2" data-testid="tcb-server-hint">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`, { defaultValue: state.hint.reason })}
                {/* **識別子をそのまま出さない。**訳が無ければ識別子に落とす
                    (キー文字列は出さない)。 */}
                {state.hint.action != null &&
                  ` (${t(`hint.${state.hint.action}`, { defaultValue: state.hint.action })})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="threecardbrag-action-buttons">
              {isHumanBetTurn && (
                <>
                  {!humanPlayer?.seen && (
                    <button type="button" className={btnSecondary} onClick={handleSee} disabled={loading}>
                      {t('seeButton')}
                    </button>
                  )}
                  <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                    {t('betButton', { amount: betCost })}
                  </button>
                  <div className="flex items-center gap-1" data-testid="tcb-raise-controls">
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => setRaiseStake((a) => clampThreeCardBragRaise(a - 1, raiseMin, raiseMax))}
                      disabled={loading || !canRaise || raiseStake <= raiseMin}
                      aria-label={t('raiseDecrease')}
                    >
                      −
                    </button>
                    <span
                      className="text-ds-text-primary text-sm min-w-[4rem] text-center"
                      aria-live="polite"
                      data-testid="tcb-raise-amount"
                    >
                      {t('raisePrompt')} {raiseStake}
                    </span>
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => setRaiseStake((a) => clampThreeCardBragRaise(a + 1, raiseMin, raiseMax))}
                      disabled={loading || !canRaise || raiseStake >= raiseMax}
                      aria-label={t('raiseIncrease')}
                    >
                      ＋
                    </button>
                    <button
                      type="button"
                      className={btnWarning}
                      onClick={() => handleRaise(raiseStake)}
                      disabled={loading || !canRaise || raiseStake < raiseMin || raiseStake > raiseMax}
                    >
                      {t('raiseButton', { amount: raiseCost })}
                    </button>
                    <span className="text-ds-text-muted text-xs ml-1" data-testid="tcb-raise-range">
                      {canRaise ? t('raiseRange', { min: raiseMin, max: raiseMax }) : t('raiseUnavailable')}
                    </span>
                  </div>
                  <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                    {t('foldButton')}
                  </button>
                  {state.canShow && (
                    <button type="button" className={btnPrimary} onClick={handleShow} disabled={loading}>
                      {t('showButton')}
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
                dataTutorial="threecardbrag-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
