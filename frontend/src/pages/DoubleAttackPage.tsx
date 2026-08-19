import { useCallback, useMemo, useState } from 'react';
import { doubleattackApi } from '../api/gameApi';
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
import type { DoubleAttackResponse } from '../types/card';
import { DOUBLE_ATTACK_RESULT } from '../types/games/doubleattack';
import { DoubleAttackPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DOUBLEATTACK_CLI_HELP, parseDoubleAttackCommand } from '../utils/cli/commands/doubleattackCommands';
import { formatDoubleAttackState } from '../utils/cli/formatters/doubleattackFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const DA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="da-bet"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="da-dealer"]', messageKey: 'tutorial.attack', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="da-hands"]', messageKey: 'tutorial.hands', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Extra Bet Blackjack game page (#5261). */
export const DoubleAttackPage = withTutorial(DoubleAttackPageContent, 'doubleattack', DA_TUTORIAL_STEPS);

function DoubleAttackPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('doubleattack');

  const [ante, setAnte] = useState(50);
  const [bustIt, setBustIt] = useState(0);
  const [attack, setAttack] = useState(0);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(doubleattackApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('doubleattack');
  const cliConfig: CliGameConfig<DoubleAttackResponse, Parameters<typeof doubleattackApi.exec>> = useMemo(
    () => ({
      gameName: 'doubleattack',
      parseCommand: parseDoubleAttackCommand,
      formatResponse: formatDoubleAttackState,
      helpText: DOUBLEATTACK_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetPhase = phase === DoubleAttackPhase.BET;
  const isAttackPhase = phase === DoubleAttackPhase.ATTACK;
  const isPlayPhase = phase === DoubleAttackPhase.PLAY;
  const isResultPhase = phase === DoubleAttackPhase.RESULT;
  const gameOver = !!state?.gameEndFlag;

  const handleDeal = useCallback(() => execApi('bet', { ante, bustIt }), [execApi, ante, bustIt]);
  const handleAttack = useCallback(() => execApi('attack', { amount: attack }), [execApi, attack]);
  // **見送りは 0 を送ること。** 送らないのとは違う (サーバは 400 を返す)。
  const handleDecline = useCallback(() => execApi('attack', { amount: 0 }), [execApi]);

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: () => execApi('hit'), enabled: isPlayPhase },
      { key: 's', action: () => execApi('stand'), enabled: isPlayPhase },
      { key: 'n', action: () => execApi('next'), enabled: isResultPhase && !gameOver },
    ],
    [execApi, isPlayPhase, isResultPhase, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('doubleattack', state);

  if (!state) return <GameSkeleton gameKey="doubleattack" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [DoubleAttackPhase.BET]: t('phase.bet'),
      [DoubleAttackPhase.ATTACK]: t('phase.attack'),
      [DoubleAttackPhase.PLAY]: t('phase.play'),
      [DoubleAttackPhase.RESULT]: t('phase.result'),
    }[state.phase] ?? '';

  const staked = state.hands.reduce((sum, h) => sum + h.bet, 0) + state.bustItBet;
  const net = state.payout - staked;
  const won = state.hands.some(
    (h) => h.result === DOUBLE_ATTACK_RESULT.win || h.result === DOUBLE_ATTACK_RESULT.blackjack,
  );
  const resultKeyOf = (r: number) => Object.entries(DOUBLE_ATTACK_RESULT).find(([, v]) => v === r)?.[0] ?? 'none';

  return (
    <GamePageShell
      title={tc('nav.doubleattack')}
      gameThemeBg={gameTheme.doubleattack.bg}
      phaseName={phaseName}
      gamePath="/doubleattack"
      gameEndFlag={gameOver}
      winShow={isResultPhase && won}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="da-chips">
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

            <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="da-bet-line">
              {t('label.round')}: {state.roundNumber}
              {state.anteBet > 0 && (
                <>
                  {' · '}
                  {t('label.ante')}: {state.anteBet} · {t('label.attack')}: {state.attackBet} · {t('label.bustIt')}:{' '}
                  {state.bustItBet}
                </>
              )}
            </div>

            {state.dealerCards.length > 0 && (
              <div className="mb-3" data-tutorial="da-dealer">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">
                  {state.dealerHoleDealt ? t('label.dealer') : t('label.upCard')}
                </div>
                <div className="flex justify-center gap-1 flex-wrap" data-testid="da-dealer-cards">
                  {state.dealerCards.map((card, i) => (
                    <AnimatedCard key={`dealer-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
                {state.dealerHoleDealt ? (
                  <div className="text-ds-text-primary text-center text-sm mt-1" data-testid="da-dealer-score">
                    {state.dealerScore} {t('label.score')}
                  </div>
                ) : (
                  <p className="text-ds-text-muted text-center text-xs mt-1" data-testid="da-dealer-hidden">
                    {t('label.hidden')}
                  </p>
                )}
              </div>
            )}

            {state.hands.length > 0 && (
              <div data-tutorial="da-hands">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.yourHands')}</div>
                {state.hands.map((h, i) => (
                  <div
                    key={`hand-${i}-${h.cards.length}-${h.score}`}
                    data-testid={`da-hand-${i}`}
                    className={`mb-2 rounded px-2 py-1 ${
                      isPlayPhase && i === state.activeHand ? 'ring-2 ring-ds-success' : ''
                    }`}
                  >
                    <div className="flex justify-center gap-1 flex-wrap">
                      {h.cards.map((card, k) => (
                        <AnimatedCard key={`h${i}-${card.design}-${card.value}-${k}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                    <div className="text-ds-text-primary text-center text-sm mt-1">
                      {state.hands.length > 1 && `${t('label.hand', { idx: i + 1 })} · `}
                      {h.score} {t('label.score')} · {t('label.bet')} {h.bet}
                      {isResultPhase && ` · ${t(`result.${resultKeyOf(h.result)}`)}`}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {isResultPhase && (
              <div className="text-center mb-2" data-testid="da-result">
                <div className={`text-sm font-medium ${net >= 0 ? 'text-ds-success' : 'text-ds-error'}`}>
                  {t('label.net')}: {net}
                </div>
                {/* **賭けたのに結果が見えない状態をなくす** (#5776)。合計収支だけ
                    だと、Bust It が当たったのか外れたのかが読めない。 */}
                {state.bustItBet > 0 && (
                  <div
                    className={`text-sm ${state.bustItPayout > 0 ? 'text-ds-success' : 'text-ds-text-muted'}`}
                    data-testid="da-bustit-result"
                  >
                    {state.bustItPayout > 0 ? t('result.bustItHit', { n: state.bustItPayout }) : t('result.bustItMiss')}
                  </div>
                )}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.doubleattack.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2">
              {isBetPhase && !gameOver && (
                <div className="flex flex-col items-center gap-2" data-tutorial="da-bet">
                  <p className="text-ds-text-muted text-sm">{t('betGuide')}</p>
                  <ChipBetInput
                    id="doubleattack-ante"
                    label={t('label.ante')}
                    value={ante}
                    onChange={setAnte}
                    max={state.chips}
                  />
                  <ChipBetInput
                    id="doubleattack-bustit"
                    label={t('label.bustIt')}
                    value={bustIt}
                    onChange={setBustIt}
                    max={Math.max(0, state.chips - ante)}
                  />
                  <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
                    {t('button.deal')}
                  </button>
                </div>
              )}

              {isAttackPhase && (
                <>
                  <p className="text-ds-text-muted text-sm">{t('attackGuide')}</p>
                  <p className="text-ds-text-muted text-xs" data-testid="da-attack-notice">
                    {t('attackNotice')}
                  </p>
                  {/* **上限はサーバの値に従う。** アンティから計算し直さない。 */}
                  <ChipBetInput
                    id="doubleattack-attack"
                    label={`${t('label.attack')} (${t('label.maxAttack')} ${state.maxAttackBet})`}
                    value={attack}
                    onChange={setAttack}
                    max={state.maxAttackBet}
                  />
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className={btnPrimary}
                      data-hint-action="attack"
                      onClick={handleAttack}
                      disabled={loading}
                    >
                      {t('button.attack')}
                    </button>
                    <button
                      type="button"
                      className={btnSecondary}
                      data-hint-action="decline"
                      onClick={handleDecline}
                      disabled={loading}
                    >
                      {t('button.decline')}
                    </button>
                  </div>
                </>
              )}

              {isPlayPhase && (
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
                    <button
                      type="button"
                      className={btnSecondary}
                      data-hint-action="stand"
                      onClick={() => execApi('stand')}
                      disabled={loading}
                    >
                      {t('button.stand')}
                    </button>
                    {/* **押せるかどうかはサーバが決める。** 手札から計算し直さない。 */}
                    {state.canDouble && (
                      <button
                        type="button"
                        className={btnWarning}
                        data-testid="da-double"
                        onClick={() => execApi('double')}
                        disabled={loading}
                      >
                        {t('button.double')}
                      </button>
                    )}
                    {state.canSplit && (
                      <button
                        type="button"
                        className={btnWarning}
                        data-testid="da-split"
                        onClick={() => execApi('split')}
                        disabled={loading}
                      >
                        {t('button.split')}
                      </button>
                    )}
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
