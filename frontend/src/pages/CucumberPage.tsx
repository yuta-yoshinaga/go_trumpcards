import { useCallback, useEffect, useMemo, useState } from 'react';
import { cucumberApi } from '../api/gameApi';
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
import { btnDanger, btnPrimary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CucumberResponse } from '../types/card';
import { CucumberPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CUCUMBER_HELP, parseCucumberCommand } from '../utils/cli/commands/cucumberCommands';
import { formatCucumberState } from '../utils/cli/formatters/cucumberFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guided tutorial steps (the comparison rule, the penalty, the strategy, your hand). */
const CUCUMBER_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="cu-rule"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cu-seats"]', messageKey: 'tutorial.penalty', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cu-threshold"]', messageKey: 'tutorial.strategy', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cu-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Cucumber page (wrapped by `withTutorial`).
 *
 * **The rank to beat is the whole game state.** Suits are irrelevant, so the
 * board tells you nothing a player can act on unless the threshold is spelled
 * out — the page leads with it.
 *
 * It also distinguishes "choose which high card to spend" from "your lowest
 * card is the only legal play". Those look identical when only one move is
 * legal, and the server's `forced` flag is the only reliable way to tell them
 * apart: a hand can have exactly one card that *does* beat the trick.
 */
function CucumberPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cucumber');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<CucumberResponse, Parameters<typeof cucumberApi.exec>>(cucumberApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('cucumber', state);
  const [playerCnt, setPlayerCnt] = useState(4);
  const [targetScore, setTargetScore] = useState(30);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cucumber');
  const cliConfig: CliGameConfig<CucumberResponse, Parameters<typeof cucumberApi.exec>> = useMemo(
    () => ({
      gameName: 'cucumber',
      parseCommand: parseCucumberCommand,
      formatResponse: formatCucumberState,
      helpText: CUCUMBER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', undefined, { playerCnt, targetScore });
  }, [dispatch, hideActionLog, playerCnt, targetScore]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNext = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="cucumber" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 7 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === CucumberPhase.GAME_END || state.gameEndFlag;
  const isRoundEnd = state.phase === CucumberPhase.ROUND_END && !isGameEnd;
  const isHumanTurn =
    state.phase === CucumberPhase.PLAY && !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const seatName = (idx: number) => (idx === 0 ? t('header.you') : t('header.cpu', { idx: String(idx) }));

  const phaseName = (() => {
    if (isGameEnd) return t('phase.gameEnd');
    if (isRoundEnd) return t('phase.roundEnd');
    return t('phase.play');
  })();

  const resultBanner = (() => {
    if (!isGameEnd || state.winnerIdx < 0) return null;
    const n = String(state.players[state.winnerIdx]?.penalty ?? 0);
    return state.winnerIdx === 0 ? t('result.you', { n }) : t('result.cpu', { name: seatName(state.winnerIdx), n });
  })();

  return (
    <GamePageShell
      title={tc('nav.cucumber')}
      gameThemeBg={gameTheme.cucumber.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/cucumber"
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
            <div className="text-ds-text-primary text-center mb-2" data-testid="cu-header">
              <span className="mr-4">{t('header.round', { n: String(state.roundNumber) })}</span>
              <span className="mr-4">{t('header.trick', { n: String(state.trickNumber + 1) })}</span>
              <span>{t('header.target', { n: String(state.config.targetScore) })}</span>
            </div>

            {/* **スート無関係・失点は最終トリックだけ、が規則そのもの。** */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="cu-rule"
              data-tutorial="cu-rule"
            >
              {t('header.rule')}
            </div>

            {/* **超えるべきランクが盤面の全て。** スートが無い以上、これが唯一の手がかり。 */}
            <div
              className="mb-3 text-center text-ds-accent font-semibold"
              data-testid="cu-threshold"
              data-tutorial="cu-threshold"
            >
              {state.highestInTrick > 0 ? t('header.highest', { n: String(state.highestInTrick) }) : t('header.lead')}
            </div>

            {/* **失点がそのまま順位。** 少ないほうが良い。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="cu-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`cu-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">{seatName(p.id)}</span>
                  {p.id === state.lastTrickWinnerIdx && state.lastPenalty > 0 && (
                    <span className="ml-1 text-ds-warning">
                      {t('header.lastTrick', { n: String(state.lastPenalty) })}
                    </span>
                  )}
                  {': '}
                  <span>{t('header.cards', { n: String(p.cardCount) })}</span>
                  {' / '}
                  <span className="text-ds-accent">{t('header.penalty', { n: String(p.penalty) })}</span>
                </div>
              ))}
            </div>

            <div data-tutorial="cu-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="cu-result"
              >
                {resultBanner}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <ErrorAlert message={error} onRetry={retry} />

            {/* **失点はラウンドに1回だけの出来事。** 配り直す前に読ませます。 */}
            {isRoundEnd && state.lastTrickWinnerIdx >= 0 && (
              <div className="mt-3 text-center text-ds-warning" role="status" data-testid="cu-round-end">
                {t('status.roundEnd', {
                  name: seatName(state.lastTrickWinnerIdx),
                  n: String(state.lastPenalty),
                })}
              </div>
            )}

            {/* **「選べる」と「決まっている」を言い分けます。** */}
            {isHumanTurn && (
              <div className="mt-3 text-center text-ds-text-muted" role="status" data-testid="cu-status">
                {state.forced
                  ? t('status.forced')
                  : state.highestInTrick > 0
                    ? t('status.beat', { n: String(state.highestInTrick) })
                    : t('status.lead')}
              </div>
            )}

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="cu-hand">
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
                      aria-label={t('actions.playAria', { card: cardAlt(card) })}
                      className={`disabled:opacity-50 ${legalRing.has(idx) ? 'rounded-lg ring-2 ring-ds-success' : ''}`}
                    >
                      <CardImage card={card} width={cardWidth} />
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
              {isRoundEnd && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleNext}
                  disabled={loading}
                  data-testid="cu-next-btn"
                >
                  {t('actions.next')}
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
              groups={[
                {
                  items: [
                    {
                      type: 'select',
                      id: 'cucumber-players',
                      label: t('actions.players'),
                      value: String(playerCnt),
                      options: [3, 4, 5, 6].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setPlayerCnt(Number(v)),
                      testId: 'cu-players-select',
                    },
                    {
                      type: 'select',
                      id: 'cucumber-target',
                      label: t('actions.target'),
                      value: String(targetScore),
                      options: [20, 30, 50].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setTargetScore(Number(v)),
                      testId: 'cu-target-select',
                    },
                    hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                  ],
                },
              ]}
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

/** Cucumber page wrapped with TutorialProvider. */
export const CucumberPage = withTutorial(CucumberPageContent, 'cucumber', CUCUMBER_TUTORIAL_STEPS);
