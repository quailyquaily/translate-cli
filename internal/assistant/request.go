package assistant

import (
	"context"
	"errors"

	"github.com/lyricat/goutils/ai"
)

func (a *Assistant) AIRequestJSON(ctx context.Context, inst string) (*ai.Result, error) {
	if a.cfg.Provider == ai.ProviderSusanoo {
		rp := &ai.SusanoParams{
			Format: "json",
			Conditions: ai.SusanoParamsConditions{
				PreferredProvider: "bedrock",
			},
		}

		ret, err := a.aiInst.OneTimeRequestWithParams(ctx, inst, rp.ToMap())
		if err != nil {
			return nil, err
		}
		return ret, nil
	} else if a.cfg.Provider == ai.ProviderOpenAI {
		ret, err := a.aiInst.OneTimeRequestWithParams(ctx, inst, map[string]any{})
		if err != nil {
			return nil, err
		}
		// extract json from the response
		json, err := a.aiInst.GrabJsonOutput(ctx, ret.Text)
		if err != nil {
			return nil, err
		}
		ret.Json = json
		return ret, nil
	}

	return nil, errors.New("unsupported provider")
}

func (a *Assistant) AIRequestText(ctx context.Context, inst string) (string, error) {
	if a.cfg.Provider == ai.ProviderSusanoo {
		rp := &ai.SusanoParams{
			Format: "plaintext",
			Conditions: ai.SusanoParamsConditions{
				PreferredProvider: "bedrock",
			},
		}

		ret, err := a.aiInst.OneTimeRequestWithParams(ctx, inst, rp.ToMap())
		if err != nil {
			return "", err
		}
		return ret.Text, nil
	} else if a.cfg.Provider == ai.ProviderOpenAI {
		ret, err := a.aiInst.OneTimeRequestWithParams(ctx, inst, map[string]any{})
		if err != nil {
			return "", err
		}
		return ret.Text, nil
	}

	return "", errors.New("unsupported provider")
}
