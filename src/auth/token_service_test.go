package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"testing"
	"time"
	"viral-game-network/src/cache"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func tokenFixture(t *testing.T) {
	t.Helper()
	support.Redis(t)
	previous := vnet_key
	vnet_key = []byte("test-signing-key")
	t.Cleanup(func() { vnet_key = previous })
	cache.Set("system-configs", []dbtype.SystemConfig{{Key: "VNET_NAME", Val: "Test Network"}, {Key: "VNET_TOKEN_EXPIRE", Val: "30"}}, time.Hour)
}

func TestTokenRoundTrip(t *testing.T) {
	tokenFixture(t)
	encoded, err := GenerateToken("alice")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidateToken(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Username != "alice" || claims.Issuer != "Test Network" || time.Until(claims.ExpiresAt.Time) < 29*time.Minute {
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
		{"expired", jwt.SigningMethodHS256, vnet_key, time.Now().Add(-time.Minute)},
		{"wrong key", jwt.SigningMethodHS256, []byte("wrong"), time.Now().Add(time.Hour)},
		{"wrong algorithm", jwt.SigningMethodHS384, vnet_key, time.Now().Add(time.Hour)},
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
	if _, err := GenerateToken("alice"); err == nil {
		t.Fatal("invalid expiry accepted")
	}
	t.Setenv("VNET_TOKEN_EXPIRE", "10")
	cache.Set("system-configs", []dbtype.SystemConfig{{Key: "VNET_NAME"}, {Key: "VNET_TOKEN_EXPIRE"}}, time.Hour)
	encoded, err := GenerateToken("alice")
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
