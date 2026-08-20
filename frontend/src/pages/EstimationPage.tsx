import { useCallback, useEffect, useMemo } from 'react';
import { estimationApi } from '../api/gameApi';
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
import type { EstimationPlayer, EstimationResponse } from '../types/card';
import { EstimationCall, EstimationPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { ESTIMATION_HELP, parseEstimationCommand } from '../utils/cli/commands/estimationCommands';
import { formatEstimationState } from '../utils/cli/formatters/estimationFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Signed round delta: the plus sign is ours to add, and 0 reads as "no change". */
function signedScore(n: number): string {
  if (n > 0) return `+${n.toString()}`;
  if (n === 0) return '±0';
  return n.toString();
}

/** Tricks per round (thirteen cards each). */
const TRICKS_PER_ROUND = 13;

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol, for the trump readout. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the trump buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/** Every call a player may make, from a Dash Call up to the whole hand. */
const BIDS: readonly number[] = Array.from({ length: TRICKS_PER_ROUND + 1 }, (_, i) => i);

/** Guided tutorial steps (scoring, trump, calling, hand). */
const ESTIMATION_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="est-score"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="est-actions"]', messageKey: 'tutorial.call', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="est-seats"]', messageKey: 'tutorial.seats', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="est-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Estimation page (wrapped by `withTutorial`).
 *
 * Renders the Gulf's household Oh Hell: 52 cards, four players each for
 * themselves, thirteen cards apiece. **Only an exact call scores** — one short
 * and five over lose the same amount — so the page states the scoring outright
 * rather than leaving the player to infer it, and marks the barred call before
 * the player can press it rather than after the server rejects it.
 */
function EstimationPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('estimation');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<EstimationResponse, Parameters<typeof estimationApi.exec>>(estimationApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('estimation', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('estimation');
  const cliConfig: CliGameConfig<EstimationResponse, Parameters<typeof estimationApi.exec>> = useMemo(
    () => ({
      gameName: 'estimation',
      parseCommand: parseEstimationCommand,
      formatResponse: formatEstimationState,
      helpText: ESTIMATION_HELP,
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

  const handleBid = useCallback(
    (bid: number) => {
      void dispatch('bid', undefined, undefined, undefined, bid);
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
    return <GameSkeleton gameKey="estimation" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isTrump = state.phase === EstimationPhase.TRUMP;
  const isBid = state.phase === EstimationPhase.BID;
  const isRoundEnd = state.phase === EstimationPhase.ROUND_END;
  const isGameEnd = state.phase === EstimationPhase.GAME_END || state.gameEndFlag;
  const isHumanTrumpTurn = isTrump && !isGameEnd && state.dealerIdx === 0;
  const isHumanBidTurn = isBid && !isGameEnd && state.bidPlayerIdx === 0;
  const isHumanTurn =
    !isGameEnd && !isRoundEnd && !isTrump && !isBid && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isTrump
        ? t('phase.trump')
        : isBid
          ? t('phase.bid')
          : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  /** A seat's call, spelled out with its kind rather than as a bare number. */
  const bidStr = (p: EstimationPlayer): string => {
    if (p.bid < 0) return t('bid.none');
    if (p.callType === EstimationCall.DASH) return t('bid.dash');
    if (p.callType === EstimationCall.RISK) return t('bid.risk', { n: String(p.bid) });
    return t('bid.normal', { n: String(p.bid) });
  };

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx === 0) return t('result.win', { score: String(state.players[0]?.totalScore ?? 0) });
    if (state.winnerIdx < 0) return t('result.tie');
    return t('result.lose', { idx: String(state.winnerIdx) });
  })();

  return (
    <GamePageShell
      title={tc('nav.estimation')}
      gameThemeBg={gameTheme.estimation.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/estimation"
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
              <span className="mr-4" data-testid="est-round">
                {t('header.round')}: {state.roundNumber}/{state.config.rounds}
              </span>
              <span className="mr-4" data-testid="est-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
              <span data-testid="est-trump">
                {state.trumpSuit > 0
                  ? t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })
                  : t('header.trumpUndecided')}
              </span>
            </div>

            {/* **得点表は盤面から読めない。** 的中だけが得点になることを明示する。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
              data-testid="est-score"
              data-tutorial="est-score"
            >
              {t('header.scoreTable')}
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="est-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`est-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {': '}
                  {bidStr(p)}
                  {' / '}
                  {t('header.took', { n: String(p.trickCount) })}
                  {' / '}
                  <span className="text-ds-accent">{t('header.total', { n: String(p.totalScore) })}</span>
                  {/* **得点式が複雑（10+宣言 / Dash Call ±23 / Risk 2倍）なので、
                      累計の差分を暗算させない** (#5751)。増減が確定するラウンド
                      終了時にだけ出す。 */}
                  {isRoundEnd && (
                    <span
                      className={`ml-2 ${p.roundScore < 0 ? 'text-ds-error' : 'text-ds-success'}`}
                      data-testid={`est-round-delta-${p.id.toString()}`}
                    >
                      {t('header.roundDelta', { delta: signedScore(p.roundScore) })}
                    </span>
                  )}
                </div>
              ))}
            </div>

            <div data-tutorial="est-trick">
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
              <div className="mt-4" data-tutorial="est-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="est-actions">
              {isHumanTrumpTurn &&
                SUITS.map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnWarning}
                    onClick={() => handleTrump(suit)}
                    disabled={loading}
                    data-testid={`est-trump-${suit.toString()}-btn`}
                  >
                    {t('actions.trump', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
                  </button>
                ))}
              {isHumanBidTurn &&
                BIDS.map((bid) => {
                  // **禁止値は押せなくする。** 合計が13になる宣言はサーバが必ず
                  // 拒否するので、押させてから断るより出さないほうが正直。
                  const barred = bid === state.restrictedBid;
                  return (
                    <button
                      key={bid}
                      type="button"
                      className={btnWarning}
                      onClick={() => handleBid(bid)}
                      disabled={loading || barred}
                      aria-disabled={barred}
                      data-testid={`est-bid-${bid.toString()}-btn`}
                    >
                      {bid === 0 ? t('actions.dash') : String(bid)}
                    </button>
                  );
                })}
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

/** Estimation page wrapped with TutorialProvider. */
export const EstimationPage = withTutorial(EstimationPageContent, 'estimation', ESTIMATION_TUTORIAL_STEPS);
