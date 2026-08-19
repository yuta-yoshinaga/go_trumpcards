import { useCallback, useEffect, useMemo } from 'react';
import { mendikotApi } from '../api/gameApi';
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
import type { MendikotResponse } from '../types/card';
import { MendikotPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MENDIKOT_HELP, parseMendikotCommand } from '../utils/cli/commands/mendikotCommands';
import { formatMendikotState } from '../utils/cli/formatters/mendikotFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Guided tutorial steps (the race for the tens, trump, the hand, your cards). */
const MENDIKOT_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="md-tens"]', messageKey: 'tutorial.tens', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="md-trump"]', messageKey: 'tutorial.trump', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="md-seats"]', messageKey: 'tutorial.mendikot', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="md-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Mendikot page (wrapped by `withTutorial`).
 *
 * Renders India's household partnership trick-taker: 52 cards, four players in
 * two partnerships with opposite seats allied, thirteen each.
 *
 * **What decides the hand is neither points nor tricks but the four tens.**
 * Three of them takes it outright; two apiece falls back to the trick count.
 * A trick counter alone says nothing about that, so the page leads with the
 * tens. **There is also no trump-choosing step** — the suit is fixed by
 * whichever card the first player who cannot follow chooses to play, which the
 * page states outright because otherwise trump appears from nowhere mid-hand.
 */
function MendikotPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('mendikot');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<MendikotResponse, Parameters<typeof mendikotApi.exec>>(mendikotApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('mendikot', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('mendikot');
  const cliConfig: CliGameConfig<MendikotResponse, Parameters<typeof mendikotApi.exec>> = useMemo(
    () => ({
      gameName: 'mendikot',
      parseCommand: parseMendikotCommand,
      formatResponse: formatMendikotState,
      helpText: MENDIKOT_HELP,
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

  const handleNextHand = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="mendikot" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isHandEnd = state.phase === MendikotPhase.HAND_END;
  const isGameEnd = state.phase === MendikotPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && !isHandEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd ? t('phase.gameEnd') : isHandEnd ? t('phase.handEnd') : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const params = { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) };
    if (state.winnerTeam === 0) return t('result.youWin', params);
    if (state.winnerTeam === 1) return t('result.theyWin', params);
    return t('result.tie', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.mendikot')}
      gameThemeBg={gameTheme.mendikot.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/mendikot"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerTeam === 0}
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
              <span className="mr-4" data-testid="md-hand">
                {t('header.hand')}: {state.handNumber}
              </span>
              <span className="mr-4">{t('header.target', { target: String(state.config.target) })}</span>
              <span data-testid="md-trump" data-tutorial="md-trump">
                {state.trumpSuit > 0
                  ? t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })
                  : t('header.trumpUndecided')}
              </span>
            </div>

            {/* **切り札は宣言ではなく事故で決まる** (#5755)。フォローできない
                手番はハンド全体を左右する一度きりの選択なのに、警告が無かった。 */}
            {state.willSetTrump && (
              <div
                className="mb-3 rounded bg-black/30 border border-ds-warning px-3 py-2 text-ds-text-primary text-center"
                role="status"
                data-testid="md-sets-trump-warning"
              >
                {t('warn.setsTrump')}
              </div>
            )}

            {/* **勝敗は 10 の枚数で決まる。** 盤面から読めないので先頭に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="md-tens"
              data-tutorial="md-tens"
            >
              {t('header.tens', {
                t0: String(state.teamTens[0] ?? 0),
                t1: String(state.teamTens[1] ?? 0),
                total: String(state.tensInDeck),
              })}
            </div>

            <div className="text-ds-text-muted text-sm text-center mb-1" data-testid="md-tricks">
              {t('header.tricks', {
                t0: String(state.teamTricks[0] ?? 0),
                t1: String(state.teamTricks[1] ?? 0),
              })}
            </div>

            <div className="text-ds-text-muted text-sm text-center mb-3" data-testid="md-score">
              {t('header.score', { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) })}
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="md-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`md-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  <span className="ml-1 text-ds-accent">{t('header.team', { team: String(p.team) })}</span>
                  {p.id === state.trumpChooserIdx && <span className="ml-1 text-ds-accent">{t('header.chooser')}</span>}
                  {': '}
                  {t('header.seatTens', { n: String(p.tens) })}
                  {' / '}
                  {t('header.took', { n: String(p.trickCount) })}
                </div>
              ))}
            </div>

            <div data-tutorial="md-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {/* **決まり方で 1/2/3 点と変わる。** 言わないと得点が飛んで見える。 */}
            {isHandEnd && state.lastHandWinner >= 0 && (
              <div className="text-center my-3 text-ds-accent" role="status" data-testid="md-hand-result">
                {t(`handEnd.${state.lastHandKind}`, { team: String(state.lastHandWinner) })}
              </div>
            )}

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
              <div className="mt-4" data-tutorial="md-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="md-actions">
              {isHandEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextHand} disabled={loading}>
                  {t('actions.nextHand')}
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

/** Mendikot page wrapped with TutorialProvider. */
export const MendikotPage = withTutorial(MendikotPageContent, 'mendikot', MENDIKOT_TUTORIAL_STEPS);
