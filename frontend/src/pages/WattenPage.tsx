import type { TFunction } from 'i18next';
import { useEffect, useMemo, useRef, useState } from 'react';
import type { wattenApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useActionLog } from '../hooks/useActionLog';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useWattenGame } from '../hooks/useWattenGame';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { WattenResponse } from '../types/card';
import { WattenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseWattenCommand, WATTEN_HELP } from '../utils/cli/commands/wattenCommands';
import { formatWattenState } from '../utils/cli/formatters/wattenFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { buildWattenStakeHistory, WATTEN_BASE_STAKE } from '../utils/wattenStakeHistory';
import { wattenTrumpCards, wattenTrumpInfo } from '../utils/wattenTrumps';

/** Watten tutorial step definitions. */
const WATTEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="watten-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="watten-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="watten-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="watten-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="watten-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const WATTEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [WattenPhase.DECLARE]: 'declare',
  [WattenPhase.PLAY]: 'play',
  [WattenPhase.RESPOND]: 'respond',
  [WattenPhase.TRICK_END]: 'trickEnd',
  [WattenPhase.ROUND_END]: 'roundEnd',
  [WattenPhase.GAME_END]: 'gameEnd',
};

/** Trump/critical-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS = ['suitNone', 'suitSpade', 'suitClub', 'suitHeart', 'suitDiamond'] as const;

/** Selectable critical (trump) suits with their playing-card symbols (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_CHOICES = [
  { code: 1, symbol: '♠' },
  { code: 2, symbol: '♣' },
  { code: 3, symbol: '♥' },
  { code: 4, symbol: '♦' },
] as const;

/** Selectable Schlag ranks (card value + short label). Watten allows A and 7..K. */
const SCHLAG_CHOICES = [
  { value: 7, label: '7' },
  { value: 8, label: '8' },
  { value: 9, label: '9' },
  { value: 10, label: '10' },
  { value: 11, label: 'J' },
  { value: 12, label: 'Q' },
  { value: 13, label: 'K' },
  { value: 1, label: 'A' },
] as const;

/** Maps a Schlag rank value to its short display label. */
function schlagLabel(rank: number): string {
  return SCHLAG_CHOICES.find((c) => c.value === rank)?.label ?? '-';
}

type WattenGame = ReturnType<typeof useWattenGame>;
type WattenView = ReturnType<typeof useWattenView>;
type WattenContext = GamePageContext<WattenResponse, WattenGame, WattenView>;

function useWattenView({ game, t }: { game: WattenGame; t: TFunction }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('watten', WATTEN_PHASE_KEYS);

  // Pending Declare selections (null until the dealer picks each).
  const [selectedRank, setSelectedRank] = useState<number | null>(null);
  const [selectedSuit, setSelectedSuit] = useState<number | null>(null);
  const previousScores = useRef<number[] | null>(null);
  const [scoreAnnouncement, setScoreAnnouncement] = useState('');

  const state = game.state;

  useEffect(() => {
    if (!state) return;
    if (previousScores.current) {
      const changedTeams = state.teamScores
        .map((score, team) => ({ score, team }))
        .filter(({ score, team }) => previousScores.current?.[team] !== score);
      if (changedTeams.length > 0) {
        setScoreAnnouncement(
          changedTeams
            .map(({ score, team }) => t('scoreUpdated', { team: t('team', { n: team }), score }))
            .join(t('listSeparator')),
        );
      }
    }
    previousScores.current = [...state.teamScores];
  }, [state, t]);

  // Dedicated action-log fetch used only to reconstruct the stake-escalation
  // mini-history. Kept separate from the page's shared action-log panel so the
  // history stays fresh during play (the shared log is opened manually at end).
  const { actionLog: stakeLog, showActionLog: refreshStakeLog } = useActionLog('watten');
  // Re-fetch the log whenever a stake-relevant field changes so the mini-history
  // tracks the latest raise/respond exchange; these fields are triggers, not
  // values read inside the effect body.
  // biome-ignore lint/correctness/useExhaustiveDependencies: stake fields are re-fetch triggers, not deps read in the callback.
  useEffect(() => {
    void refreshStakeLog();
  }, [refreshStakeLog, state?.stake, state?.pendingStake, state?.raiseCount, state?.roundNumber]);
  const stakeHistory = useMemo(() => buildWattenStakeHistory(stakeLog), [stakeLog]);

  const declare = () => {
    if (selectedRank === null || selectedSuit === null) return;
    game.handleDeclare(selectedRank, selectedSuit);
    setSelectedRank(null);
    setSelectedSuit(null);
  };

  return {
    cardWidth,
    isMobile,
    phaseNames,
    selectedRank,
    setSelectedRank,
    selectedSuit,
    setSelectedSuit,
    scoreAnnouncement,
    stakeHistory,
    declare,
  };
}

