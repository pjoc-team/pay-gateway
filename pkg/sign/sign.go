package sign

import (
	"bytes"
	"context"
	"crypto"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blademainer/commons/pkg/util"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/validator"
	"github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

func init() {
	initCheckSignMap()
	validator.RegisterValidator(NewCheckSignValidator())
}

// CheckSignValidator validator of sign
type CheckSignValidator struct {
	ParamsCompacter ParamsCompacter
}

// Validate implements validator
func (validator *CheckSignValidator) Validate(ctx context.Context, request pay.PayRequest, cfg validator.GetMerchantConfigFunc) (e error) {
	paramsString := validator.ParamsCompacter.ParamsToString(request)
	log := logger.ContextLog(ctx)
	if log.IsDebugEnabled() {
		log.Debugf("Build interface: %v to string: %v", request, paramsString)
	}
	config, err := cfg(ctx, request.AppId)
	if err != nil {
		e = fmt.Errorf("could'nt found config of appID: %v", request.AppId)
		log.Errorf("couldn't found config of appID: %v request: %v", request.AppId, request)
		return e
	}
	e = CheckSign(ctx, request.GetCharset(), paramsString, request.GetSign(), config, Type(request.SignType))

	return
}

// NewCheckSignValidator new
func NewCheckSignValidator() *CheckSignValidator {
	v := &CheckSignValidator{}
	v.ParamsCompacter = NewParamsCompacter(&pay.PayRequest{}, "json", []string{"sign"}, true, "&", "=")
	return v
}

// CheckSignInterface check sign interface
type CheckSignInterface interface {
	checkSign(ctx context.Context, source []byte, signMsg string, key string) error
	sign(ctx context.Context, source []byte, key string) (string, error)
	getCheckSignKey(ctx context.Context, config *configclient.MerchantConfig) string
	getSignKey(ctx context.Context, config *configclient.MerchantConfig) string
	signType() Type
}

var checkSignMap = make(map[Type]CheckSignInterface)

func initCheckSignMap() {
	checkSignMap[TypeMd5] = &Md5{}
	checkSignMap[TypeSha256WithRSA] = &Sha256WithRSA{}
}

// CheckSign check sign
func CheckSign(ctx context.Context, charset string, source string, signMsg string, config *configclient.MerchantConfig, signType Type) (err error) {
	if signType == "" {
		signType = TypeSha256WithRSA
	}
	log := logger.ContextLog(ctx)
	signFunc := checkSignMap[signType]
	var sourceBytes []byte
	if key := signFunc.getCheckSignKey(ctx, config); key == "" {
		err = errors.New("couldn't found key")
		log.Errorf("couldn't get key from config: %v", config)
		return err
	} else if sourceBytes, err = stringToBytes(source, charset); err != nil {
		log.Errorf("failed to get charset: %s, error: %s", charset, err.Error())
		return fmt.Errorf("unknown charset: %s", charset)
	} else if signFunc == nil {
		log.Errorf("failed to get signType: %s, error: %s", signType, err.Error())
		e := fmt.Errorf("unknown signtype: %s", charset)
		return e
	} else if err = signFunc.checkSign(ctx, sourceBytes, signMsg, key); err != nil {
		log.Errorf("failed to check sign! error: %s", err.Error())
		e := fmt.Errorf("failed to check sign")
		return e
	} else {
		return nil
	}
}

// GenerateSign generate sign
func GenerateSign(ctx context.Context, charset string, source string, config *configclient.MerchantConfig, signType Type) (sign string, err error) {
	log := logger.ContextLog(ctx)
	signFunc := checkSignMap[signType]
	var sourceBytes []byte
	if key := signFunc.getSignKey(ctx, config); key == "" {
		err = errors.New("couldn't found key")
		log.Errorf("couldn't get key from config: %v", config)
		return
	} else if sourceBytes, err = stringToBytes(source, charset); err != nil {
		log.Errorf("failed to get charset: %s, error: %s", charset, err.Error())
		err = fmt.Errorf("unknown charset: %s", charset)
		return
	} else if signFunc == nil {
		log.Errorf("failed to get signType: %s, error: %s", signType, err.Error())
		err = fmt.Errorf("unknown signtype: %s", charset)
		return
	} else if sign, err = signFunc.sign(ctx, sourceBytes, key); err != nil {
		log.Errorf("failed to sign! error: %s", err.Error())
		err = fmt.Errorf("failed to sign")
		return
	} else {
		return
	}
}

// Md5 md5 sign
type Md5 struct {
}

func (m *Md5) getCheckSignKey(ctx context.Context, config *configclient.MerchantConfig) string {
	return config.Md5Key
}

func (m *Md5) getSignKey(ctx context.Context, config *configclient.MerchantConfig) string {
	return config.Md5Key
}

func (m *Md5) sign(ctx context.Context, source []byte, key string) (string, error) {
	buffer := bytes.NewBuffer(source)
	buffer.Write([]byte(key))
	b := buffer.Bytes()
	sum := md5.Sum(b)
	s := hex.EncodeToString(sum[:])
	return s, nil
}

func (m *Md5) checkSign(ctx context.Context, source []byte, signMsg string, key string) error {
	log := logger.ContextLog(ctx)
	generated, e := m.sign(ctx, source, key)
	if e != nil {
		log.Errorf("failed to generate sign! error: %v", e.Error())
		return e
	}
	if !util.EqualsIgnoreCase(generated, signMsg) {
		e := errors.New("check sign error")
		log.Warnf("failed to check sign! ours: %v actual: %v", generated, signMsg)
		return e
	}

	return nil
}

func (*Md5) signType() Type {
	return TypeMd5
}

// Sha256WithRSA rsa sign
type Sha256WithRSA struct {
}

func (s *Sha256WithRSA) getCheckSignKey(ctx context.Context, config *configclient.MerchantConfig) string {
	return config.MerchantRSAPublicKey
}

func (s *Sha256WithRSA) getSignKey(ctx context.Context, config *configclient.MerchantConfig) string {
	return config.GatewayRSAPrivateKey
}

func (s *Sha256WithRSA) sign(ctx context.Context, source []byte, key string) (sign string, err error) {
	log := logger.ContextLog(ctx)

	signBytes, err := PKCS1v15WithStringKey(source, key, crypto.SHA256)
	if err != nil {
		log.Errorf("failed to sign! error: %v key: %v", err.Error(), key)
		return
	}
	sign = base64.StdEncoding.EncodeToString(signBytes)
	log.Debugf("encode source: %v to sign: %v", string(source), sign)
	return
}

func (*Sha256WithRSA) checkSign(ctx context.Context, source []byte, signMsg string, key string) (err error) {
	log := logger.ContextLog(ctx)

	sign, err := base64.StdEncoding.DecodeString(signMsg)
	if err != nil {
		log.Errorf("failed to check sign! decode sign: %v with error: %v", signMsg, err.Error())
		return
	}
	err = VerifyPKCS1v15WithStringKey(source, sign, key, crypto.SHA256)
	if err != nil {
		log.Errorf("failed to check sign! check source: %v sign: %v with error: %v", string(source), signMsg, err.Error())
		return
	}
	return err
}

func (*Sha256WithRSA) signType() Type {
	return TypeSha256WithRSA
}
