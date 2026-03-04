package scriptgenerator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type BedrockClaude struct {
	logger  logger.Logger
	client  *bedrockruntime.Client
	modelID string
}

func NewBedrockClaude(logger logger.Logger, region, modelID string) (BedrockClaude, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return BedrockClaude{}, fmt.Errorf("unable to load AWS config for Bedrock: %v", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	logger.Infof("Bedrock Claude script generator initialized. Region: %s, Model: %s", region, modelID)

	return BedrockClaude{
		logger:  logger,
		client:  client,
		modelID: modelID,
	}, nil
}

type bedrockMessage struct {
	Role    string           `json:"role"`
	Content []bedrockContent `json:"content"`
}

type bedrockContent struct {
	Type   string              `json:"type"`
	Text   string              `json:"text,omitempty"`
	Source *bedrockImageSource `json:"source,omitempty"`
}

type bedrockImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type bedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens        int              `json:"max_tokens"`
	Messages         []bedrockMessage `json:"messages"`
}

type bedrockResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (b BedrockClaude) GenerateDescription(ctx context.Context, slides []SlideImage) (string, error) {
	b.logger.Info("Generating deck description from slide images via Bedrock")

	var contentBlocks []bedrockContent
	for _, slide := range slides {
		contentBlocks = append(contentBlocks, bedrockContent{
			Type: "image",
			Source: &bedrockImageSource{
				Type:      "base64",
				MediaType: "image/png",
				Data:      base64.StdEncoding.EncodeToString(slide.Content),
			},
		})
	}
	contentBlocks = append(contentBlocks, bedrockContent{
		Type: "text",
		Text: "These are all the slides from a presentation deck. Please provide a concise description (2-4 sentences) of what this presentation is about, including the main topic, key themes, and target audience. This description will be used as context when generating narration scripts for each slide.",
	})

	req := bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1024,
		Messages: []bedrockMessage{
			{
				Role:    "user",
				Content: contentBlocks,
			},
		},
	}

	return b.invoke(ctx, req)
}

func (b BedrockClaude) GenerateScript(ctx context.Context, description string, slide SlideImage) (string, error) {
	b.logger.Infof("Generating narration script for slide order=%d via Bedrock", slide.Order)

	contentBlocks := []bedrockContent{
		{
			Type: "text",
			Text: fmt.Sprintf("You are generating narration scripts for a presentation. The presentation is about: %s\n\nGenerate a natural-sounding narration script for the following slide (slide %d). The script should be suitable for text-to-speech conversion. Keep it concise and conversational. Output only the narration text, nothing else.", description, slide.Order),
		},
		{
			Type: "image",
			Source: &bedrockImageSource{
				Type:      "base64",
				MediaType: "image/png",
				Data:      base64.StdEncoding.EncodeToString(slide.Content),
			},
		},
	}

	req := bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages: []bedrockMessage{
			{
				Role:    "user",
				Content: contentBlocks,
			},
		},
	}

	return b.invoke(ctx, req)
}

func (b BedrockClaude) invoke(ctx context.Context, req bedrockRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("unable to marshal Bedrock request: %v", err)
	}

	output, err := b.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.modelID),
		ContentType: aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		return "", fmt.Errorf("Bedrock InvokeModel failed: %v", err)
	}

	var resp bedrockResponse
	err = json.Unmarshal(output.Body, &resp)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal Bedrock response: %v", err)
	}

	for _, c := range resp.Content {
		if c.Type == "text" {
			return c.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in Bedrock response")
}
