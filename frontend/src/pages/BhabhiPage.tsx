import { useCallback, useEffect, useMemo } from 'react';
import { bhabhiApi } from '../api/gameApi';
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
import { btnDanger, btnPrimary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BhabhiResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BHABHI_HELP, parseBhabhiCommand } from '../utils/cli/commands/bhabhiCommands';
import { formatBhabhiState } from '../utils/cli/formatters/bhabhiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit code (1=♠ 2=♣ 3=♥ 4=♦) to its symbol. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Table sizes the game accepts (sync: domain.BhabhiMin/MaxPlayers). */
const PLAYER_COUNTS: readonly number[] = [3, 4, 5, 6, 7];

/** Guided tutorial steps (the goal, the pile, getting out, your hand). */
const BHABHI_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bh-goal"]', messageKey: 'tutorial.goal', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bh-pile"]', messageKey: 'tutorial.pile', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bh-seats"]', messageKey: 'tutorial.out', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bh-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Bhabhi page (wrapped by `withTutorial`).
 *
 * Renders India and Pakistan's household avoidance game: the whole 52-card
 * deck dealt among three to seven players, no trumps and no scoring.
 *
 * **The game names a loser, not a winner.** Players who empty their hand are
 * simply safe; the one still holding cards at the end is the Bhabhi. Failing
 * to follow the led suit means taking **every card on the table** into your
 * hand, so the page leads with the pile size rather than with a trick count —
 * the pile is the only number that can hurt you.
 */
function BhabhiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bhabhi');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<BhabhiResponse, Parameters<typeof bhabhiApi.exec>>(bhabhiApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('bhabhi', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bhabhi');
  const cliConfig: CliGameConfig<BhabhiResponse, Parameters<typeof bhabhiApi.exec>> = useMemo(
    () => ({
      gameName: 'bhabhi',
      parseCommand: parseBhabhiCommand,
      formatResponse: formatBhabhiState,
      helpText: BHABHI_HELP,
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

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  const handlePlayerCnt = useCallback(
    (value: string) => {
      hideActionLog();
      void dispatch('reset', undefined, { playerCnt: Number(value) });
    },
    [dispatch, hideActionLog],
  );

  if (!state) {
    return <GameSkeleton gameKey="bhabhi" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd ? t('phase.gameEnd') : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const name =
      state.bhabhiIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(Math.max(state.bhabhiIdx, 0)) });
    // **膠着で終わったことは盤面から読めない。** 別の文言で言う。
    if (state.stalemate) return t('result.stalemate', { name, tricks: String(state.trickNumber) });
    return state.bhabhiIdx === 0 ? t('result.you') : t('result.cpu', { name });
  })();

  return (
    <GamePageShell
      title={tc('nav.bhabhi')}
      gameThemeBg={gameTheme.bhabhi.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/bhabhi"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.bhabhiIdx !== 0}
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
            {/* **勝者ではなく敗者を決めるゲーム。** 目的を先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="bh-goal"
              data-tutorial="bh-goal"
            >
              {t('header.goal')}
            </div>

            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4" data-testid="bh-trick">
                {t('header.trick')}: {state.trickNumber + 1}
              </span>
              <span data-testid="bh-alive">
                {t('header.alive', { n: String(state.aliveCount), total: String(state.players.length) })}
              </span>
            </div>

            {/* **場札の枚数が、フォローできなかったときの罰そのもの。** */}
            <div className="text-ds-text-muted text-sm text-center mb-3" data-testid="bh-pile" data-tutorial="bh-pile">
              {state.leadSuit > 0
                ? t('header.led', { suit: SUIT_SYMBOLS[state.leadSuit] ?? '?', n: String(state.pile.length) })
                : t('header.noLead')}
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="bh-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`bh-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {': '}
                  {/* **順位は上がった順であって強さではない。** */}
                  {p.rank > 0 ? (
                    <span className="text-ds-accent">{t('header.out', { rank: String(p.rank) })}</span>
                  ) : (
                    t('header.cards', { n: String(p.cardCount) })
                  )}
                  {' / '}
                  {t('header.pickups', { n: String(p.pickups) })}
                </div>
              ))}
            </div>

            <div data-tutorial="bh-trick">
              <TrickDisplay currentTrick={state.pile} players={state.players} cardWidth={cardWidth} label={t('pile')} />
            </div>

            {/* **直前の引き取りは盤面に痕跡が残らない。** */}
            {state.lastPickupIdx >= 0 && !isGameEnd && (
              <div className="text-center my-2 text-ds-accent text-sm" role="status" data-testid="bh-last-pickup">
                {t('lastPickup', {
                  name:
                    state.lastPickupIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(state.lastPickupIdx) }),
                  n: String(state.lastPickupSize),
                })}
              </div>
            )}

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="bh-result"
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
              <div className="mt-4" data-tutorial="bh-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="bh-actions">
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
                    hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                    {
                      type: 'select',
                      id: 'bh-player-cnt',
                      label: t('settings.playerCnt'),
                      tooltip: t('settings.playerCntHelp'),
                      value: state.config.playerCnt,
                      options: PLAYER_COUNTS.map((n) => ({ value: n, label: String(n) })),
                      onSelect: handlePlayerCnt,
                      disabled: loading,
                      testId: 'bh-player-cnt',
                    },
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

/** Bhabhi page wrapped with TutorialProvider. */
export const BhabhiPage = withTutorial(BhabhiPageContent, 'bhabhi', BHABHI_TUTORIAL_STEPS);
