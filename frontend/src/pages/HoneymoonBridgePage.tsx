import { useCallback, useEffect, useMemo, useState } from 'react';
import { honeymoonbridgeApi } from '../api/gameApi';
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
import type { HoneymoonBridgeResponse } from '../types/card';
import { HoneymoonBridgePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { HONEYMOONBRIDGE_HELP, parseHoneymoonBridgeCommand } from '../utils/cli/commands/honeymoonbridgeCommands';
import { formatHoneymoonBridgeState } from '../utils/cli/formatters/honeymoonbridgeFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Contract denominations. **`0` is no-trump**, which is a bid, not a missing value. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 0: 'NT', 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The five denominations, weakest first — the order the bid buttons are offered. */
const DENOMINATIONS: readonly number[] = [1, 2, 3, 4, 0];

/** Highest contract level. Level n needs 6 + n tricks, so 7 is all thirteen. */
const MAX_LEVEL = 7;

/**
 * The auction's denomination order (sync: honeymoonBridgeSuitRank in
 * internal/domain/HoneymoonBridge.go).
 *
 * **No-trump ranks above diamonds**, which the 1..4 suit codes cannot express.
 */
const denominationRank = (suit: number): number => (suit === 0 ? 5 : suit);

/**
 * Whether the domain would accept this bid, given the lowest bid it reported.
 *
 * **Derived from `minBidLevel`/`minBidSuit` rather than rebuilt from the
 * contract**, so the page can never offer a value the server rejects.
 */
const outbids = (level: number, suit: number, minLevel: number, minSuit: number): boolean => {
  if (minLevel === 0) return false; // 7NT に張り付いている：通る宣言は無い
  if (level !== minLevel) return level > minLevel;
  return denominationRank(suit) >= denominationRank(minSuit);
};

/** Guided tutorial steps (the scoreless first half, the auction, the contract, your hand). */
const HONEYMOONBRIDGE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="hb-rule"]', messageKey: 'tutorial.draw', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="hb-contract"]', messageKey: 'tutorial.bid', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="hb-seats"]', messageKey: 'tutorial.contract', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="hb-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Honeymoon Bridge page (wrapped by `withTutorial`).
 *
 * Bridge for two. Thirteen cards each leaves 26 in the stock, and **the first
 * thirteen tricks score nothing** — they are played without trumps purely so
 * that the winner, then the loser, draws a card after each one. 13 x 2 = 26
 * empties the stock exactly and both hands are back to thirteen for the
 * auction, which is why the page labels that half rather than showing a score.
 */
function HoneymoonBridgePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('honeymoonbridge');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<HoneymoonBridgeResponse, Parameters<typeof honeymoonbridgeApi.exec>>(honeymoonbridgeApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('honeymoonbridge', state);
  const [level, setLevel] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('honeymoonbridge');
  const cliConfig: CliGameConfig<HoneymoonBridgeResponse, Parameters<typeof honeymoonbridgeApi.exec>> = useMemo(
    () => ({
      gameName: 'honeymoonbridge',
      parseCommand: parseHoneymoonBridgeCommand,
      formatResponse: formatHoneymoonBridgeState,
      helpText: HONEYMOONBRIDGE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    setLevel(null);
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handleBid = useCallback(
    (bidLevel: number, suit: number) => {
      setLevel(null);
      void dispatch('bid', undefined, undefined, bidLevel, suit);
    },
    [dispatch],
  );

  const handlePass = useCallback(() => {
    setLevel(null);
    void dispatch('pass');
  }, [dispatch]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNextRound = useCallback(() => {
    setLevel(null);
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return (
      <GameSkeleton gameKey="honeymoonbridge" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === HoneymoonBridgePhase.BID;
  const isDraw = state.phase === HoneymoonBridgePhase.DRAW;
  const isRoundEnd = state.phase === HoneymoonBridgePhase.ROUND_END;
  const isGameEnd = state.phase === HoneymoonBridgePhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBid && !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanTurn = !isGameEnd && !isRoundEnd && !isBid && state.players[state.currentPlayerIdx]?.isHuman === true;

  // **通る宣言が無ければレベルも選ばせない。** 選択肢は minBidLevel から上だけ。
  const selectedLevel = level ?? (state.minBidLevel > 0 ? state.minBidLevel : 0);
  const levels =
    state.minBidLevel > 0
      ? Array.from({ length: MAX_LEVEL - state.minBidLevel + 1 }, (_, i) => state.minBidLevel + i)
      : [];

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isBid
        ? t('phase.bid')
        : isDraw
          ? t('phase.draw')
          : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx === 0) return t('result.you');
    if (state.winnerIdx < 0) return t('result.tie');
    return t('result.cpu', { name: t('header.cpu', { idx: String(state.winnerIdx) }) });
  })();

  const roundResult = (() => {
    if (!isRoundEnd) return null;
    if (state.contractLevel === 0) return t('roundResult.passedOut');
    const params = { need: String(state.requiredTricks), took: String(state.lastTricks) };
    return state.lastMade ? t('roundResult.made', params) : t('roundResult.down', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.honeymoonbridge')}
      gameThemeBg={gameTheme.honeymoonbridge.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/honeymoonbridge"
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
              <span className="mr-4" data-testid="hb-round">
                {t('header.round', { round: String(state.roundNumber) })}
              </span>
              <span data-testid="hb-target">{t('header.target', { n: String(state.config.target) })}</span>
            </div>

            {/* **前半と後半で意味が変わる。** 規則を先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="hb-rule"
              data-tutorial="hb-rule"
            >
              {t('header.rule')}
            </div>

            {/* **前半のトリックは得点にならない。** 山札の残りだけが意味を持つ。 */}
            {isDraw && (
              <div className="text-center mb-3 text-ds-accent text-sm" role="status" data-testid="hb-stock">
                {t('header.stock', { n: String(state.stockSize) })}
              </div>
            )}

            <div
              className="text-center mb-3 text-ds-text-primary"
              data-testid="hb-contract"
              data-tutorial="hb-contract"
            >
              {state.contractLevel > 0
                ? t('header.contract', {
                    level: String(state.contractLevel),
                    suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?',
                    name:
                      state.declarerIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(state.declarerIdx) }),
                    need: String(state.requiredTricks),
                  })
                : t('header.contractUndecided')}
            </div>

            {/* **サーバが必ず拒否する値を出させない。** 通る最小の宣言を明示する。 */}
            {isHumanBidTurn && (
              <div className="text-center mb-3 text-ds-accent text-sm" role="status" data-testid="hb-minbid">
                {state.minBidLevel > 0
                  ? t('header.minBid', {
                      level: String(state.minBidLevel),
                      suit: SUIT_SYMBOLS[state.minBidSuit] ?? '?',
                    })
                  : t('header.minBidCapped')}
              </div>
            )}

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="hb-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`hb-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.id === state.declarerIdx && <span className="ml-1 text-ds-accent">{t('header.declarer')}</span>}
                  {': '}
                  <span className="text-ds-accent">
                    {p.bidLevel > 0
                      ? t('header.bid', { level: String(p.bidLevel), suit: SUIT_SYMBOLS[p.bidSuit] ?? '?' })
                      : t('header.noBid')}
                  </span>
                  {' / '}
                  {t('header.took', { n: String(p.trickCount) })}
                  {' / '}
                  {t('header.score', { n: String(p.score) })}
                </div>
              ))}
            </div>

            <div data-tutorial="hb-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {roundResult && (
              <div className="text-center my-3 text-ds-text-primary" role="status" data-testid="hb-round-result">
                {roundResult}
              </div>
            )}

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="hb-result"
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
              <div className="mt-4" data-tutorial="hb-hand">
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

            <div className="mt-4 flex flex-wrap gap-2 items-center" data-tutorial="hb-actions">
              {isHumanBidTurn && levels.length > 0 && (
                <>
                  <label className="text-ds-text-muted text-sm" htmlFor="hb-level">
                    {t('actions.level')}
                  </label>
                  {/* 折りたたみに入れない：閉じた details の中はクリックできない。 */}
                  <select
                    id="hb-level"
                    className="rounded bg-black/40 text-ds-text-primary px-2 py-1"
                    value={selectedLevel}
                    onChange={(e) => setLevel(Number(e.target.value))}
                    disabled={loading}
                    data-testid="hb-level-select"
                  >
                    {levels.map((lv) => (
                      <option key={lv} value={lv}>
                        {lv}
                      </option>
                    ))}
                  </select>
                  {DENOMINATIONS.map((suit) => {
                    const legal = outbids(selectedLevel, suit, state.minBidLevel, state.minBidSuit);
                    return (
                      <button
                        key={suit}
                        type="button"
                        className={btnWarning}
                        onClick={() => handleBid(selectedLevel, suit)}
                        disabled={loading || !legal}
                        aria-disabled={!legal}
                        data-testid={`hb-bid-${suit.toString()}-btn`}
                      >
                        {t('actions.bid', { level: String(selectedLevel), suit: SUIT_SYMBOLS[suit] ?? '?' })}
                      </button>
                    );
                  })}
                </>
              )}
              {isHumanBidTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePass}
                  disabled={loading}
                  data-testid="hb-pass-btn"
                >
                  {t('actions.pass')}
                </button>
              )}
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

/** Honeymoon Bridge page wrapped with TutorialProvider. */
export const HoneymoonBridgePage = withTutorial(
  HoneymoonBridgePageContent,
  'honeymoonbridge',
  HONEYMOONBRIDGE_TUTORIAL_STEPS,
);