function LiveRegion({ view }: WattenContext) {
  return (
    <div className="sr-only" role="status" aria-live="polite" data-testid="watten-score-live">
      {view.scoreAnnouncement}
    </div>
  );
}

function Board({ state, t, busy, view }: WattenContext) {
  const { cardWidth, isMobile, selectedRank, setSelectedRank, selectedSuit, setSelectedSuit, stakeHistory } = view;
  const loading = busy;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isDeclarePhase = state.phase === WattenPhase.DECLARE;
  const isRoundEnd = state.phase === WattenPhase.ROUND_END;
  const isGameEnd = state.phase === WattenPhase.GAME_END || state.gameEndFlag;

  const canDeclare = isDeclarePhase && state.dealerIdx === humanIdx;

  const schlagText = state.schlagRank > 0 ? schlagLabel(state.schlagRank) : t('schlagNone');
  const criticalText = state.criticalSuit >= 1 ? t(SUIT_KEYS[state.criticalSuit] ?? 'suitNone') : t('suitNone');

  const messageBox = (
    <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
  );

  return (
    <>
      {canDeclare && (
        <div className="mb-3 flex flex-col gap-1">
          <fieldset className="grid grid-cols-3 sm:grid-cols-6 gap-1 border-0 p-0 m-0 min-w-0">
            <legend className="text-ds-text-muted text-sm">{t('chooseSchlag')}</legend>
            {SCHLAG_CHOICES.map((c) => (
              <button
                key={c.value}
                type="button"
                className={selectedRank === c.value ? btnPrimary : btnSecondary}
                onClick={() => setSelectedRank(c.value)}
                disabled={loading}
                aria-pressed={selectedRank === c.value}
              >
                {c.label}
              </button>
            ))}
          </fieldset>
          <fieldset className="grid grid-cols-4 sm:grid-cols-8 gap-1 border-0 p-0 m-0 min-w-0">
            <legend className="text-ds-text-muted text-sm">{t('chooseCritical')}</legend>
            {SUIT_CHOICES.map((c) => (
              <button
                key={c.code}
                type="button"
                className={selectedSuit === c.code ? btnPrimary : btnSecondary}
                onClick={() => setSelectedSuit(c.code)}
                disabled={loading}
                aria-label={t(SUIT_KEYS[c.code])}
                aria-pressed={selectedSuit === c.code}
              >
                {c.symbol}
              </button>
            ))}
          </fieldset>
        </div>
      )}
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('schlag', { rank: schlagText })}</span>
        <span>{t('critical', { suit: criticalText })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="watten-trick-display"
          />
          {state.lastTrick.length > 0 && (
            <details open className="my-3" data-testid="watten-trick-history">
              <summary className="text-ds-text-primary text-sm cursor-pointer py-2">{t('trickHistory.title')}</summary>
              <TrickDisplay
                currentTrick={[]}
                lastTrick={state.lastTrick}
                lastTrickWinner={state.lastTrickWinner}
                players={state.players}
                cardWidth={cardWidth}
                label={t('trickHistory.cards')}
                winnerLabel={t('trickHistory.winner')}
              />
            </details>
          )}
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="watten-info">
          {/* Stake + team scores/tricks */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            <div className="text-ds-text-primary mb-1">
              {state.pendingStake > 0
                ? t('stakePending', { stake: state.stake, pending: state.pendingStake })
                : t('stake', { stake: state.stake })}
            </div>
            <table className="w-full">
              <thead>
                <tr>
                  <th scope="col" className="text-left">
                    {t('team', { n: 0 })}
                    {state.players[humanIdx]?.team === 0 ? ` (${t('you')})` : ''}
                  </th>
                  <th scope="col" className="text-center">
                    {t('team', { n: 1 })}
                    {state.players[humanIdx]?.team === 1 ? ` (${t('you')})` : ''}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td className="text-ds-accent">{t('score', { score: state.teamScores[0] })}</td>
                  <td className="text-center">{t('score', { score: state.teamScores[1] })}</td>
                </tr>
                <tr>
                  <td className="text-xs">{t('tricks', { count: state.teamTricks[0] })}</td>
                  <td className="text-center text-xs">{t('tricks', { count: state.teamTricks[1] })}</td>
                </tr>
              </tbody>
            </table>
          </div>

          {/* Stake-escalation mini-history (raise/hold/fold sequence) */}
          {stakeHistory.length > 0 && (
            <div
              className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              data-testid="watten-stake-history"
              // 賭け金の吊り上げはこのゲームの核心なので、履歴が伸びるたびに
              // 読み上げる。視覚レイアウトは変えない。
              role="status"
              aria-live="polite"
            >
              <div className="text-ds-text-primary mb-1">{t('stakeHistory.title')}</div>
              <div className="flex flex-wrap items-center gap-x-1.5 gap-y-1">
                <span>{t('stakeHistory.start', { stake: WATTEN_BASE_STAKE })}</span>
                {stakeHistory.map((ev) => (
                  <span key={ev.key} className="flex items-center gap-x-1.5">
                    <span aria-hidden="true" className="text-ds-text-muted">
                      →
                    </span>
                    <span className={ev.type === 'fold' ? 'text-ds-warning' : 'text-ds-accent'}>
                      {t(`stakeHistory.${ev.type}`, {
                        player: findPlayerName(state.players, ev.playerIdx),
                        stake: ev.stake,
                      })}
                    </span>
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Players: cards / tricks */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('teamShort', { n: p.team })} |{' '}
                    {t('cards', { count: p.cardCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('teamShort', { n: p.team })} | {t('cards', { count: p.cardCount })}
                </div>
              ))}
            </div>
          )}

          {/* Deal result */}
          {(isRoundEnd || isGameEnd) && state.dealWinnerTeam >= 0 && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
              <div>{t('dealResult.winner', { team: state.dealWinnerTeam, stake: state.stake })}</div>
            </div>
          )}
        </div>
      </div>
      {isMobile && messageBox}
    </>
  );
}

function FooterLead({ state, t, view }: WattenContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const canDeclare = state.phase === WattenPhase.DECLARE && state.dealerIdx === humanIdx;

  const effectiveRank = view.selectedRank ?? state.schlagRank;
  const effectiveSuit = view.selectedSuit ?? state.criticalSuit;
  const humanTrumps = humanPlayer ? wattenTrumpCards(humanPlayer.cards, effectiveRank, effectiveSuit) : [];

  return (
    <>
      {canDeclare && (
        <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="watten-trump-panel">
          <div className="text-ds-text-primary mb-1">{t('trumpPanel.title')}</div>
          <p className="mb-2 text-xs leading-relaxed">{t('trumpPanel.rule')}</p>
          {humanTrumps.length > 0 ? (
            <>
              <div className="text-ds-accent font-semibold mb-1" data-testid="watten-trump-count">
                {t('trumpPanel.summary', { count: humanTrumps.length })}
              </div>
              <ul className="flex flex-wrap gap-x-3 gap-y-0.5">
                {humanTrumps.map((tc) => (
                  <li key={`${tc.card.design}-${tc.card.value}`} className="whitespace-nowrap">
                    <span className="text-ds-text-primary">{cardAlt(tc.card)}</span>{' '}
                    <span className="text-ds-text-muted">— {t(`trumpCat.${tc.category}`)}</span>
                  </li>
                ))}
              </ul>
            </>
          ) : (
            <div className="text-xs" data-testid="watten-trump-count">
              {t('trumpPanel.none')}
            </div>
          )}
        </div>
      )}

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="watten-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndex !== undefined && ` ([${state.hint.cardIndex}])`}
          </div>
        )}
      </div>
    </>
  );
}

