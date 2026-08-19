import { useCallback, useEffect, useMemo, useState } from 'react';
import { stealingbundlesApi } from '../api/gameApi';
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
import { btnDanger, btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { StealingBundlesResponse } from '../types/card';
import { StealingBundlesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseStealingBundlesCommand, STEALINGBUNDLES_HELP } from '../utils/cli/commands/stealingbundlesCommands';
import { formatStealingBundlesState } from '../utils/cli/formatters/stealingbundlesFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guided tutorial steps (capturing, stealing, the compulsory capture, your hand). */
const STEALINGBUNDLES_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sb-table"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sb-seats"]', messageKey: 'tutorial.steal', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sb-status"]', messageKey: 'tutorial.mustCapture', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sb-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Stealing Bundles page (wrapped by `withTutorial`).
 *
 * **Every seat's bundle top is public, including the CPUs'.** It is the one
 * thing opponents aim at, so the page shows it next to each bundle count.
 *
 * Choosing a card is a two-step: pick a card, then pick what to do with it.
 * A single click cannot express the choice, because the same card may both
 * capture from the table and steal from more than one seat — and the server
 * only accepts one of those per turn.
 */
function StealingBundlesPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('stealingbundles');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<StealingBundlesResponse, Parameters<typeof stealingbundlesApi.exec>>(stealingbundlesApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('stealingbundles', state);
  const [playerCnt, setPlayerCnt] = useState(4);
  const [selected, setSelected] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('stealingbundles');
  const cliConfig: CliGameConfig<StealingBundlesResponse, Parameters<typeof stealingbundlesApi.exec>> = useMemo(
    () => ({
      gameName: 'stealingbundles',
      parseCommand: parseStealingBundlesCommand,
      formatResponse: formatStealingBundlesState,
      helpText: STEALINGBUNDLES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    setSelected(null);
    void dispatch('reset', undefined, undefined, { playerCnt });
  }, [dispatch, hideActionLog, playerCnt]);

  const handleTake = useCallback(
    (idx: number) => {
      setSelected(null);
      void dispatch('take', idx);
    },
    [dispatch],
  );

  const handleSteal = useCallback(
    (idx: number, victim: number) => {
      setSelected(null);
      void dispatch('steal', idx, victim);
    },
    [dispatch],
  );

  const handleTrail = useCallback(
    (idx: number) => {
      setSelected(null);
      void dispatch('trail', idx);
    },
    [dispatch],
  );

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return (
      <GameSkeleton gameKey="stealingbundles" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === StealingBundlesPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  const seatName = (idx: number) => (idx === 0 ? t('header.you') : t('header.cpu', { idx: String(idx) }));

  const selectedCard = selected === null ? undefined : human?.cards[selected];
  const selectedTakes = selected === null ? [] : (state.tableMatches[String(selected)] ?? []);
  const selectedSteals = selected === null ? [] : (state.stealTargets[String(selected)] ?? []);
  // **取れるときは置けません。** サーバが必ず拒否するので、ボタンも出しません。
  const canTrailSelected = selected !== null && !state.canCapture;

  const resultBanner = (() => {
    if (!isGameEnd || state.winnerIdx < 0) return null;
    const n = String(state.players[state.winnerIdx]?.bundleSize ?? 0);
    return state.winnerIdx === 0 ? t('result.you', { n }) : t('result.cpu', { name: seatName(state.winnerIdx), n });
  })();

  return (
    <GamePageShell
      title={tc('nav.stealingbundles')}
      gameThemeBg={gameTheme.stealingbundles.bg}
      phaseName={isGameEnd ? t('phase.gameEnd') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/stealingbundles"
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
            <div className="text-ds-text-primary text-center mb-2" data-testid="sb-header">
              <span className="mr-4">{t('header.turn', { n: String(state.turnNumber + 1) })}</span>
              <span>{t('header.deck', { n: String(state.deckRemaining) })}</span>
            </div>

            {/* **束の一番上が弱点、というのが規則そのもの。** 先に出す。 */}
            <div className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center" data-testid="sb-rule">
              {t('header.rule')}
            </div>

            {/* **空の場も情報。** 消さずに「なし」と書きます。 */}
            <div className="mb-4" data-testid="sb-table" data-tutorial="sb-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('header.table')}</div>
              {state.tableCards.length === 0 ? (
                <div className="text-ds-text-muted">{t('header.tableEmpty')}</div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {state.tableCards.map((card, idx) => (
                    <CardImage key={`${card.design}-${card.value}-${idx}`} card={card} width={cardWidth} />
                  ))}
                </div>
              )}
            </div>

            {/* **束の一番上は全員に見えます。** そこが狙われる場所だからです。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="sb-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`sb-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">{seatName(p.id)}</span>
                  {/* **盗みは相手の束を丸ごと消す。** 場から取っただけの手と
                      同じ印にすると、痕跡の残らないこの盤面では区別が付かない
                      (#5767)。 */}
                  {p.id === state.lastCaptureIdx && (
                    <span className="ml-1 text-ds-warning" data-testid={`sb-capture-${p.id.toString()}`}>
                      {state.lastCaptureKind === 'steal'
                        ? t('header.lastCaptureSteal', { name: seatName(state.lastCaptureVictimIdx) })
                        : t('header.lastCaptureTake')}
                    </span>
                  )}
                  {': '}
                  <span>{t('header.cards', { n: String(p.cardCount) })}</span>
                  {' / '}
                  <span className="text-ds-accent">{t('header.bundle', { n: String(p.bundleSize) })}</span>
                  {' / '}
                  <span>
                    {t('header.bundleTop')}:{' '}
                    {p.bundleTop ? (
                      <span className="text-ds-accent">{cardAlt(p.bundleTop)}</span>
                    ) : (
                      t('header.bundleEmpty')
                    )}
                  </span>
                </div>
              ))}
            </div>

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="sb-result"
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

            {/* **取れるときは置けない、を必ず言う。** 黙っていると詰まって見えます。 */}
            {isHumanTurn && (
              <div
                className="mt-3 text-center text-ds-text-muted"
                role="status"
                data-testid="sb-status"
                data-tutorial="sb-status"
              >
                {state.canCapture ? t('status.mustCapture') : t('status.trailOnly')}
              </div>
            )}

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="sb-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => {
                    const usable =
                      (state.tableMatches[String(idx)]?.length ?? 0) > 0 ||
                      (state.stealTargets[String(idx)]?.length ?? 0) > 0 ||
                      !state.canCapture;
                    return (
                      <button
                        key={`${card.design}-${card.value}-${idx}`}
                        type="button"
                        onClick={() => setSelected(idx)}
                        disabled={loading || !isHumanTurn}
                        aria-label={t('actions.selectAria', { card: cardAlt(card) })}
                        aria-pressed={selected === idx}
                        className={`disabled:opacity-50 ${
                          selected === idx
                            ? 'rounded-lg ring-2 ring-ds-accent'
                            : usable && isHumanTurn
                              ? 'rounded-lg ring-2 ring-ds-success'
                              : ''
                        }`}
                      >
                        <CardImage card={card} width={cardWidth} />
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            {/* **選んだ札でできることだけを出す。** できない手を押させません。 */}
            {selected !== null && isHumanTurn && (
              <div className="mt-3 flex flex-wrap gap-2 items-center" data-testid="sb-actions">
                <span className="text-ds-text-muted text-sm">
                  {t('status.selected', { card: selectedCard ? cardAlt(selectedCard) : '' })}
                </span>
                {selectedTakes.length > 0 && (
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleTake(selected)}
                    disabled={loading}
                    data-testid="sb-take-btn"
                  >
                    {t('actions.take')}
                  </button>
                )}
                {selectedSteals.map((victim) => (
                  <button
                    key={victim}
                    type="button"
                    className={btnWarning}
                    onClick={() => handleSteal(selected, victim)}
                    disabled={loading}
                    data-testid={`sb-steal-btn-${victim.toString()}`}
                  >
                    {t('actions.steal', { name: seatName(victim) })}
                  </button>
                ))}
                {canTrailSelected && (
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => handleTrail(selected)}
                    disabled={loading}
                    data-testid="sb-trail-btn"
                  >
                    {t('actions.trail')}
                  </button>
                )}
                <button type="button" className={btnSecondary} onClick={() => setSelected(null)} disabled={loading}>
                  {t('actions.cancel')}
                </button>
              </div>
            )}

            {hintEnabled && hint && (
              <div className="mt-3 flex justify-center">
                <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />
              </div>
            )}

            <div className="mt-4 flex flex-wrap gap-2">
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
                      id: 'stealingbundles-players',
                      label: t('actions.players'),
                      value: String(playerCnt),
                      options: [2, 3, 4].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setPlayerCnt(Number(v)),
                      testId: 'sb-players-select',
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

/** Stealing Bundles page wrapped with TutorialProvider. */
export const StealingBundlesPage = withTutorial(
  StealingBundlesPageContent,
  'stealingbundles',
  STEALINGBUNDLES_TUTORIAL_STEPS,
);
