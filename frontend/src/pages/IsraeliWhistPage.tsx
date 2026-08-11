import { useCallback, useEffect, useMemo } from 'react';
import { israeliwhistApi } from '../api/gameApi';
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
import type { IsraeliWhistPlayer, IsraeliWhistResponse } from '../types/card';
import { IsraeliWhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { ISRAELIWHIST_HELP, parseIsraeliWhistCommand } from '../utils/cli/commands/israeliwhistCommands';
import { formatIsraeliWhistState } from '../utils/cli/formatters/israeliwhistFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tricks per round (thirteen cards each). */
const TRICKS_PER_ROUND = 13;

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the auction buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/** The auction opens at five. */
const AUCTION_MIN = 5;

/** Every call a player may make in the second round. */
const BIDS: readonly number[] = Array.from({ length: TRICKS_PER_ROUND + 1 }, (_, i) => i);

/** Guided tutorial steps (the two bidding rounds, scoring, seats, hand). */
const ISRAELIWHIST_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="iw-score"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="iw-actions"]', messageKey: 'tutorial.twoStage', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="iw-seats"]', messageKey: 'tutorial.seats', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="iw-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Israeli Whist page (wrapped by `withTutorial`).
 *
 * Renders the Israeli two-stage bidder: 52 cards, four players each for
 * themselves, thirteen apiece. **Bidding happens twice** — an auction settles
 * trump and a quota for whoever wins it, and then everyone calls their own
 * target separately — so the page keeps both rounds' state visible on every
 * seat rather than replacing one with the other. Calls the server is certain
 * to reject (below the winner's quota, or the one that would make the total
 * 13) are disabled rather than offered and refused.
 */
function IsraeliWhistPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('israeliwhist');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<IsraeliWhistResponse, Parameters<typeof israeliwhistApi.exec>>(israeliwhistApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('israeliwhist', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('israeliwhist');
  const cliConfig: CliGameConfig<IsraeliWhistResponse, Parameters<typeof israeliwhistApi.exec>> = useMemo(
    () => ({
      gameName: 'israeliwhist',
      parseCommand: parseIsraeliWhistCommand,
      formatResponse: formatIsraeliWhistState,
      helpText: ISRAELIWHIST_HELP,
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

  const handleAuction = useCallback(
    (bid: number, suit: number) => {
      void dispatch('auction', undefined, undefined, suit, bid);
    },
    [dispatch],
  );

  const handleAuctionPass = useCallback(() => {
    void dispatch('pass');
  }, [dispatch]);

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
    return (
      <GameSkeleton gameKey="israeliwhist" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isAuction = state.phase === IsraeliWhistPhase.AUCTION;
  const isBid = state.phase === IsraeliWhistPhase.BID;
  const isRoundEnd = state.phase === IsraeliWhistPhase.ROUND_END;
  const isGameEnd = state.phase === IsraeliWhistPhase.GAME_END || state.gameEndFlag;
  const humanPassed = state.players[0]?.passed === true;
  const isHumanAuctionTurn = isAuction && !isGameEnd && state.auctionPlayerIdx === 0 && !humanPassed;
  const isHumanBidTurn = isBid && !isGameEnd && state.bidPlayerIdx === 0;
  const isHumanTurn =
    !isGameEnd && !isRoundEnd && !isAuction && !isBid && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isAuction
        ? t('phase.auction')
        : isBid
          ? t('phase.bid')
          : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  /** The lowest auction call that would actually beat the standing one. */
  const nextAuctionBid = Math.max(AUCTION_MIN, state.highBid);

  /** A seat's standing in the auction, which the calling round does not replace. */
  const roleStr = (p: IsraeliWhistPlayer): string => {
    if (p.id === state.declarerIdx) return t('role.declarer', { n: String(p.auctionBid) });
    return p.passed ? t('role.passed') : t('role.active');
  };

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx === 0) return t('result.win', { score: String(state.players[0]?.totalScore ?? 0) });
    if (state.winnerIdx < 0) return t('result.tie');
    return t('result.lose', { idx: String(state.winnerIdx) });
  })();

  return (
    <GamePageShell
      title={tc('nav.israeliwhist')}
      gameThemeBg={gameTheme.israeliwhist.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/israeliwhist"
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
              <span className="mr-4" data-testid="iw-round">
                {t('header.round')}: {state.roundNumber}/{state.config.rounds}
              </span>
              <span className="mr-4" data-testid="iw-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
              <span data-testid="iw-trump">
                {state.trumpSuit > 0
                  ? t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?', n: String(state.highBid) })
                  : t('header.auction', {
                      n: String(state.highBid),
                      suit: SUIT_SYMBOLS[state.highSuit] ?? '-',
                    })}
              </span>
            </div>

            {/* **的中が2乗で跳ねることと全員一致の倍率は盤面から読めない。** */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
              data-testid="iw-score"
              data-tutorial="iw-score"
            >
              {t('header.scoreTable')}
            </div>

            {/* **2段階ぶんの状態を同時に出す。** 片方を消すとどちらの段か読めない。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="iw-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`iw-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  <span className="ml-1 text-ds-accent">[{roleStr(p)}]</span>
                  {': '}
                  {p.bid < 0 ? t('bid.none') : t('bid.value', { n: String(p.bid) })}
                  {' / '}
                  {t('header.took', { n: String(p.trickCount) })}
                  {' / '}
                  {t('header.total', { n: String(p.totalScore) })}
                </div>
              ))}
            </div>

            <div data-tutorial="iw-trick">
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
              <div className="mt-4" data-tutorial="iw-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="iw-actions">
              {isHumanAuctionTurn && (
                <>
                  {/* **入札は数とスートの両方で1つ。** スートごとにボタンを出す。 */}
                  {SUITS.map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnWarning}
                      onClick={() => handleAuction(nextAuctionBid, suit)}
                      disabled={loading}
                      data-testid={`iw-auction-${suit.toString()}-btn`}
                    >
                      {t('actions.auction', { n: String(nextAuctionBid), suit: SUIT_SYMBOLS[suit] ?? '?' })}
                    </button>
                  ))}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleAuctionPass}
                    disabled={loading}
                    data-testid="iw-pass-btn"
                  >
                    {t('actions.pass')}
                  </button>
                </>
              )}
              {isHumanBidTurn &&
                BIDS.map((bid) => {
                  // **サーバが必ず拒否する宣言は出さない。** 落札者のノルマ未満と、
                  // 合計が13になる値がそれにあたる。
                  const barred = bid < state.minimumBid || bid === state.restrictedBid;
                  return (
                    <button
                      key={bid}
                      type="button"
                      className={btnWarning}
                      onClick={() => handleBid(bid)}
                      disabled={loading || barred}
                      aria-disabled={barred}
                      data-testid={`iw-bid-${bid.toString()}-btn`}
                    >
                      {String(bid)}
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

/** Israeli Whist page wrapped with TutorialProvider. */
export const IsraeliWhistPage = withTutorial(IsraeliWhistPageContent, 'israeliwhist', ISRAELIWHIST_TUTORIAL_STEPS);
