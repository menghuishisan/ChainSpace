package handler

import (
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
)

func buildChallengeManageResponse(challenge *model.Challenge) response.ChallengeResponse {
	return response.BuildChallengeResponse(challenge, &response.BuildChallengeResponseOptions{
		IncludeManageFields: true,
	})
}

func buildChallengeManageResponses(challenges []model.Challenge) []response.ChallengeResponse {
	if len(challenges) == 0 {
		return []response.ChallengeResponse{}
	}

	result := make([]response.ChallengeResponse, 0, len(challenges))
	for i := range challenges {
		result = append(result, buildChallengeManageResponse(&challenges[i]))
	}
	return result
}
