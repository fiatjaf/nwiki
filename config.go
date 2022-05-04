package main

var config Config

type Config struct {
	DataDir    string            `json:"-"`
	Relays     map[string]Policy `json:"relays,flow"`
	PrivateKey string            `json:"privatekey,omitempty"`
}

func (c Config) RelaysList() []string {
	list := make([]string, len(c.Relays))
	i := 0
	for u, _ := range c.Relays {
		list[i] = u
		i++
	}
	return list
}

type Policy struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

func (p Policy) String() string {
	var ret string
	if p.Read {
		ret += "r"
	}
	if p.Write {
		ret += "w"
	}
	return ret
}

func (c *Config) Init() {
	if c.Relays == nil {
		c.Relays = make(map[string]Policy)
	}
}
