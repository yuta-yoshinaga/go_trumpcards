import { useCallback, useMemo, useState } from 'react';
import { banluckApi } from '../api/gameApi';
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
import type { BanLuckResponse } from '../types/card';
import { BAN_LUCK_RANK } from '../types/games/banluck';
import { BanLuckPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BANLUCK_CLI_HELP, parseBanLuckCommand } from '../utils/cli/commands/banluckCommands';
import { formatBanLuckState } from '../utils/cli/formatters/banluckFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const BL_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bl-bet"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bl-seats"]', messageKey: 'tutorial.seats', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bl-banker"]', messageKey: 'tutorial.banker', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Ban Luck game page (#5263). */
export const BanLuckPage = withTutorial(BanLuckPageContent, 'banluck', BL_TUTORIAL_STEPS);

/** Rank key lookup for the result label. */
const rankKeyOf = (r: number) => Object.entries(BAN_LUCK_RANK).find(([, v]) => v === r)?.[0] ?? 'bust';

function BanLuckPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('banluck');

  const [bet, setBet] = useState(50);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(banluckApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('banluck');
  const cliConfig: CliGameConfig<BanLuckResponse, Parameters<typeof banluckApi.exec>> = useMemo(
    () => ({
      gameName: 'banluck',
      parseCommand: parseBanLuckCommand,
      formatResponse: formatBanLuckState,
      helpText: BANLUCK_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetPhase = phase === BanLuckPhase.BET;
  const isPlayPhase = phase === BanLuckPhase.PLAY;
  const isRoundEnd = phase === BanLuckPhase.ROUND_END;
  const gameOver = !!state?.gameEndFlag;
  const humanIsBanker = !!state && state.bankerSeat === state.humanSeat;

  // **親のラウンドは 0 を送る。** 送らないのとは違う (サーバが 400 を返す)。
  const handleDeal = useCallback(() => execApi('bet', { bet: humanIsBanker ? 0 : bet }), [execApi, bet, humanIsBanker]);

  const canAct = !!state?.isHumanTurn && isPlayPhase;
  const actionBindings = useMemo(
    () => [
      { key: 'h', action: () => execApi('hit'), enabled: canAct },
      // **義務があるときは止められない。** 鍵盤からも塞ぐ。
      { key: 's', action: () => execApi('stand'), enabled: canAct && !state?.mustHit },
      { key: 'n', action: () => execApi('next'), enabled: isRoundEnd && !gameOver },
    ],
    [execApi, canAct, state?.mustHit, isRoundEnd, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('banluck', state);

  if (!state) return <GameSkeleton gameKey="banluck" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [BanLuckPhase.BET]: t('phase.bet'),
      [BanLuckPhase.PLAY]: t('phase.play'),
      [BanLuckPhase.ROUND_END]: t('phase.roundEnd'),
      [BanLuckPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const human = state.seats[state.humanSeat];
  const humanWon = gameOver && state.winnerSeat === state.humanSeat;

  return (
    <GamePageShell
      title={tc('nav.banluck')}
      gameThemeBg={gameTheme.banluck.bg}
      phaseName={phaseName}
      gamePath="/banluck"
      gameEndFlag={gameOver}
      isHumanTurn={state.isHumanTurn}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="bl-chips">
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

            <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="bl-round-line">
              {/* **親が全席を一巡するまで何ラウンドあるか**は CUI にだけ出て
                  いた (#5778)。config が来る前の初回描画でも壊れないよう、
                  総数が読めないうちは番号だけ出す。 */}
              {state.config
                ? t('label.roundOf', { n: state.roundNumber, total: state.config.rounds })
                : `${t('label.round')}: ${state.roundNumber}`}
              {' · '}
              <span data-testid="bl-banker">
                {t('label.banker')}: {state.seats[state.bankerSeat]?.name ?? '?'}
              </span>
            </div>

            {/* **義務は名指しで出す。** 押せない理由が読めないと規則が伝わらない。 */}
            {state.mustHit && (
              <p className="text-ds-text-muted text-center text-xs mb-2" data-testid="bl-must-hit">
                {t('mustHitNotice')}
              </p>
            )}

            <div data-tutorial="bl-seats">
              {state.seats.map((seat, i) => (
                <div
                  key={`seat-${seat.name}-${i}`}
                  data-testid={`bl-seat-${i}`}
                  className={`mb-2 rounded px-2 py-1 ${seat.isTurn ? 'ring-2 ring-ds-success' : ''}`}
                >
                  <div
                    className="text-ds-text-primary text-center text-sm font-bold mb-1"
                    {...(seat.isBanker ? { 'data-tutorial': 'bl-banker' } : {})}
                  >
                    {seat.name}
                    {seat.isBanker && <span data-testid={`bl-banker-mark-${i}`}> · {t('label.banker')}</span>}
                    {' · '}
                    {t('label.chips')} {seat.chips}
                    {seat.bet > 0 && ` · ${t('label.bet')} ${seat.bet}`}
                  </div>
                  <div className="flex justify-center gap-1 flex-wrap">
                    {seat.cards.map((card, k) => (
                      <AnimatedCard key={`s${i}-${card.design}-${card.value}-${k}`} card={card} width={cardWidth} />
                    ))}
                  </div>
                  {seat.cards.length > 0 && (
                    <div className="text-ds-text-primary text-center text-sm mt-1">
                      {seat.score} {t('label.score')}
                      {(isRoundEnd || gameOver) && (
                        <span data-testid={`bl-result-${i}`}>
                          {' · '}
                          {t(`rank.${rankKeyOf(seat.rank)}`)}
                          {' · '}
                          {t('label.delta')} {seat.delta}
                        </span>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {gameOver && (
              <div className="text-ds-text-primary text-center text-sm font-bold" data-testid="bl-winner">
                {t('label.winner')}: {state.seats[state.winnerSeat]?.name ?? '?'}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.banluck.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2">
              {isBetPhase && !gameOver && (
                <div className="flex flex-col items-center gap-2" data-tutorial="bl-bet">
                  <p className="text-ds-text-muted text-sm">{t('betGuide')}</p>
                  {/* **親は賭けない。** 入力そのものを出さず、そう明示する。 */}
                  {humanIsBanker ? (
                    <p className="text-ds-text-muted text-xs" data-testid="bl-banker-notice">
                      {t('bankerBetNotice')}
                    </p>
                  ) : (
                    <ChipBetInput
                      id="banluck-bet"
                      label={t('label.bet')}
                      value={bet}
                      onChange={setBet}
                      max={human?.chips ?? 0}
                    />
                  )}
                  <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
                    {t('button.deal')}
                  </button>
                </div>
              )}

              {isPlayPhase && state.isHumanTurn && (
                <>
                  <p className="text-ds-text-muted text-sm">{t('playGuide')}</p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    <button
                      type="button"
                      className={btnSuccess}
                      data-hint-action="hit"
                      onClick={() => execApi('hit')}
                      disabled={loading}
                    >
                      {t('button.hit')}
                    </button>
                    {/* **止まれるかどうかはサーバが決める。** 点数から計算し直さない。 */}
                    {!state.mustHit && (
                      <button
                        type="button"
                        className={btnSecondary}
                        data-testid="bl-stand"
                        data-hint-action="stand"
                        onClick={() => execApi('stand')}
                        disabled={loading}
                      >
                        {t('button.stand')}
                      </button>
                    )}
                  </div>
                </>
              )}

              {isRoundEnd && !gameOver && (
                <button type="button" className={btnPrimary} onClick={() => execApi('next')} disabled={loading}>
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
