import { useCallback, useMemo, useState } from 'react';
import { chemindeferApi } from '../api/gameApi';
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
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ChemindeFerResponse } from '../types/card';
import { CHEMIN_DE_FER_HUMAN_SEAT, CHEMIN_DE_FER_RESULT } from '../types/games/chemindefer';
import { ChemindeFerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CHEMINDEFER_CLI_HELP, parseChemindeFerCommand } from '../utils/cli/commands/chemindeferCommands';
import { formatChemindeFerState } from '../utils/cli/formatters/chemindeferFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const CDF_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="cdf-seats"]', messageKey: 'tutorial.seats', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cdf-table"]', messageKey: 'tutorial.draw', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cdf-controls"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
];

/** Renders the Chemin de Fer game page (#5259). */
export const ChemindeFerPage = withTutorial(ChemindeFerPageContent, 'chemindefer', CDF_TUTORIAL_STEPS);

function ChemindeFerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('chemindefer');

  const [stakeAmount, setStakeAmount] = useState(100);
  const [betAmount, setBetAmount] = useState(50);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(chemindeferApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('chemindefer');
  const cliConfig: CliGameConfig<ChemindeFerResponse, Parameters<typeof chemindeferApi.exec>> = useMemo(
    () => ({
      gameName: 'chemindefer',
      parseCommand: parseChemindeFerCommand,
      formatResponse: formatChemindeFerState,
      helpText: CHEMINDEFER_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  // **どの操作が出せるかはサーバの状態だけで決める。** 規則をページ側で組み直すと
  // 必ずドメインとずれる。とくに子の引き方は 5 のときしか選べないので、
  // `punterMayChoose` をそのまま使う (合計から自分で判定し直さない)。
  const isMyTurn = !!state?.isHumanTurn;
  const phase = state?.phase;
  const canStake = isMyTurn && phase === ChemindeFerPhase.STAKE;
  const canBet = isMyTurn && phase === ChemindeFerPhase.BET;
  const canPunterDecide = isMyTurn && phase === ChemindeFerPhase.PUNTER_DRAW && !!state?.punterMayChoose;
  const canBankerDecide = isMyTurn && phase === ChemindeFerPhase.BANKER_DRAW;
  const isRoundEnd = phase === ChemindeFerPhase.ROUND_END;
  const gameOver = !!state?.gameEndFlag;

  const handleStake = useCallback(() => execApi('stake', { stake: stakeAmount }), [execApi, stakeAmount]);
  const handleBet = useCallback(() => execApi('bet', { amount: betAmount }), [execApi, betAmount]);
  // **降りるのは賭け額 0 を送ること。** 送らないのとは違う。
  const handlePassBet = useCallback(() => execApi('bet', { amount: 0 }), [execApi]);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: () => execApi(canPunterDecide ? 'pd' : 'bd'), enabled: canPunterDecide || canBankerDecide },
      { key: 's', action: () => execApi(canPunterDecide ? 'ps' : 'bs'), enabled: canPunterDecide || canBankerDecide },
      { key: 'n', action: () => execApi('next'), enabled: isRoundEnd && !gameOver },
      { key: 'g', action: () => execApi('giveup'), enabled: !gameOver },
    ],
    [execApi, canPunterDecide, canBankerDecide, isRoundEnd, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('chemindefer', state);

  if (!state) return <GameSkeleton gameKey="chemindefer" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [ChemindeFerPhase.STAKE]: t('phase.stake'),
      [ChemindeFerPhase.BET]: t('phase.bet'),
      [ChemindeFerPhase.PUNTER_DRAW]: t('phase.punterDraw'),
      [ChemindeFerPhase.BANKER_DRAW]: t('phase.bankerDraw'),
      [ChemindeFerPhase.ROUND_END]: t('phase.roundEnd'),
    }[state.phase] ?? '';

  const resultName =
    {
      [CHEMIN_DE_FER_RESULT.none]: '',
      [CHEMIN_DE_FER_RESULT.banker]: t('result.banker'),
      [CHEMIN_DE_FER_RESULT.punter]: t('result.punter'),
      [CHEMIN_DE_FER_RESULT.tie]: t('result.tie'),
    }[state.result] ?? '';

  const me = state.players.find((p) => p.id === CHEMIN_DE_FER_HUMAN_SEAT);
  const iAmBanker = state.bankerIdx === CHEMIN_DE_FER_HUMAN_SEAT;

  const handSide = (label: string, cards: ChemindeFerResponse['bankerHand'], total: number, testId: string) => (
    <div className="flex-1 min-w-0">
      <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{label}</div>
      <div className="flex justify-center gap-1 flex-wrap" data-testid={testId}>
        {cards.length === 0 ? (
          <span className="text-ds-text-muted text-xs">—</span>
        ) : (
          cards.map((card, i) => (
            <AnimatedCard key={`${testId}-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
          ))
        )}
      </div>
      {cards.length > 0 && (
        <div className="text-ds-text-primary text-center text-lg font-bold mt-1" data-testid={`${testId}-total`}>
          {total}
        </div>
      )}
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.chemindefer')}
      gameThemeBg={gameTheme.chemindefer.bg}
      phaseName={phaseName}
      gamePath="/chemindefer"
      gameEndFlag={gameOver}
      isHumanTurn={isMyTurn}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="cdf-my-chips">
            {t('label.chips')}: {me?.chips ?? 0}
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

            <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="cdf-bank-line">
              {t('label.round')}: {state.roundNumber}
              {state.config ? ` / ${state.config.rounds}` : ''} · {t('label.banker')}: #{state.bankerIdx} ·{' '}
              {t('label.stake')}: {state.stake}
            </div>

            {state.stake > 0 && (
              <div className="text-ds-text-muted text-center text-xs mb-3" data-testid="cdf-bet-line">
                {t('label.totalBet')}: {state.totalBet} · {t('label.remaining')}: {state.remainingStake}
                {state.representativeIdx >= 0 && <> · {`${t('label.representative')}: #${state.representativeIdx}`}</>}
              </div>
            )}

            <div className="flex gap-4 mb-4 flex-wrap sm:flex-nowrap" data-tutorial="cdf-table">
              {handSide(t('label.punterSide'), state.punterHand, state.punterTotal, 'cdf-punter-hand')}
              {handSide(t('label.banker'), state.bankerHand, state.bankerTotal, 'cdf-banker-hand')}
            </div>

            {state.phase === ChemindeFerPhase.PUNTER_DRAW && !state.punterMayChoose && (
              <p className="text-ds-text-muted text-center text-sm mb-2" data-testid="cdf-forced-notice">
                {t('forcedNotice')}
              </p>
            )}

            {resultName && (
              <div className="text-ds-text-primary text-center text-base font-bold mb-2" data-testid="cdf-result">
                {resultName}
              </div>
            )}

            <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 mb-4" data-tutorial="cdf-seats">
              {state.players.map((p) => (
                <div
                  key={`seat-${p.id}`}
                  data-testid={`cdf-seat-${p.id}`}
                  className={`rounded border px-2 py-1 text-xs ${
                    p.isBanker ? 'border-ds-warning' : 'border-ds-border'
                  } ${p.id === state.betTurn ? 'ring-2 ring-ds-success' : ''}`}
                >
                  <div className="font-bold text-ds-text-primary">
                    {p.isHuman ? t('label.you') : p.name}
                    {p.isBanker && ' ★'}
                    {p.isRepresentative && ' ◆'}
                  </div>
                  <div className="text-ds-text-muted">
                    {t('label.chips')}: {p.chips}
                    {p.bet > 0 && ` (${t('label.bet')}: ${p.bet})`}
                  </div>
                </div>
              ))}
            </div>

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.chemindefer.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="cdf-controls">
              {canStake && (
                <>
                  <p className="text-ds-text-muted text-sm">
                    {t('stakeGuide', { min: state.stakeMin, max: state.stakeMax })}
                  </p>
                  <ChipBetInput
                    id="chemindefer-stake"
                    label={t('label.amount')}
                    value={stakeAmount}
                    onChange={setStakeAmount}
                    max={state.stakeMax}
                  />
                  <button type="button" className={btnPrimary} onClick={handleStake} disabled={loading}>
                    {t('button.stake')}
                  </button>
                </>
              )}

              {canBet && (
                <>
                  <p className="text-ds-text-muted text-sm">{t('betGuide', { max: state.betMax })}</p>
                  <ChipBetInput
                    id="chemindefer-bet"
                    label={t('label.amount')}
                    value={betAmount}
                    onChange={setBetAmount}
                    max={state.betMax}
                  />
                  <div className="flex gap-2">
                    <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                      {t('button.bet')}
                    </button>
                    <button type="button" className={btnSecondary} onClick={handlePassBet} disabled={loading}>
                      {t('button.pass')}
                    </button>
                  </div>
                </>
              )}

              {(canPunterDecide || canBankerDecide) && (
                <>
                  <p className="text-ds-text-muted text-sm">{canPunterDecide ? t('punterGuide') : t('bankerGuide')}</p>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className={btnSuccess}
                      data-hint-action="draw"
                      onClick={() => execApi(canPunterDecide ? 'pd' : 'bd')}
                      disabled={loading}
                    >
                      {t('button.draw')}
                    </button>
                    <button
                      type="button"
                      className={btnSecondary}
                      onClick={() => execApi(canPunterDecide ? 'ps' : 'bs')}
                      disabled={loading}
                    >
                      {t('button.stand')}
                    </button>
                  </div>
                </>
              )}

              {isRoundEnd && !gameOver && (
                <div className="flex gap-2">
                  <button type="button" className={btnPrimary} onClick={() => execApi('next')} disabled={loading}>
                    {t('button.next')}
                  </button>
                  {iAmBanker && (
                    <button type="button" className={btnSecondary} onClick={() => execApi('pb')} disabled={loading}>
                      {t('button.passBank')}
                    </button>
                  )}
                </div>
              )}

              {!isMyTurn && !isRoundEnd && !gameOver && (
                <p className="text-ds-text-muted text-sm" data-testid="cdf-wait-notice">
                  {t('waitNotice')}
                </p>
              )}

              <div className="flex gap-2">
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('button.actionLog')}
                </button>
                <button type="button" className={btnWarning} onClick={() => execApi('giveup')} disabled={gameOver}>
                  {t('button.giveUp')}
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
