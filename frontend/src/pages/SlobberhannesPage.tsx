import { useCallback, useEffect, useMemo } from 'react';
import { slobberhannesApi } from '../api/gameApi';
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
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SlobberhannesPlayer, SlobberhannesResponse } from '../types/card';
import { SlobberhannesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSlobberhannesCommand, SLOBBERHANNES_HELP } from '../utils/cli/commands/slobberhannesCommands';
import { formatSlobberhannesState } from '../utils/cli/formatters/slobberhannesFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tricks per round. The last one carries a penalty, so this count is load-bearing. */
const TRICKS_PER_ROUND = 8;

/** Guided tutorial steps (penalty warning, trick, hand, actions). */
const SLOBBERHANNES_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="slobberhannes-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="slobberhannes-trick"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="slobberhannes-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="slobberhannes-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Inner content for the Slobberhannes page (wrapped by `withTutorial`).
 *
 * Renders the 4-player avoidance trick-taking game: a 32-card piquet pack, no
 * trump, and three penalties of one point each — taking the first trick, the
 * last trick, or the trick holding the Q of clubs. Two of the three attach to
 * a trick's **position** rather than its contents, so the page calls out the
 * opening and closing tricks explicitly; nothing on the table would otherwise
 * show that those two are dangerous.
 */
function SlobberhannesPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('slobberhannes');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<SlobberhannesResponse, Parameters<typeof slobberhannesApi.exec>>(slobberhannesApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('slobberhannes', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('slobberhannes');
  const cliConfig: CliGameConfig<SlobberhannesResponse, Parameters<typeof slobberhannesApi.exec>> = useMemo(
    () => ({
      gameName: 'slobberhannes',
      parseCommand: parseSlobberhannesCommand,
      formatResponse: formatSlobberhannesState,
      helpText: SLOBBERHANNES_HELP,
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
    return (
      <GameSkeleton gameKey="slobberhannes" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isRoundEnd = state.phase === SlobberhannesPhase.ROUND_END;
  const isGameEnd = state.phase === SlobberhannesPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && !isRoundEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd ? t('phase.gameEnd') : isRoundEnd ? t('phase.roundEnd') : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない。サーバが必ず検証するし、
  // 個別に disabled にすると e2e の「先頭の札を押す」が動く的になる。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const isFirstTrick = state.trickNumber === 0;
  const isLastTrick = state.trickNumber === TRICKS_PER_ROUND - 1;

  /** The three penalty marks a seat has already taken this round. */
  const penaltyMarks = (p: SlobberhannesPlayer): string => {
    const marks = [
      p.tookFirstTrick && t('marks.first'),
      p.tookLastTrick && t('marks.last'),
      p.tookQueen && t('marks.queen'),
    ].filter(Boolean);
    return marks.length > 0 ? marks.join(', ') : t('marks.clean');
  };

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx < 0) return t('result.tie');
    const winner = state.players[state.winnerIdx];
    return state.winnerIdx === 0
      ? t('result.youWin', { score: String(winner?.score ?? 0) })
      : t('result.cpuWin', { idx: String(state.winnerIdx), score: String(winner?.score ?? 0) });
  })();

  return (
    <GamePageShell
      title={tc('nav.slobberhannes')}
      gameThemeBg={gameTheme.slobberhannes.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/slobberhannes"
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
              <span className="mr-4" data-testid="sh-round">
                {t('header.round')}: {state.roundNumber}/{state.config.rounds}
              </span>
              <span className="mr-4" data-testid="sh-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
              <span className="text-ds-accent">{t('header.noTrump')}</span>
            </div>

            {/* 最初と最後のトリックは中身に関係なく罰点対象。盤面には出ない情報。 */}
            {!isGameEnd && !isRoundEnd && (isFirstTrick || isLastTrick) && (
              <div
                className="mb-3 rounded bg-black/30 border border-ds-warning px-3 py-2 text-ds-text-primary text-sm text-center"
                role="status"
                data-testid="sh-position-warning"
              >
                {isFirstTrick ? t('warn.first') : t('warn.last')}
              </div>
            )}

            {/* 得点と、そのラウンドで受けている罰の内訳 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="slobberhannes-scores">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`sh-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {': '}
                  {t('header.score', { score: String(p.score) })} [{penaltyMarks(p)}]
                </div>
              ))}
            </div>

            <div data-tutorial="slobberhannes-trick">
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
              <div className="mt-4" data-tutorial="slobberhannes-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="slobberhannes-actions">
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

/** Slobberhannes page wrapped with TutorialProvider. */
export const SlobberhannesPage = withTutorial(SlobberhannesPageContent, 'slobberhannes', SLOBBERHANNES_TUTORIAL_STEPS);
