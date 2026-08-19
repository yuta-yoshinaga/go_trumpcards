import { useCallback, useEffect, useMemo, useState } from 'react';
import { minibridgeApi } from '../api/gameApi';
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
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MinibridgeResponse } from '../types/card';
import { MinibridgePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MINIBRIDGE_HELP, parseMinibridgeCommand } from '../utils/cli/commands/minibridgeCommands';
import { formatMinibridgeState } from '../utils/cli/formatters/minibridgeFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Contract denominations. **`0` is no-trump**, which is a choice, not a blank. */
const SUIT_SYMBOLS: Readonly<Record<number, string>> = { 0: 'NT', 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** The five denominations, in the order the contract buttons are offered. */
const DENOMINATIONS: readonly number[] = [1, 2, 3, 4, 0];

/** Highest contract level. Level n needs 6 + n tricks, so 7 is all thirteen. */
const MAX_LEVEL = 7;

/** Guided tutorial steps (the HCP announcement, the declarer, the contract, your hand). */
const MINIBRIDGE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="mb-rule"]', messageKey: 'tutorial.hcp', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mb-seats"]', messageKey: 'tutorial.declarer', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mb-contract"]', messageKey: 'tutorial.contract', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mb-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Minibridge page (wrapped by `withTutorial`).
 *
 * Bridge with the auction replaced by an open count. **Every seat's HCP is on
 * the board because that is the whole mechanic** — the pair holding more
 * declares, and since the four hands always total exactly 40, a 20-20 split is
 * possible and goes to the dealer's side.
 *
 * The declarer plays the dummy as well as their own hand, so the page renders
 * the dummy as a second pressable row and follows `validPlays` to whichever
 * seat is currently being controlled.
 */
function MinibridgePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('minibridge');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<MinibridgeResponse, Parameters<typeof minibridgeApi.exec>>(minibridgeApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('minibridge', state);
  const [level, setLevel] = useState(1);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('minibridge');
  const cliConfig: CliGameConfig<MinibridgeResponse, Parameters<typeof minibridgeApi.exec>> = useMemo(
    () => ({
      gameName: 'minibridge',
      parseCommand: parseMinibridgeCommand,
      formatResponse: formatMinibridgeState,
      helpText: MINIBRIDGE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    setLevel(1);
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handleContract = useCallback(
    (suit: number) => {
      void dispatch('contract', undefined, undefined, level, suit);
    },
    [dispatch, level],
  );

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNextRound = useCallback(() => {
    setLevel(1);
    void dispatch('next');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="minibridge" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  // 味方かどうかは人間のチーム番号との一致で決まる。
  const humanTeam = human?.team;
  const isContract = state.phase === MinibridgePhase.CONTRACT;
  const isRoundEnd = state.phase === MinibridgePhase.ROUND_END;
  const isGameEnd = state.phase === MinibridgePhase.GAME_END || state.gameEndFlag;
  const isHumanContractTurn = isContract && !isGameEnd && state.players[state.declarerIdx]?.isHuman === true;
  // **ダミーの席かどうかを先に見る（sync: Minibridge.IsHumanTurn）。** ダミーは
  // 落札者の相方なので、落札者が席 2 ならダミーは席 0 ——**自分の席**になる。
  // 「その席が人間か」を先に見ると、CPU が落札したダミーまで自分の番と判定され、
  // サーバは受け付けないのに押せる札が出てしまう。
  const humanIsDeclarer = state.players[state.declarerIdx]?.isHuman === true;
  const isDummyTurn = state.phase === MinibridgePhase.PLAY && state.currentPlayerIdx === state.dummyIdx;
  const seatIsControllable = isDummyTurn ? humanIsDeclarer : state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanTurn = !isGameEnd && !isRoundEnd && !isContract && seatIsControllable;
  const isHumanDummyTurn = isHumanTurn && isDummyTurn;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isContract
        ? t('phase.contract')
        : t('phase.play');

  // 出せる札に緑の枠を足すだけで、押せなくはしない（サーバが必ず検証する）。
  // **validPlays はいま操作している席のもの**なので、ダミーの手番ではダミー側に付く。
  const legalRing = new Set(isHumanTurn ? state.validPlays : []);

  const pairHcp = state.players
    .filter((p) => p.team === state.players[state.declarerIdx]?.team)
    .reduce((sum, p) => sum + p.hcp, 0);

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerTeam === 0) return t('result.you');
    if (state.winnerTeam < 0) return t('result.tie');
    return t('result.cpu');
  })();

  const roundResult = (() => {
    if (!isRoundEnd) return null;
    const params = { need: String(state.requiredTricks), took: String(state.lastTricks) };
    return state.lastMade ? t('roundResult.made', params) : t('roundResult.down', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.minibridge')}
      gameThemeBg={gameTheme.minibridge.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/minibridge"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerTeam === 0}
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
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4" data-testid="mb-round">
                {t('header.round', { round: String(state.roundNumber), total: String(state.config.rounds) })}
              </span>
              <span data-testid="mb-score">
                {t('header.score', {
                  // **`ns` は i18next の予約オプション。** 補間名に使うと
                  // 名前空間の指定と解釈され、キーがそのまま表示される。
                  ours: String(state.teamScores[0] ?? 0),
                  theirs: String(state.teamScores[1] ?? 0),
                })}
              </span>
            </div>

            {/* **競りが無いこと自体が規則。** 先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="mb-rule"
              data-tutorial="mb-rule"
            >
              {t('header.rule')}
            </div>

            <div
              className="text-center mb-3 text-ds-text-primary"
              data-testid="mb-contract"
              data-tutorial="mb-contract"
            >
              {state.contractLevel > 0
                ? t('header.contract', {
                    level: String(state.contractLevel),
                    suit: SUIT_SYMBOLS[state.contractSuit] ?? '?',
                    name:
                      state.declarerIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(state.declarerIdx) }),
                    need: String(state.requiredTricks),
                  })
                : t('header.contractUndecided')}
            </div>

            {isContract && (
              <div className="text-center mb-3 text-ds-accent text-sm" role="status" data-testid="mb-pair-hcp">
                {t('header.pairHcp', { n: String(pairHcp) })}
              </div>
            )}

            {/* **HCP は公開情報。** 4 席ぶん常に出す。 */}
            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="mb-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`mb-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {/* **競りが無いぶん、味方が誰かは席表示でしか分からない** (#5761)。
                      CUI は最初から team を出しているのに、Web は契約が決まって
                      デクレアラー/ダミーのタグが付くまで何も出していなかった。 */}
                  <span
                    className={`ml-1 ${p.team === humanTeam ? 'text-ds-info' : 'text-ds-text-muted'}`}
                    data-testid={`mb-team-${p.id.toString()}`}
                  >
                    <span aria-hidden="true">{t('header.teamTag', { team: String(p.team) })}</span>
                    <span className="sr-only">
                      {p.team === humanTeam ? t('header.teamAllyAria') : t('header.teamFoeAria')}
                    </span>
                  </span>
                  {p.id === state.declarerIdx && <span className="ml-1 text-ds-accent">{t('header.declarer')}</span>}
                  {p.id === state.dummyIdx && <span className="ml-1 text-ds-accent">{t('header.dummy')}</span>}
                  {': '}
                  <span className="text-ds-accent">{t('header.hcp', { n: String(p.hcp) })}</span>
                  {' / '}
                  {t('header.took', { n: String(p.trickCount) })}
                </div>
              ))}
            </div>

            <div data-tutorial="mb-trick">
              <TrickDisplay
                currentTrick={state.currentTrick}
                players={state.players}
                cardWidth={cardWidth}
                label={t('currentTrick')}
              />
            </div>

            {roundResult && (
              <div className="text-center my-3 text-ds-text-primary" role="status" data-testid="mb-round-result">
                {roundResult}
              </div>
            )}

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="mb-result"
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

            {/* **ダミーは契約が決まると公開され、デクレアラーが操作する。** */}
            {state.dummyHand.length > 0 && (
              <div className="mt-4" data-testid="mb-dummy">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.dummyHand')}
                  {isHumanDummyTurn && <span className="ml-2 text-ds-accent">{t('header.dummyTurn')}</span>}
                </div>
                <div className="flex flex-wrap gap-2">
                  {state.dummyHand.map((card, idx) => (
                    <button
                      key={`dummy-${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => handlePlay(idx)}
                      disabled={loading || !isHumanDummyTurn}
                      aria-label={t('actions.playDummyAria', { card: cardAlt(card) })}
                      className={`disabled:opacity-50 ${
                        isHumanDummyTurn && legalRing.has(idx) ? 'rounded-lg ring-2 ring-ds-success' : ''
                      }`}
                    >
                      <CardImage card={card} width={cardWidth} />
                    </button>
                  ))}
                </div>
              </div>
            )}

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="mb-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                  {isHumanTurn && !isHumanDummyTurn && (
                    <span className="ml-2 text-ds-accent">{t('header.yourTurn')}</span>
                  )}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => handlePlay(idx)}
                      disabled={loading || !isHumanTurn || isHumanDummyTurn}
                      aria-label={t('actions.playAria', { card: cardAlt(card) })}
                      className={`disabled:opacity-50 ${
                        !isHumanDummyTurn && legalRing.has(idx) ? 'rounded-lg ring-2 ring-ds-success' : ''
                      }`}
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

            <div className="mt-4 flex flex-wrap gap-2 items-center" data-tutorial="mb-actions">
              {isHumanContractTurn && (
                <>
                  <label className="text-ds-text-muted text-sm" htmlFor="mb-level">
                    {t('actions.level')}
                  </label>
                  {/* 折りたたみに入れない：閉じた details の中はクリックできない。 */}
                  <select
                    id="mb-level"
                    className="rounded bg-black/40 text-ds-text-primary px-2 py-1"
                    value={level}
                    onChange={(e) => setLevel(Number(e.target.value))}
                    disabled={loading}
                    data-testid="mb-level-select"
                  >
                    {Array.from({ length: MAX_LEVEL }, (_, i) => i + 1).map((lv) => (
                      <option key={lv} value={lv}>
                        {lv}
                      </option>
                    ))}
                  </select>
                  {/* **競りが無いので上回る必要はない。** どの組み合わせも選べる。 */}
                  {DENOMINATIONS.map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnWarning}
                      onClick={() => handleContract(suit)}
                      disabled={loading}
                      data-testid={`mb-contract-${suit.toString()}-btn`}
                    >
                      {t('actions.contract', { level: String(level), suit: SUIT_SYMBOLS[suit] ?? '?' })}
                    </button>
                  ))}
                </>
              )}
              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('actions.nextRound')}
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
              groups={[{ items: [hintCheckboxItem(tc, hintEnabled, setHintEnabled)] }]}
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

/** Minibridge page wrapped with TutorialProvider. */
export const MinibridgePage = withTutorial(MinibridgePageContent, 'minibridge', MINIBRIDGE_TUTORIAL_STEPS);
