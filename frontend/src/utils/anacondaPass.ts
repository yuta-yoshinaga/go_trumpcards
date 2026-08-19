/** The minimum a seat must expose for {@link anacondaPassRecipient}. */
export interface AnacondaPassSeat {
  /** Whether this seat is the human player. */
  isHuman: boolean;
  /** Whether this seat is out of the round (eliminated or folded). */
  out: boolean;
}

/**
 * Seat index the human passes their cards to.
 *
 * Anaconda passes to the left, but **skips seats that are out**, so with
 * eliminations "left" is not simply `id + 1`. Mirrors `anacondaPassRecipient` in
 * `internal/adapter/presenter/AnacondaCuiPresenter.go`; the two are held together
 * by `frontend/src/utils/__fixtures__/anacondaPassRecipient.golden.json`, which
 * both test suites assert against.
 *
 * @param seats - Every seat at the table, in table order.
 * @returns The receiving seat index, or -1 when the human is out or alone.
 */
export function anacondaPassRecipient(seats: readonly AnacondaPassSeat[]): number {
  let humanPos = -1;
  const participants: number[] = [];
  seats.forEach((seat, index) => {
    if (seat.out) return;
    if (seat.isHuman) humanPos = participants.length;
    participants.push(index);
  });
  if (humanPos < 0 || participants.length < 2) return -1;
  return participants[(humanPos + 1) % participants.length];
}
