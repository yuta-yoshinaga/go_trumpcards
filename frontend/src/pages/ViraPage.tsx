import { useEffect } from 'react';
import type { viraApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useViraGame } from '../hooks/useViraGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { ViraResponse } from '../types/card';
import { ViraContract, ViraPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { PREFERENCE_HELP, parseViraCommand } from '../utils/cli/commands/viraCommands';
import { formatViraState } from '../utils/cli/formatters/viraFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Contract i18n key suffixes indexed by contract value
 * (0=Pass 1=Gask 2=Solo 3=Misère 4=Vira).
 *
 * **Order matters and it is not Préférence's.** Misère sits third here, not
 * second, so a table copied from that game renders Solo as "Misère" and
 * silently changes the contract the player thinks they are watching.
 */
const CONTRACT_KEYS = ['pass', 'gask', 'solo', 'misere', 'vira'] as const;

/**
 * Target trick count for each contract, indexed by contract value
 * (0=Pass 1=Gask 2=Solo 3=Misère 4=Vira). Mirrors the Go domain
 * `ViraBidTarget`; Misère targets 0 tricks won.
 */
const CONTRACT_TARGET_TRICKS = [0, 7, 8, 0, 10] as const;

/** Whether the declarer has made their contract, failed it, or is still in progress. */
type ContractStatus = 'made' | 'failed' | 'progress';

/** Tailwind text color per contract status (made=success, in-progress=warning, failed=error). */
const CONTRACT_STATUS_COLOR: Readonly<Record<ContractStatus, string>> = {
  made: 'text-ds-success',
  failed: 'text-ds-error-text',
  progress: 'text-ds-warning',
};

/** The declarer's contract progress derived from tricks won and cards still in hand. */
interface ContractProgress {
  /** Tricks the declarer has won so far this round. */
  won: number;
  /** Tricks the declarer needs (0 for Misère). */
  needed: number;
  /** Tricks still available in the round. */
  remaining: number;
  /** Tricks still needed to reach the target, clamped at zero. */
  short: number;
  /** Made / failed / still in progress. */
  status: ContractStatus;
  /** Whether the contract is Misère (win no tricks). */
  isMisere: boolean;
}

/**
 * Computes the declarer's contract progress from tricks won and remaining tricks.
 * `remaining` is the declarer's card count (each remaining card equals one trick still to play).
 * Matches the Go domain `contractMade`: Misère succeeds on zero tricks, others on reaching the target.
 */
function computeContractProgress(contract: number, won: number, remaining: number): ContractProgress {
  const isMisere = contract === ViraContract.MISERE;
  const needed = CONTRACT_TARGET_TRICKS[contract] ?? 0;
  let status: ContractStatus;
  if (isMisere) {
    // Misère fails the instant a trick is won; it is only made once the round completes clean.
    if (won > 0) status = 'failed';
    else if (remaining === 0) status = 'made';
    else status = 'progress';
  } else if (won >= needed) {
    status = 'made';
  } else if (won + remaining < needed) {
    // Not enough tricks left to reach the target — failure is mathematically certain.
    status = 'failed';
  } else {
    status = 'progress';
  }
  return { won, needed, remaining, short: Math.max(needed - won, 0), status, isMisere };
}

/** Bid button options (Pass/Six/Misère/Seven/Eight). */
const BIDS: { value: number; key: string }[] = [
  // **並びは階梯そのもの。**Misère は Solo と Vira の間で、下位ではない。
  { value: ViraContract.PASS, key: 'bid.pass' },
  { value: ViraContract.GASK, key: 'bid.gask' },
  { value: ViraContract.SOLO, key: 'bid.solo' },
  { value: ViraContract.MISERE, key: 'bid.misere' },
  { value: ViraContract.VIRA, key: 'bid.vira' },
];

/** Vira tutorial step definitions. */
const PREFERENCE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="vira-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="vira-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vira-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vira-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PREFERENCE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ViraPhase.BID]: 'bid',
  [ViraPhase.PLAY]: 'play',
  [ViraPhase.TRICK_END]: 'trickEnd',
  [ViraPhase.ROUND_END]: 'roundEnd',
  [ViraPhase.GAME_END]: 'gameEnd',
};

type ViraGame = ReturnType<typeof useViraGame>;
type ViraView = ReturnType<typeof useViraView>;
type ViraContext = GamePageContext<ViraResponse, ViraGame, ViraView>;