function FooterStatus({ frontendHint, t }: WattenContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function FooterHand({ game, state, t, view }: WattenContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isDeclarePhase = state.phase === WattenPhase.DECLARE;
  const isPlayPhase = state.phase === WattenPhase.PLAY;

  const canDeclare = isDeclarePhase && state.dealerIdx === humanIdx;
  const canPlay = isPlayPhase && state.currentPlayerIdx === humanIdx;

  const effectiveRank = view.selectedRank ?? state.schlagRank;
  const effectiveSuit = view.selectedSuit ?? state.criticalSuit;
  const humanTrumps = humanPlayer ? wattenTrumpCards(humanPlayer.cards, effectiveRank, effectiveSuit) : [];
  const trumpIndices = humanTrumps.map((tc) => tc.index);

  const leadCard = state.currentTrick[0]?.card;
  const leadIsTrump = leadCard != null && wattenTrumpInfo(leadCard, effectiveRank, effectiveSuit) !== null;
  const followSuit = leadCard?.design;
  const canFollowTrump =
    humanPlayer != null &&
    leadIsTrump &&
    humanPlayer.cards.some((card) => wattenTrumpInfo(card, effectiveRank, effectiveSuit) !== null);
  const canFollowSuit =
    humanPlayer != null &&
    !leadIsTrump &&
    followSuit != null &&
    humanPlayer.cards.some(
      (card) => card.design === followSuit && wattenTrumpInfo(card, effectiveRank, effectiveSuit) === null,
    );
  const legalIndices =
    canPlay && humanPlayer
      ? humanPlayer.cards.flatMap((card, idx) => {
          const isTrump = wattenTrumpInfo(card, effectiveRank, effectiveSuit) !== null;
          const legal = canFollowTrump ? isTrump : canFollowSuit ? !isTrump && card.design === followSuit : true;
          return legal ? [idx] : [];
        })
      : undefined;
  const restrictedReason =
    canPlay && canFollowTrump ? t('mustFollowTrump') : canPlay && canFollowSuit ? t('mustFollowSuit') : undefined;

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="watten-prompt-live" role="status" aria-live="polite">
        {canDeclare && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="watten-declare-prompt">
            {t('declarePhase')}
          </div>
        )}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="watten"
          validIndices={legalIndices}
          legalIndices={legalIndices}
          restrictedTooltip={restrictedReason}
          cardStatusFor={(idx) => (legalIndices != null && !legalIndices.includes(idx) ? restrictedReason : undefined)}
          trumpIndices={trumpIndices.length > 0 ? trumpIndices : undefined}
          trumpTitle={t('trumpRing')}
          trumpAccessibleLabel={t('trumpCardDescription')}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, tc, busy, view }: WattenContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isDeclarePhase = state.phase === WattenPhase.DECLARE;
  const isPlayPhase = state.phase === WattenPhase.PLAY;
  const isRespondPhase = state.phase === WattenPhase.RESPOND;
  const isRoundEnd = state.phase === WattenPhase.ROUND_END;

  const canDeclare = isDeclarePhase && state.dealerIdx === humanIdx;
  const canPlay = isPlayPhase && state.currentPlayerIdx === humanIdx;
  const canRespond = isRespondPhase && state.responderIdx === humanIdx;

  if (!canDeclare && !canPlay && !canRespond && !isRoundEnd) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="watten-action-buttons">
      {canDeclare && (
        <button
          type="button"
          className={btnPrimary}
          onClick={view.declare}
          disabled={busy || view.selectedRank === null || view.selectedSuit === null}
        >
          {t('declareButton')}
        </button>
      )}
      {canPlay && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handlePlay}
            disabled={busy || game.selectedCardIndices.length !== 1}
          >
            {t('playButton')}
          </button>
          {state.canRaise && (
            <button type="button" className={btnSecondary} onClick={game.handleRaise} disabled={busy}>
              {t('raiseButton')}
            </button>
          )}
          <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy}>
            {tc('button.hint')}
          </button>
        </>
      )}
      {canRespond && (
        <>
          <span className="text-ds-text-muted text-sm">{t('respondPhase', { stake: state.pendingStake })}</span>
          {/* **吊り上げたのが誰かで hold/fold の判断は変わる。**サーバは
          raiserTeam を毎回送っているのに画面が読まず、人間は
          pendingStake だけを見て決めるしかなかった (#6497)。
          まだ誰も吊り上げていない (-1) 局面では出さない。 */}
          {state.raiserTeam >= 0 && (
            <span className="text-ds-text-muted text-sm" data-testid="watten-raiser-info">
              {state.raiserTeam === humanPlayer?.team ? t('raiserOwn') : t('raiserOpponent')}
            </span>
          )}
          <span className="text-ds-warning text-sm" data-testid="watten-respond-loss">
            {t('respondLoss', { stake: state.stake })}
          </span>
          <button type="button" className={btnPrimary} onClick={() => game.handleRespond(true)} disabled={busy}>
            {t('holdButton')}
          </button>
          <button type="button" className={btnSecondary} onClick={() => game.handleRespond(false)} disabled={busy}>
            {t('foldButton')}
          </button>
        </>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const wattenDefinition: GameDefinition<WattenResponse, WattenGame, Parameters<typeof wattenApi.exec>, WattenView> = {
  key: 'watten',
  useGame: useWattenGame,
  useView: useWattenView,
  tutorial: WATTEN_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseWattenCommand,
    formatResponse: formatWattenState,
    helpText: WATTEN_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === WattenPhase.GAME_END || state.gameEndFlag;
    const humanPlayer = state.players.find((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && humanPlayer != null && state.winnerTeam === humanPlayer.team,
    };
  },
  shell: ({ state }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const canDeclare = state.phase === WattenPhase.DECLARE && state.dealerIdx === humanIdx;
    const canPlay = state.phase === WattenPhase.PLAY && state.currentPlayerIdx === humanIdx;
    const canRespond = state.phase === WattenPhase.RESPOND && state.responderIdx === humanIdx;
    return {
      isHumanTurn: canDeclare || canPlay || canRespond,
    };
  },
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.wattenConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.wattenConfig.targetScore,
          options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: (ctx) =>
    !ctx.view.isMobile && (
      <GameMessageBox
        message={ctx.state.message}
        messageCode={ctx.state.messageCode}
        messageParams={ctx.state.messageParams}
      />
    ),
  resetTutorialId: 'watten-reset-button',
  actionsClassName: '',
  Board,
  FooterHand,
  FooterLead,
  FooterStatus,
  LiveRegion,
};

/** Renders the Watten (ヴァッテン) game page: a Bavarian/Austrian 4-player, 2-team trick-taker with a raise/bluff stake mechanic. */
export const WattenPage = createGamePage(wattenDefinition);
