import { useCallback, useMemo, useState } from 'react';
import { kingoApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
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
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KingoResponse } from '../types/card';
import { KingoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KINGO_CLI_HELP, parseKingoCommand } from '../utils/cli/commands/kingoCommands';
import { formatKingoState } from '../utils/cli/formatters/kingoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const KINGO_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="kingo-seats"]', messageKey: 'tutorial.rule', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="kingo-actions"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="kingo-banker"]', messageKey: 'tutorial.banker', placement: 'bottom', advanceOn: 'next' },
];

/** Rank keys, matching the Go domain's order. */
const RANK_KEYS = ['rank.none', 'rank.pair', 'rank.arashi'] as const;

/** Renders the Kingo game page (#5282). */
export const KingoPage = withTutorial(KingoPageContent, 'kingo', KINGO_TUTORIAL_STEPS);

function KingoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('kingo');

  const [amount, setAmount] = useState(10);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(kingoApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('kingo');
  const cliConfig: CliGameConfig<KingoResponse, Parameters<typeof kingoApi.exec>> = useMemo(
    () => ({
      gameName: 'kingo',
      parseCommand: parseKingoCommand,
      formatResponse: formatKingoState,
      helpText: KINGO_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetting = phase === KingoPhase.BET;
  const isResult = phase === KingoPhase.RESULT;
  const gameOver = !!state?.gameEndFlag;
  // **親か子かで求められる操作が変わる。** サーバの判定に従う。
  const isBanker = !!state?.isHumanBanker;
  const canAct = isBetting && !gameOver;

  const handleBet = useCallback(() => execApi('bet', { amount }), [execApi, amount]);

  const actionBindings = useMemo(
    () => [{ key: 'n', action: () => execApi('next'), enabled: isResult && !gameOver }],
    [execApi, isResult, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('kingo', state);

  if (!state) return <GameSkeleton gameKey="kingo" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [KingoPhase.BET]: t('phase.bet'),
      [KingoPhase.RESULT]: t('phase.result'),
      [KingoPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const human = state.seats[state.humanSeat];
  const humanWon = gameOver && state.winnerSeat === state.humanSeat;
  const minBet = state.config?.minBet ?? 10;

  return (
    <GamePageShell
      title={tc('nav.kingo')}
      gameThemeBg={gameTheme.kingo.bg}
      phaseName={phaseName}
      gamePath="/kingo"
      gameEndFlag={gameOver}
      isHumanTurn={state.isHumanTurn}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="kingo-chips">
            {t('label.chips')}: {human?.chips ?? 0}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div data-testid="card-area" className={`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="kingo-round">
              {t('label.round', { round: state.roundNumber, total: state.rounds })}
            </div>

            <div className="text-ds-text-primary text-center text-sm mb-1" data-tutorial="kingo-banker">
              <span data-testid="kingo-banker">
                {t('label.banker')}: {state.seats[state.bankerSeat]?.name ?? '?'}
                {isBanker && ` · ${t('label.youAreBanker')}`}
              </span>
            </div>

            {/* **配当はサーバが送ってくる。** 画面で倍率を持たない。 */}
            <p className="text-ds-text-muted text-center text-xs mb-1" data-testid="kingo-payouts">
              {t('label.payouts', { arashi: state.payoutArashi, pair: state.payoutPair })}
            </p>
            <p className="text-ds-text-muted text-center text-xs mb-3" data-testid="kingo-notice">
              {t('notice')}
            </p>

            <div data-tutorial="kingo-seats">
              {state.seats.map((seat, i) => (
                <div
                  key={`seat-${seat.name}-${i}`}
                  data-testid={`kingo-seat-${i}`}
                  className={`mb-2 rounded px-2 py-1 text-center text-sm ${
                    seat.isBanker ? 'ring-2 ring-ds-warning' : ''
                  }`}
                >
                  <span className="text-ds-text-primary">
                    {seat.name}
                    {seat.isBanker && ` · ${t('label.banker')}`}
                    {' · '}
                    {t('label.chips')} {seat.chips}
                    {/* **親は張らないので額を出さない。** */}
                    {!seat.isBanker && seat.bet > 0 && ` · ${t('label.bet')} ${seat.bet}`}
                    {seat.wonAmount !== 0 && (
                      <span
                        data-testid={`kingo-won-${i}`}
                        className={seat.wonAmount > 0 ? 'text-ds-success' : 'text-ds-danger'}
                      >
                        {' · '}
                        {t('label.won', { amount: seat.wonAmount })}
                      </span>
                    )}
                  </span>
                  {/*
                    **配る前は手札が存在しない。** 隠しているのではなく、まだ
                    無い ── キンゴに「自分だけ見える手札」は無い。
                  */}
                  <div className="flex justify-center gap-1 flex-wrap mt-1" data-testid={`kingo-cards-${i}`}>
                    {seat.cards.length > 0 ? (
                      seat.cards.map((card, k) => (
                        <AnimatedCard
                          key={`s${i}-${card.design}-${card.value}-${k}`}
                          card={card}
                          width={Math.round(cardWidth * 0.8)}
                        />
                      ))
                    ) : (
                      <span className="text-ds-text-muted text-xs">{t('label.noCards')}</span>
                    )}
                  </div>
                  {seat.cards.length > 0 && (
                    <div className="text-ds-text-primary text-xs mt-1" data-testid={`kingo-rank-${i}`}>
                      {t(RANK_KEYS[seat.rank] ?? 'rank.none')}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {gameOver && (
              <div className="text-ds-text-primary text-center text-sm font-bold mt-2" data-testid="kingo-winner">
                {t('label.winner')}: {state.seats[state.winnerSeat]?.name ?? '?'}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.kingo.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="kingo-actions">
              {/*
                **親と子で出す操作を変える。** どちらもサーバの `isHumanBanker`
                に従う ── ページでフェーズから割り出すと、親のときに張りの
                ボタンが出て、押しても通らない操作を見せることになる。
              */}
              {canAct &&
                (isBanker ? (
                  <>
                    <p className="text-ds-text-muted text-sm" data-testid="kingo-deal-guide">
                      {t('label.dealPrompt')}
                    </p>
                    <button
                      type="button"
                      className={btnSuccess}
                      data-testid="kingo-deal"
                      data-hint-action="deal"
                      onClick={() => execApi('deal')}
                      disabled={loading}
                    >
                      {t('button.deal')}
                    </button>
                  </>
                ) : (
                  <>
                    <p className="text-ds-text-muted text-sm" data-testid="kingo-bet-guide">
                      {t('label.betPrompt', { min: minBet })}
                    </p>
                    <button
                      type="button"
                      className={btnSuccess}
                      data-testid="kingo-bet"
                      data-hint-action="bet"
                      onClick={handleBet}
                      disabled={loading}
                    >
                      {t('button.bet')}
                    </button>
                    <ChipBetInput
                      id="kingo-amount"
                      label={t('label.bet')}
                      value={amount}
                      onChange={setAmount}
                      max={human?.chips ?? 0}
                    />
                  </>
                ))}

              {isResult && !gameOver && (
                <button
                  type="button"
                  className={btnPrimary}
                  data-testid="kingo-next"
                  data-hint-action="next"
                  onClick={() => execApi('next')}
                  disabled={loading}
                >
                  {t('button.next')}
                </button>
              )}

              <div className="flex gap-2">
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('button.actionLog')}
                </button>
                <GameResetButton
                  isGameEnd={gameOver}
                  onReset={() => execApi('reset')}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
              </div>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
