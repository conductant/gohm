package auth

import (
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestToken(t *testing.T) { TestingT(t) }

type TestSuiteToken struct {
}

var _ = Suite(&TestSuiteToken{})

func (suite *TestSuiteToken) SetUpSuite(c *C) {
}

func (suite *TestSuiteToken) TearDownSuite(c *C) {
}

// Generate the key pair using openssl
// openssl genrsa -out k1.key 4096  # private key
// openssl rsa -pubout -in k1.key -out k1.pub # public key
func testPublicKey() []byte {
	return []byte(`
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAphwicPL+LFZJWD2ab9MU
Q3NEG1GE9YhSLAN5zVIPfQ6MuLGoh80TNWiZPWO/sPWlOQVtQT+cZOubdEgsqnbo
HxAQGYeulszZ25P8qfNo0rr+l/mDnsWMw90B5WcOLxorjodludxbZ8ZTApYZ5SZY
vwSpaBzTZxf/B3yO/DUixMCqjPbuaS5h4U2A/upcTBoWU6sqBnJxF/Qv16jGfLq+
mdLl1B6AIbPV1XpS5YJyDSon2YE7YpKunfbAPnZVtZQV0uPy2WRiY9axa7rs0lyZ
2Ot1cK6kc+nhcGUYHoIaYHVO12sd+R8E9TJCNfYBX8+AApaTuKWPMVrwhUI4ttVX
aTt3JJ1wTKmxNsvBoOlKB+Uqkf1BEF8IqEYmFetjEZosH0Xj+C6599hHMHYdNBfx
hFAMZO8gf8lVXASRhO05mHK/UQ09/1GrBk0g3plkNGB4qOH4VfDYeyo/MQPCYdn1
uujIjCbBAUw7v5iZ1ft7eAV0DUUisJTezwzC7FTiFUelZPOZywxyO2e7IZ91eg8J
KrYqQJGtRc8b7LzLK4i1jCbtZErd0nPkmP44MpyqelaM51s9z6xIIeSntswFzDQq
xH2LxoI4ukNXXHQhrTKjLZYO8uzl0mjmNGlanHjIbnZV5g8BbhKiAWwx1bYUEB+J
AAycPYOKRs07UER2FGnK8a8CAwEAAQ==
-----END PUBLIC KEY-----
`)
}

func testPrivateKey() []byte {
	return []byte(`
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)
}

func (suite *TestSuiteToken) TestParseKeys(c *C) {
	_, err := RsaPublicKeyFromPem(testPublicKey)
	c.Assert(err, IsNil)
	_, err = RsaPrivateKeyFromPem(testPrivateKey)
	c.Assert(err, IsNil)
}

func (suite *TestSuiteToken) TestNewToken(c *C) {

	token := NewToken(1 * time.Hour)
	token.Add("foo1", "foo1").Add("foo2", "foo2").Add("count", 2)

	signedString, err := token.SignedString(testPrivateKey)
	c.Assert(err, IsNil)

	c.Log("token=", signedString)
	parsed, err := TokenFromString(signedString, testPublicKey, time.Now)
	c.Assert(err, IsNil)
	c.Assert(parsed.HasKey("count"), Equals, true)
	c.Assert(parsed.Get("count"), Equals, float64(2))
	c.Assert(parsed.Get("foo1"), DeepEquals, "foo1")
}

func (suite *TestSuiteToken) TestAuthTokenExpiration(c *C) {

	token := NewToken(1 * time.Hour)
	encoded, err := token.SignedString(testPrivateKey)

	// decode
	_, err = TokenFromString(encoded, testPublicKey, func() time.Time { return time.Now().Add(2 * time.Hour) })
	c.Assert(err, Equals, ErrExpiredAuthToken)
}

type uuid string

func (suite *TestSuiteToken) TestGetAppAuthTokenAuthRsaKey(c *C) {

	id := uuid("1234")

	token := NewToken(1*time.Hour).Add("appKey", id)
	encoded, err := token.SignedString(testPrivateKey)

	// decode
	parsed, err := TokenFromString(encoded, testPublicKey, time.Now)
	c.Assert(err, IsNil)

	appKey := parsed.GetString("appKey")
	c.Log("appkey=", appKey)
	c.Assert(uuid(appKey), DeepEquals, id)
}
