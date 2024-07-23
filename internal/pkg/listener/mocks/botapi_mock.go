// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

// BotAPIMock is a mock implementation of listener.BotAPI.
//
//	func TestSomethingThatUsesBotAPI(t *testing.T) {
//
//		// make and configure a mocked listener.BotAPI
//		mockedBotAPI := &BotAPIMock{
//			GetUpdatesChanFunc: func(config tbapi.UpdateConfig) tbapi.UpdatesChannel {
//				panic("mock out the GetUpdatesChan method")
//			},
//		}
//
//		// use mockedBotAPI in code that requires listener.BotAPI
//		// and then make assertions.
//
//	}
type BotAPIMock struct {
	// GetUpdatesChanFunc mocks the GetUpdatesChan method.
	GetUpdatesChanFunc func(config tbapi.UpdateConfig) tbapi.UpdatesChannel

	// calls tracks calls to the methods.
	calls struct {
		// GetUpdatesChan holds details about calls to the GetUpdatesChan method.
		GetUpdatesChan []struct {
			// Config is the config argument value.
			Config tbapi.UpdateConfig
		}
	}
	lockGetUpdatesChan sync.RWMutex
}

// GetUpdatesChan calls GetUpdatesChanFunc.
func (mock *BotAPIMock) GetUpdatesChan(config tbapi.UpdateConfig) tbapi.UpdatesChannel {
	if mock.GetUpdatesChanFunc == nil {
		panic("BotAPIMock.GetUpdatesChanFunc: method is nil but BotAPI.GetUpdatesChan was just called")
	}
	callInfo := struct {
		Config tbapi.UpdateConfig
	}{
		Config: config,
	}
	mock.lockGetUpdatesChan.Lock()
	mock.calls.GetUpdatesChan = append(mock.calls.GetUpdatesChan, callInfo)
	mock.lockGetUpdatesChan.Unlock()
	return mock.GetUpdatesChanFunc(config)
}

// GetUpdatesChanCalls gets all the calls that were made to GetUpdatesChan.
// Check the length with:
//
//	len(mockedBotAPI.GetUpdatesChanCalls())
func (mock *BotAPIMock) GetUpdatesChanCalls() []struct {
	Config tbapi.UpdateConfig
} {
	var calls []struct {
		Config tbapi.UpdateConfig
	}
	mock.lockGetUpdatesChan.RLock()
	calls = mock.calls.GetUpdatesChan
	mock.lockGetUpdatesChan.RUnlock()
	return calls
}

// ResetGetUpdatesChanCalls reset all the calls that were made to GetUpdatesChan.
func (mock *BotAPIMock) ResetGetUpdatesChanCalls() {
	mock.lockGetUpdatesChan.Lock()
	mock.calls.GetUpdatesChan = nil
	mock.lockGetUpdatesChan.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *BotAPIMock) ResetCalls() {
	mock.lockGetUpdatesChan.Lock()
	mock.calls.GetUpdatesChan = nil
	mock.lockGetUpdatesChan.Unlock()
}
