import { useCallback, useEffect, useMemo } from 'react';
import { polignacApi } from '../api/gameApi';
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
import type { PolignacResponse } from '../types/card';
import { PolignacPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { POLIGNAC_HELP, parsePolignacCommand } from '../utils/cli/commands/polignacCommands';
import { formatPolignacState } from '../utils/cli/formatters/polignacFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by the domain's design constant (1..4). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;
/** Domain constant for spades — the jack that costs two points. */
const SPADE_DESIGN = 1;
/** Suit names for the accessible reading, indexed like SUIT_SYMBOLS. */
const SUIT_ARIA_KEYS = ['', 'spade', 'clover', 'heart', 'diamond'] as const;

/** Tricks per round. Capot means taking all of them. */
const TRICKS_PER_ROUND = 8;

/** Guided tutorial steps (penalty rule, capot, trick, hand). */
const POLIGNAC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="polignac-penalty"]',
    messageKey: 'tutorial.penalty',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="polignac-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="polignac-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="polignac-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Polignac page (wrapped by `withTutorial`).
 *
 * Renders the 4-player French avoidance game: a 32-card piquet pack, no trump,
 * and penalties that fall on **the four jacks only** — one point each except
 * the jack of spades, which costs two. Neither the "jacks only" rule nor the
 * doubled spade is visible from the table, so the page states both in a
 * standing banner. A player may also declare **capot** (win all eight tricks)
 * before play, which inverts everyone else's goal from avoiding jacks to
 * taking a single trick.
 */
function PolignacPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('polignac');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<PolignacResponse, Parameters<typeof polignacApi.exec>>(polignacApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('polignac', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('polignac');
  const cliConfig: CliGameConfig<PolignacResponse, Parameters<typeof polignacApi.exec>> = useMemo(
    () => ({
      gameName: 'polignac',
      parseCommand: parsePolignacCommand,
      formatResponse: formatPolignacState,
      helpText: POLIGNAC_HELP,
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

  const handleCapot = useCallback(() => {
    void dispatch('capot');
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
    return <GameSkeleton gameKey="polignac" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isDeclare = state.phase === PolignacPhase.DECLARE;
  const isRoundEnd = state.phase === PolignacPhase.ROUND_END;
  const isGameEnd = state.phase === PolignacPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    !isGameEnd && !isRoundEnd && !isDeclare && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isDeclare
        ? t('phase.declare')
        : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx < 0) return t('result.tie');
    const winner = state.players[state.winnerIdx];
    return state.winnerIdx === 0
      ? t('result.youWin', { score: String(winner?.score ?? 0) })
      : t('result.cpuWin', { idx: String(state.winnerIdx), score: String(winner?.score ?? 0) });
  })();

  return (
    <GamePageShell
      title={tc('nav.polignac')}
      gameThemeBg={gameTheme.polignac.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/polignac"
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
              <span className="mr-4" data-testid="pg-round">
                {t('header.round')}: {state.roundNumber}/{state.config.rounds}
              </span>
              <span className="mr-4" data-testid="pg-trick">
                {t('header.trick')}: {state.trickNumber + 1}/{TRICKS_PER_ROUND}
              </span>
            </div>

            {/* 失点するのはジャックだけ、♠J は2倍。盤面には出ない規則。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-muted text-sm text-center"
              data-testid="pg-penalty-rule"
              data-tutorial="polignac-penalty"
            >
              {t('header.penaltyRule')}
            </div>

            {/* capot 宣言中は全員の狙いが変わる。 */}
            {state.capotIdx >= 0 && !isGameEnd && (
              <div
                className="mb-3 rounded bg-black/30 border border-ds-warning px-3 py-2 text-ds-text-primary text-sm text-center"
                role="status"
                data-testid="pg-capot-banner"
              >
                {t('capot.active', {
                  name: state.capotIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(state.capotIdx) }),
                  tricks: String(state.capotTricks),
                })}
              </div>
            )}

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="polignac-scores">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`pg-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.declaredCapot && <span className="ml-1 text-ds-warning">{t('capot.mark')}</span>}
                  {': '}
                  {t('header.score', { score: String(p.score), round: String(p.roundPenalty) })}
                  {/* **合計だけでは ♠J を踏んだのか他を 2 枚拾ったのかが分からない** (#5746)。
                      姉妹ゲームの Slobberhannes / Reversis は取った印付き札を個別に出している。 */}
                  {(p.takenJackSuits?.length ?? 0) > 0 && (
                    <span className="ml-2" data-testid={`pg-jacks-${p.id.toString()}`}>
                      {t('jacks.label')}:{' '}
                      {p.takenJackSuits?.map((suit) => (
                        <span
                          key={`${p.id.toString()}-jack-${suit.toString()}`}
                          className={suit === SPADE_DESIGN ? 'ml-1 font-bold text-ds-error' : 'ml-1'}
                        >
                          <span aria-hidden="true">
                            {suit === SPADE_DESIGN
                              ? t('jacks.spade')
                              : t('jacks.other', { suit: SUIT_SYMBOLS[suit] ?? '?' })}
                          </span>
                          <span className="sr-only">
                            {suit === SPADE_DESIGN
                              ? t('jacks.spadeAria')
                              : t('jacks.otherAria', {
                                  suitName: t(`jacks.suitName.${SUIT_ARIA_KEYS[suit] ?? 'spade'}`),
                                })}
                          </span>
                        </span>
                      ))}
                    </span>
                  )}
                </div>
              ))}
            </div>

            <div data-tutorial="polignac-trick">
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
              <div className="mt-4" data-tutorial="polignac-hand">
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

            <div className="mt-4 flex flex-wrap gap-2">
              {/* 宣言は配り直後の一度きり。ここを逃すと capot は打てない。 */}
              {isDeclare && !isGameEnd && (
                <>
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={handleCapot}
                    disabled={loading}
                    data-testid="pg-capot-btn"
                  >
                    {t('actions.declareCapot')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handlePass}
                    disabled={loading}
                    data-testid="pg-pass-btn"
                  >
                    {t('actions.pass')}
                  </button>
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

/** Polignac page wrapped with TutorialProvider. */
export const PolignacPage = withTutorial(PolignacPageContent, 'polignac', POLIGNAC_TUTORIAL_STEPS);
