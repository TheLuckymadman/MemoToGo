package llm

import (
	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"go.uber.org/zap"
)

type State struct {
	store             map[int64][]llmtype.Message
	msgCleanThreshold int
	maxRemoveMsg      int
	deps              *deps.Deps
}

func NewState(maxMsg int, maxRemoveMsg int, deps *deps.Deps) *State {
	return &State{
		store:             make(map[int64][]llmtype.Message),
		msgCleanThreshold: maxMsg,
		maxRemoveMsg:      maxRemoveMsg,
		deps:              deps,
	}
}

func (s *State) AddMessage(userID int64, msgs []llmtype.Message) {
	userState, ok := s.store[userID]
	if !ok {
		s.store[userID] = msgs
		return
	}

	if len(userState) >= s.msgCleanThreshold {
		s.deps.Logger.Info("State.AddMessage: message threshold reached", zap.Int("Msg count", len(userState)))
		userState = s.CleanMsgState(userState)
		s.deps.Logger.Info("State.AddMessage: message clean finished", zap.Int("Msg count", len(userState)))
	}
	userState = append(userState, msgs...)
	s.store[userID] = userState

}

func (s *State) CleanMsgState(userState []llmtype.Message) []llmtype.Message {
	systemMsg := userState[0]
	recentMsgsIdx := getRecentMsgIdx(userState, "user")
	recentMsgs := userState[recentMsgsIdx:]
	msgForRemove := userState[1:recentMsgsIdx]
	userState = []llmtype.Message{systemMsg}

	if len(msgForRemove) > s.maxRemoveMsg {
		userState = append(userState, msgForRemove[:s.maxRemoveMsg]...)
	}
	userState = append(userState, recentMsgs...)
	return userState
}

func getRecentMsgIdx(messages []llmtype.Message, role string) int {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == role {
			return i
		}
	}
	return 0
}
