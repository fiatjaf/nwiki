package main

var config Config

type Config struct {
	DataDir    string   `json:"-"`
	Relays     []string `json:"relays,flow"`
	PrivateKey string   `json:"privatekey,omitempty"`
}
