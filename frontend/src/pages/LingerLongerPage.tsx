import { useCallback, useEffect, useMemo, useState } from 'react';
import { lingerlongerApi } from '../api/gameApi';
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
import type { LingerLongerResponse } from '../types/card';
import { LingerLongerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { LINGERLONGER_HELP, parseLingerLongerCommand } from '../utils/cli/commands/lingerlongerCommands';
import { formatLingerLongerState } from '../utils/cli/formatters/lingerlongerFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guided tutorial steps (the prize, the inverted goal, the stock, your hand). */
const LINGERLONGER_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ll-rule"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ll-seats"]', messageKey: 'tutorial.goal', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ll-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ll-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Linger Longer page (wrapped by `withTutorial`).
 *
 * **A trick is worth one card from the stock, and nothing else.** Because your
 * hand does not shrink while you keep winning, the page leads with hand sizes
 * and the stock count rather than any score — and it says so explicitly when
 * the stock empties, since that is the moment the game turns terminal.
 *
 * It also has to handle something no other game here does: **the human can be
 * eliminated while the game goes on.** A silent board with no prompt reads as a
 * bug, so an out player is told so directly.
 */
function LingerLongerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('lingerlonger');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<LingerLongerResponse, Parameters<typeof lingerlongerApi.exec>>(lingerlongerApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('lingerlonger', state);
  const [playerCnt, setPlayerCnt] = useState(4);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('lingerlonger');
  const cliConfig: CliGameConfig<LingerLongerResponse, Parameters<typeof lingerlongerApi.exec>> = useMemo(
    () => ({
      gameName: 'lingerlonger',
      parseCommand: parseLingerLongerCommand,
      formatResponse: formatLingerLongerState,
      helpText: LINGERLONGER_HELP,
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

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return (
      <GameSkeleton gameKey="lingerlonger" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === LingerLongerPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;
  // **脱落しても局は続く。** 手番が二度と来ないので、そうと言わないと固まって見える。
  const isEliminated = !isGameEnd && (human?.eliminatedAt ?? 0) > 0;

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd || state.winnerIdx < 0) return null;
    if (state.winnerIdx === 0) return t('result.you');
    return t('result.cpu', { name: t('header.cpu', { idx: String(state.winnerIdx) }) });
  })();

  return (
    <GamePageShell
      title={tc('nav.lingerlonger')}
      gameThemeBg={gameTheme.lingerlonger.bg}
      phaseName={isGameEnd ? t('phase.gameEnd') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/lingerlonger"
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
            <div className="text-ds-text-primary text-center mb-2" data-testid="ll-stock" data-tutorial="ll-stock">
              <span className="mr-4">{t('header.trick', { n: String(state.trickNumber + 1) })}</span>
              <span>{t('header.stock', { n: String(state.stockSize) })}</span>
            </div>

            {/* **取っても得点にならない規則が要。** 先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="ll-rule"
              data-tutorial="ll-rule"
            >
              {t('header.rule')}
            </div>

            {/* **山札が尽きたら誰も補充できない。** 盤面からは読み取れない。 */}
            {state.stockSize === 0 && !isGameEnd && (
              <div className="mb-3 text-center text-ds-warning" role="status" data-testid="ll-no-stock">
                {t('header.noStock')}
              </div>
            )}

            {/* **手札の枚数が生死そのもの。** 得点表示は無い。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="ll-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`ll-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.eliminatedAt > 0 && (
                    <span className="ml-1 text-ds-accent">
                      {t('header.eliminated', { rank: String(p.eliminatedAt) })}
                    </span>
                  )}
                  {p.eliminatedAt === 0 && p.id === state.lastDrawIdx && (
                    <span className="ml-1 text-ds-warning">{t('header.drew')}</span>
                  )}
                  {': '}
                  <span className="text-ds-accent">{t('header.cards', { n: String(p.cardCount) })}</span>
                  {' / '}
                  {t('header.tricks', { n: String(p.tricksWon) })}
                </div>
              ))}
            </div>

            <div data-tutorial="ll-trick">
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
                data-testid="ll-result"
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

            {/* **脱落しても局は続く。** 黙っていると操作不能に見える。 */}
            {isEliminated && (
              <div className="mt-3 text-center text-ds-warning" role="status" data-testid="ll-eliminated">
                {t('result.eliminated')}
              </div>
            )}

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="ll-hand">
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="ll-actions">
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
                      id: 'lingerlonger-players',
                      label: t('actions.players'),
                      value: String(playerCnt),
                      options: [4, 5, 6].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setPlayerCnt(Number(v)),
                      testId: 'll-players-select',
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

/** Linger Longer page wrapped with TutorialProvider. */
export const LingerLongerPage = withTutorial(LingerLongerPageContent, 'lingerlonger', LINGERLONGER_TUTORIAL_STEPS);
