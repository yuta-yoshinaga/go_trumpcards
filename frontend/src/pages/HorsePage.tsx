import { useCallback, useEffect, useMemo, useState } from 'react';
import type { horseApi as HorseApi } from '../api/gameApi';
import { horseApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { BettingControls } from '../components/BettingControls';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { HorseResponse } from '../types/card';
import { HorsePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { HORSE_HELP, parseHorseCommand } from '../utils/cli/commands/horseCommands';
import { formatHorseState } from '../utils/cli/formatters/horseFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Round names by discipline family, indexed by the running discipline's phase.
 *
 * **The five disciplines do not share a betting structure** (#5788): the
 * community-card games run pre-flop→river, the stud games run 3rd→7th street.
 * Index 0 is "not dealt yet", so it has no label.
 *
 * Phase numbers come from the domain (`HoldemPhase*` / `SevenCardStudPhase*`);
 * `horse_round_labels_test.go` pins these two arrays against those constants.
 */
const COMMUNITY_ROUND_KEYS = ['', 'preflop', 'flop', 'turn', 'river', 'showdown'] as const;
const STUD_ROUND_KEYS = ['', 'third', 'fourth', 'fifth', 'sixth', 'seventh', 'showdown'] as const;

/** Disciplines whose betting rounds are named after community cards. */
const COMMUNITY_DISCIPLINES = new Set(['holdem', 'omahaHiLo']);

/** H.O.R.S.E. tutorial step definitions. */
const HORSE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ho-discipline"]',
    messageKey: 'tutorial.discipline',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ho-seats"]', messageKey: 'tutorial.seats', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ho-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
];

/**
 * Table sizes the disciplines accept.
 *
 * **Only 4 / 6 / 9 exist.** The underlying poker engines silently fall back to
 * a four-handed table for anything else, so offering 5 would quietly seat a
 * different number of players than the one shown.
 */
const SEAT_OPTIONS = [4, 6, 9] as const;

/** Hands played before the table moves on to the next discipline. */
const HANDS_OPTIONS = [1, 2, 3, 5, 10] as const;

/** Renders the H.O.R.S.E. page: five poker disciplines rotating at one table. */
export const HorsePage = withTutorial(HorsePageContent, 'horse', HORSE_TUTORIAL_STEPS);

/** Inner content of the H.O.R.S.E. page, wrapped by TutorialWrapper. */
function HorsePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('horse');
  const [seats, setSeats] = useState(4);
  const [handsPerDiscipline, setHandsPerDiscipline] = useState(2);
  const [betAmount, setBetAmount] = useState(20);

  const { loading, error, state, exec: callApi, retry } = useGameApi(horseApi.exec);
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('horse', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const resetWithConfig = useCallback(() => {
    callApi('reset', { config: { seats, handsPerDiscipline } });
  }, [callApi, seats, handsPerDiscipline]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('horse');
  const cliConfig: CliGameConfig<HorseResponse, Parameters<typeof HorseApi.exec>> = useMemo(
    () => ({
      gameName: 'horse',
      parseCommand: parseHorseCommand,
      formatResponse: formatHorseState,
      helpText: HORSE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state) {
    return (
      <GameSkeleton
        gameKey="horse"
        layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 0, footerHandSize: 2 }}
      />
    );
  }

  const isGameEnd = state.gameEndFlag;
  const isHandEnd = state.phase === HorsePhase.HAND_END && !isGameEnd;
  const isHumanTurn = state.isHumanTurn && !isGameEnd && state.phase === HorsePhase.HAND;
  const humanWon = isGameEnd && state.winnerSeat === state.humanSeat;
  const phaseName = isGameEnd ? t('phase.gameEnd') : isHandEnd ? t('phase.handEnd') : t('phase.play');
  const disciplineName = t(`discipline.${state.disciplineName}`, { defaultValue: state.disciplineName });
  // **いま何回戦目のベットなのかを出す** (#5788)。生の数字ではなく種目に
  // 応じた名前にする——同じ 2 でもホールデムはフロップ、スタッドは 4th street。
  const roundKeys = COMMUNITY_DISCIPLINES.has(state.disciplineName) ? COMMUNITY_ROUND_KEYS : STUD_ROUND_KEYS;
  const roundKey: string | undefined = roundKeys[state.tablePhase];
  const roundLabel = roundKey ? t(`round.${roundKey}`) : '';

  return (
    <GamePageShell
      title={tc('nav.horse')}
      gameThemeBg={gameTheme.horse.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/horse"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
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
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            <div
              className="text-center text-sm text-ds-text-muted"
              data-testid="ho-discipline"
              data-tutorial="ho-discipline"
              // **手番かどうかを DOM に出す。** 出さないと E2E は「押しても
              // 何も起きない」を待つことになり、配り次第で落ちる。
              data-human-turn={isHumanTurn || undefined}
            >
              <span className="mr-2 text-lg text-ds-text-primary" data-testid="ho-letter">
                {state.disciplineLetter}
              </span>
              <span className="mr-3 text-ds-text-primary">{disciplineName}</span>
              <span className="mr-3">
                {t('hand', { n: state.handInDiscipline, total: state.config.handsPerDiscipline })}
              </span>
              {roundLabel && (
                <span className="mr-3 text-ds-accent" data-testid="ho-round">
                  {roundLabel}
                </span>
              )}
              <span className="mr-3">{t('handTotal', { n: state.handNumber })}</span>
              <span data-testid="ho-pot">{t('pot', { pot: state.pot })}</span>
            </div>

            {state.communityCards.length > 0 && (
              <div className="flex flex-col items-center gap-1" data-testid="ho-community">
                <div className="text-xs text-ds-text-muted">{t('community')}</div>
                <div className="flex flex-wrap justify-center gap-2">
                  {state.communityCards.map((card) => (
                    <AnimatedCard key={`${card.design}-${card.value}`} card={card} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            <div className="flex flex-wrap justify-center gap-4" data-testid="ho-seats" data-tutorial="ho-seats">
              {state.seats.map((s) => (
                <div key={s.id} className="text-center text-xs text-ds-text-muted" data-testid={`ho-seat-${s.id}`}>
                  <div className={s.id === state.currentTurn ? 'text-ds-warning' : 'text-ds-text-primary'}>
                    {s.isHuman ? t('you') : s.name}
                  </div>
                  <div data-testid={`ho-seat-${s.id}-chips`}>{t('chips', { chips: s.chips })}</div>
                  {/* **見えている札だけが届く。** CPU の伏せ札はサーバが返さない。 */}
                  <div className="mt-1 flex justify-center gap-1">
                    {s.cards.map((card) => (
                      <AnimatedCard
                        key={`${card.design}-${card.value}`}
                        card={card}
                        width={Math.round(cardWidth * 0.8)}
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            {isGameEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="ho-result">
                <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
                <div className="text-ds-success mb-1">
                  {t('result.winner', {
                    name: state.seats[state.winnerSeat]?.isHuman
                      ? t('you')
                      : (state.seats[state.winnerSeat]?.name ?? ''),
                  })}
                </div>
                {state.seats.map((s) => (
                  <div key={s.id}>{t('result.chips', { name: s.isHuman ? t('you') : s.name, chips: s.chips })}</div>
                ))}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <ErrorAlert message={error} onRetry={retry} />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'seats',
                    label: t('settings.seats'),
                    value: String(seats),
                    options: SEAT_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                    onSelect: (v: string) => setSeats(Number.parseInt(v, 10)),
                  },
                  {
                    type: 'select' as const,
                    id: 'handsPerDiscipline',
                    label: t('settings.handsPerDiscipline'),
                    value: String(handsPerDiscipline),
                    options: HANDS_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                    onSelect: (v: string) => setHandsPerDiscipline(Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.horse.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="ho-actions">
              {isHumanTurn && (
                <BettingControls
                  inputId="ho-bet"
                  betAmount={betAmount}
                  onBetAmountChange={setBetAmount}
                  minRaise={Math.max(state.minRaise, 1)}
                  maxBetAmount={state.seats[state.humanSeat]?.chips ?? 0}
                  potSize={state.pot}
                  // **「賭けられているか」はサーバが決める。** 固定すると、
                  // チェックできる場面でチェックが出ず、逆も起きる。
                  hasOutstandingBet={state.toCall > 0}
                  loading={loading}
                  onCall={() => callApi('action', { action: 'call' })}
                  onRaise={() => callApi('action', { action: 'raise', amount: betAmount })}
                  onBet={() => callApi('action', { action: 'bet', amount: betAmount })}
                  onCheck={() => callApi('action', { action: 'check' })}
                  onFold={() => callApi('action', { action: 'fold' })}
                  onAllIn={() => callApi('action', { action: 'allin' })}
                />
              )}
              {isHandEnd && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => callApi('next')}
                  disabled={loading}
                  data-testid="ho-next-hand"
                >
                  {t('nextHand')}
                </button>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={resetWithConfig} disabled={loading}>
                  {t('newGame')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={resetWithConfig}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ho-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
