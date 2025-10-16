package notifier

import "context"

type Interface interface {
    // 対象は自由に増やしてOK（ユーザ宛/ルーム宛 など）
    SendToUser(ctx context.Context, userID int, typ string, data any) error
    BroadcastRoom(ctx context.Context, roomID int, typ string, data any) error
}