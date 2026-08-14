import { useCallback, useEffect, useMemo, useState } from 'react';
import { pigApi } from '../api/gameApi';
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
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PigResponse } from '../types/card';
import { PigPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PIG_HELP, parsePigCommand } from '../utils/cli/commands/pigCommands';
import { formatPigState } from '../utils/cli/formatters/pigFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guided tutorial steps (the silent signal, the penalty, the pass, your hand). */
const PIG_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pig-rule"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pig-seats"]', messageKey: 'tutorial.penalty', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pig-round"]', messageKey: 'tutorial.pass', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pig-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Pig page (wrapped by `withTutorial`).
 *
 * **The signal is silent, and being slow is the only way to lose.** Nothing on
 * the board records that somebody put a hand to their nose, so the page has to
 * announce it and offer the one button that answers it. Everything else here
 * follows from the same problem: which seats have already chosen a card (all
 * passes happen at once, so there is a wait), who took the letter at the end of
 * a round, and that an eliminated human keeps watching rather than being stuck.
 */
function PigPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pig');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<PigResponse, Parameters<typeof pigApi.exec>>(pigApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('pig', state);
  const [playerCnt, setPlayerCnt] = useState(4);
  const [difficulty, setDifficulty] = useState(1);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pig');
  const cliConfig: CliGameConfig<PigResponse, Parameters<typeof pigApi.exec>> = useMemo(
    () => ({
      gameName: 'pig',
      parseCommand: parsePigCommand,
      formatResponse: formatPigState,
      helpText: PIG_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', undefined, { playerCnt, cpuDifficulty: difficulty });
  }, [dispatch, hideActionLog, playerCnt, difficulty]);

  const handlePass = useCallback(
    (idx: number) => {
      void dispatch('pass', idx);
    },
    [dispatch],
  );

  const handleSignal = useCallback(() => {
    void dispatch('signal');
  }, [dispatch]);

  const handleNext = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="pig" layout={{ kind: 'trick-taking', trickArea: false, footerHandSize: 4 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === PigPhase.GAME_END || state.gameEndFlag;
  const isEliminated = !isGameEnd && human?.eliminated === true;
  // **合図が出ている場面は、押すべきボタンが1つだけ。** 遅れることが負けです。
  const canSignal = state.phase === PigPhase.SIGNAL && !isEliminated && human?.hasSignalled === false;
  const isRoundEnd = state.phase === PigPhase.ROUND_END && !isGameEnd;
  const canPass = state.phase === PigPhase.PASS && !isEliminated && human?.hasChosenPass === false;

  const phaseName = (() => {
    if (isGameEnd) return t('phase.gameEnd');
    if (state.phase === PigPhase.SIGNAL) return t('phase.signal');
    if (state.phase === PigPhase.ROUND_END) return t('phase.roundEnd');
    return t('phase.pass');
  })();

  const seatName = (idx: number) => (idx === 0 ? t('header.you') : t('header.cpu', { idx: String(idx) }));

  const resultBanner = (() => {
    if (!isGameEnd || state.winnerIdx < 0) return null;
    return state.winnerIdx === 0 ? t('result.you') : t('result.cpu', { name: seatName(state.winnerIdx) });
  })();

  return (
    <GamePageShell
      title={tc('nav.pig')}
      gameThemeBg={gameTheme.pig.bg}
      phaseName={phaseName}
      isHumanTurn={canPass || canSignal}
      gamePath="/pig"
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
            <div className="text-ds-text-primary text-center mb-2" data-testid="pig-round" data-tutorial="pig-round">
              <span className="mr-4">{t('header.round', { n: String(state.roundNumber) })}</span>
              <span className="mr-4">{t('header.deck', { n: String(state.deckSize) })}</span>
              <span>{t('header.pass', { n: String(state.passCount) })}</span>
            </div>

            {/* **取り合うものが何もないのが規則そのもの。** 先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="pig-rule"
              data-tutorial="pig-rule"
            >
              {t('header.rule')}
            </div>

            {/* **合図は声に出さない。** 盤面からは絶対に読み取れません。 */}
            {canSignal && (
              <div
                className={`mb-3 rounded px-3 py-2 text-center font-semibold ${badgeWarningColors}`}
                role="status"
                data-testid="pig-signal-alert"
              >
                {t('status.signal')}
              </div>
            )}

            {state.phase === PigPhase.SIGNAL && human?.hasSignalled === true && (
              <div className="mb-3 text-center text-ds-text-muted" role="status" data-testid="pig-signal-done">
                {t('status.signalDone', { n: String(state.noticedCnt) })}
              </div>
            )}

            {/* **罰は1ラウンドに1回の出来事。** 配り直す前に読ませます。 */}
            {isRoundEnd && state.roundLoserIdx >= 0 && (
              <div className="mb-3 text-center text-ds-accent" role="status" data-testid="pig-round-end">
                {t('status.roundEnd', {
                  name: seatName(state.roundLoserIdx),
                  word: state.players[state.roundLoserIdx]?.letterWord ?? '',
                })}
              </div>
            )}

            {/* **文字がそのまま残機。** 得点表示はありません。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="pig-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`pig-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">{seatName(p.id)}</span>
                  {p.eliminated && <span className="ml-1 text-ds-danger">{t('header.out')}</span>}
                  {!p.eliminated && p.noticedOrder > 0 && (
                    <span className="ml-1 text-ds-success">
                      {t('header.noticed', { order: String(p.noticedOrder) })}
                    </span>
                  )}
                  {!p.eliminated && p.noticedOrder === 0 && p.hasChosenPass && (
                    <span className="ml-1 text-ds-warning">{t('header.chosen')}</span>
                  )}
                  {': '}
                  <span className="text-ds-accent">{t('header.cards', { n: String(p.cardCount) })}</span>
                  {' / '}
                  {t('header.letters', { word: p.letterWord || '-' })}
                </div>
              ))}
            </div>

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="pig-result"
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

            {isEliminated && (
              <div className="mt-3 text-center text-ds-warning" role="status" data-testid="pig-eliminated">
                {t('status.eliminated')}
              </div>
            )}

            {!isEliminated && state.phase === PigPhase.PASS && human?.hasChosenPass === true && (
              <div className="mt-3 text-center text-ds-text-muted" role="status" data-testid="pig-waiting">
                {t('status.waiting')}
              </div>
            )}

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="pig-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => handlePass(idx)}
                      disabled={loading || !canPass}
                      aria-label={t('actions.passAria', { card: cardAlt(card) })}
                      className={`disabled:opacity-50 ${canPass ? 'rounded-lg ring-2 ring-ds-success' : ''}`}
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

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="pig-actions">
              {canSignal && (
                <button
                  type="button"
                  className={btnWarning}
                  onClick={handleSignal}
                  disabled={loading}
                  data-testid="pig-signal-btn"
                >
                  {t('actions.signal')}
                </button>
              )}
              {isRoundEnd && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleNext}
                  disabled={loading}
                  data-testid="pig-next-btn"
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
                      id: 'pig-players',
                      label: t('actions.players'),
                      value: String(playerCnt),
                      options: [3, 4, 5, 6].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setPlayerCnt(Number(v)),
                      testId: 'pig-players-select',
                    },
                    {
                      type: 'select',
                      id: 'pig-difficulty',
                      label: t('actions.difficulty'),
                      value: String(difficulty),
                      options: [
                        { value: '0', label: t('actions.difficultyEasy') },
                        { value: '1', label: t('actions.difficultyNormal') },
                        { value: '2', label: t('actions.difficultyHard') },
                      ],
                      onSelect: (v: string) => setDifficulty(Number(v)),
                      testId: 'pig-difficulty-select',
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

/** Pig page wrapped with TutorialProvider. */
export const PigPage = withTutorial(PigPageContent, 'pig', PIG_TUTORIAL_STEPS);
