package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/pkg/jsonmessage"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progressui"
	"google.golang.org/protobuf/proto"
)

// StreamBuildOutput streams Docker build output with progress display
func StreamBuildOutput(ctx context.Context, r io.Reader, w io.Writer) error {
	dec := json.NewDecoder(r)

	display, err := progressui.NewDisplay(w, progressui.AutoMode)
	if err != nil {
		return fmt.Errorf("error creating display: %w", err)
	}

	statusCh := make(chan *client.SolveStatus)
	doneCh := make(chan struct{})
	go func() {
		_, _ = display.UpdateFrom(ctx, statusCh)
		close(doneCh)
	}()

	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("decode json: %w", err)
		}

		if jm.Error != nil {
			return jm.Error
		}

		if jm.Aux != nil {
			switch jm.ID {
			case "moby.buildkit.trace":
				var raw []byte
				if err := json.Unmarshal(*jm.Aux, &raw); err != nil {
					return fmt.Errorf("unmarshal aux: %w", err)
				}
				var resp controlapi.StatusResponse
				if err := proto.Unmarshal(raw, &resp); err != nil {
					return fmt.Errorf("proto unmarshal: %w", err)
				}
				statusCh <- client.NewSolveStatus(&resp)
			default:
				fmt.Fprintf(w, "%s: %s\n", jm.ID, string(*jm.Aux))
			}
			continue
		}

		_ = jm.Display(w, true)
	}

	close(statusCh)
	<-doneCh
	return nil
}
