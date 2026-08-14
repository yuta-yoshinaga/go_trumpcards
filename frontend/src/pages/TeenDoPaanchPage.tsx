import { useCallback, useEffect, useMemo } from 'react';
import { teendopaanchApi } from '../api/gameApi';
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
import type { TeenDoPaanchResponse } from '../types/card';
import { TeenDoPaanchPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTeenDoPaanchCommand, TEENDOPAANCH_HELP } from '../utils/cli/commands/teendopaanchCommands';
import { formatTeenDoPaanchState } from '../utils/cli/formatters/teendopaanchFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the trump buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/** Guided tutorial steps (the assigned targets, trump, the exchange, your hand). */
const TEENDOPAANCH_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="td-targets"]', messageKey: 'tutorial.targets', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="td-trump"]', messageKey: 'tutorial.trump', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="td-seats"]', messageKey: 'tutorial.exchange', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="td-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the 3-2-5 page (wrapped by `withTutorial`).
 *
 * Renders northern India's three-handed trick-taker: a 30-card pack (eights
 * through aces plus the ♠7 and ♥7) dealt ten each, for exactly ten tricks.
 *
 * **The targets are assigned, not bid.** The three seats owe 3, 2 and 5 tricks
 * and the roles rotate, so the page leads with who owes what rather than with
 * a score. **Taking extra tricks scores nothing** — the surplus instead buys
 * that many of a short player's best cards next round, which nothing on the
 * board would otherwise reveal.
 */
function TeenDoPaanchPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('teendopaanch');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<TeenDoPaanchResponse, Parameters<typeof teendopaanchApi.exec>>(teendopaanchApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('teendopaanch', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('teendopaanch');
  const cliConfig: CliGameConfig<TeenDoPaanchResponse, Parameters<typeof teendopaanchApi.exec>> = useMemo(
    () => ({
      gameName: 'teendopaanch',
      parseCommand: parseTeenDoPaanchCommand,
      formatResponse: formatTeenDoPaanchState,
      helpText: TEENDOPAANCH_HELP,
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

  const handleNextRound = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return (
      <GameSkeleton gameKey="teendopaanch" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 10 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isTrump = state.phase === TeenDoPaanchPhase.TRUMP;
  const isRoundEnd = state.phase === TeenDoPaanchPhase.ROUND_END;
  const isGameEnd = state.phase === TeenDoPaanchPhase.GAME_END || state.gameEndFlag;
  const isHumanTrumpTurn = isTrump && !isGameEnd && state.fivePlayerIdx === 0;
  const isHumanTurn = !isGameEnd && !isRoundEnd && !isTrump && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isTrump
        ? t('phase.trump')
        : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx === 0) return t('result.you');
    if (state.winnerIdx < 0) return t('result.tie');
    return t('result.cpu', { name: t('header.cpu', { idx: String(state.winnerIdx) }) });
  })();

  return (
    <GamePageShell
      title={tc('nav.teendopaanch')}
      gameThemeBg={gameTheme.teendopaanch.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/teendopaanch"
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
              <span className="mr-4" data-testid="td-round">
                {t('header.round', { round: String(state.roundNumber), total: String(state.config.rounds) })}
              </span>
              <span data-testid="td-trump" data-tutorial="td-trump">
                {state.trumpSuit > 0
                  ? t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })
                  : t('header.trumpUndecided')}
              </span>
            </div>

            {/* **ノルマは宣言ではなく割り当て。** 何を負っているかを先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="td-targets"
              data-tutorial="td-targets"
            >
              {t('header.targets')}
            </div>

            {/* **前ラウンドの札のやり取りは盤面に痕跡が残らない。** */}
            {state.lastExchange > 0 && (
              <div className="text-center mb-3 text-ds-accent text-sm" role="status" data-testid="td-exchange">
                {t('header.exchange', { n: String(state.lastExchange) })}
              </div>
            )}

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="td-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`td-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.id === state.fivePlayerIdx && <span className="ml-1 text-ds-accent">{t('header.declarer')}</span>}
                  {': '}
                  {/* **あと何トリック要るかが読めないと打ち方が決まらない。** */}
                  <span className="text-ds-accent">
                    {t('header.target', { target: String(p.target), took: String(p.trickCount) })}
                  </span>
                  {' / '}
                  {t('header.met', { n: String(p.met) })}
                </div>
              ))}
            </div>

            <div data-tutorial="td-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="td-result"
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
              <div className="mt-4" data-tutorial="td-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="td-actions">
              {isHumanTrumpTurn &&
                SUITS.map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnWarning}
                    onClick={() => handleTrump(suit)}
                    disabled={loading}
                    data-testid={`td-trump-${suit.toString()}-btn`}
                  >
                    {t('actions.trump', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
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

/** 3-2-5 page wrapped with TutorialProvider. */
export const TeenDoPaanchPage = withTutorial(TeenDoPaanchPageContent, 'teendopaanch', TEENDOPAANCH_TUTORIAL_STEPS);
