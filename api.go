package main

type QueryTiming struct {
	api_call QueryApi
	name     string
}

func (l *QueryTiming) GetName() string {
	return "query/" + l.name
}

func (l *QueryTiming) GetUnits() string {
	return "ms"
}

func (l *QueryTiming) GetValue() (float64, error) {
	duration, err := l.api_call.RequestTest()
	if err != nil {
		return 0, err
	}

	return float64(duration), nil
}
