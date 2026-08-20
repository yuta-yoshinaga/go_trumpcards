import { useCallback, useEffect, useMemo } from 'react';
import { reversisApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeErrorColors, badgeInfoColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ReversisPlayer, ReversisResponse } from '../types/card';
import { ReversisPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseReversisCommand, REVERSIS_HELP } from '../utils/cli/commands/reversisCommands';
import { formatReversisState } from '../utils/cli/formatters/reversisFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { reversisCardPoints } from '../utils/reversisPoints';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tricks per round (12 cards each on a 48-card pack). */
const TRICKS_PER_ROUND = 12;

/** Guided tutorial steps (pool, penalty rule, trick, hand). */
const REVERSIS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="reversis-pool"]', messageKey: 'tutorial.pool', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="reversis-penalty"]',
    messageKey: 'tutorial.penalty',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="reversis-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="reversis-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Reversis page (wrapped by `withTutorial`).
 *
 * Renders the 4-player French avoidance game on a 48-card pack (a standard 52
 * with the four 10s removed), no trump. Taking tricks costs card points, and
 * two marked cards — the **Quinola (J♥)** and the **A♦** — cost extra points
 * *and* chips into a pool that the fewest-penalty player takes whole.
 *
 * Two things the table cannot show, so the page states them outright: the
 * point scale, and how many chips are riding on the round.
 */
function ReversisPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('reversis');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<ReversisResponse, Parameters<typeof reversisApi.exec>>(reversisApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('reversis', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('reversis');
  const cliConfig: CliGameConfig<ReversisResponse, Parameters<typeof reversisApi.exec>> = useMemo(
    () => ({
      gameName: 'reversis',
      parseCommand: parseReversisCommand,
      formatResponse: formatReversisState,
      helpText: REVERSIS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNextRound = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="reversis" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isRoundEnd = state.phase === ReversisPhase.ROUND_END;
  const isGameEnd = state.phase === ReversisPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && !isRoundEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd ? t('phase.gameEnd') : isRoundEnd ? t('phase.roundEnd') : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  /** Which marked cards a seat has been landed with this round. */
  const markStr = (p: ReversisPlayer): string => {
    const marks = [p.tookQuinola && t('marks.quinola'), p.tookDiamondAce && t('marks.diamondAce')].filter(Boolean);
    return marks.length > 0 ? marks.join(', ') : t('marks.none');
  };

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx < 0) return t('result.tie');
    const winner = state.players[state.winnerIdx];
    return state.winnerIdx === 0
      ? t('result.youWin', { chips: String(winner?.chips ?? 0) })
      : t('result.cpuWin', { idx: String(state.winnerIdx), chips: String(winner?.chips ?? 0) });
  })();

  return (
    <GamePageShell
      title={tc('nav.reversis')}
      gameThemeBg={gameTheme.reversis.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/reversis"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerIdx === 0}
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
          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4" data-testid="rv-round">
                {t('header.round')}: {state.roundNumber}/{state.config.rounds}
              </span>
              <span className="mr-4" data-testid="rv-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
            </div>

            {/* 何を取り合っているのかは盤面から読めない。 */}
            <div
              className="mb-2 rounded bg-black/30 border border-ds-warning px-3 py-2 text-ds-text-primary text-sm text-center"
              data-testid="rv-pool"
              data-tutorial="reversis-pool"
            >
              {t('header.pool', { pool: String(state.pool) })}
            </div>

            {/* 失点の配分と印付きの2枚も盤面には出ない。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
              data-testid="rv-penalty-rule"
              data-tutorial="reversis-penalty"
            >
              {t('header.penaltyRule')}
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`rv-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {': '}
                  {t('header.seat', { chips: String(p.chips), penalty: String(p.roundPenalty) })} [{markStr(p)}]
                </div>
              ))}
            </div>

            <div data-tutorial="reversis-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {resultBanner && (
              <div className="text-center text-xl my-4 text-ds-accent font-semibold" role="status">
                {resultBanner}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <ErrorAlert message={error} onRetry={retry} />

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="reversis-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => handlePlay(idx)}
                      disabled={loading || !isHumanTurn}
                      // **点を取り合うのが核なのに、どの札が何点かは出ていなかった**
                      // (#5747)。A=4 / K=3 / Q=2 / J=1 を暗算し続けることになる。
                      aria-label={t('actions.playAriaWithPoints', {
                        card: cardAlt(card),
                        points: reversisCardPoints(card),
                      })}
                      className={`relative disabled:opacity-50 ${
                        legalRing.has(idx) ? 'rounded-lg ring-2 ring-ds-success' : ''
                      }`}
                    >
                      <CardImage card={card} width={cardWidth} />
                      <span
                        data-testid={`rv-points-${idx.toString()}`}
                        aria-hidden="true"
                        className={`absolute top-0 right-0 rounded-bl px-1 text-[10px] leading-tight ${
                          reversisCardPoints(card) > 0 ? badgeErrorColors : badgeInfoColors
                        }`}
                      >
                        {reversisCardPoints(card)}
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {hintEnabled && hint && (
              <div className="mt-3 flex justify-center">
                <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />
              </div>
            )}

            <div className="mt-4 flex flex-wrap gap-2">
              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('actions.nextRound')}
                </button>
              )}
              <button
                type="button"
                className={btnPrimary}
                onClick={() => requestConfirm(handleReset)}
                disabled={loading}
              >
                {t('actions.reset')}
              </button>
              {!isGameEnd && (
                <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                  {t('actions.giveUp')}
                </button>
              )}
            </div>

            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, hintEnabled, setHintEnabled)] }]}
            />
          </div>

          <ActionLogSection
            isEndPhase={isGameEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />
        </>
      )}
    </GamePageShell>
  );
}

/** Reversis page wrapped with TutorialProvider. */
export const ReversisPage = withTutorial(ReversisPageContent, 'reversis', REVERSIS_TUTORIAL_STEPS);
