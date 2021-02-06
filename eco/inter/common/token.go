package common

import (
	"encoding/json"
	"fmt"
	"time"

	std "github.com/wednesdaysunny/onerpc/eco/inter"
	tc "github.com/wednesdaysunny/onerpc/eco/inter/toolkit/crypto"
)

const (
	TokenKey               = "token_secrect"
)


type Token struct {
	AppType              string
	UID                  int64
	Expiration           int64
	DeviceType           string
	ClientType           string
}

func NewTokenFromString(content string) (*Token, error) {
	data, err := tc.Base64Decode(content)
	if err != nil {
		return nil, std.ErrIllegalToken
	}
	data, err = tc.Decrypt(data, []byte(TokenKey))
	if err != nil {
		return nil, std.ErrIllegalToken
	}
	token := new(Token)
	if err = json.Unmarshal(data, &token); err != nil {
		return nil, std.ErrIllegalToken
	}
	return token, nil
}

func NewToken(appType string, uid int64, outTime int64, deviceType string, clientType string, userType string) *Token {
	token := new(Token)
	token.AppType = appType
	token.UID = uid
	token.Expiration = time.Now().Unix() + outTime
	token.DeviceType = deviceType
	token.ClientType = clientType
	return token
}


func (o *Token) GenTokenString() string {
	data, err := json.Marshal(o)
	if err != nil {
		fmt.Println("GenTokenString Error:", err)
		return ""
	}
	data, err = tc.Encrypt(data, []byte(TokenKey))
	if err != nil {
		fmt.Println("GenTokenString Encrypt Error:", err)
		return ""
	}
	return tc.Base64Encode(data)
}
