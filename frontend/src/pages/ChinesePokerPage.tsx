import { useCallback, useEffect, useMemo, useState } from 'react';
import { chinesepokerApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
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
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ChinesePokerResponse } from '../types/card';
import { ChinesePokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig } from '../utils/cli/types';

const FRONT_RANK_KEYS: Record<number, string> = {
  0: '',
  1: 'frontRank.1',
  2: 'frontRank.2',
  3: 'frontRank.3',
  4: 'frontRank.4',
  5: 'frontRank.5',
  6: 'frontRank.6',
};

const FIVE_CARD_RANK_KEYS: Record<number, string> = {
  0: 'fiveCardRank.0',
  1: 'fiveCardRank.1',
  2: 'fiveCardRank.2',
  3: 'fiveCardRank.3',
  4: 'fiveCardRank.4',
  5: 'fiveCardRank.5',
  6: 'fiveCardRank.6',
  7: 'fiveCardRank.7',
  8: 'fiveCardRank.8',
  9: 'fiveCardRank.9',
};

const CP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cp-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cp-set-hands"]',
    messageKey: 'tutorial.setHands',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cp-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

type HandAssignment = 'front' | 'middle' | undefined;

/** Renders the Chinese Poker game page. */
export const ChinesePokerPage = withTutorial(ChinesePokerPageContent, 'chinesepoker', CP_TUTORIAL_STEPS);

