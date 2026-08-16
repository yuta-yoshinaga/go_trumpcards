import { useCallback, useEffect, useMemo } from 'react';
import { germanwhistApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
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
import type { GermanWhistResponse } from '../types/card';
import { GermanWhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GERMANWHIST_HELP, parseGermanWhistCommand } from '../utils/cli/commands/germanwhistCommands';
import { formatGermanWhistState } from '../utils/cli/formatters/germanwhistFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit code to its symbol, for the trump readout. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Guided tutorial steps (face-up card, trick, hand, actions). */
const GERMANWHIST_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="germanwhist-upcard"]',
    messageKey: 'tutorial.upCard',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="germanwhist-trick"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="germanwhist-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="germanwhist-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Inner content for the German Whist page (wrapped by `withTutorial`).
 *
 * Renders the 2-player English trick-taking game: 13 cards each, the next card
 * turned face up with its suit as trump for the whole hand, then 26 tricks in
 * two halves. In the first half the winner takes the face-up card and the
 * loser draws blind, and **those tricks do not score** — so the page shows the
 * face-up card prominently and reports the two trick counts separately, since
 * only the second-half count decides the game.
 */
function GermanWhistPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('germanwhist');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<GermanWhistResponse, Parameters<typeof germanwhistApi.exec>>(germanwhistApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('germanwhist', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('germanwhist');
  const cliConfig: CliGameConfig<GermanWhistResponse, Parameters<typeof germanwhistApi.exec>> = useMemo(
    () => ({
      gameName: 'germanwhist',
      parseCommand: parseGermanWhistCommand,
      formatResponse: formatGermanWhistState,
      helpText: GERMANWHIST_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Initial reset on mount.
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

  if (!state) {
    return (
      <GameSkeleton gameKey="germanwhist" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const cpu = state.players.find((p) => !p.isHuman);
  const isFirstHalf = state.phase === GermanWhistPhase.DRAW;
  const isGameEnd = state.phase === GermanWhistPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd ? t('phase.gameEnd') : isFirstHalf ? t('phase.firstHalf') : t('phase.secondHalf');

  // Following suit is compulsory in both halves, so the legal set is always
  // meaningful. As on the Schnapsen page this is an **additive ring**, not a
  // disabled state: the server validates every play, and disabling cards makes
  // the first clickable card in the hand a moving target for the e2e suite.
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const params = { p0: String(human?.scoringTricks ?? 0), p1: String(cpu?.scoringTricks ?? 0) };
    if (state.winnerIdx === 0) return t('result.youWin', params);
    if (state.winnerIdx === 1) return t('result.cpuWin', params);
    return t('result.tie', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.germanwhist')}
      gameThemeBg={gameTheme.germanwhist.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/germanwhist"
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
            <div className="text-ds-text-primary text-center mb-3">
              <span className="mr-4" data-testid="gw-trick">
                {t('header.trick')}: {state.trickNumber}/26
              </span>
              <span className="mr-4">
                {t('header.stock')}: {state.stockCount}
              </span>
              <span className="mr-4">
                {t('header.trump')}: {SUIT_SYMBOLS[state.trumpSuit] ?? '?'}
              </span>
              <span className="text-ds-accent" data-testid="gw-phase">
                {isFirstHalf ? t('header.phase1') : t('header.phase2')}
              </span>
            </div>

            {/* The face-up card is what the first half is played for. */}
            <div className="flex flex-wrap items-start gap-4 mb-4">
              <div className="p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                {/* **前半は得点が両者 0 のまま。** trickCount まで出さないと、
                    13 トリックがどちらに有利に進んでいるか画面から読めない (#5744)。
                    CUI の germanWhistPlayerStr は最初から両方出している。 */}
                <span data-testid="gw-cpu-tricks">
                  {t('header.cpu')}: {cpu?.cardCount ?? 0} / {t('header.trickCount')}: {cpu?.trickCount ?? 0} /{' '}
                  {t('header.scoringTricks')}: {cpu?.scoringTricks ?? 0}
                </span>
              </div>
              <div
                className="flex items-center gap-2 rounded bg-black/30 p-2"
                data-testid="gw-upcard"
                data-tutorial="germanwhist-upcard"
              >
                <span className="text-ds-text-muted text-sm">
                  {state.upCard
                    ? t('header.upCard')
                    : t('header.upCardNone', { suit: SUIT_SYMBOLS[state.trumpSuit] ?? '?' })}
                </span>
                {state.upCard ? (
                  <div className="flex items-center gap-1">
                    <CardImage card={state.upCard} width={cardWidth} />
                    {state.stockCount > 0 && <AnimatedCardBack width={Math.round(cardWidth * 0.7)} />}
                  </div>
                ) : null}
              </div>
            </div>

            {/* Current trick */}
            <div data-tutorial="germanwhist-trick">
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

            {/* Human hand */}
            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="germanwhist-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  <span data-testid="gw-human-tricks">
                    {t('header.you')}: {human.cardCount} / {t('header.trickCount')}: {human.trickCount} /{' '}
                    {t('header.scoringTricks')}: {human.scoringTricks}
                  </span>
                  {isFirstHalf && <span className="ml-2 text-ds-accent">{t('header.firstHalfNote')}</span>}
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

            {/* Frontend hint */}
            {hintEnabled && hint && (
              <div className="mt-3 flex justify-center">
                <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />
              </div>
            )}

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="germanwhist-actions">
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

/** German Whist page wrapped with TutorialProvider. */
export const GermanWhistPage = withTutorial(GermanWhistPageContent, 'germanwhist', GERMANWHIST_TUTORIAL_STEPS);
