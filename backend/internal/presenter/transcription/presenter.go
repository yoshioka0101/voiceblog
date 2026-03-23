package presenter

import (
	"encoding/json"
	"fmt"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
)

func Transcription(value *entity.Transcription) (api.TranscriptionResponse, error) {
	segments, err := decodeSegmentsJSON(value.SegmentsJSON)
	if err != nil {
		return api.TranscriptionResponse{}, err
	}

	return api.TranscriptionResponse{
		Id:           value.ID,
		UserId:       value.UserID,
		FullText:     value.FullText,
		SegmentsJson: segments,
		CreatedAt:    value.CreatedAt,
		UpdatedAt:    value.UpdatedAt,
	}, nil
}

func Transcriptions(values []*entity.Transcription) (api.TranscriptionsResponse, error) {
	responses := make(api.TranscriptionsResponse, 0, len(values))
	for _, value := range values {
		response, err := Transcription(value)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func decodeSegmentsJSON(raw json.RawMessage) ([]map[string]interface{}, error) {
	var segments []map[string]interface{}
	if err := json.Unmarshal(raw, &segments); err != nil {
		return nil, fmt.Errorf("unmarshal segments_json: %w", err)
	}
	return segments, nil
}