function useViraView({ game }: { game: ViraGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('vira', PREFERENCE_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ game, state, t, view, busy }: ViraContext) {
  const loading = busy;
  const { cardWidth, isMobile } = view;

  const isBidPhase = state.phase === ViraPhase.BID;
  const isRoundEnd = state.phase === ViraPhase.ROUND_END;
  const isGameEnd = state.phase === ViraPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = state.isHumanBidTurn;

  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : suitSymbolAt(state.trumpSuit, '');

  // The current highest (non-pass) bid; a new non-pass bid must beat it.
  const highestBid = Math.max(0, ...state.bids);

  const contractName =
    state.declarerIdx >= 0 ? t(`contractName.${CONTRACT_KEYS[state.contract] ?? 'pass'}`) : t('contractUndecided');

  // Declarer's progress toward the contract, derived from tricks won and cards still in hand.
  const declarer = state.declarerIdx >= 0 ? state.players[state.declarerIdx] : undefined;
  const contractProgress =
    declarer && state.contract !== ViraContract.PASS
      ? computeContractProgress(state.contract, declarer.trickCount, declarer.cardCount)
      : undefined;

  // The settlement line names a declarer only when there was a contract (PASS has none).
  const settlementDeclarer = state.contract !== ViraContract.PASS ? state.players[state.declarerIdx] : undefined;
  const settlementContractKey = settlementDeclarer ? CONTRACT_KEYS[state.contract] : undefined;

  return (
    <>
      {isBidPhase && (
        <div className="mb-2">
          {/* The running bid stays visible while the CPUs bid; otherwise the auction is invisible until it is over (#4835). */}
          <div className="text-xs font-semibold text-ds-text-primary" data-testid="vira-highest-bid">
            {t('bidHighest', {
              name: highestBid > 0 ? t(`bid.${CONTRACT_KEYS[highestBid]}`) : t('bidNone'),
            })}
          </div>
          <div className="text-xs text-ds-text-muted" data-testid="vira-bid-status">
            <span className="font-semibold text-ds-text-primary">{t('bidStatusTitle')}</span>
            {state.players.map((player) => (
              <div key={player.id}>
                {playerName(player.id, player.isHuman)}:{' '}
                {state.bidDone[player.id] ? t(`bid.${CONTRACT_KEYS[state.bids[player.id]]}`) : t('bidNotYet')}
              </div>
            ))}
          </div>
          {isHumanBidTurn && (
            <div className="grid grid-cols-3 sm:grid-cols-5 gap-1 items-center" data-tutorial="vira-bid-choices">
              <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
              {BIDS.filter((bid) => bid.value !== ViraContract.PASS).map((b) => {
                // A non-pass bid must beat the current highest.
                const tooLow = b.value <= highestBid;
                const disabled = loading || tooLow;
                const isMisere = b.value === ViraContract.MISERE;
                const reason = tooLow ? t('bidTooLow') : undefined;
                return (
                  <span key={b.value} title={reason} className="inline-flex">
                    <button
                      type="button"
                      className={`px-3 py-2 rounded-lg text-ds-text-on-accent text-sm disabled:opacity-40 ${
                        isMisere ? 'bg-ds-warning ring-1 ring-ds-warning' : 'bg-ds-info'
                      }`}
                      onClick={() => game.handleBid(b.value)}
                      disabled={disabled}
                      aria-disabled={disabled}
                      aria-label={reason ? `${t(b.key)} — ${reason}` : undefined}
                      aria-describedby={isMisere ? 'vira-misere-desc' : undefined}
                      data-testid={`bid-${b.value}`}
                    >
                      {t(b.key)}
                      {isMisere && <span className="ml-1 text-[10px] opacity-80">{t('misereBadge')}</span>}
                    </button>
                    {isMisere && (
                      <span id="vira-misere-desc" className="sr-only">
                        {t('misereDesc')}
                      </span>
                    )}
                  </span>
                );
              })}
            </div>
          )}
        </div>
      )}
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="vira-info">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
        {/* ポットはラウンドをまたいで持ち越され、全員パスの流局でも積み上がる。 */}
        <span className="mr-4" data-testid="vira-pot" title={t('potHint')}>
          {t('pot', { pot: state.pot })}
        </span>
        <span>{t('target', { points: state.config.targetRounds })}</span>
      </div>

      <div className="text-ds-text-muted text-center mb-2 text-sm" data-testid="vira-declarer-line">
        {state.declarerIdx >= 0
          ? t('declarerLine', {
              name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
              contract: contractName,
            })
          : t('contractUndecided')}
      </div>

      {/* 切り札は選ぶものではなく、契約成立と同時に宣言者の最長スートへ
          自動設定される (Vira.go の resolveBidding)。切り札選択ボタンの
          あるゲームと並ぶと、いつ選んだのか分からない。姉妹の Solo Whist と
          Préférence は #5652 で同じ注記を入れてある (#7313)。
          ミゼールは切り札そのものが無いので出さない。 */}
      {state.contract !== ViraContract.PASS && state.contract !== ViraContract.MISERE && (
        <div className="text-ds-text-muted text-center mb-2 text-xs" data-testid="vira-trump-note">
          {t('trumpAutoNote')}
        </div>
      )}

      {contractProgress && (
        <div
          className={`text-center mb-2 text-sm font-semibold ${CONTRACT_STATUS_COLOR[contractProgress.status]}`}
          data-testid="vira-contract-progress"
        >
          {contractProgress.isMisere
            ? t('progress.misere', { won: contractProgress.won, remaining: contractProgress.remaining })
            : t('progress.line', {
                won: contractProgress.won,
                needed: contractProgress.needed,
                short: contractProgress.short,
              })}
          {contractProgress.status === 'made' && ` — ${t('progress.made')}`}
          {contractProgress.status === 'failed' && ` — ${t('progress.failed')}`}
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
            dataTutorial="vira-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* Per-player match scores with declarer badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
                {p.isDeclarer && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('declarerBadge')}</span>
                )}
              </div>
            ))}
          </div>

          {/* Players: cards / tricks */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                </div>
              ))}
            </div>
          )}

          {/* Round result: tricks won per player */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              {state.players.map((p) => (
                <div key={p.id}>
                  {t('roundResult.tricks', {
                    name: playerName(p.id, p.isHuman),
                    count: state.roundTricks[p.id] ?? 0,
                  })}
                </div>
              ))}
              {settlementDeclarer && settlementContractKey && (
                <div className="mt-1 text-ds-text-primary" data-testid="vira-contract-settlement">
                  {t(state.lastRoundMade ? 'roundResult.contractMade' : 'roundResult.contractFailed', {
                    name: playerName(state.declarerIdx, settlementDeclarer.isHuman),
                    contract: t(`contractName.${settlementContractKey}`),
                  })}
                </div>
              )}
              {/* 達成側は `pot` ではなく `lastRoundPotWon` を読む ── 精算が
                  `pot` を 0 に潰すので、取った額はそこには残っていない。 */}
              <div className="mt-1 text-ds-text-primary" data-testid="vira-pot-settlement">
                {state.lastRoundMade
                  ? t('roundResult.potWon', { chips: state.lastRoundPotWon })
                  : t('roundResult.potCarried', { chips: state.pot })}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: ViraContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="vira-hint-live" role="status" aria-live="polite">
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

function primaryActions({ game, state, t, busy }: ViraContext) {
  const isBidPhase = state.phase === ViraPhase.BID;
  const isPlayPhase = state.phase === ViraPhase.PLAY;
  const isTrickEnd = state.phase === ViraPhase.TRICK_END;
  const isRoundEnd = state.phase === ViraPhase.ROUND_END;
  const isHumanTurn = state.isHumanTurn;
  const isHumanBidTurn = state.isHumanBidTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  if (!((isBidPhase && isHumanBidTurn) || canPlay || isTrickEnd || isRoundEnd)) return null;

  return (
    <div className="grid grid-cols-3 sm:grid-cols-5 gap-1 items-center" data-tutorial="vira-action-buttons">
      {isBidPhase && isHumanBidTurn && (
        <button
          type="button"
          className={btnSecondary}
          onClick={() => game.handleBid(ViraContract.PASS)}
          disabled={busy}
          aria-disabled={busy}
          data-testid={`bid-${ViraContract.PASS}`}
        >
          {t('bid.pass')}
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

function FooterHand({ game, state, t, view }: ViraContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === ViraPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;

  if (!humanPlayer) return null;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="vira"
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  );
}

const viraDefinition: GameDefinition<ViraResponse, ViraGame, Parameters<typeof viraApi.exec>, ViraView> = {
  key: 'vira',
  useGame: useViraGame,
  useView: useViraView,
  tutorial: PREFERENCE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 10 },
  cli: () => ({
    parseCommand: parseViraCommand,
    formatResponse: formatViraState,
    helpText: PREFERENCE_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === ViraPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
    };
  },
  shell: ({ state }) => {
    const isGameEnd = state.phase === ViraPhase.GAME_END || state.gameEndFlag;
    return {
      isHumanTurn: (state.isHumanTurn || state.isHumanBidTurn) && !isGameEnd,
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
          value: game.viraConfig.cpuDifficulty,
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
          value: game.viraConfig.targetRounds,
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
  resetTutorialId: 'vira-reset-button',
  // The hand-wired page rendered a bare reset button; a class-less div lays out the same, unlike the default flex row.
  actionsClassName: '',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Vira game page: a Russian/Austrian 3-player trick-taker with a bidding phase. */
export const ViraPage = createGamePage(viraDefinition);
