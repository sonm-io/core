package main

type LoggingConfig struct {
	Level string `config:"level"`
}

type EthereumConfig struct {
	Endpoint    string `config:"endpoint"`
	AccountType string `config:"account_type"`
	AccountPath string `config:"account_path"`
	AccountPass string `config:"account_pass"`
	Registry    string `config:"registry"`
}
