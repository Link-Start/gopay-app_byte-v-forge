package app

type ClientSet struct {
	Auth     *AuthClient
	Customer *CustomerClient
	Gojek    *GojekClient

	session *session
}

func NewClientSet(cfg Config) (*ClientSet, error) {
	sess, err := newSession(cfg)
	if err != nil {
		return nil, err
	}
	return &ClientSet{
		Auth:     &AuthClient{domain: newDomainClient(sess, AuthBaseURL, "gopay-auth", authHeaderPolicy)},
		Customer: &CustomerClient{domain: newDomainClient(sess, CustomerBaseURL, "gopay-customer", customerHeaderPolicy)},
		Gojek:    &GojekClient{domain: newDomainClient(sess, GojekBaseURL, "gopay-gojek", gojekHeaderPolicy)},
		session:  sess,
	}, nil
}

func (c *ClientSet) Device() DeviceFingerprint {
	if c == nil || c.session == nil {
		return DeviceFingerprint{}
	}
	return c.session.device
}
