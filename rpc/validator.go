package rpc

import (
	enLocale "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "gopkg.in/go-playground/validator.v9"
	enTrans "gopkg.in/go-playground/validator.v9/translations/en"
)

type requestValidator struct {
	validator  *validator.Validate
	translator *ut.Translator
}

func newRequestValidator() (*requestValidator, error) {
	lt := enLocale.New()
	en := &lt

	uniTranslator := ut.New(*en, *en)
	uniEn, _ := uniTranslator.GetTranslator("en")
	translator := &uniEn

	validate := validator.New()
	err := enTrans.RegisterDefaultTranslations(validate, *translator)
	if err != nil {
		return nil, err
	}

	return &requestValidator{
		validator:  validate,
		translator: translator,
	}, nil

}
