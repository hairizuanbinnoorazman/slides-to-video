package image2videoconverter

import (
	"context"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type GoogleTextToSpeech struct {
	logger            logger.Logger
	text2speechClient *texttospeech.Client
}

func NewGoogleTextToSpeech(l logger.Logger, t *texttospeech.Client) GoogleTextToSpeech {
	return GoogleTextToSpeech{
		logger:            l,
		text2speechClient: t,
	}
}

func (g *GoogleTextToSpeech) Generate(text string) ([]byte, error) {
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}
	resp, err := g.text2speechClient.SynthesizeSpeech(context.Background(), req)
	if err != nil {
		return []byte{}, err
	}
	return resp.AudioContent, nil
}
