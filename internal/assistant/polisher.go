package assistant

import (
	"context"
	"fmt"
	"time"
)

type PolishInput struct {
	Content  string
	Lang     string
	LangCode string
}

func (a *Assistant) Polish(ctx context.Context, input *PolishInput) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	jaPrompt := fmt.Sprintf(`
	あなたはプロの日本語ライティングアシスタントです。
	ユーザーが日常や仕事で作成する日本語のメッセージを、以下の基準に基づいて校正し、より効果的で自然な表現に仕上げます。
	メッセージを短くわかりやすく、丁寧かつフレンドリーに、そして失礼のないトーンで整えてください。
	嫌味や誤解を避け、柔らかい表現に修正することを心がけます。
	ユーザーの意図を正確に汲み取り、より伝わりやすい文章に仕上げることを目標とします。
	この文章を簡潔で伝わりやすくしてください。
	詳細は説明せず、単に書き換えた結果を出力してください。

	以下の文章を校正してください：
%s
`, input.Content)

	inst := ""
	switch input.LangCode {
	case "ja", "ja-JP":
		inst = jaPrompt
	}

	if inst == "" {
		return input.Content, nil
	}

	return a.AIRequestText(ctx, inst)
}
