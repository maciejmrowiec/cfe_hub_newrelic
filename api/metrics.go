package api

type QueryTiming struct {
	apiCall QueryApi
	name    string
}

func NewQueryTiming(apiCall QueryApi, name string) *QueryTiming {
	return &QueryTiming{
		apiCall: apiCall,
		name:    name,
	}
}

func (l *QueryTiming) GetName() string {
	return "query/" + l.name
}

func (l *QueryTiming) GetUnits() string {
	return "ms"
}

func (l *QueryTiming) GetValue() (float64, error) {
	duration, err := l.apiCall.RequestTest()
	if err != nil {
		return 0, err
	}

	return float64(duration), nil
}
