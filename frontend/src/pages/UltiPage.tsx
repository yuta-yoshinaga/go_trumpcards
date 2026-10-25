import { useEffect, useState } from 'react';
import type { ultiApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useUltiGame } from '../hooks/useUltiGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { UltiResponse } from '../types/card';
import { UltiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseUltiCommand, ULTI_HELP } from '../utils/cli/commands/ultiCommands';
import { formatUltiState } from '../utils/cli/formatters/ultiFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Ulti tutorial step definitions. */
const ULTI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ulti-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ulti-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ulti-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ulti-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ulti-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const ULTI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [UltiPhase.BID]: 'bid',
  [UltiPhase.DISCARD]: 'discard',
  [UltiPhase.PLAY]: 'play',
  [UltiPhase.TRICK_END]: 'trickEnd',
  [UltiPhase.ROUND_END]: 'roundEnd',
  [UltiPhase.GAME_END]: 'gameEnd',
};

/** Contract i18n keys indexed by contract value (0=none, 1=Party, 2=Betli, 3=Durchmarsch, 4=Ulti). */
const CONTRACT_KEYS = [
  'contractNone',
  'contractParty',
  'contractBetli',
  'contractDurchmarsch',
  'contractUlti',
] as const;

/** Trump-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS = ['suitNone', 'suitSpade', 'suitClub', 'suitHeart', 'suitDiamond'] as const;

/** Outcome i18n keys indexed by outcome value (0=none, 1=Win/made, 2=Loss/failed). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Number of talon cards the declarer must discard in the Discard phase (matches `UltiDiscardSize` in `internal/domain/Ulti.go`). */
const DISCARD_COUNT = 2;

/** Selectable trump suits with their playing-card symbols (1=♠ 2=♣ 3=♥ 4=♦). */
const TRUMP_CHOICES = [
  { code: 1, symbol: '♠' },
  { code: 2, symbol: '♣' },
  { code: 3, symbol: '♥' },
  { code: 4, symbol: '♦' },
] as const;

type UltiGame = ReturnType<typeof useUltiGame>;
type UltiView = ReturnType<typeof useUltiView>;
type UltiContext = GamePageContext<UltiResponse, UltiGame, UltiView>;

