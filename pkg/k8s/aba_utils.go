package k8s

type StringUtils struct {
	Value     string
	Defaulted string
}

func NewStringUtils(value string) *StringUtils {
	return &StringUtils{Value: value}
}

func NewDefaultStringUtils(value, defaulted string) *StringUtils {
	return &StringUtils{Value: value, Defaulted: defaulted}
}

func (s *StringUtils) IsEmpty() bool {
	return s.Value == ""
}

func (s *StringUtils) ValueOrDefault() string {
	if s.IsEmpty() {
		return s.Defaulted
	}
	return s.Value
}

func LambdaCompare(f func() bool, a, b interface{}) interface{} {
	if f() {
		return a
	} else {
		return b
	}
}
