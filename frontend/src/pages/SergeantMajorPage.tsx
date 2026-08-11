import { useCallback, useEffect, useMemo, useState } from 'react';
import { sergeantmajorApi } from '../api/gameApi';
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
import type { SergeantMajorResponse } from '../types/card';
import { SergeantMajorPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSergeantMajorCommand, SERGEANTMAJOR_HELP } from '../utils/cli/commands/sergeantmajorCommands';
import { formatSergeantMajorState } from '../utils/cli/formatters/sergeantmajorFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The four suits, in the order the trump buttons are offered. */
const SUITS: readonly number[] = [1, 2, 3, 4];

/** Guided tutorial steps (the fixed targets, the kitty, the exchange, your hand). */
const SERGEANTMAJOR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sm-rule"]', messageKey: 'tutorial.targets', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sm-trump"]', messageKey: 'tutorial.kitty', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sm-seats"]', messageKey: 'tutorial.exchange', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sm-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Sergeant Major (8-5-3) page (wrapped by `withTutorial`).
 *
 * Renders the British army's three-handed trick-taker: 52 cards dealt sixteen
 * each, with **the four left over going to the dealer as a kitty** — 52 does
 * not divide by three, which is exactly why the kitty exists.
 *
 * **The targets are fixed by seat, never bid**: the dealer owes 8, the seat to
 * their left 5, the next 3, and 8+5+3 = 16 matches the trick count. Every
 * trick past your target scores and every one short is paid for next round in
 * your best cards, so the page shows target against tricks taken rather than a
 * bare score.
 */
function SergeantMajorPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sergeantmajor');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<SergeantMajorResponse, Parameters<typeof sergeantmajorApi.exec>>(sergeantmajorApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('sergeantmajor', state);
  const [picked, setPicked] = useState<number[]>([]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sergeantmajor');
  const cliConfig: CliGameConfig<SergeantMajorResponse, Parameters<typeof sergeantmajorApi.exec>> = useMemo(
    () => ({
      gameName: 'sergeantmajor',
      parseCommand: parseSergeantMajorCommand,
      formatResponse: formatSergeantMajorState,
      helpText: SERGEANTMAJOR_HELP,
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

  const handleTrump = useCallback(
    (suit: number) => {
      void dispatch('trump', undefined, undefined, suit);
    },
    [dispatch],
  );

  const handleDiscard = useCallback(() => {
    const indices = picked;
    setPicked([]);
    void dispatch('discard', undefined, undefined, undefined, indices);
  }, [dispatch, picked]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNextRound = useCallback(() => {
    setPicked([]);
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return (
      <GameSkeleton gameKey="sergeantmajor" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 16 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isTrump = state.phase === SergeantMajorPhase.TRUMP;
  const isDiscard = state.phase === SergeantMajorPhase.DISCARD;
  const isRoundEnd = state.phase === SergeantMajorPhase.ROUND_END;
  const isGameEnd = state.phase === SergeantMajorPhase.GAME_END || state.gameEndFlag;
  const isHumanTrumpTurn = isTrump && !isGameEnd && state.dealerIdx === 0;
  const isHumanDiscardTurn = isDiscard && !isGameEnd && state.dealerIdx === 0;
  const isHumanTurn =
    !isGameEnd && !isRoundEnd && !isTrump && !isDiscard && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isTrump
        ? t('phase.trump')
        : isDiscard
          ? t('phase.discard')
          : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const togglePicked = (idx: number) => {
    setPicked((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx === 0) return t('result.you');
    if (state.winnerIdx < 0) return t('result.tie');
    return t('result.cpu', { name: t('header.cpu', { idx: String(state.winnerIdx) }) });
  })();

  return (
    <GamePageShell
      title={tc('nav.sergeantmajor')}
      gameThemeBg={gameTheme.sergeantmajor.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/sergeantmajor"
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
              <span className="mr-4" data-testid="sm-round">
                {t('header.round', { round: String(state.roundNumber), total: String(state.config.rounds) })}
              </span>
              <span data-testid="sm-trump" data-tutorial="sm-trump">
                {state.trumpSuit > 0
                  ? t('header.trump', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })
                  : t('header.trumpUndecided', { kitty: String(state.kittySize) })}
              </span>
            </div>

            {/* **ノルマは席順で決まる。** 規則を先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="sm-rule"
              data-tutorial="sm-rule"
            >
              {t('header.rule')}
            </div>

            {/* **前ラウンドの札のやり取りは盤面に痕跡が残らない。** */}
            {state.lastExchange > 0 && (
              <div className="text-center mb-3 text-ds-accent text-sm" role="status" data-testid="sm-exchange">
                {t('header.exchange', { n: String(state.lastExchange) })}
              </div>
            )}

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="sm-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`sm-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.id === state.dealerIdx && <span className="ml-1 text-ds-accent">{t('header.dealer')}</span>}
                  {': '}
                  <span className="text-ds-accent">
                    {t('header.target', { target: String(p.target), took: String(p.trickCount) })}
                  </span>
                  {' / '}
                  {t('header.score', { n: String(p.score) })}
                </div>
              ))}
            </div>

            <div data-tutorial="sm-trick">
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
                data-testid="sm-result"
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
              <div className="mt-4" data-tutorial="sm-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => (isHumanDiscardTurn ? togglePicked(idx) : handlePlay(idx))}
                      disabled={loading || (!isHumanTurn && !isHumanDiscardTurn)}
                      aria-pressed={isHumanDiscardTurn ? picked.includes(idx) : undefined}
                      aria-label={
                        isHumanDiscardTurn
                          ? t('actions.discardAria', { card: cardAlt(card) })
                          : t('actions.playAria', { card: cardAlt(card) })
                      }
                      className={`disabled:opacity-50 ${
                        picked.includes(idx)
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="sm-actions">
              {isHumanTrumpTurn &&
                SUITS.map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnWarning}
                    onClick={() => handleTrump(suit)}
                    disabled={loading}
                    data-testid={`sm-trump-${suit.toString()}-btn`}
                  >
                    {t('actions.trump', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
                  </button>
                ))}
              {/* **ちょうど 4 枚選ぶまで確定できない。** サーバが必ず拒否する操作は出さない。 */}
              {isHumanDiscardTurn && (
                <button
                  type="button"
                  className={btnWarning}
                  onClick={handleDiscard}
                  disabled={loading || picked.length !== state.discardCount}
                  aria-disabled={picked.length !== state.discardCount}
                  data-testid="sm-discard-btn"
                >
                  {t('actions.discard', { n: String(picked.length), total: String(state.discardCount) })}
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

/** Sergeant Major page wrapped with TutorialProvider. */
export const SergeantMajorPage = withTutorial(SergeantMajorPageContent, 'sergeantmajor', SERGEANTMAJOR_TUTORIAL_STEPS);
