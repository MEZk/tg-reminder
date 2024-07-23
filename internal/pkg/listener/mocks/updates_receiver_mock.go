// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"sync"
)

// UpdateReceiverMock is a mock implementation of listener.UpdateReceiver.
//
//	func TestSomethingThatUsesUpdateReceiver(t *testing.T) {
//
//		// make and configure a mocked listener.UpdateReceiver
//		mockedUpdateReceiver := &UpdateReceiverMock{
//			OnCallbackQueryFunc: func(ctx context.Context, callback domain.TgCallbackQuery) error {
//				panic("mock out the OnCallbackQuery method")
//			},
//			OnMessageFunc: func(ctx context.Context, message domain.TgMessage) error {
//				panic("mock out the OnMessage method")
//			},
//		}
//
//		// use mockedUpdateReceiver in code that requires listener.UpdateReceiver
//		// and then make assertions.
//
//	}
type UpdateReceiverMock struct {
	// OnCallbackQueryFunc mocks the OnCallbackQuery method.
	OnCallbackQueryFunc func(ctx context.Context, callback domain.TgCallbackQuery) error

	// OnMessageFunc mocks the OnMessage method.
	OnMessageFunc func(ctx context.Context, message domain.TgMessage) error

	// calls tracks calls to the methods.
	calls struct {
		// OnCallbackQuery holds details about calls to the OnCallbackQuery method.
		OnCallbackQuery []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Callback is the callback argument value.
			Callback domain.TgCallbackQuery
		}
		// OnMessage holds details about calls to the OnMessage method.
		OnMessage []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Message is the message argument value.
			Message domain.TgMessage
		}
	}
	lockOnCallbackQuery sync.RWMutex
	lockOnMessage       sync.RWMutex
}

// OnCallbackQuery calls OnCallbackQueryFunc.
func (mock *UpdateReceiverMock) OnCallbackQuery(ctx context.Context, callback domain.TgCallbackQuery) error {
	if mock.OnCallbackQueryFunc == nil {
		panic("UpdateReceiverMock.OnCallbackQueryFunc: method is nil but UpdateReceiver.OnCallbackQuery was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Callback domain.TgCallbackQuery
	}{
		Ctx:      ctx,
		Callback: callback,
	}
	mock.lockOnCallbackQuery.Lock()
	mock.calls.OnCallbackQuery = append(mock.calls.OnCallbackQuery, callInfo)
	mock.lockOnCallbackQuery.Unlock()
	return mock.OnCallbackQueryFunc(ctx, callback)
}

// OnCallbackQueryCalls gets all the calls that were made to OnCallbackQuery.
// Check the length with:
//
//	len(mockedUpdateReceiver.OnCallbackQueryCalls())
func (mock *UpdateReceiverMock) OnCallbackQueryCalls() []struct {
	Ctx      context.Context
	Callback domain.TgCallbackQuery
} {
	var calls []struct {
		Ctx      context.Context
		Callback domain.TgCallbackQuery
	}
	mock.lockOnCallbackQuery.RLock()
	calls = mock.calls.OnCallbackQuery
	mock.lockOnCallbackQuery.RUnlock()
	return calls
}

// ResetOnCallbackQueryCalls reset all the calls that were made to OnCallbackQuery.
func (mock *UpdateReceiverMock) ResetOnCallbackQueryCalls() {
	mock.lockOnCallbackQuery.Lock()
	mock.calls.OnCallbackQuery = nil
	mock.lockOnCallbackQuery.Unlock()
}

// OnMessage calls OnMessageFunc.
func (mock *UpdateReceiverMock) OnMessage(ctx context.Context, message domain.TgMessage) error {
	if mock.OnMessageFunc == nil {
		panic("UpdateReceiverMock.OnMessageFunc: method is nil but UpdateReceiver.OnMessage was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Message domain.TgMessage
	}{
		Ctx:     ctx,
		Message: message,
	}
	mock.lockOnMessage.Lock()
	mock.calls.OnMessage = append(mock.calls.OnMessage, callInfo)
	mock.lockOnMessage.Unlock()
	return mock.OnMessageFunc(ctx, message)
}

// OnMessageCalls gets all the calls that were made to OnMessage.
// Check the length with:
//
//	len(mockedUpdateReceiver.OnMessageCalls())
func (mock *UpdateReceiverMock) OnMessageCalls() []struct {
	Ctx     context.Context
	Message domain.TgMessage
} {
	var calls []struct {
		Ctx     context.Context
		Message domain.TgMessage
	}
	mock.lockOnMessage.RLock()
	calls = mock.calls.OnMessage
	mock.lockOnMessage.RUnlock()
	return calls
}

// ResetOnMessageCalls reset all the calls that were made to OnMessage.
func (mock *UpdateReceiverMock) ResetOnMessageCalls() {
	mock.lockOnMessage.Lock()
	mock.calls.OnMessage = nil
	mock.lockOnMessage.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *UpdateReceiverMock) ResetCalls() {
	mock.lockOnCallbackQuery.Lock()
	mock.calls.OnCallbackQuery = nil
	mock.lockOnCallbackQuery.Unlock()

	mock.lockOnMessage.Lock()
	mock.calls.OnMessage = nil
	mock.lockOnMessage.Unlock()
}