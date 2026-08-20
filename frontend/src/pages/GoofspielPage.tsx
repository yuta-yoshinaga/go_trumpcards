import { useCallback, useEffect, useMemo, useState } from 'react';
import { goofspielApi } from '../api/gameApi';
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
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GoofspielResponse } from '../types/card';
import { GoofspielPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GOOFSPIEL_HELP, parseGoofspielCommand } from '../utils/cli/commands/goofspielCommands';
import { formatGoofspielState } from '../utils/cli/formatters/goofspielFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guided tutorial steps (simultaneous bidding, open hands, the pot, your cards). */
const GOOFSPIEL_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="gs-rule"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="gs-seats"]', messageKey: 'tutorial.open', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="gs-prize"]', messageKey: 'tutorial.prize', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="gs-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Goofspiel page (wrapped by `withTutorial`).
 *
 * **Every seat's remaining cards are shown, CPUs included.** Spent cards are
 * public, so a remainder is just a suit minus what has been played — hiding it
 * would make the player do arithmetic instead of thinking.
 *
 * What *is* hidden is the bid in flight: a seat shows only that it has
 * committed until the reveal turns the cards face up.
 */
function GoofspielPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('goofspiel');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<GoofspielResponse, Parameters<typeof goofspielApi.exec>>(goofspielApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('goofspiel', state);
  const [playerCnt, setPlayerCnt] = useState(2);
  const [tieRule, setTieRule] = useState(0);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('goofspiel');
  const cliConfig: CliGameConfig<GoofspielResponse, Parameters<typeof goofspielApi.exec>> = useMemo(
    () => ({
      gameName: 'goofspiel',
      parseCommand: parseGoofspielCommand,
      formatResponse: formatGoofspielState,
      helpText: GOOFSPIEL_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', undefined, { playerCnt, tieRule });
  }, [dispatch, hideActionLog, playerCnt, tieRule]);

  const handleBid = useCallback(
    (idx: number) => {
      void dispatch('bid', idx);
    },
    [dispatch],
  );

  const handleNext = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="goofspiel" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === GoofspielPhase.GAME_END || state.gameEndFlag;
  const isReveal = state.phase === GoofspielPhase.REVEAL && !isGameEnd;
  const canBid = state.phase === GoofspielPhase.BID && !isGameEnd && human?.hasBid === false;

  const seatName = (idx: number) => (idx === 0 ? t('header.you') : t('header.cpu', { idx: String(idx) }));

  const phaseName = (() => {
    if (isGameEnd) return t('phase.gameEnd');
    if (isReveal) return t('phase.reveal');
    return t('phase.bid');
  })();

  const resultBanner = (() => {
    if (!isGameEnd || state.winnerIdx < 0) return null;
    const n = String(state.players[state.winnerIdx]?.score ?? 0);
    return state.winnerIdx === 0 ? t('result.you', { n }) : t('result.cpu', { name: seatName(state.winnerIdx), n });
  })();

  return (
    <GamePageShell
      title={tc('nav.goofspiel')}
      gameThemeBg={gameTheme.goofspiel.bg}
      phaseName={phaseName}
      isHumanTurn={canBid}
      gamePath="/goofspiel"
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
            <div className="text-ds-text-primary text-center mb-2" data-testid="gs-header">
              <span className="mr-4">{t('header.round', { n: String(state.roundNumber) })}</span>
              <span>{t('header.left', { n: String(state.prizeRemaining) })}</span>
            </div>

            {/* **同時入札であることが規則そのもの。** */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="gs-rule"
              data-tutorial="gs-rule"
            >
              {t('header.rule')}
            </div>

            {/* **懸かっている点は賞札のランク（＋持ち越し）。** */}
            {state.currentPrize && (
              <div className="mb-4 flex flex-col items-center gap-1" data-testid="gs-prize" data-tutorial="gs-prize">
                <div className="text-ds-text-muted text-sm">{t('header.prize')}</div>
                <CardImage card={state.currentPrize} width={cardWidth} />
                <div className="text-ds-accent font-semibold">
                  {t('header.prizeValue', { n: String(state.prizeValue) })}
                </div>
                {(state.carriedPrizes?.length ?? 0) > 0 && (
                  <div className="text-ds-warning text-sm" data-testid="gs-carried">
                    {t('header.carried', { n: String(state.carriedPrizes?.length ?? 0) })}
                  </div>
                )}
              </div>
            )}

            {/* **残り札は全員分を公開。** 使った札は場に出るので隠せていません。 */}
            <div className="flex flex-col gap-2 mb-4" data-tutorial="gs-seats">
              {state.players.map((p) => (
                <div key={p.id} className="rounded bg-black/30 px-3 py-2" data-testid={`gs-seat-${p.id.toString()}`}>
                  <div className="text-sm text-ds-text-muted">
                    <span className="text-ds-text-primary">{seatName(p.id)}</span>
                    {p.revealedBid ? (
                      <span className="ml-1 text-ds-accent">{t('header.revealed')}</span>
                    ) : (
                      p.hasBid && <span className="ml-1 text-ds-warning">{t('header.bidDone')}</span>
                    )}
                    {': '}
                    <span>{t('header.cards', { n: String(p.cardCount) })}</span>
                    {' / '}
                    <span className="text-ds-accent">{t('header.score', { n: String(p.score) })}</span>
                  </div>
                  {p.revealedBid && (
                    <div className="mt-1">
                      <CardImage card={p.revealedBid} width={cardWidth} />
                    </div>
                  )}
                  {/* **勝負はランクの大小比較そのもの** (#5769)。CPU の残り札も
                      公開情報なので、alt 文字列の羅列ではなく自分の手札と同じ絵で
                      並べる。枚数が多い局面 (13枚) でも折り返せるよう flex-wrap。 */}
                  {!p.isHuman && p.cards.length > 0 && (
                    <div className="mt-1 flex flex-wrap gap-1" data-testid={`gs-hand-${p.id.toString()}`}>
                      {p.cards.map((c) => (
                        <CardImage
                          key={`${c.design}-${c.value.toString()}`}
                          card={c}
                          width={Math.round(cardWidth * 0.5)}
                        />
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="gs-result"
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

            {/* **同点は誰も取らない。** 勝者が居ない結果を言い分けます。 */}
            {isReveal && (
              <div className="mt-3 text-center text-ds-warning" role="status" data-testid="gs-round-end">
                {state.lastWinnerIdx < 0
                  ? t('status.tie')
                  : t('status.roundEnd', {
                      name: seatName(state.lastWinnerIdx),
                      n: String(state.lastGained),
                    })}
              </div>
            )}

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="gs-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => handleBid(idx)}
                      disabled={loading || !canBid}
                      aria-label={t('actions.bidAria', { card: cardAlt(card) })}
                      className={`disabled:opacity-50 ${canBid ? 'rounded-lg ring-2 ring-ds-success' : ''}`}
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
              {isReveal && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleNext}
                  disabled={loading}
                  data-testid="gs-next-btn"
                >
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
                      id: 'goofspiel-players',
                      label: t('actions.players'),
                      value: String(playerCnt),
                      options: [2, 3].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setPlayerCnt(Number(v)),
                      testId: 'gs-players-select',
                    },
                    {
                      type: 'select',
                      id: 'goofspiel-tie',
                      label: t('actions.tieRule'),
                      value: String(tieRule),
                      options: [
                        { value: '0', label: t('actions.tieDiscard') },
                        { value: '1', label: t('actions.tieCarry') },
                      ],
                      onSelect: (v: string) => setTieRule(Number(v)),
                      testId: 'gs-tie-select',
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

/** Goofspiel page wrapped with TutorialProvider. */
export const GoofspielPage = withTutorial(GoofspielPageContent, 'goofspiel', GOOFSPIEL_TUTORIAL_STEPS);
