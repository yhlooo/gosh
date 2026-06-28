package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
)

// agentOutputHandler 处理 Agent 输出
func (ctl *Controller) agentOutputHandler() func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
	curKind := ai.PartKind(-1)
	return func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		raw, _ := json.Marshal(chunk)
		ctl.logger.V(1).Info(fmt.Sprintf("model chunk: %s", string(raw)))

		for _, part := range chunk.Content {
			resetPrefix := ""
			if curKind != -1 {
				resetPrefix = "\x1b[0m\r\n"
			}
			switch {
			case part.IsReasoning():
				if curKind != ai.PartReasoning {
					_, _ = ctl.output.Write([]byte(resetPrefix + "\x1b[2m"))
					curKind = ai.PartReasoning
				}
				_, _ = ctl.output.Write([]byte(strings.ReplaceAll(part.Text, "\n", "\r\n")))

			case part.IsText() || part.IsData():
				if curKind != ai.PartText {
					_, _ = ctl.output.Write([]byte(resetPrefix))
					curKind = ai.PartText
				}
				_, _ = ctl.output.Write([]byte(strings.ReplaceAll(part.Text, "\n", "\r\n")))

			case part.IsToolRequest() && part.ToolRequest != nil:
				curKind = ai.PartToolRequest
				inputRaw, _ := json.Marshal(part.ToolRequest.Input)
				inputRawStr := string(inputRaw)
				if len(inputRawStr) > 80 {
					inputRawStr = inputRawStr[:80] + "..."
				}
				_, _ = ctl.output.Write([]byte(fmt.Sprintf(
					"%s\x1b[2;34mToolCall: %s %s\x1b[0m",
					resetPrefix,
					part.ToolResponse.Ref,
					inputRawStr,
				)))

			case part.IsToolResponse() && part.ToolResponse != nil:

				curKind = ai.PartToolResponse
				outputRaw, _ := json.Marshal(part.ToolResponse.Output)
				outputRawStr := string(outputRaw)
				if len(outputRawStr) > 80 {
					outputRawStr = outputRawStr[:80] + "..."
				}
				_, _ = ctl.output.Write([]byte(fmt.Sprintf(
					"%s\x1b[2;34m%s %s\x1b[0m",
					resetPrefix,
					strings.Repeat(" ", len(part.ToolResponse.Ref)+11),
					outputRawStr,
				)))
			}
		}

		return nil
	}
}
