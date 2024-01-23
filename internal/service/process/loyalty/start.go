package loyalty

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
)

func Start(ctx context.Context, cfg *config.Config, repository repository.Repository) {

	ticker := time.NewTicker(time.Duration(1) * time.Second)

	for {
		select {
		case <-ticker.C:
			LoyaltyRun(ctx, cfg, repository, ticker)
		case <-ctx.Done():
			fmt.Println("Loyalty Stop")
			return
		}
	}
}
