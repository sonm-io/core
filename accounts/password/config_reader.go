package password

type EthConfig struct {
	Passphrases map[string]string `required:"false" default:"" yaml:"pass_phrases"`
	Keystore    string            `required:"false" default:"" yaml:"key_store"`
}