/** Inner content of the Chinese Poker page. */
function ChinesePokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('chinesepoker');

  const [betAmount, setBetAmount] = useState(100);
  const [assignments, setAssignments] = useState<HandAssignment[]>([]);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(chinesepokerApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('chinesepoker', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('chinesepoker');
  const cliConfig: CliGameConfig<ChinesePokerResponse, Parameters<typeof chinesepokerApi.exec>> = useMemo(
    () => ({
      gameName: 'chinesepoker',
      parseCommand: (cmd: string) => {
        const parts = cmd.trim().split(/\s+/);
        const c = parts[0]?.toLowerCase();
        if (c === 'r' || c === 'reset') return { args: ['reset'] as Parameters<typeof chinesepokerApi.exec> };
        if (c === 'l' || c === 'log') return { args: ['log'] as Parameters<typeof chinesepokerApi.exec> };
        if (c === 'b' || c === 'bet') {
          const amt = Number.parseInt(parts[1] ?? '', 10);
          if (Number.isNaN(amt)) return { error: 'Usage: b <amount>' };
          return { args: ['bet', amt] as Parameters<typeof chinesepokerApi.exec> };
        }
        if (c === 's' || c === 'set') {
          if (parts.length < 9) return { error: 'Usage: s <f0 f1 f2 m0 m1 m2 m3 m4>' };
          const fi = parts.slice(1, 4).map(Number);
          const mi = parts.slice(4, 9).map(Number);
          if (fi.some(Number.isNaN) || mi.some(Number.isNaN)) return { error: 'Invalid indices' };
          return { args: ['set', undefined, fi, mi] as Parameters<typeof chinesepokerApi.exec> };
        }
        return { error: `Unknown command: ${c}` };
      },
      formatResponse: () => '',
      helpText: 'b <amount>  Bet\ns <f0 f1 f2 m0 m1 m2 m3 m4>  Set hands\nr  Reset\nl  Action log',
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const cardCount = state?.playerCards?.length ?? 0;
  useEffect(() => {
    if (state?.phase !== ChinesePokerPhase.SET_HANDS) {
      setAssignments([]);
    } else {
      setAssignments(new Array(cardCount).fill(undefined));
    }
  }, [state?.phase, cardCount]);

  const isBetPhase = state?.phase === ChinesePokerPhase.BET;
  const isSetHandsPhase = state?.phase === ChinesePokerPhase.SET_HANDS;
  const isEndPhase = state?.phase === ChinesePokerPhase.END;

  const frontIndices = useMemo(
    () => assignments.map((a, i) => (a === 'front' ? i : -1)).filter((i) => i >= 0),
    [assignments],
  );
  const middleIndices = useMemo(
    () => assignments.map((a, i) => (a === 'middle' ? i : -1)).filter((i) => i >= 0),
    [assignments],
  );
  const canSet = frontIndices.length === 3 && middleIndices.length === 5;

  const toggleCard = useCallback((index: number) => {
    setAssignments((prev) => {
      const next = [...prev];
      const cur = next[index];
      if (cur === 'front') {
        const midCount = next.filter((a) => a === 'middle').length;
        next[index] = midCount < 5 ? 'middle' : undefined;
      } else if (cur === 'middle') {
        next[index] = undefined;
      } else {
        const frontCount = next.filter((a) => a === 'front').length;
        const midCount = next.filter((a) => a === 'middle').length;
        next[index] = frontCount < 3 ? 'front' : midCount < 5 ? 'middle' : undefined;
      }
      return next;
    });
  }, []);

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: () => execApi('bet', betAmount), enabled: isBetPhase },
      {
        key: 's',
        action: () => {
          if (canSet) execApi('set', undefined, frontIndices, middleIndices);
        },
        enabled: isSetHandsPhase && canSet,
      },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, betAmount, frontIndices, middleIndices, isBetPhase, isSetHandsPhase, isEndPhase, canSet],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) return <GameSkeleton gameKey="chinesepoker" layout={{ kind: 'casino-table', sections: [13, 13] }} />;

  const handleBet = () => execApi('bet', betAmount);
  const handleSet = () => {
    if (canSet) execApi('set', undefined, frontIndices, middleIndices);
  };
  const handleReset = () => execApi('reset');

  const phaseName = isBetPhase ? t('phase.bet') : isSetHandsPhase ? t('phase.setHands') : t('phase.end');

  const ringClass = (a: HandAssignment) => {
    if (a === 'front') return '-translate-y-3 ring-2 ring-ds-info rounded';
    if (a === 'middle') return '-translate-y-3 ring-2 ring-ds-success rounded';
    return '';
  };

  const badgeText = (a: HandAssignment) => {
    if (a === 'front') return 'F';
    if (a === 'middle') return 'M';
    return '';
  };

  return (
    <GamePageShell
      title={tc('nav.chinesepoker')}
      gameThemeBg={gameTheme.chinesepoker.bg}
      phaseName={phaseName}
      gamePath="/chinesepoker"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.result > 0}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
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
          <div
            data-testid="card-area"
            className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
              .filter(Boolean)
              .join(' ')}
          >
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
              </div>
            )}

            {isSetHandsPhase && state.playerCards.length > 0 && (
              <div className="mb-4" data-tutorial="cp-set-hands">
                <p className="text-ds-text-muted text-center text-sm mb-2">{t('selectCards')}</p>
                <div className="flex justify-center gap-1 text-sm mb-1">
                  <span className="text-ds-info">{t('selectedFront', { count: frontIndices.length })}</span>
                  <span className="text-ds-text-muted mx-1">|</span>
                  <span className="text-ds-success">{t('selectedMiddle', { count: middleIndices.length })}</span>
                </div>
                <div className="flex justify-center gap-1 flex-wrap">
                  {state.playerCards.map((card, i) => (
                    <button
                      key={`p-${card.design}-${card.value}-${i}`}
                      type="button"
                      onClick={() => toggleCard(i)}
                      className={`relative transition-transform ${ringClass(assignments[i])}`}
                      aria-pressed={!!assignments[i]}
                      aria-label={`Card ${i}`}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
                      {assignments[i] && (
                        <span
                          className={`absolute -top-2 -right-1 text-xs font-bold rounded-full w-5 h-5 flex items-center justify-center ${
                            assignments[i] === 'front' ? 'bg-ds-info text-white' : 'bg-ds-success text-white'
                          }`}
                        >
                          {badgeText(assignments[i])}
                        </span>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {isEndPhase && (
              <div data-tutorial="cp-results">
                {/* Player hands */}
                <HandSection
                  label={`${t('label.front')}`}
                  cards={state.playerFront}
                  rankKey={FRONT_RANK_KEYS[state.playerFrontRank]}
                  result={state.frontResult}
                  cardWidth={cardWidth}
                  t={t}
                  isPlayer
                />
                <HandSection
                  label={`${t('label.middle')}`}
                  cards={state.playerMiddle}
                  rankKey={FIVE_CARD_RANK_KEYS[state.playerMiddleRank]}
                  result={state.middleResult}
                  cardWidth={cardWidth}
                  t={t}
                  isPlayer
                />
                <HandSection
                  label={`${t('label.back')}`}
                  cards={state.playerBack}
                  rankKey={FIVE_CARD_RANK_KEYS[state.playerBackRank]}
                  result={state.backResult}
                  cardWidth={cardWidth}
                  t={t}
                  isPlayer
                />

                {/* Dealer hands */}
                <HandSection
                  label={`${t('label.front')}`}
                  cards={state.dealerFront}
                  rankKey={FRONT_RANK_KEYS[state.dealerFrontRank]}
                  result={state.frontResult}
                  cardWidth={cardWidth}
                  t={t}
                  isPlayer={false}
                />
                <HandSection
                  label={`${t('label.middle')}`}
                  cards={state.dealerMiddle}
                  rankKey={FIVE_CARD_RANK_KEYS[state.dealerMiddleRank]}
                  result={state.middleResult}
                  cardWidth={cardWidth}
                  t={t}
                  isPlayer={false}
                />
                <HandSection
                  label={`${t('label.back')}`}
                  cards={state.dealerBack}
                  rankKey={FIVE_CARD_RANK_KEYS[state.dealerBackRank]}
                  result={state.backResult}
                  cardWidth={cardWidth}
                  t={t}
                  isPlayer={false}
                />

                <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
                  {state.payout !== 0 && (
                    <div>
                      {t('label.bet')}: {state.bet} | {tc('common:label.payout', { defaultValue: 'Payout' })}:{' '}
                      {state.payout}
                    </div>
                  )}
                  {(state.playerRoyalty > 0 || state.dealerRoyalty > 0) && (
                    <div>
                      {t('label.playerRoyalty')}: {state.playerRoyalty} | {t('label.dealerRoyalty')}:{' '}
                      {state.dealerRoyalty}
                    </div>
                  )}
                  {state.scoop && <div className="font-bold text-ds-warning">{t('label.scoop')}!</div>}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.chinesepoker.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'chinesepoker-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="cp-bet-controls">
                <div className="flex items-center gap-2">
                  <label htmlFor="cp-bet-amount" className="text-ds-text-primary text-sm">
                    {t('label.bet')}
                  </label>
                  <input
                    id="cp-bet-amount"
                    type="number"
                    min={10}
                    max={state.chips}
                    step={10}
                    value={betAmount}
                    onChange={(e) => setBetAmount(Number(e.target.value))}
                    className="w-24 px-2 py-1 rounded text-sm"
                  />
                </div>
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isSetHandsPhase && (
              <div className="flex flex-col items-center gap-1 pb-2">
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleSet}
                  disabled={loading || !canSet}
                  data-testid="set-hands-button"
                >
                  {t('button.set')}
                </button>
              </div>
            )}
            {isEndPhase && (
              <div className="flex justify-center gap-2 pb-2">
                <GameResetButton
                  isGameEnd={isEndPhase}
                  onReset={handleReset}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('actionLog.view')}
                </button>
              </div>
            )}
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Renders a single hand section (front/middle/back) for player or dealer. */
function HandSection({
  label,
  cards,
  rankKey,
  result,
  cardWidth,
  t,
  isPlayer,
}: {
  label: string;
  cards: { design: number; value: number }[];
  rankKey: string | undefined;
  result: number;
  cardWidth: number;
  t: (key: string) => string;
  isPlayer: boolean;
}) {
  if (!cards || cards.length === 0) return null;
  const colorClass = isPlayer ? 'text-ds-warning' : 'text-ds-error';
  const icon = isPlayer ? '🟡' : '🔴';
  const resultIcon = isPlayer ? (result > 0 ? '✅' : '❌') : result > 0 ? '❌' : '✅';

  return (
    <div className="mb-3">
      <div className={`${colorClass} font-bold text-center text-sm mb-1`}>
        <span aria-hidden="true">{icon}</span> {label}
        {rankKey && <span className="ml-1 text-xs font-normal">({t(rankKey)})</span>}
        <span className="ml-1">{resultIcon}</span>
      </div>
      <div className="flex justify-center gap-1">
        {cards.map((card, i) => (
          <AnimatedCard
            key={`${isPlayer ? 'p' : 'd'}-${label}-${card.design}-${card.value}-${i}`}
            card={card}
            width={cardWidth}
          />
        ))}
      </div>
    </div>
  );
}
