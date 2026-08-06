import { useCallback, useEffect, useMemo } from 'react';
import { schnapsenApi } from '../api/gameApi';
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
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SchnapsenResponse } from '../types/card';
import { SchnapsenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, suitSymbol } from '../utils/cardAlt';
import { parseSchnapsenCommand, SCHNAPSEN_HELP } from '../utils/cli/commands/schnapsenCommands';
import { formatSchnapsenState } from '../utils/cli/formatters/schnapsenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { computeSchnapsenLegalRing } from '../utils/schnapsenLegal';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Card design → Schnapsen trump-suit id (1=♠ 2=♣ 3=♥ 4=♦). */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/**
 * Suit code (1=♠ 2=♣ 3=♥ 4=♦) back to a design, for the phase-2 header.
 *
 * ストックが尽きて切り札の現物が手札に入ると、画面から切り札スートが消えていた。
 * フェーズ2 はマストフォローの厳格ルールなので、スートを覚えていないと合法手の
 * リングの意味が読めない。CUI は同じ場面でスート名を出している (#4810)。
 */
const SUIT_TO_DESIGN: Readonly<Record<number, string>> = { 1: 'SPADE', 2: 'CLOVER', 3: 'HEART', 4: 'DIAMOND' };

/** Guided tutorial steps for the Schnapsen page (trump/stock, trick, hand, actions). */
const SCHNAPSEN_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="schnapsen-trump"]', messageKey: 'tutorial.trump', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="schnapsen-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="schnapsen-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="schnapsen-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Inner content for the Schnapsen / Sixty-Six page (wrapped by `withTutorial`).
 *
 * Renders the 2-player Austrian trick-taking game: a 20-card deck, a face-up
 * trump upcard, marriages (K+Q same suit, +20 / +40 trump), Briscola-style
 * draw after each trick, and a must-follow second phase once the stock is
 * exhausted. First to 66 card points wins the round.
 */
function SchnapsenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('schnapsen');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<SchnapsenResponse, Parameters<typeof schnapsenApi.exec>>(schnapsenApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('schnapsen', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('schnapsen');
  const cliConfig: CliGameConfig<SchnapsenResponse, Parameters<typeof schnapsenApi.exec>> = useMemo(
    () => ({
      gameName: 'schnapsen',
      parseCommand: parseSchnapsenCommand,
      formatResponse: formatSchnapsenState,
      helpText: SCHNAPSEN_HELP,
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

  const handleMarriage = useCallback(
    (idx: number) => {
      void dispatch('marriage', idx);
    },
    [dispatch],
  );

  const handleNext = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="schnapsen" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const cpu = state.players.find((p) => !p.isHuman);
  const isPlayPhase = state.phase === SchnapsenPhase.PLAY;
  const isTrickEnd = state.phase === SchnapsenPhase.TRICK_END;
  const isGameEnd = state.phase === SchnapsenPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseName = isGameEnd ? t('phase.gameEnd') : isTrickEnd ? t('phase.trickEnd') : t('phase.play');

  const marriageSet = new Set(state.marriagePlays ?? []);
  // In the second phase (stock exhausted) strict must-follow rules apply. We
  // *guide* the human with an additive success ring on the legal cards instead
  // of disabling illegal ones — the backend still validates every play, so
  // illegal cards stay clickable (no hard block). Phase 1 shows no ring.
  const legalRing = computeSchnapsenLegalRing(state.isEndgame, isHumanTurn, state.validPlays);
  const showLegalGuide = isHumanTurn && state.isEndgame;

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const p0 = state.players[0]?.points ?? 0;
    const p1 = state.players[1]?.points ?? 0;
    const params = { p0: String(p0), p1: String(p1) };
    if (state.winnerIdx === 0) return t('result.youWin', params);
    if (state.winnerIdx === 1) return t('result.cpuWin', params);
    return t('result.tie', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.schnapsen')}
      gameThemeBg={gameTheme.schnapsen.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/schnapsen"
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
              <span className="mr-4">
                {t('header.trick')}: {state.trickNumber}
              </span>
              <span className="mr-4">
                {t('header.stock')}: {state.stockRemaining}
              </span>
              <span className="mr-4">
                {t('header.points')} — {t('header.you')}: {human?.points ?? 0} / {t('header.cpu')}: {cpu?.points ?? 0}
              </span>
              <span className="text-ds-accent" data-testid="schnapsen-phase">
                {state.isEndgame ? t('header.phase2') : t('header.phase1')}
              </span>
            </div>

            {/* CPU info + trump upcard */}
            <div className="flex flex-wrap items-start gap-4 mb-4">
              <div className="p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                {t('header.cpu')}: {cpu?.cardCount ?? 0} / {t('header.tricks')}: {cpu?.trickCount ?? 0}
              </div>
              <div
                className="flex items-center gap-2 rounded bg-black/30 p-2"
                data-testid="schnapsen-stock"
                data-tutorial="schnapsen-trump"
              >
                <span className="text-ds-text-muted text-sm">
                  {state.trumpCard
                    ? t('header.trump')
                    : t('header.trumpNoneWithSuit', { suit: suitSymbol(SUIT_TO_DESIGN[state.trumpSuit] ?? '') })}
                </span>
                {state.trumpCard ? (
                  <div
                    className="relative"
                    style={{ width: Math.round(cardWidth * 1.1), height: Math.round(cardWidth * 0.95) }}
                  >
                    <div
                      className="absolute left-0 top-1/2"
                      style={{ transform: 'translateY(-50%) rotate(90deg)', transformOrigin: 'left center' }}
                    >
                      <CardImage card={state.trumpCard} width={Math.round(cardWidth * 0.7)} />
                    </div>
                    {state.stockRemaining > 0 && (
                      <div
                        className="absolute left-1/2 top-1/2 -translate-y-1/2"
                        style={{ transform: 'translate(0,-50%)' }}
                      >
                        <div className="relative">
                          {Array.from({ length: Math.min(state.stockRemaining, 4) }, (_, i) => (
                            <div
                              key={`back-${i.toString()}`}
                              className="absolute"
                              style={{ top: i * -1.5, left: i * 1.5 }}
                            >
                              <AnimatedCardBack width={Math.round(cardWidth * 0.7)} />
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                ) : null}
              </div>
            </div>

            {/* Current trick */}
            <div data-tutorial="schnapsen-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {/* Result banner */}
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
              <div className="mt-4" data-tutorial="schnapsen-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount} / {t('header.tricks')}: {human.trickCount}
                </div>
                {showLegalGuide && (
                  <div
                    className="mb-2 rounded bg-black/30 border border-ds-success px-3 py-2 text-ds-text-primary text-sm"
                    role="status"
                    data-testid="schnapsen-endgame-guide"
                  >
                    {t('endgameRuleGuide')}
                  </div>
                )}
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => {
                    const legal = legalRing.has(idx);
                    return (
                      <div key={`${card.design}-${card.value}-${idx}`} className="flex flex-col items-center gap-1">
                        <button
                          type="button"
                          onClick={() => handlePlay(idx)}
                          disabled={loading || !isHumanTurn}
                          aria-label={t('actions.playAria', { card: cardAlt(card) })}
                          className={`disabled:opacity-50 ${legal ? 'rounded-lg ring-2 ring-ds-success' : ''}`}
                        >
                          <CardImage card={card} width={cardWidth} />
                        </button>
                        {isHumanTurn && marriageSet.has(idx) && (
                          <button
                            type="button"
                            onClick={() => handleMarriage(idx)}
                            disabled={loading}
                            className={`${btnWarning} text-xs px-2 py-1`}
                            data-testid={`schnapsen-marriage-${idx.toString()}`}
                            aria-label={t('actions.marriageAria', {
                              suit: suitSymbol(card.design),
                              points: DESIGN_TO_SUIT[card.design] === state.trumpSuit ? 40 : 20,
                            })}
                          >
                            👑 {t('actions.marriage')}
                          </button>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Frontend hint */}
            {hintEnabled && hint && (
              <div className="mt-3 flex justify-center">
                <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />
              </div>
            )}

            {/* Phase-specific controls */}
            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="schnapsen-actions">
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
                  {t('actions.next')}
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
            </div>

            <SettingsPanel
              title={tc('settings.title')}
              groups={[
                {
                  items: [hintCheckboxItem(tc, hintEnabled, setHintEnabled)],
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

/** Schnapsen page wrapped with TutorialProvider. */
export const SchnapsenPage = withTutorial(SchnapsenPageContent, 'schnapsen', SCHNAPSEN_TUTORIAL_STEPS);
