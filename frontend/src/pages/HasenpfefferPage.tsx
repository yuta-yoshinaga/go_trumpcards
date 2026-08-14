import { useCallback, useEffect, useMemo, useState } from 'react';
import { hasenpfefferApi } from '../api/gameApi';
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
import type { HasenpfefferResponse } from '../types/card';
import { HasenpfefferPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { HASENPFEFFER_HELP, parseHasenpfefferCommand } from '../utils/cli/commands/hasenpfefferCommands';
import { formatHasenpfefferState } from '../utils/cli/formatters/hasenpfefferFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the trump buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/** Highest bid the server accepts (sync: domain.HasenpfefferMaxBid). */
const BID_MAX = 6;

/** Guided tutorial steps (the joker, compulsory bidding, the blind, your hand). */
const HASENPFEFFER_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="hpf-rule"]', messageKey: 'tutorial.joker', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="hpf-actions"]', messageKey: 'tutorial.bidding', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="hpf-seats"]', messageKey: 'tutorial.blind', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="hpf-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Hasenpfeffer page (wrapped by `withTutorial`).
 *
 * Renders the euchre variant carried to America by German immigrants: 25 cards
 * (the 24-card euchre pack plus a joker) dealt six each with one card left
 * face down, and six tricks per hand.
 *
 * **The joker is the highest trump of all** — the Best Bower, above both the
 * right and left bowers — so the page states the ranking outright rather than
 * leaving a player to discover it by losing to it. **Bidding is compulsory**:
 * once three players pass, the dealer cannot, so the page hides the pass button
 * rather than offering a move the server will refuse.
 */
function HasenpfefferPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('hasenpfeffer');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<HasenpfefferResponse, Parameters<typeof hasenpfefferApi.exec>>(hasenpfefferApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('hasenpfeffer', state);
  const [picked, setPicked] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('hasenpfeffer');
  const cliConfig: CliGameConfig<HasenpfefferResponse, Parameters<typeof hasenpfefferApi.exec>> = useMemo(
    () => ({
      gameName: 'hasenpfeffer',
      parseCommand: parseHasenpfefferCommand,
      formatResponse: formatHasenpfefferState,
      helpText: HASENPFEFFER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    setPicked(null);
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handleBid = useCallback(
    (bid: number) => {
      void dispatch('bid', undefined, undefined, undefined, bid);
    },
    [dispatch],
  );

  const handleDiscard = useCallback(
    (suit: number) => {
      if (picked === null) return;
      const idx = picked;
      setPicked(null);
      void dispatch('discard', idx, undefined, suit);
    },
    [dispatch, picked],
  );

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNextHand = useCallback(() => {
    setPicked(null);
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return (
      <GameSkeleton gameKey="hasenpfeffer" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 6 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === HasenpfefferPhase.BID;
  const isDiscard = state.phase === HasenpfefferPhase.DISCARD;
  const isHandEnd = state.phase === HasenpfefferPhase.HAND_END;
  const isGameEnd = state.phase === HasenpfefferPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBid && !isGameEnd && state.currentPlayerIdx === 0;
  const isHumanDiscardTurn = isDiscard && !isGameEnd && state.declarerIdx === 0;
  const isHumanTurn =
    !isGameEnd && !isHandEnd && !isBid && !isDiscard && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isHandEnd
      ? t('phase.handEnd')
      : isBid
        ? t('phase.bid')
        : isDiscard
          ? t('phase.discard')
          : t('phase.play');

  // **サーバが必ず拒否する額は出さない (#5304)。** minBid が 0 なら宣言できない。
  const bidChoices: number[] = [];
  if (isHumanBidTurn && state.minBid > 0) {
    for (let n = state.minBid; n <= BID_MAX; n++) bidChoices.push(n);
  }

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
      title={tc('nav.hasenpfeffer')}
      gameThemeBg={gameTheme.hasenpfeffer.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/hasenpfeffer"
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
              <span className="mr-4" data-testid="hpf-hand">
                {t('header.hand')}: {state.handNumber}
              </span>
              <span className="mr-4">{t('header.target', { target: String(state.config.target) })}</span>
              <span data-testid="hpf-trump">
                {state.trumpSuit > 0
                  ? t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })
                  : state.blindSize > 0
                    ? t('header.blind', { n: String(state.blindSize) })
                    : t('header.trumpUndecided')}
              </span>
            </div>

            {/* **ジョーカーが全カード中最強。** 知らないと打ち方が変わる。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="hpf-rule"
              data-tutorial="hpf-rule"
            >
              {t('header.rule')}
            </div>

            <div className="text-ds-text-muted text-sm text-center mb-3" data-testid="hpf-score">
              {t('header.score', { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) })}
            </div>

            {/* **親は降りられないことがある。** 選択肢が無い場面を言う。 */}
            {isHumanBidTurn && state.mustBid && (
              <div className="text-center mb-3 text-ds-accent text-sm" role="status" data-testid="hpf-must-bid">
                {t('header.mustBid')}
              </div>
            )}

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="hpf-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`hpf-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  <span className="ml-1 text-ds-accent">{t('header.team', { team: String(p.team) })}</span>
                  {p.id === state.declarerIdx && <span className="ml-1 text-ds-accent">{t('header.declarer')}</span>}
                  {': '}
                  {p.bid < 0
                    ? t('header.bidNone')
                    : p.bid === 0
                      ? t('header.bidPassed')
                      : t('header.bidValue', { n: String(p.bid) })}
                  {' / '}
                  {t('header.took', { n: String(p.trickCount) })}
                </div>
              ))}
            </div>

            <div data-tutorial="hpf-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {/* **落としたのか達成したのかは盤面から読めない。** */}
            {isHandEnd && (
              <div className="text-center my-3 text-ds-accent" role="status" data-testid="hpf-hand-result">
                {state.lastHandEuchred
                  ? t('handEnd.euchred', { contract: String(state.contract), took: String(state.lastHandTricks) })
                  : t('handEnd.made', { contract: String(state.contract), took: String(state.lastHandTricks) })}
              </div>
            )}

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="hpf-result"
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

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="hpf-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => (isHumanDiscardTurn ? setPicked(idx) : handlePlay(idx))}
                      disabled={loading || (!isHumanTurn && !isHumanDiscardTurn)}
                      aria-pressed={isHumanDiscardTurn ? picked === idx : undefined}
                      aria-label={
                        isHumanDiscardTurn
                          ? t('actions.discardAria', { card: cardAlt(card) })
                          : t('actions.playAria', { card: cardAlt(card) })
                      }
                      className={`disabled:opacity-50 ${
                        picked === idx
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="hpf-actions">
              {bidChoices.map((n) => (
                <button
                  key={n}
                  type="button"
                  className={btnWarning}
                  onClick={() => handleBid(n)}
                  disabled={loading}
                  data-testid={`hpf-bid-${n.toString()}-btn`}
                >
                  {t('actions.bid', { n: String(n) })}
                </button>
              ))}
              {/* **親が降りられない場面では降りるボタンを出さない。** */}
              {isHumanBidTurn && !state.mustBid && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => handleBid(0)}
                  disabled={loading}
                  data-testid="hpf-pass-btn"
                >
                  {t('actions.pass')}
                </button>
              )}
              {isHumanDiscardTurn &&
                SUITS.map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnWarning}
                    onClick={() => handleDiscard(suit)}
                    disabled={loading || picked === null}
                    aria-disabled={picked === null}
                    data-testid={`hpf-discard-${suit.toString()}-btn`}
                  >
                    {t('actions.discard', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
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

/** Hasenpfeffer page wrapped with TutorialProvider. */
export const HasenpfefferPage = withTutorial(HasenpfefferPageContent, 'hasenpfeffer', HASENPFEFFER_TUTORIAL_STEPS);
