package mynats

import (
	"context"
	"encoding/json"

	"github.com/gianglt2198/platforms/pkg/utils"
)

func UseCase[T any, U any](
	f func(context.Context, T) (U, error),
) func(context.Context, []byte) ([]byte, error) {
	return func(ctx context.Context, payload []byte) ([]byte, error) {
		input, err := utils.TransformToType[T](payload)
		if err != nil {
			return nil, err
		}

		output, aerr := f(ctx, *input)
		var replyData []byte
		if aerr != nil {
			replyData, _ = json.Marshal(aerr)
		} else {
			replyData, _ = json.Marshal(output)
		}

		return replyData, nil
	}
}

func Subscriber[T any](
	f func(context.Context, T) error,
) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		input, err := utils.TransformToType[T](payload)
		if err != nil {
			return err
		}

		return f(ctx, *input)
	}
}
