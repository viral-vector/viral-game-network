package auth

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
	"testing"
	"time"
	"viral-game-network/src/cache"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func randomSigningKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(key)
}

func tokenFixture(t *testing.T) {
	t.Helper()
	support.Redis(t)
	t.Setenv("VNET_TOKEN_KEY", randomSigningKey(t))
	t.Setenv("VNET_KEY", "test-api-key")
	cache.Set("system-configs", []dbtype.SystemConfig{{Key: "VNET_NAME", Val: "Test Network"}, {Key: "VNET_TOKEN_EXPIRE", Val: "30"}}, time.Hour)
}

func TestTokenRoundTrip(t *testing.T) {
	tokenFixture(t)
	encoded, err := GenerateToken("alice", "User:alice", UserAudience)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidateToken(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "User:alice" || len(claims.Audience) != 1 || claims.Audience[0] != string(UserAudience) || claims.Username != "alice" || claims.Issuer != "Test Network" || time.Until(claims.ExpiresAt.Time) < 29*time.Minute {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestRejectInvalidTokens(t *testing.T) {
	tokenFixture(t)
	for _, tc := range []struct {
		name   string
		method jwt.SigningMethod
		key    []byte
		expiry time.Time
	}{
		{"expired", jwt.SigningMethodHS256, []byte(os.Getenv("VNET_TOKEN_KEY")), time.Now().Add(-time.Minute)},
		{"wrong key", jwt.SigningMethodHS256, []byte("wrong"), time.Now().Add(time.Hour)},
		{"wrong algorithm", jwt.SigningMethodHS384, []byte(os.Getenv("VNET_TOKEN_KEY")), time.Now().Add(time.Hour)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwt.NewWithClaims(tc.method, Claims{Username: "alice", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(tc.expiry)}}).SignedString(tc.key)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ValidateToken(token); err == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
	for _, token := range []string{"", "broken", "a.b.c"} {
		if _, err := ValidateToken(token); err == nil {
			t.Errorf("accepted %q", token)
		}
	}
}

func TestTokenExpirationConfiguration(t *testing.T) {
	tokenFixture(t)
	cache.Set("system-configs", []dbtype.SystemConfig{{Key: "VNET_NAME"}, {Key: "VNET_TOKEN_EXPIRE", Val: "invalid"}}, time.Hour)
	if _, err := GenerateToken("alice", "User:alice", UserAudience); err == nil {
		t.Fatal("invalid expiry accepted")
	}
	t.Setenv("VNET_TOKEN_EXPIRE", "10")
	cache.Set("system-configs", []dbtype.SystemConfig{{Key: "VNET_NAME"}, {Key: "VNET_TOKEN_EXPIRE"}}, time.Hour)
	encoded, err := GenerateToken("alice", "User:alice", UserAudience)
	if err != nil {
		t.Fatal(err)
	}
	claims, _ := ValidateToken(encoded)
	if claims.Issuer != "VNet" {
		t.Fatal(claims.Issuer)
	}
}

func TestAppKeys(t *testing.T) {
	support.Redis(t)
	cache.Set("system-configs-api-keys", []dbtype.SystemConfig{{Val: "allowed-key"}}, time.Hour)
	for _, tc := range []struct {
		key   string
		valid bool
	}{{"allowed-key", true}, {"wrong", false}, {"", false}} {
		valid, err := ValidateAppKey(tc.key)
		if valid != tc.valid || (err == nil) != tc.valid {
			t.Errorf("%q: %v %v", tc.key, valid, err)
		}
	}
	a, b := GenerateAppKey(20), GenerateAppKey(20)
	if len(a) != 20 || strings.Contains(a, "-") || a == b {
		t.Fatal("invalid generated keys")
	}
}

func TestPasswordHashing(t *testing.T) {
	hash, err := HashGenerate("secret-password")
	if err != nil || hash == "secret-password" {
		t.Fatalf("%s %v", hash, err)
	}
	if ok, err := HashValidate("secret-password", hash); !ok || err != nil {
		t.Fatalf("%v %v", ok, err)
	}
	if ok, err := HashValidate("wrong", hash); ok || err == nil {
		t.Fatal("wrong password accepted")
	}
	if _, err := HashGenerate(strings.Repeat("a", 73)); err == nil {
		t.Fatal("bcrypt length limit ignored")
	}
}

func TestTokenSigningKeyIsPrivateAndConfigured(t *testing.T) {
	tokenFixture(t)
	for _, key := range []string{"", "short", strings.Repeat("a", 32)} {
		t.Setenv("VNET_TOKEN_KEY", key)
		if len(key) >= 32 {
			t.Setenv("VNET_KEY", key)
		}
		if err := ValidateSigningKey(); err == nil {
			t.Fatal("unsafe signing configuration accepted")
		}
		if _, err := GenerateToken("alice", "User:alice", UserAudience); err == nil {
			t.Fatal("minted token with unsafe signing configuration")
		}
	}
	t.Setenv("VNET_KEY", "test-api-key")
	t.Setenv("VNET_TOKEN_KEY", randomSigningKey(t))
	claims := Claims{Username: "admin", RegisteredClaims: jwt.RegisteredClaims{Subject: "Admin:admin", Audience: jwt.ClaimStrings{string(AdminAudience)}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("VNET_KEY")))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateTokenFor(forged, AdminAudience); err == nil {
		t.Fatal("API key forged an admin session")
	}
}

func TestScopedTokensRequireAnIdentityAndExpiration(t *testing.T) {
	tokenFixture(t)
	for _, tc := range []struct {
		subject  string
		audience TokenAudience
	}{{"User:alice", UserAudience}, {"Admin:alice", AdminAudience}} {
		token, err := GenerateToken("alice", tc.subject, tc.audience)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ValidateTokenFor(token, tc.audience); err != nil {
			t.Fatal(err)
		}
		other := UserAudience
		if tc.audience == UserAudience {
			other = AdminAudience
		}
		if _, err := ValidateTokenFor(token, other); err == nil {
			t.Fatal("token accepted for another audience")
		}
	}
	valid := jwt.RegisteredClaims{Subject: "User:alice", Audience: jwt.ClaimStrings{string(UserAudience)}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
	for _, tc := range []struct {
		name   string
		change func(*jwt.RegisteredClaims)
	}{
		{"no expiration", func(c *jwt.RegisteredClaims) { c.ExpiresAt = nil }},
		{"no subject", func(c *jwt.RegisteredClaims) { c.Subject = "" }},
		{"wrong subject table", func(c *jwt.RegisteredClaims) { c.Subject = "Admin:alice" }},
		{"no audience", func(c *jwt.RegisteredClaims) { c.Audience = nil }},
		{"multiple audiences", func(c *jwt.RegisteredClaims) {
			c.Audience = jwt.ClaimStrings{string(UserAudience), string(AdminAudience)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := valid
			tc.change(&claims)
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{Username: "alice", RegisteredClaims: claims}).SignedString([]byte(os.Getenv("VNET_TOKEN_KEY")))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ValidateToken(token); err == nil {
				t.Fatal("invalid claims accepted")
			}
		})
	}
}

func TestRejectNonpositiveAndOverflowingTokenLifetimes(t *testing.T) {
	tokenFixture(t)
	for _, expiry := range []string{"0", "-1", "153722867280912931", "9223372036854775807"} {
		cache.Set("system-configs", []dbtype.SystemConfig{{Key: "VNET_TOKEN_EXPIRE", Val: expiry}}, time.Hour)
		if _, err := GenerateToken("alice", "User:alice", UserAudience); err == nil {
			t.Fatalf("accepted lifetime %s", expiry)
		}
	}
}

func TestTokenConfigurationFailureReturnsError(t *testing.T) {
	tokenFixture(t)
	cache.Del("system-configs")
	if _, err := GenerateToken("alice", "User:alice", UserAudience); err == nil {
		t.Fatal("database failure was silently replaced by default configuration")
	}
}
