package zconfig

type TestProvider struct {
	name   string
	values map[string]string
}

func (p TestProvider) Retrieve(key string) (raw interface{}, found bool, err error) {
	raw, found = p.values[key]
	return raw, found, nil
}

func (TestProvider) Priority() int {
	return 1
}

func (p TestProvider) Name() string {
	return p.name
}
