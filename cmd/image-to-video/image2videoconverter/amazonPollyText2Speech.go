package image2videoconverter

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type AmazonPollyTextToSpeech struct {
	logger      logger.Logger
	pollyClient *polly.Client
	voiceID     types.VoiceId
	engine      types.Engine
}

func NewAmazonPollyTextToSpeech(l logger.Logger, client *polly.Client, voiceID, engine string) AmazonPollyTextToSpeech {
	vid := types.VoiceIdJoanna
	if voiceID != "" {
		vid = types.VoiceId(voiceID)
	}

	eng := types.EngineNeural
	if engine != "" {
		eng = types.Engine(engine)
	}

	return AmazonPollyTextToSpeech{
		logger:      l,
		pollyClient: client,
		voiceID:     vid,
		engine:      eng,
	}
}

func (a *AmazonPollyTextToSpeech) Generate(text string) ([]byte, error) {
	input := &polly.SynthesizeSpeechInput{
		Text:         &text,
		OutputFormat: types.OutputFormatMp3,
		VoiceId:      a.voiceID,
		Engine:       a.engine,
	}

	resp, err := a.pollyClient.SynthesizeSpeech(context.Background(), input)
	if err != nil {
		return []byte{}, err
	}
	defer resp.AudioStream.Close()

	audioBytes, err := io.ReadAll(resp.AudioStream)
	if err != nil {
		return []byte{}, err
	}

	return audioBytes, nil
}
