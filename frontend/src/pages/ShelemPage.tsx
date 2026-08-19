import { useCallback, useEffect, useMemo, useState } from 'react';
import { shelemApi } from '../api/gameApi';
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
import type { ShelemPlayer, ShelemResponse } from '../types/card';
import { ShelemPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseShelemCommand, SHELEM_HELP } from '../utils/cli/commands/shelemCommands';
import { formatShelemState } from '../utils/cli/formatters/shelemFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Card points on the table each round (`ShelemHandPoints` in the domain). */
const SHELEM_HAND_POINTS = 100;

/** Tricks per round (twelve cards each). */
const TRICKS_PER_ROUND = 12;

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the trump buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/**
 * Bidding tops out at the whole round's card points (sync: `ShelemMaxBid` in
 * `internal/domain/Shelem.go`). **A contract above 100 could never be made**,
 * so those buttons are not offered at all.
 */
const BID_MAX = 100;

/** Guided tutorial steps (the point cards, bidding, the widow, hand). */
const SHELEM_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sh-points"]', messageKey: 'tutorial.points', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sh-actions"]', messageKey: 'tutorial.bid', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sh-seats"]', messageKey: 'tutorial.widow', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sh-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Shelem page (wrapped by `withTutorial`).
 *
 * Renders Iran's bridge-shaped bidder: 52 cards as twelve each plus a
 * four-card widow, four players in two partnerships.
 *
 * **What is bid is the score itself**, not a number of tricks, over a hand
 * that holds exactly 100 card points — and only three ranks carry any (A and
 * 10 are 10, the 5 is 5). Neither fact is visible on the table, so the page
 * states the point table outright and shows the contract against the card
 * points taken so far. The bid buttons offer only amounts that would actually
 * beat the standing bid.
 */
function ShelemPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('shelem');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<ShelemResponse, Parameters<typeof shelemApi.exec>>(shelemApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('shelem', state);
  // 捨てる4枚は押して選ぶ。サーバに送るまではこちらで保持する。
  const [picked, setPicked] = useState<number[]>([]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('shelem');
  const cliConfig: CliGameConfig<ShelemResponse, Parameters<typeof shelemApi.exec>> = useMemo(
    () => ({
      gameName: 'shelem',
      parseCommand: parseShelemCommand,
      formatResponse: formatShelemState,
      helpText: SHELEM_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    setPicked([]);
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleBid = useCallback(
    (bid: number) => {
      void dispatch('bid', undefined, undefined, undefined, bid);
    },
    [dispatch],
  );

  const handleShelem = useCallback(() => {
    void dispatch('shelem');
  }, [dispatch]);

  const handlePass = useCallback(() => {
    void dispatch('pass');
  }, [dispatch]);

  const handleDiscard = useCallback(
    (suit: number) => {
      void dispatch('discard', undefined, undefined, suit, undefined, picked);
      setPicked([]);
    },
    [dispatch, picked],
  );

  const togglePick = useCallback((idx: number) => {
    setPicked((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  }, []);

  const handleNextRound = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="shelem" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === ShelemPhase.BID;
  const isDiscard = state.phase === ShelemPhase.DISCARD;
  const isRoundEnd = state.phase === ShelemPhase.ROUND_END;
  const isGameEnd = state.phase === ShelemPhase.GAME_END || state.gameEndFlag;
  const humanPassed = state.players[0]?.passed === true;
  const isHumanBidTurn = isBid && !isGameEnd && state.bidPlayerIdx === 0 && !humanPassed;
  const isHumanDiscardTurn = isDiscard && !isGameEnd && state.declarerIdx === 0;
  const isHumanTurn =
    !isGameEnd && !isRoundEnd && !isBid && !isDiscard && state.players[state.currentPlayerIdx]?.isHuman === true;

  // **誰も入札しないまま最後の1人になったら降りられない。** 契約が決まらないため。
  const mustBid = state.contract === 0 && state.players.filter((p) => !p.passed).length === 1;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isBid
        ? t('phase.bid')
        : isDiscard
          ? t('phase.discard')
          : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  /** Bids that would actually beat the standing one, capped at the maximum. */
  const bidChoices: number[] = [];
  for (let b = state.minBid; b <= BID_MAX; b += 5) bidChoices.push(b);

  /** A seat's standing in the bidding. */
  const roleStr = (p: ShelemPlayer): string => {
    if (p.id === state.declarerIdx) {
      return p.declaredShelem ? t('role.shelem') : t('role.declarer', { n: String(p.bid) });
    }
    if (p.passed) return t('role.passed');
    return p.bid >= 0 ? t('role.bid', { n: String(p.bid) }) : t('role.active');
  };

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const params = { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) };
    if (state.winnerTeam === 0) return t('result.youWin', params);
    if (state.winnerTeam === 1) return t('result.theyWin', params);
    return t('result.tie', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.shelem')}
      gameThemeBg={gameTheme.shelem.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/shelem"
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
              <span className="mr-4" data-testid="sh-round">
                {t('header.round')}: {state.roundNumber}
              </span>
              <span className="mr-4" data-testid="sh-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
              <span>{t('header.target', { target: String(state.config.target) })}</span>
            </div>

            {/* **点になるのは A/10/5 だけ。** 盤面から読めない。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
              data-testid="sh-points"
              data-tutorial="sh-points"
            >
              {t('header.pointTable')}
            </div>

            <div className="text-ds-text-primary text-center mb-3" data-testid="sh-contract">
              {state.declarerIdx < 0
                ? t('header.contractUndecided', { min: String(state.minBid) })
                : state.shelemBid
                  ? t('header.contractShelem')
                  : t('header.contract', {
                      n: String(state.contract),
                      got: String(state.roundPoints[state.declarerIdx % 2] ?? 0),
                    })}
              {/* **守備側の点も出す** (#5754)。契約を阻止できているかは、
                  宣言側の点だけ見ていても分からない。合計は必ず 100 点。 */}
              {state.declarerIdx >= 0 && (
                <span className="ml-2 text-ds-text-muted" data-testid="sh-defenders">
                  {t('header.defenders', {
                    got: String(state.roundPoints[(state.declarerIdx + 1) % 2] ?? 0),
                    total: String(SHELEM_HAND_POINTS),
                  })}
                </span>
              )}
              {state.trumpSuit > 0 && ` / ${t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })}`}
            </div>

            <div className="text-ds-text-muted text-sm text-center mb-3" data-testid="sh-score">
              {t('header.score', { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) })}
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="sh-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`sh-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  <span className="ml-1 text-ds-accent">{t('header.team', { team: String(p.team) })}</span>
                  {': '}
                  {roleStr(p)}
                  {' / '}
                  {t('header.took', { n: String(p.trickCount) })}
                </div>
              ))}
            </div>

            <div data-tutorial="sh-trick">
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
              <div className="mt-4" data-tutorial="sh-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                  {isHumanDiscardTurn &&
                    ` — ${t('discard.picked', { n: String(picked.length), of: String(state.discardCount) })}`}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => (isHumanDiscardTurn ? togglePick(idx) : handlePlay(idx))}
                      disabled={loading || (!isHumanTurn && !isHumanDiscardTurn)}
                      aria-pressed={isHumanDiscardTurn ? picked.includes(idx) : undefined}
                      aria-label={
                        isHumanDiscardTurn
                          ? t('discard.pickAria', { card: cardAlt(card) })
                          : t('actions.playAria', { card: cardAlt(card) })
                      }
                      className={`disabled:opacity-50 ${
                        isHumanDiscardTurn && picked.includes(idx)
                          ? 'rounded-lg ring-2 ring-ds-warning'
                          : legalRing.has(idx)
                            ? 'rounded-lg ring-2 ring-ds-success'
                            : ''
                      }`}
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="sh-actions">
              {isHumanBidTurn && (
                <>
                  {/* **上回れる額だけを出す。** サーバが必ず拒否する額は出さない。 */}
                  {bidChoices.map((bid) => (
                    <button
                      key={bid}
                      type="button"
                      className={btnWarning}
                      onClick={() => handleBid(bid)}
                      disabled={loading}
                      data-testid={`sh-bid-${bid.toString()}-btn`}
                    >
                      {String(bid)}
                    </button>
                  ))}
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={handleShelem}
                    disabled={loading}
                    data-testid="sh-shelem-btn"
                  >
                    {t('actions.shelem')}
                  </button>
                  {/* **最後の1人は降りられない。** 押せないボタンを出さない。 */}
                  {mustBid ? (
                    <span className="self-center text-ds-text-muted text-sm" data-testid="sh-must-bid">
                      {t('mustBid')}
                    </span>
                  ) : (
                    <button
                      type="button"
                      className={btnSuccess}
                      onClick={handlePass}
                      disabled={loading}
                      data-testid="sh-pass-btn"
                    >
                      {t('actions.pass')}
                    </button>
                  )}
                </>
              )}
              {isHumanDiscardTurn &&
                SUITS.map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnWarning}
                    onClick={() => handleDiscard(suit)}
                    // **ちょうど4枚選ぶまで押させない。** サーバが必ず拒否する。
                    disabled={loading || picked.length !== state.discardCount}
                    aria-disabled={picked.length !== state.discardCount}
                    data-testid={`sh-discard-${suit.toString()}-btn`}
                  >
                    {t('actions.discard', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
                  </button>
                ))}
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

/** Shelem page wrapped with TutorialProvider. */
export const ShelemPage = withTutorial(ShelemPageContent, 'shelem', SHELEM_TUTORIAL_STEPS);
