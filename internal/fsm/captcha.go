package fsm

import (
	"context"
	"sync"
	"time"
)

// CaptchaState represents the FSM state for captcha verification
type CaptchaState string

const (
	StateIdle          CaptchaState = "idle"
	StateWaitingAnswer CaptchaState = "waiting_answer"
)

// CaptchaData holds the verification data for a user
type CaptchaData struct {
	ChatID         int64
	UserID         int64
	Answer         string
	ExpiresAt      time.Time
	PhotoMessageID int
}

// CaptchaFSM manages captcha verification states
type CaptchaFSM struct {
	mu     sync.RWMutex
	states map[int64]*CaptchaData // key: userID
}

// NewCaptchaFSM creates a new captcha FSM manager
func NewCaptchaFSM() *CaptchaFSM {
	return &CaptchaFSM{
		states: make(map[int64]*CaptchaData),
	}
}

// SetState sets the captcha state for a user
func (f *CaptchaFSM) SetState(userID int64, data *CaptchaData) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states[userID] = data
}

// GetState gets the captcha state for a user
func (f *CaptchaFSM) GetState(userID int64) (*CaptchaData, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	data, ok := f.states[userID]
	return data, ok
}

// DeleteState removes the captcha state for a user
func (f *CaptchaFSM) DeleteState(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.states, userID)
}

// IsExpired checks if the captcha has expired
func (f *CaptchaFSM) IsExpired(userID int64) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	data, ok := f.states[userID]
	if !ok {
		return true
	}
	return time.Now().After(data.ExpiresAt)
}

// CleanupExpired removes all expired states
func (f *CaptchaFSM) CleanupExpired() {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	for userID, data := range f.states {
		if now.After(data.ExpiresAt) {
			delete(f.states, userID)
		}
	}
}

type captchaFSMKey struct{}

// WithCaptchaFSM adds CaptchaFSM to context
func WithCaptchaFSM(ctx context.Context, fsm *CaptchaFSM) context.Context {
	return context.WithValue(ctx, captchaFSMKey{}, fsm)
}

// GetCaptchaFSM retrieves CaptchaFSM from context
func GetCaptchaFSM(ctx context.Context) (*CaptchaFSM, bool) {
	fsm, ok := ctx.Value(captchaFSMKey{}).(*CaptchaFSM)
	return fsm, ok
}
