package main

type LoggingConfig struct {
	Level string `config:"level"`
}

type EthereumConfig struct {
	Endpoint    string `config:"endpoint"`
	AccountPath string `config:"account_path"`
	AccountPass string `config:"account_pass"`
}
