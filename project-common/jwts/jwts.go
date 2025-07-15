package jwts

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zeebo/errs"
	"time"
)

type JwtToken struct {
	AccessToken  string // 访问令牌
	RefreshToken string // 刷新令牌
	AccessExp    int64  // 访问令牌过期时间
	RefreshExp   int64  // 刷新令牌过期时间
}

// CreateToken val: 要存储在Token中的值，exp：访问令牌过期的时间，secret：访问令牌签名密钥，refreshExp：刷新令牌过期时间，refreshSecret：刷新令牌签名密钥
func CreateToken(val string, exp time.Duration, secret string, refreshExp time.Duration, refreshSecret string, ip string) *JwtToken {
	// 创建访问令牌
	aExp := time.Now().Add(exp).Unix() // 计算过期时间
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   aExp,
		"ip":    ip,
	})
	aToken, _ := accessToken.SignedString([]byte(secret)) // 签名生成token

	// 创建刷新令牌
	rExp := time.Now().Add(refreshExp).Unix()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   rExp,
	})
	rToken, _ := refreshToken.SignedString([]byte(refreshSecret))

	return &JwtToken{
		AccessExp:    aExp,
		AccessToken:  aToken,
		RefreshExp:   rExp,
		RefreshToken: rToken,
	}
}

func ParseToken(tokenString string, secret string, ip string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Printf("%v \n", claims)
		val := claims["token"].(string)
		exp := int64(claims["exp"].(float64))
		if exp <= time.Now().Unix() {
			return "", errs.New("token过期了")
		}
		//// 暂时警用一下
		//if claims["ip"] != ip {
		//	return "", errors.New("ip不合法")
		//}
		return val, nil
	} else {
		return "", err
	}

}
