// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package sender

import (
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

// Ensure, that BotAPIMock does implement BotAPI.
// If this is not the case, regenerate this file with moq.
var _ BotAPI = &BotAPIMock{}

// BotAPIMock is a mock implementation of BotAPI.
//
//	func TestSomethingThatUsesBotAPI(t *testing.T) {
//
//		// make and configure a mocked BotAPI
//		mockedBotAPI := &BotAPIMock{
//			SendFunc: func(c tbapi.Chattable) (tbapi.Message, error) {
//				panic("mock out the Send method")
//			},
//		}
//
//		// use mockedBotAPI in code that requires BotAPI
//		// and then make assertions.
//
//	}
type BotAPIMock struct {
	// SendFunc mocks the Send method.
	SendFunc func(c tbapi.Chattable) (tbapi.Message, error)

	// calls tracks calls to the methods.
	calls struct {
		// Send holds details about calls to the Send method.
		Send []struct {
			// C is the c argument value.
			C tbapi.Chattable
		}
	}
	lockSend sync.RWMutex
}

// Send calls SendFunc.
func (mock *BotAPIMock) Send(c tbapi.Chattable) (tbapi.Message, error) {
	if mock.SendFunc == nil {
		panic("BotAPIMock.SendFunc: method is nil but BotAPI.Send was just called")
	}
	callInfo := struct {
		C tbapi.Chattable
	}{
		C: c,
	}
	mock.lockSend.Lock()
	mock.calls.Send = append(mock.calls.Send, callInfo)
	mock.lockSend.Unlock()
	return mock.SendFunc(c)
}

// SendCalls gets all the calls that were made to Send.
// Check the length with:
//
//	len(mockedBotAPI.SendCalls())
func (mock *BotAPIMock) SendCalls() []struct {
	C tbapi.Chattable
} {
	var calls []struct {
		C tbapi.Chattable
	}
	mock.lockSend.RLock()
	calls = mock.calls.Send
	mock.lockSend.RUnlock()
	return calls
}

// ResetSendCalls reset all the calls that were made to Send.
func (mock *BotAPIMock) ResetSendCalls() {
	mock.lockSend.Lock()
	mock.calls.Send = nil
	mock.lockSend.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *BotAPIMock) ResetCalls() {
	mock.lockSend.Lock()
	mock.calls.Send = nil
	mock.lockSend.Unlock()
}
