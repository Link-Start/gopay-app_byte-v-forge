package appsvc

import "time"

const (
	signupRateLimitScopeProbe    = "signup_login_methods"
	signupRateLimitScopeMethods  = "signup_cvs_methods"
	signupRateLimitScopeInitiate = "signup_cvs_initiate"
)

func (s *Server) signupInitiateDelay() time.Duration {
	minDelay := s.cfg.SignupInitiateJitterMin
	maxDelay := s.cfg.SignupInitiateJitterMax
	if maxDelay <= 0 {
		return 0
	}
	if minDelay < 0 {
		minDelay = 0
	}
	if maxDelay < minDelay {
		minDelay, maxDelay = maxDelay, minDelay
	}
	if maxDelay == minDelay {
		return maxDelay
	}
	span := int64((maxDelay - minDelay) / time.Second)
	if span <= 0 {
		return minDelay
	}
	return minDelay + time.Duration(randomInt64(span+1))*time.Second
}
