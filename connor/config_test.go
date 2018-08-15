package connor

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/connor/antifraud"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestConfigContainerEnv_Empty(t *testing.T) {
	cfg := &Config{}
	env := cfg.containerEnv(sonm.NewBigIntFromInt(1))
	// empty config with no extra params should provide empty env map
	assert.Len(t, env, 0)
}

func TestConfigContainerEnv_KnownParams(t *testing.T) {
	cfg := &Config{
		Mining: miningConfig{
			Token:  "ETH",
			Wallet: common.HexToAddress("0xB9aF252ec7D84f0feAD1DBB893b02cfD99EF47C6"),
		},
		AntiFraud: antifraud.Config{
			PoolProcessorConfig: antifraud.ProcessorConfig{
				Format: antifraud.PoolFormatDwarf,
			},
		},
	}
	env := cfg.containerEnv(sonm.NewBigIntFromInt(1))
	// config for eth mining on dwarfPool should provide wallet and pool addr
	assert.Len(t, env, 2)
	assert.Contains(t, env, "WALLET")
	assert.Contains(t, env, "POOL")
}

func TestConfigContainerEnv_ExtraParams(t *testing.T) {
	cfg := &Config{
		Mining: miningConfig{
			Token:  "ETH",
			Wallet: common.HexToAddress("0xB9aF252ec7D84f0feAD1DBB893b02cfD99EF47C6"),
		},
		AntiFraud: antifraud.Config{
			PoolProcessorConfig: antifraud.ProcessorConfig{
				Format: antifraud.PoolFormatDwarf,
			},
		},
		Engine: engineConfig{
			ContainerEnv: map[string]string{
				"KEY1": "value1",
				"key2": "value2",
			},
		},
	}

	env := cfg.containerEnv(sonm.NewBigIntFromInt(1))
	// config should contain params for dwarfpool + two extra params
	assert.Len(t, env, 4)
	assert.Contains(t, env, "WALLET")
	assert.Contains(t, env, "POOL")
	assert.Contains(t, env, "KEY1")
	assert.Contains(t, env, "key2")
}

func TestConfigContainerEnv_OverrideParams(t *testing.T) {
	cfg := &Config{
		Mining: miningConfig{
			Token:  "ETH",
			Wallet: common.HexToAddress("0xB9aF252ec7D84f0feAD1DBB893b02cfD99EF47C6"),
		},
		AntiFraud: antifraud.Config{
			PoolProcessorConfig: antifraud.ProcessorConfig{
				Format: antifraud.PoolFormatDwarf,
			},
		},
		Engine: engineConfig{
			ContainerEnv: map[string]string{"WALLET": "foo"},
		},
	}

	env := cfg.containerEnv(sonm.NewBigIntFromInt(1))
	// wallet should be overriden with extra env-vars
	assert.Len(t, env, 2)
	assert.Contains(t, env, "WALLET")
	assert.Contains(t, env, "POOL")
	assert.Equal(t, "foo", env["WALLET"])
}
