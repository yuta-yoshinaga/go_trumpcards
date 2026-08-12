import { useCallback, useMemo, useState } from 'react';
import { crazyfourpokerApi } from '../api/gameApi';
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
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CrazyFourPokerResponse } from '../types/card';
import { CRAZY_FOUR_POKER_ANTE_UNIT, CRAZY_FOUR_POKER_RESULT } from '../types/games/crazyfourpoker';
import { CrazyFourPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CRAZYFOURPOKER_CLI_HELP, parseCrazyFourPokerCommand } from '../utils/cli/commands/crazyfourpokerCommands';
import { formatCrazyFourPokerState } from '../utils/cli/formatters/crazyfourpokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const C4P_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="c4p-bet"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="c4p-hand"]', messageKey: 'tutorial.hand', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="c4p-actions"]', messageKey: 'tutorial.multiplier', placement: 'top', advanceOn: 'next' },
];

/** Renders the Crazy 4 Poker game page (#5260). */
export const CrazyFourPokerPage = withTutorial(CrazyFourPokerPageContent, 'crazyfourpoker', C4P_TUTORIAL_STEPS);

function CrazyFourPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('crazyfourpoker');

  const [ante, setAnte] = useState(50);
  const [queensUp, setQueensUp] = useState(0);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(crazyfourpokerApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('crazyfourpoker');
  const cliConfig: CliGameConfig<CrazyFourPokerResponse, Parameters<typeof crazyfourpokerApi.exec>> = useMemo(
    () => ({
      gameName: 'crazyfourpoker',
      parseCommand: parseCrazyFourPokerCommand,
      formatResponse: formatCrazyFourPokerState,
      helpText: CRAZYFOURPOKER_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetPhase = phase === CrazyFourPokerPhase.BET;
  const isDecidePhase = phase === CrazyFourPokerPhase.DECIDE;
  const isResultPhase = phase === CrazyFourPokerPhase.RESULT;
  const gameOver = !!state?.gameEndFlag;

  const handleDeal = useCallback(() => execApi('bet', { ante, queensUp }), [execApi, ante, queensUp]);

  // **置ける倍率はサーバが決める。** 手役から計算し直すと、このゲームの本体である
  // 「3 倍はエースのペア以上だけ」という規則が 2 か所に増えてずれる。
  const multipliers = useMemo(() => {
    const max = state?.maxMultiplier ?? 1;
    return Array.from({ length: max }, (_, i) => i + 1);
  }, [state?.maxMultiplier]);

  const actionBindings = useMemo(
    () => [
      { key: 'f', action: () => execApi('fold'), enabled: isDecidePhase },
      { key: 'n', action: () => execApi('next'), enabled: isResultPhase && !gameOver },
    ],
    [execApi, isDecidePhase, isResultPhase, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('crazyfourpoker', state);

  if (!state) return <GameSkeleton gameKey="crazyfourpoker" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [CrazyFourPokerPhase.BET]: t('phase.bet'),
      [CrazyFourPokerPhase.DECIDE]: t('phase.decide'),
      [CrazyFourPokerPhase.RESULT]: t('phase.result'),
    }[state.phase] ?? '';

  const resultKey = Object.entries(CRAZY_FOUR_POKER_RESULT).find(([, v]) => v === state.result)?.[0] ?? 'none';
  const staked = state.anteBet + state.superBet + state.queensUpBet + state.playBet;
  const net = state.payout - staked;
  const won = state.result === CRAZY_FOUR_POKER_RESULT.win;

  const handRow = (label: string, cards: CrazyFourPokerResponse['playerHand'], testId: string) => (
    <div className="mb-2">
      <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{label}</div>
      <div className="flex justify-center gap-1 flex-wrap" data-testid={testId}>
        {cards.map((card, i) => (
          <AnimatedCard key={`${testId}-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
        ))}
      </div>
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.crazyfourpoker')}
      gameThemeBg={gameTheme.crazyfourpoker.bg}
      phaseName={phaseName}
      gamePath="/crazyfourpoker"
      gameEndFlag={gameOver}
      winShow={isResultPhase && won}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="c4p-chips">
            {t('label.chips')}: {state.chips}
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

            <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="c4p-bet-line">
              {t('label.round')}: {state.roundNumber}
              {state.anteBet > 0 && (
                <>
                  {' · '}
                  {t('label.ante')}: {state.anteBet} · {t('label.superBonus')}: {state.superBet} · {t('label.queensUp')}
                  : {state.queensUpBet}
                </>
              )}
            </div>

            {state.playerHand.length > 0 && (
              <div data-tutorial="c4p-hand">
                {handRow(t('label.yourHand'), state.playerHand, 'c4p-player-hand')}
                <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="c4p-player-rank">
                  {t('label.yourBest')}: {t(`rank.${state.playerHandRank}`)}
                </div>
              </div>
            )}

            {/* **決着まではディーラーの手を出さない。** サーバも送っていない。 */}
            {isResultPhase && state.dealerHand.length > 0 ? (
              <>
                {handRow(t('label.dealerHand'), state.dealerHand, 'c4p-dealer-hand')}
                <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="c4p-dealer-rank">
                  {t('label.dealerBest')}: {t(`rank.${state.dealerHandRank}`)}
                  {!state.dealerQualifies && ` · ${t('result.dealerNotQualified')}`}
                </div>
              </>
            ) : (
              state.playerHand.length > 0 && (
                <p className="text-ds-text-muted text-center text-xs mb-2" data-testid="c4p-dealer-hidden">
                  {t('label.hidden')}
                </p>
              )
            )}

            {isResultPhase && (
              <div className="text-center mb-2" data-testid="c4p-result">
                <div className="text-ds-text-primary text-base font-bold">{t(`result.${resultKey}`)}</div>
                <div className={`text-sm font-medium ${net >= 0 ? 'text-ds-success' : 'text-ds-error'}`}>
                  {t('label.net')}: {net}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.crazyfourpoker.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="c4p-actions">
              {isBetPhase && !gameOver && (
                <div className="flex flex-col items-center gap-2" data-tutorial="c4p-bet">
                  <p className="text-ds-text-muted text-sm">{t('betGuide')}</p>
                  <ChipBetInput
                    id="crazyfourpoker-ante"
                    label={t('label.ante')}
                    value={ante}
                    onChange={setAnte}
                    max={state.chips}
                    step={CRAZY_FOUR_POKER_ANTE_UNIT}
                  />
                  {/* **アンティと Super Bonus を引いた残りしか置けない。** 上限を
                      チップ全額にすると、合計が持ち金を超える組み合わせを選べてしまう。 */}
                  <ChipBetInput
                    id="crazyfourpoker-queensup"
                    label={t('label.queensUp')}
                    value={queensUp}
                    onChange={setQueensUp}
                    max={Math.max(0, state.chips - ante * 2)}
                    step={CRAZY_FOUR_POKER_ANTE_UNIT}
                  />
                  <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
                    {t('button.deal')}
                  </button>
                </div>
              )}

              {isDecidePhase && (
                <>
                  <p className="text-ds-text-muted text-sm">{t('decideGuide')}</p>
                  <p
                    className={`text-sm font-bold ${state.hasAcesOrBetter ? 'text-ds-success' : 'text-ds-text-muted'}`}
                    data-testid="c4p-multiplier-notice"
                  >
                    {state.hasAcesOrBetter ? t('acesNotice') : t('normalNotice')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    {multipliers.map((m) => (
                      <button
                        key={`mult-${m}`}
                        type="button"
                        className={btnPrimary}
                        data-hint-action={m === state.maxMultiplier && state.hasAcesOrBetter ? 'raise' : 'play'}
                        data-testid={`c4p-play-${m}`}
                        onClick={() => execApi('play', { multiplier: m })}
                        disabled={loading}
                      >
                        {t('button.play', { multiplier: m })}
                      </button>
                    ))}
                    <button type="button" className={btnWarning} onClick={() => execApi('fold')} disabled={loading}>
                      {t('button.fold')}
                    </button>
                  </div>
                </>
              )}

              {isResultPhase && !gameOver && (
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
