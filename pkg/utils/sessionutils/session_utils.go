package sessionutils

import (
	"context"
	"fmt"
	"hw11_shopql/pkg/session"
)

func IdFromContex(ctx context.Context) (int, error) {
	if ctx.Value("tokens") == nil {
		return -1, fmt.Errorf("failed to fetch id")
	}
	userID := ctx.Value("tokens").(*session.Session).UserID
	return int(userID), nil
}