function useUltiView({ game }: { game: UltiGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  // Trump suit chosen for a pending Party declaration (null until picked).
  const [selectedTrump, setSelectedTrump] = useState<number | null>(null);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ulti', ULTI_PHASE_KEYS);

  return {
    selectedTrump,
    setSelectedTrump,
    cardWidth,
    isMobile,
    phaseNames,
  };
}

function Board({ state, t, view }: UltiContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isPlayPhase = state.phase === UltiPhase.PLAY;
  const isTrickEnd = state.phase === UltiPhase.TRICK_END;
  const isRoundEnd = state.phase === UltiPhase.ROUND_END;
  const isGameEnd = state.phase === UltiPhase.GAME_END || state.gameEndFlag;

  const contractLabel = t(CONTRACT_KEYS[state.contract] ?? 'contractNone');
  const trumpLabel = state.trumpSuit >= 1 ? t(SUIT_KEYS[state.trumpSuit] ?? 'suitNone') : t('suitNone');
  const declarer = state.players[state.declarerIdx];
  const contractProgress = (() => {
    if (!(isPlayPhase || isTrickEnd) || !declarer) return null;

    switch (state.contract) {
      case 1:
        return t('contractProgress.party', { current: declarer.cardPoints, required: state.contractRequirement });
      case 2:
        return t('contractProgress.betli');
      case 3:
        return t('contractProgress.durchmarsch', { current: declarer.trickCount, required: state.contractRequirement });
      case 4:
        return t('contractProgress.ulti', { suit: trumpLabel });
      default:
        return null;
    }
  })();

  // Per-player coin change at settlement, straight from the deal that applied it.
  // The running balance alone cannot show it, and the backend clears this back to
  // zeros when the next deal starts.
  const coinDeltas = state.lastDealCoins ?? null;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('contract', { contract: contractLabel })}</span>
        <span>{t('trump', { suit: trumpLabel })}</span>
      </div>
      <div className="sr-only" role="status" aria-live="polite" data-testid="ulti-contract-progress-live">
        {contractProgress}
      </div>
      {contractProgress && (
        <div
          className="text-ds-text-muted text-sm text-center mb-2"
          data-testid="ulti-contract-progress"
          aria-hidden="true"
        >
          {contractProgress}
        </div>
      )}

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
            winnerLabel={
              isTrickEnd
                ? t('previousTrickWinner', { name: findPlayerName(state.players, state.lastTrickWinner) })
                : undefined
            }
            dataTutorial="ulti-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="ulti-info">
          {/* Per-player coin balances with a declarer badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p, i) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('coins', { coins: p.coins })}
                </span>
                {(isRoundEnd || isGameEnd) && coinDeltas && coinDeltas[i] !== 0 && (
                  <span
                    className={`text-xs font-semibold ${coinDeltas[i] > 0 ? 'text-ds-success' : 'text-ds-error-text'}`}
                    data-testid={`ulti-coin-delta-${p.id}`}
                  >
                    {t('coinDelta', { delta: formatSignedDelta(coinDeltas[i]) })}
                  </span>
                )}
                <span
                  className={`px-1.5 py-0.5 rounded text-xs ${p.isDeclarer ? badgeWarningColors : 'bg-ds-surface-elevated text-ds-text-muted'}`}
                >
                  {p.isDeclarer ? t('declarerBadge') : t('roleCoalition')}
                </span>
              </div>
            ))}
          </div>

          {/* Players: cards / tricks / captured points */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                </div>
              ))}
            </div>
          )}

          {/* Round result: the deal outcome (contract made / failed) */}
          {(isRoundEnd || isGameEnd) && state.outcome > 0 && (
            <div
              className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              role="status"
              aria-live="polite"
              data-testid="ulti-round-result"
            >
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
              {state.declarerIdx >= 0 && (
                <div>
                  {t('roundResult.declarer', {
                    name: playerName(state.declarerIdx, state.declarerIdx === humanIdx),
                  })}
                </div>
              )}
              {(isRoundEnd || isGameEnd) && coinDeltas && humanIdx >= 0 && (
                <div
                  className={
                    coinDeltas[humanIdx] > 0 ? 'text-ds-success' : coinDeltas[humanIdx] < 0 ? 'text-ds-error-text' : ''
                  }
                >
                  {t('roundResult.yourCoins', { delta: formatSignedDelta(coinDeltas[humanIdx] ?? 0) })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: UltiContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="ulti-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, t, view }: UltiContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === UltiPhase.BID;
  const isDiscardPhase = state.phase === UltiPhase.DISCARD;
  const isPlayPhase = state.phase === UltiPhase.PLAY;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canDiscard = isDiscardPhase && isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。
          宣言中と捨て札中は排他なので領域は 1 つでよい。タロンの案内も同じ群。 */}
      <div data-testid="ulti-prompt-live" role="status" aria-live="polite">
        {canBid && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="ulti-bid-prompt">
            {t('bidPhase')}
          </div>
        )}
        {/* **宣言と同時に手札が 10 → 12 枚に増える。**`applyBid` が伏せられた
            タロンを自動で加えるのに、その存在がどこにも出ておらず、宣言直後に
            手札が突然増えて見えた (#6486)。 */}
        {isBidPhase && !state.talonTaken && state.talonCount > 0 && (
          <div className="mb-1 text-center text-xs text-ds-text-muted" data-testid="ulti-talon-pending">
            {t('talonPending', { count: state.talonCount })}
          </div>
        )}
        {canDiscard && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="ulti-discard-prompt">
            <div>{t('discardPhase')}</div>
            <div className="text-ds-text-muted" data-testid="ulti-discard-progress">
              {t('discardProgress', { selected: game.selectedCardIndices.length, required: DISCARD_COUNT })}
            </div>
          </div>
        )}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={cardWidth}
          isMobile={isMobile}
          stackRows
          dataTutorialPrefix="ulti"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, busy, view }: UltiContext) {
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === UltiPhase.BID;
  const isDiscardPhase = state.phase === UltiPhase.DISCARD;
  const isPlayPhase = state.phase === UltiPhase.PLAY;
  const isTrickEnd = state.phase === UltiPhase.TRICK_END;
  const isRoundEnd = state.phase === UltiPhase.ROUND_END;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canDiscard = isDiscardPhase && isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  if (!(canBid || canDiscard || canPlay || isTrickEnd || isRoundEnd)) return null;

  const declareParty = () => {
    if (view.selectedTrump === null) return;
    game.handleBid('party', view.selectedTrump);
    view.setSelectedTrump(null);
  };

  const declareUlti = () => {
    if (view.selectedTrump === null) return;
    game.handleBid('ulti', view.selectedTrump);
    view.setSelectedTrump(null);
  };

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="ulti-action-buttons">
      {canBid && (
        <>
          <fieldset className="flex items-center gap-2 border-0 p-0 m-0 min-w-0">
            <legend className="text-ds-text-muted text-sm">{t('chooseTrump')}</legend>
            {TRUMP_CHOICES.map((c) => (
              <button
                key={c.code}
                type="button"
                className={view.selectedTrump === c.code ? btnPrimary : btnSecondary}
                onClick={() => view.setSelectedTrump(c.code)}
                disabled={busy}
                aria-label={t(SUIT_KEYS[c.code])}
                aria-pressed={view.selectedTrump === c.code}
              >
                {c.symbol}
              </button>
            ))}
          </fieldset>
          <button
            type="button"
            className={btnPrimary}
            onClick={declareParty}
            disabled={busy || view.selectedTrump === null}
            title={t('bidDesc.party')}
            aria-describedby="ulti-bid-desc-party"
          >
            {t('bidParty')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={declareUlti}
            disabled={busy || view.selectedTrump === null}
            title={t('bidDesc.ulti')}
            aria-describedby="ulti-bid-desc-ulti"
          >
            {t('bidUlti')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => game.handleBid('betli')}
            disabled={busy}
            title={t('bidDesc.betli')}
            aria-describedby="ulti-bid-desc-betli"
          >
            {t('bidBetli')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => game.handleBid('durchmarsch')}
            disabled={busy}
            title={t('bidDesc.durchmarsch')}
            aria-describedby="ulti-bid-desc-durchmarsch"
          >
            {t('bidDurchmarsch')}
          </button>
          {/* The win conditions the CUI states in promptBidHelp; Ulti's own
          (trump 7 takes the last trick) is the name of the game. */}
          <div className="sr-only" data-testid="ulti-bid-descriptions">
            {(['party', 'ulti', 'betli', 'durchmarsch'] as const).map((k) => (
              <span key={k} id={`ulti-bid-desc-${k}`}>
                {t(`bidDesc.${k}`)}
              </span>
            ))}
          </div>
        </>
      )}
      {canDiscard && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handleDiscard}
          disabled={busy || game.selectedCardIndices.length !== 2}
        >
          {t('discardButton')}
        </button>
      )}
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const ultiDefinition: GameDefinition<UltiResponse, UltiGame, Parameters<typeof ultiApi.exec>, UltiView> = {
  key: 'ulti',
  useGame: useUltiGame,
  useView: useUltiView,
  tutorial: ULTI_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 10 },
  cli: () => ({
    parseCommand: parseUltiCommand,
    formatResponse: formatUltiState,
    helpText: ULTI_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === UltiPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
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
          value: game.ultiConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.ultiConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  // Hand plus actions outgrow the 45vh cap on phones (#11655).
  footerTall: true,
  resetTutorialId: 'ulti-reset-button',
  actionsClassName: '',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Ulti (Ultimo) game page: a 3-player Hungarian contract 32-card trick-taker with a Party/Betli/Durchmarsch bid, a 2-card talon discard, and coin settlement. */
export const UltiPage = createGamePage(ultiDefinition);
