import { useCallback, useEffect, useMemo } from 'react';
import { balootApi } from '../api/gameApi';
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
import type { BalootResponse } from '../types/card';
import { BalootMode, BalootPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BALOOT_HELP, parseBalootCommand } from '../utils/cli/commands/balootCommands';
import { formatBalootState } from '../utils/cli/formatters/balootFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tricks per round (eight cards each). */
const TRICKS_PER_ROUND = 8;

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol, for the trump readout. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the Hokom buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/** Guided tutorial steps (the two orders, declaring, Baloot, hand). */
const BALOOT_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="baloot-order"]', messageKey: 'tutorial.order', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="baloot-declare"]', messageKey: 'tutorial.declare', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="baloot-seats"]', messageKey: 'tutorial.baloot', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="baloot-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Baloot page (wrapped by `withTutorial`).
 *
 * Renders the Gulf's most-played trick-taker: 32 cards, four players in two
 * partnerships with the seats opposite each other allied, eight cards each.
 *
 * The thing that cannot be read off the table is that **the rank order itself
 * changes with the declared mode** — Sun runs A>10>K>Q>J>9>8>7 with no trump,
 * Hokom gives the chosen trump suit J>9>A>10>K>Q>8>7 — so the page prints
 * whichever order is in force rather than both. The pass button is hidden when
 * the player is the dealer, because the dealer cannot pass.
 */
function BalootPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('baloot');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<BalootResponse, Parameters<typeof balootApi.exec>>(balootApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('baloot', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('baloot');
  const cliConfig: CliGameConfig<BalootResponse, Parameters<typeof balootApi.exec>> = useMemo(
    () => ({
      gameName: 'baloot',
      parseCommand: parseBalootCommand,
      formatResponse: formatBalootState,
      helpText: BALOOT_HELP,
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

  const handleSun = useCallback(() => {
    void dispatch('sun');
  }, [dispatch]);

  const handleHokom = useCallback(
    (suit: number) => {
      void dispatch('hokom', undefined, undefined, suit);
    },
    [dispatch],
  );

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
    return <GameSkeleton gameKey="baloot" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isDeclare = state.phase === BalootPhase.DECLARE;
  const isRoundEnd = state.phase === BalootPhase.ROUND_END;
  const isGameEnd = state.phase === BalootPhase.GAME_END || state.gameEndFlag;
  const isHumanDeclareTurn = isDeclare && !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanTurn =
    !isGameEnd && !isRoundEnd && !isDeclare && state.players[state.currentPlayerIdx]?.isHuman === true;
  // **親は見送れない。** 押せないボタンを出すより、出さないほうが正直。
  const humanIsDealer = state.dealerIdx === 0;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isDeclare
        ? t('phase.declare')
        : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const declarerName = state.declarerIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(state.declarerIdx) });

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const params = { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) };
    if (state.winnerTeam === 0) return t('result.youWin', params);
    if (state.winnerTeam === 1) return t('result.theyWin', params);
    return t('result.tie', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.baloot')}
      gameThemeBg={gameTheme.baloot.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/baloot"
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
              <span className="mr-4" data-testid="bl-round">
                {t('header.round')}: {state.roundNumber}
              </span>
              <span className="mr-4" data-testid="bl-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
              <span>{t('header.target', { target: String(state.config.target) })}</span>
            </div>

            <div className="text-ds-text-primary text-center mb-2" data-testid="bl-score">
              {t('header.score', { t0: String(state.scores[0] ?? 0), t1: String(state.scores[1] ?? 0) })}
            </div>

            {/* **有効な序列だけを出す。** モードで入れ替わるので、両方出すと
                どちらが効いているのか読み取れない。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
              data-testid="bl-order"
              data-tutorial="baloot-order"
            >
              {state.mode === BalootMode.SUN && (
                <>
                  <div className="text-ds-text-primary" data-testid="bl-mode">
                    {t('header.modeSun', { name: declarerName })}
                  </div>
                  <div>{t('header.orderSun')}</div>
                </>
              )}
              {state.mode === BalootMode.HOKOM && (
                <>
                  <div className="text-ds-text-primary" data-testid="bl-mode">
                    {t('header.modeHokom', {
                      suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?',
                      name: declarerName,
                    })}
                  </div>
                  <div>{t('header.orderHokom')}</div>
                </>
              )}
              {state.mode === BalootMode.NONE && <div data-testid="bl-mode">{t('header.modeUndecided')}</div>}
            </div>

            {/* Baloot（切り札の K+Q）は Hokom のときだけ成立する。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="baloot-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`bl-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  <span className="ml-1 text-ds-accent">{t('header.team', { team: String(p.team) })}</span>
                  {': '}
                  {/* **配られた瞬間に相手の手の内が割れるのは体験を壊す** (#5750)。
                      切り札の K か Q が実際に出る (かラウンドが終わる) まで伏せる。 */}
                  <span data-testid={`bl-baloot-${p.id.toString()}`}>
                    {p.balootRevealed === false
                      ? t('baloot.hidden')
                      : p.hasBaloot
                        ? t('baloot.held')
                        : t('baloot.none')}
                  </span>
                </div>
              ))}
            </div>

            <div data-tutorial="baloot-trick">
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
              <div className="mt-4" data-tutorial="baloot-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="baloot-declare">
              {isHumanDeclareTurn && (
                <>
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={handleSun}
                    disabled={loading}
                    data-testid="bl-sun-btn"
                  >
                    {t('actions.sun')}
                  </button>
                  {/* **Hokom はスートまで選ばせる。** モードだけでは序列が決まらない。 */}
                  {SUITS.map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnWarning}
                      onClick={() => handleHokom(suit)}
                      disabled={loading}
                      data-testid={`bl-hokom-${suit.toString()}-btn`}
                    >
                      {t('actions.hokom', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
                    </button>
                  ))}
                  {/* **親は見送れない。** 押せないボタンを出さない。 */}
                  {humanIsDealer ? (
                    <span className="self-center text-ds-text-muted text-sm" data-testid="bl-dealer-stuck">
                      {t('dealerStuck')}
                    </span>
                  ) : (
                    <button
                      type="button"
                      className={btnSuccess}
                      onClick={handlePass}
                      disabled={loading}
                      data-testid="bl-pass-btn"
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

/** Baloot page wrapped with TutorialProvider. */
export const BalootPage = withTutorial(BalootPageContent, 'baloot', BALOOT_TUTORIAL_STEPS);
