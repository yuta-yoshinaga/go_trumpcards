import { useCallback, useEffect, useMemo, useState } from 'react';
import { rollingstoneApi } from '../api/gameApi';
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
import { btnDanger, btnPrimary, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { RollingStoneResponse } from '../types/card';
import { RollingStonePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseRollingStoneCommand, ROLLINGSTONE_HELP } from '../utils/cli/commands/rollingstoneCommands';
import { formatRollingStoneState } from '../utils/cli/formatters/rollingstoneFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guided tutorial steps (the inverted goal, the penalty, the deck, your hand). */
const ROLLINGSTONE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="rs-rule"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="rs-seats"]', messageKey: 'tutorial.penalty', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="rs-deck"]', messageKey: 'tutorial.deck', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="rs-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Rolling Stone page (wrapped by `withTutorial`).
 *
 * The goal runs backwards: **a trick is worth nothing but the lead, and
 * failing to follow suit puts the whole trick into your hand.** So the page
 * leads with hand sizes rather than any score, and reads `mustPickUp` before
 * `validPlays` — an empty list of legal plays also means "not your turn", and
 * the two call for opposite things.
 */
function RollingStonePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('rollingstone');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<RollingStoneResponse, Parameters<typeof rollingstoneApi.exec>>(rollingstoneApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('rollingstone', state);
  const [playerCnt, setPlayerCnt] = useState(4);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('rollingstone');
  const cliConfig: CliGameConfig<RollingStoneResponse, Parameters<typeof rollingstoneApi.exec>> = useMemo(
    () => ({
      gameName: 'rollingstone',
      parseCommand: parseRollingStoneCommand,
      formatResponse: formatRollingStoneState,
      helpText: ROLLINGSTONE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', undefined, { playerCnt });
  }, [dispatch, hideActionLog, playerCnt]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handlePickUp = useCallback(() => {
    void dispatch('pickup');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return (
      <GameSkeleton gameKey="rollingstone" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === RollingStonePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;
  // **引き取りが要る局面は、出せる札が無い局面。** 空の validPlays と区別する。
  const mustPickUp = isHumanTurn && state.mustPickUp;

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn && !mustPickUp ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd || state.winnerIdx < 0) return null;
    const winner = state.players[state.winnerIdx];
    const name = state.winnerIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(state.winnerIdx) });
    // **上限で切った局は「上がった」わけではない。** 言い分ける。
    if (winner && winner.cardCount > 0) {
      return t('result.stalemate', { name, n: String(winner.cardCount) });
    }
    return state.winnerIdx === 0 ? t('result.you') : t('result.cpu', { name });
  })();

  return (
    <GamePageShell
      title={tc('nav.rollingstone')}
      gameThemeBg={gameTheme.rollingstone.bg}
      phaseName={isGameEnd ? t('phase.gameEnd') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/rollingstone"
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
            <div className="text-ds-text-primary text-center mb-2" data-testid="rs-deck" data-tutorial="rs-deck">
              <span className="mr-4">{t('header.trick', { n: String(state.trickNumber + 1) })}</span>
              <span>
                {t('header.deck', {
                  deck: String(state.deckSize),
                  left: String(state.deckSize - state.discarded),
                })}
              </span>
            </div>

            {/* **勝利条件が逆さまなのが規則そのもの。** 先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="rs-rule"
              data-tutorial="rs-rule"
            >
              {t('header.rule')}
            </div>

            {/* **手札の枚数がそのまま順位。** 得点表示は無い。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="rs-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`rs-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.finishedAt > 0 && (
                    <span className="ml-1 text-ds-accent">{t('header.finished', { rank: String(p.finishedAt) })}</span>
                  )}
                  {p.finishedAt === 0 && p.id === state.lastPickupIdx && (
                    <span className="ml-1 text-ds-warning">{t('header.pickedUp')}</span>
                  )}
                  {': '}
                  <span className="text-ds-accent">{t('header.cards', { n: String(p.cardCount) })}</span>
                  {' / '}
                  {t('header.pickups', { n: String(p.pickups) })}
                </div>
              ))}
            </div>

            <div data-tutorial="rs-trick">
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
                data-testid="rs-result"
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

            {/* **出せる札が無いことははっきり言う。** 黙っていると打てない理由が分からない。 */}
            {mustPickUp && (
              <div className="mt-3 text-center text-ds-warning" role="status" data-testid="rs-must-pickup">
                {t('header.mustPickUp', { n: String(state.currentTrick.length) })}
              </div>
            )}

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="rs-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => handlePlay(idx)}
                      disabled={loading || !isHumanTurn || mustPickUp}
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="rs-actions">
              {mustPickUp && (
                <button
                  type="button"
                  className={btnWarning}
                  onClick={handlePickUp}
                  disabled={loading}
                  data-testid="rs-pickup-btn"
                >
                  {t('actions.pickUp', { n: String(state.currentTrick.length) })}
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
              groups={[
                {
                  items: [
                    {
                      type: 'select',
                      id: 'rollingstone-players',
                      label: t('actions.players'),
                      value: String(playerCnt),
                      options: [4, 5, 6].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setPlayerCnt(Number(v)),
                      testId: 'rs-players-select',
                    },
                    hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                  ],
                },
              ]}
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

/** Rolling Stone page wrapped with TutorialProvider. */
export const RollingStonePage = withTutorial(RollingStonePageContent, 'rollingstone', ROLLINGSTONE_TUTORIAL_STEPS);
