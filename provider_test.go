package zconfig

type TestProvider struct {
	values map[string]string
}

func (p TestProvider) Retrieve(key string) (raw string, found bool, err error) {
	raw, found = p.values[key]
	return raw, found, nil
}

func (TestProvider) Priority() int {
	return 1
}

func (TestProvider) Name() string {
	return "test"
}
