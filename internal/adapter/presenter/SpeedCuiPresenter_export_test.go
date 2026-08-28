//go:build test

package presenter

import "time"

func SetSpeedCuiPresenterClock(p *SpeedCuiPresenter, now func() time.Time) {
	p.now = now
}
