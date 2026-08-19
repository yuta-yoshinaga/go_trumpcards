import { useCallback, useEffect, useMemo } from 'react';
import { tarabishApi } from '../api/gameApi';
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
import { badgeInfoColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TarabishPlayer, TarabishResponse } from '../types/card';
import { TarabishPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTarabishCommand, TARABISH_HELP } from '../utils/cli/commands/tarabishCommands';
import { formatTarabishState } from '../utils/cli/formatters/tarabishFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { tarabishCardPoints } from '../utils/tarabishPoints';

/** Tricks per round (nine cards each). */
const TRICKS_PER_ROUND = 9;

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol, for the trump readout. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Guided tutorial steps (trump order, bidding, melds, hand). */
const TARABISH_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tarabish-order"]', messageKey: 'tutorial.order', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tarabish-bid"]', messageKey: 'tutorial.bid', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="tarabish-melds"]', messageKey: 'tutorial.meld', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tarabish-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Tarabish page (wrapped by `withTutorial`).
 *
 * Renders the Cape Breton Jass-family game: 36 cards, four players in two
 * partnerships with the seats opposite each other allied, nine cards each.
 *
 * Two things cannot be read off the table, so the page states both outright:
 * the **trump order**, where the J (Jass, 20) and 9 (Menel, 14) outrank the
 * ace in strength as well as points, and each seat's **meld**, which is
 * computed from the deal rather than declared. The pass button is hidden when
 * the player is the dealer, because the dealer cannot pass.
 */
function TarabishPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tarabish');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<TarabishResponse, Parameters<typeof tarabishApi.exec>>(tarabishApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('tarabish', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tarabish');
  const cliConfig: CliGameConfig<TarabishResponse, Parameters<typeof tarabishApi.exec>> = useMemo(
    () => ({
      gameName: 'tarabish',
      parseCommand: parseTarabishCommand,
      formatResponse: formatTarabishState,
      helpText: TARABISH_HELP,
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

  const handleTake = useCallback(() => {
    void dispatch('take');
  }, [dispatch]);

  const handlePass = useCallback(() => {
    void dispatch('pass');
  }, [dispatch]);

  const handleNextRound = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="tarabish" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === TarabishPhase.BID;
  const isRoundEnd = state.phase === TarabishPhase.ROUND_END;
  const isGameEnd = state.phase === TarabishPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBid && !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanTurn = !isGameEnd && !isRoundEnd && !isBid && state.players[state.currentPlayerIdx]?.isHuman === true;
  // **親は見送れない。** 押せないボタンを出すより、出さないほうが正直。
  const humanIsDealer = state.dealerIdx === 0;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isBid
        ? t('phase.bid')
        : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  /** A seat's meld, spelled out rather than as a bare number. */
  const meldStr = (p: TarabishPlayer): string => {
    if (p.meldPoints === 0) return t('meld.none');
    const parts = [p.runLen > 0 && t('meld.run', { len: String(p.runLen) }), p.hasBella && t('meld.bella')].filter(
      Boolean,
    );
    return t('meld.summary', { detail: parts.join('+'), points: String(p.meldPoints) });
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
      title={tc('nav.tarabish')}
      gameThemeBg={gameTheme.tarabish.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/tarabish"
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
              <span className="mr-4" data-testid="tb-round">
                {t('header.round')}: {state.roundNumber}
              </span>
              <span className="mr-4" data-testid="tb-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
              <span>{t('header.target', { target: String(state.config.target) })}</span>
            </div>

            <div className="text-ds-text-primary text-center mb-2" data-testid="tb-score">
              {t('header.score', { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) })}
            </div>

            {/* 切り札の序列はこの系統の肝。盤面からは読み取れない。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
              data-testid="tb-order"
              data-tutorial="tarabish-order"
            >
              {t('header.order')}
            </div>

            {/* 入札前は候補、決まったあとは切り札。 */}
            <div className="mb-3 flex flex-wrap justify-center items-center gap-3">
              {state.trumpTakerIdx >= 0 ? (
                <div className="rounded bg-black/30 px-3 py-2 text-ds-text-primary text-sm" data-testid="tb-trump">
                  {t('header.trumpTaken', {
                    suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?',
                    name:
                      state.trumpTakerIdx === 0
                        ? t('header.you')
                        : t('header.cpu', { idx: String(state.trumpTakerIdx) }),
                  })}
                </div>
              ) : (
                state.upCard && (
                  <div className="flex items-center gap-2 rounded bg-black/30 p-2" data-testid="tb-upcard">
                    <span className="text-ds-text-muted text-sm">{t('header.upCard')}</span>
                    <CardImage card={state.upCard} width={Math.round(cardWidth * 0.8)} />
                  </div>
                )
              )}
            </div>

            {/* メルドは自動判定。誰がいくら持っているかは盤面に出ない。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="tarabish-melds">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`tb-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  <span className="ml-1 text-ds-accent">{t('header.team', { team: String(p.team) })}</span>
                  {': '}
                  {meldStr(p)}
                </div>
              ))}
            </div>

            <div data-tutorial="tarabish-trick">
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
              <div className="mt-4" data-tutorial="tarabish-hand">
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
                      // **切り札だけ点数表が入れ替わる** (#5749)。同じ J でも
                      // 切り札なら 20 点、そうでなければ 2 点。暗算させると
                      // パートナーに寄せる札を間違える。
                      aria-label={
                        state.trumpSuit > 0
                          ? t('actions.playAriaWithPoints', {
                              card: cardAlt(card),
                              points: tarabishCardPoints(card, state.trumpSuit),
                            })
                          : t('actions.playAria', { card: cardAlt(card) })
                      }
                      className={`relative disabled:opacity-50 ${
                        legalRing.has(idx) ? 'rounded-lg ring-2 ring-ds-success' : ''
                      }`}
                    >
                      <CardImage card={card} width={cardWidth} />
                      {/* 切り札が決まるまでは点が定まらないので出さない。 */}
                      {state.trumpSuit > 0 && (
                        <span
                          data-testid={`tb-points-${idx.toString()}`}
                          aria-hidden="true"
                          className={`absolute top-0 right-0 rounded-bl px-1 text-[10px] leading-tight ${
                            tarabishCardPoints(card, state.trumpSuit) > 0 ? badgeWarningColors : badgeInfoColors
                          }`}
                        >
                          {tarabishCardPoints(card, state.trumpSuit)}
                        </span>
                      )}
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="tarabish-bid">
              {isHumanBidTurn && (
                <>
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={handleTake}
                    disabled={loading}
                    data-testid="tb-take-btn"
                  >
                    {t('actions.take')}
                  </button>
                  {/* **親は見送れない。** 押せないボタンを出さない。 */}
                  {humanIsDealer ? (
                    <span className="self-center text-ds-text-muted text-sm" data-testid="tb-dealer-stuck">
                      {t('dealerStuck')}
                    </span>
                  ) : (
                    <button
                      type="button"
                      className={btnSuccess}
                      onClick={handlePass}
                      disabled={loading}
                      data-testid="tb-pass-btn"
                    >
                      {t('actions.pass')}
                    </button>
                  )}
                </>
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

/** Tarabish page wrapped with TutorialProvider. */
export const TarabishPage = withTutorial(TarabishPageContent, 'tarabish', TARABISH_TUTORIAL_STEPS);
