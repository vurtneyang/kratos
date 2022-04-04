package binding

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
	trans    ut.Translator
}

var _ StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}
	return nil
}

func (v *defaultValidator) RegisterValidation(key string, fn validator.Func) error {
	v.lazyinit()
	return v.validate.RegisterValidation(key, fn)
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.RegisterTagNameFunc(func(field reflect.StructField) string {
			zhVal := field.Tag.Get("zh")
			if zhVal != "" {
				return zhVal
			}
			formVal := field.Tag.Get("form")
			if formVal != "" {
				return formVal
			}
			return field.Name
		})
		zhT := zh.New()
		enT := en.New()

		uni := ut.New(enT, zhT)
		v.trans, _ = uni.GetTranslator("zh")
		_ = zhTranslations.RegisterDefaultTranslations(v.validate, v.trans)
	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func (v *defaultValidator) GetValidate() *validator.Validate {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) GetTranslator() ut.Translator {
	v.lazyinit()
	return v.trans
}
