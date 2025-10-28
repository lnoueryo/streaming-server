package live_video_request

import (
	"context"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
)

func GetOfferRequest(ctx context.Context, msg interface{}) (*live_video_dto.Params, error) {
	var rawParams = &RawParams{}
	params, err := rawParams.parse(ctx);if err != nil {
		return nil, err
	}
	return params, nil
}
