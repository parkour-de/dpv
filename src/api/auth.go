package api

import (
	"crypto/sha256"
	"dpv/dpv/src/repository/t"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	authCache     sync.Map
	authCacheSize int32
	authLastSweep atomic.Int64
	rlMap         sync.Map
	rlLastSweep   atomic.Int64
)

func init() {
	authLastSweep.Store(time.Now().Unix())
	rlLastSweep.Store(time.Now().Unix())
}

func buildAuthCacheKey(username, password, storedHash string) string {
	uHex := hex.EncodeToString([]byte(username))
	pHex := hex.EncodeToString([]byte(password))
	hHex := hex.EncodeToString([]byte(storedHash))
	hash := sha256.Sum256([]byte(uHex + ":" + pHex + ":" + hHex))
	return hex.EncodeToString(hash[:])
}

func checkAuthCache(key string) bool {
	val, ok := authCache.Load(key)
	if !ok {
		return false
	}
	timestamp := val.(int64)
	if time.Now().Unix()-timestamp < 600 { // 10 minutes
		return true
	}
	return false
}

func putAuthCache(key string) {
	now := time.Now().Unix()
	_, loaded := authCache.LoadOrStore(key, now)
	if loaded {
		authCache.Store(key, now)
	} else {
		atomic.AddInt32(&authCacheSize, 1)
	}

	lastSweep := authLastSweep.Load()
	if atomic.LoadInt32(&authCacheSize) > 10000 || now-lastSweep > 300 {
		if authLastSweep.CompareAndSwap(lastSweep, now) {
			var currentSize int32
			authCache.Range(func(k, v interface{}) bool {
				ts := v.(int64)
				if now-ts > 600 { // 10 minutes
					authCache.Delete(k)
				} else {
					currentSize++
				}
				return true
			})
			atomic.StoreInt32(&authCacheSize, currentSize)
		}
	}
}

type rlState struct {
	mu         sync.Mutex
	Failures   int
	LockoutEnd time.Time
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}
	if ips := r.Header.Get("X-Forwarded-For"); ips != "" {
		tokens := strings.Split(ips, ",")
		return strings.TrimSpace(tokens[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return ip
	}
	return r.RemoteAddr
}

func checkRateLimit(ip, email string) error {
	now := time.Now()

	lastSweep := rlLastSweep.Load()
	if now.Unix()-lastSweep > 1200 {
		if rlLastSweep.CompareAndSwap(lastSweep, now.Unix()) {
			rlMap.Range(func(k, v interface{}) bool {
				state := v.(*rlState)
				state.mu.Lock()
				deletionTime := state.LockoutEnd.Add(time.Duration(state.Failures) * time.Hour)
				if now.After(deletionTime) {
					state.mu.Unlock()
					rlMap.Delete(k)
				} else {
					state.mu.Unlock()
				}
				return true
			})
		}
	}

	keys := []string{"ip:" + ip, "usr:" + email}
	for _, k := range keys {
		if val, ok := rlMap.Load(k); ok {
			state := val.(*rlState)
			state.mu.Lock()
			if state.Failures >= 5 && now.Before(state.LockoutEnd) {
				waitMinutes := int(state.LockoutEnd.Sub(now).Minutes()) + 1
				state.mu.Unlock()
				if strings.HasPrefix(k, "ip:") {
					return t.Errorf("too many failed login attempts from this IP address, try again in %d minutes", waitMinutes)
				}
				return t.Errorf("too many failed login attempts for this user, try again in %d minutes", waitMinutes)
			}
			state.mu.Unlock()
		}
	}
	return nil
}

func recordAuthFailure(ip, email string) {
	now := time.Now()

	keys := []string{"ip:" + ip, "usr:" + email}
	for _, k := range keys {
		var state *rlState
		val, ok := rlMap.Load(k)
		if ok {
			state = val.(*rlState)
		} else {
			newState := &rlState{}
			actual, _ := rlMap.LoadOrStore(k, newState)
			state = actual.(*rlState)
		}

		state.mu.Lock()
		state.Failures++
		if state.Failures >= 5 && state.Failures%5 == 0 {
			exp := (state.Failures / 5) - 1
			if exp > 10 {
				exp = 10
			}
			penalty := time.Duration(1<<exp) * 10 * time.Minute

			if state.LockoutEnd.Before(now) {
				state.LockoutEnd = now.Add(penalty)
			} else {
				state.LockoutEnd = state.LockoutEnd.Add(penalty)
			}
		}
		state.mu.Unlock()
	}
}

func recordAuthSuccessEmail(email string) {
	rlMap.Delete("usr:" + email)
}
