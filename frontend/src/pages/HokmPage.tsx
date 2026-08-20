import { useCallback, useEffect, useMemo } from 'react';
import { hokmApi } from '../api/gameApi';
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
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { HokmResponse } from '../types/card';
import { HokmPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { HOKM_HELP, parseHokmCommand } from '../utils/cli/commands/hokmCommands';
import { formatHokmState } from '../utils/cli/formatters/hokmFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the trump buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/** Guided tutorial steps (the seven-trick race, the hakem, Kot, hand). */
const HOKM_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="hk-race"]', messageKey: 'tutorial.race', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="hk-actions"]', messageKey: 'tutorial.hakem', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="hk-seats"]', messageKey: 'tutorial.kot', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="hk-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Hokm page (wrapped by `withTutorial`).
 *
 * Renders Iran's most-played trick-taker: 52 cards, four players in two
 * partnerships with opposite seats allied, thirteen each.
 *
 * **A hand does not play out all thirteen tricks** — the first partnership to
 * seven takes it and the rest of the cards are never played. Nothing on a
 * trick-counter conveys that, so the page leads with the race to seven rather
 * than with "trick n of 13". The hakem declares trump from **five cards only**
 * and keeps the role while their team keeps winning; a hand where the losers
 * take nothing (**Kot**) is worth two, which the page says outright because a
 * jump of two points is otherwise unexplained.
 */
function HokmPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('hokm');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<HokmResponse, Parameters<typeof hokmApi.exec>>(hokmApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('hokm', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('hokm');
  const cliConfig: CliGameConfig<HokmResponse, Parameters<typeof hokmApi.exec>> = useMemo(
    () => ({
      gameName: 'hokm',
      parseCommand: parseHokmCommand,
      formatResponse: formatHokmState,
      helpText: HOKM_HELP,
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

  const handleTrump = useCallback(
    (suit: number) => {
      void dispatch('trump', undefined, undefined, suit);
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
    return <GameSkeleton gameKey="hokm" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isTrump = state.phase === HokmPhase.TRUMP;
  const isHandEnd = state.phase === HokmPhase.HAND_END;
  const isGameEnd = state.phase === HokmPhase.GAME_END || state.gameEndFlag;
  const isHumanTrumpTurn = isTrump && !isGameEnd && state.hakemIdx === 0;
  const isHumanTurn = !isGameEnd && !isHandEnd && !isTrump && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isHandEnd
      ? t('phase.handEnd')
      : isTrump
        ? t('phase.trump')
        : t('phase.play');

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
      title={tc('nav.hokm')}
      gameThemeBg={gameTheme.hokm.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/hokm"
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
              <span className="mr-4" data-testid="hk-hand">
                {t('header.hand')}: {state.handNumber}
              </span>
              <span className="mr-4">{t('header.target', { target: String(state.config.target) })}</span>
              <span data-testid="hk-trump">
                {state.trumpSuit > 0
                  ? t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })
                  : t('header.trumpUndecided')}
              </span>
            </div>

            {/* **13 まで打たない。** 進捗はトリック数の競り合いのほうに出る。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="hk-race"
              data-tutorial="hk-race"
            >
              {t('header.race', {
                t0: String(state.teamTricks[0] ?? 0),
                t1: String(state.teamTricks[1] ?? 0),
                need: String(state.tricksToWin),
              })}
            </div>

            <div className="text-ds-text-muted text-sm text-center mb-3" data-testid="hk-score">
              {t('header.score', { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) })}
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="hk-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`hk-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  <span className="ml-1 text-ds-accent">{t('header.team', { team: String(p.team) })}</span>
                  {p.isHakem && <span className="ml-1 text-ds-accent">{t('header.hakem')}</span>}
                  {': '}
                  {t('header.took', { n: String(p.trickCount) })}
                </div>
              ))}
            </div>

            <div data-tutorial="hk-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {/* **Kot は 2 点。** 何が起きたか言わないと得点が飛んで見える。 */}
            {isHandEnd && state.lastHandWinner >= 0 && (
              <div className="text-center my-3 text-ds-accent" role="status" data-testid="hk-hand-result">
                {state.lastHandKot
                  ? t('handEnd.kot', { team: String(state.lastHandWinner) })
                  : t('handEnd.normal', { team: String(state.lastHandWinner), need: String(state.tricksToWin) })}
                {/* **親は負けたときだけ交代する** (#5753)。次に自分が切り札を
                    選べるかを左右するのに、次ハンドが始まって親バッジが動くまで
                    分からなかった。Kot でも通常勝利でも同じように続けて出す。 */}
                <span className="ml-2" data-testid="hk-hakem-change">
                  {state.lastHandHakemChanged ? t('handEnd.hakemMoves') : t('handEnd.hakemStays')}
                </span>
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
              <div className="mt-4" data-tutorial="hk-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="hk-actions">
              {isHumanTrumpTurn &&
                SUITS.map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnWarning}
                    onClick={() => handleTrump(suit)}
                    disabled={loading}
                    data-testid={`hk-trump-${suit.toString()}-btn`}
                  >
                    {t('actions.trump', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
                  </button>
                ))}
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

/** Hokm page wrapped with TutorialProvider. */
export const HokmPage = withTutorial(HokmPageContent, 'hokm', HOKM_TUTORIAL_STEPS);
