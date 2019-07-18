package optimus

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestPlan(price int64) *sonm.AskPlan {
	return &sonm.AskPlan{
		Price: &sonm.Price{
			PerSecond: sonm.NewBigIntFromInt(price),
		},
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := map[string]struct {
		create []*sonm.AskPlan
		remove []*sonm.AskPlan

		expectedCreate []*sonm.AskPlan
		expectedRemove []*sonm.AskPlan
	}{
		"OneEqRemoveEmpty": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
				newTestPlan(3),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(2),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(3),
			},
			expectedRemove: []*sonm.AskPlan{},
		},
		"OneEqRemoveNotEmpty": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
				newTestPlan(3),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(3),
				newTestPlan(4),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
			},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(4),
			},
		},
		"TwoEq": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(2),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(4),
				newTestPlan(4),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
			},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(4),
				newTestPlan(4),
			},
		},
		"ThreeCompletelyEq": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{},
			expectedRemove: []*sonm.AskPlan{},
		},
		"ThreeEqExtraRemove": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(4),
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(4),
			},
		},
		"TwoEqExtraCreate": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
			},
			expectedRemove: []*sonm.AskPlan{},
		},
		"CombinedNoRemoveLeft": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(2),
				newTestPlan(2),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(2),
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
			},
			expectedRemove: []*sonm.AskPlan{},
		},
		"Combined": {
			create: []*sonm.AskPlan{
				newTestPlan(2),
				newTestPlan(3),
				newTestPlan(5),
				newTestPlan(4),
				newTestPlan(1),
				newTestPlan(4),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(4),
				newTestPlan(2),
				newTestPlan(2),
				newTestPlan(3),
				newTestPlan(3),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(4),
				newTestPlan(5),
			},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(2),
				newTestPlan(3),
			},
		},
	}

	sortPlans := func(plans []*sonm.AskPlan) {
		sort.Slice(plans, func(i, j int) bool {
			return plans[i].Price.PerSecond.Cmp(plans[j].Price.PerSecond) < 0
		})
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			create, remove := removeDuplicates(test.create, test.remove)

			sortPlans(create)
			sortPlans(remove)
			sortPlans(test.expectedCreate)
			sortPlans(test.expectedRemove)

			assert.Equal(t, test.expectedCreate, create)
			assert.Equal(t, test.expectedCreate, create)
		})
	}
}

func TestBBM(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	var err error

	devicesJSON := `{"CPU":{"device":{"modelName":"Intel(R) Celeron(R) CPU G3930 @ 2.90GHz","cores":2,"sockets":2},"benchmarks":{"0":{"code":"cpu-sysbench-multi","type":1,"image":"sonm/cpu-sysbench@sha256:8eeb78e04954c07b2f72f9311ac2f7eb194456a4af77b2c883f99f8949701924","result":2239,"splittingAlgorithm":1},"1":{"ID":1,"code":"cpu-sysbench-single","type":1,"image":"sonm/cpu-sysbench@sha256:8eeb78e04954c07b2f72f9311ac2f7eb194456a4af77b2c883f99f8949701924","result":1140},"12":{"ID":12,"code":"cpu-cryptonight","type":1,"image":"sonm/cpu-cryptonight@sha256:c0cbff06d21fcb7ab156945c9b347216bf4c7336efcb892dbb9be0e41146897c","result":52,"splittingAlgorithm":1},"2":{"ID":2,"code":"cpu-cores","type":1,"result":2}}},"GPUs":[{"device":{"ID":"0000:01:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"Memory":11720982528,"hash":"cdd6d38430c1a1fbb86e68ed097b50d7","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card0","/dev/dri/renderD128","/dev/nvidia0"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":689,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":731,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33394000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:03:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":1,"Memory":11720982528,"hash":"dbadaf5cd086181570f9b3d4d67679ee","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card1","/dev/dri/renderD129","/dev/nvidia1"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":683,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":587,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":32681000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:04:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":2,"Memory":11720982528,"hash":"0be1e7a2eb8ccafe4df5340e2140ae9e","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card2","/dev/dri/renderD130","/dev/nvidia2"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":688,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":636,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33398000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:05:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":3,"Memory":11720982528,"hash":"669aefa304dc052283796637c14e38c1","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card3","/dev/dri/renderD131","/dev/nvidia3"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":645,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33387000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:06:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":4,"Memory":11720982528,"hash":"d6fe583923fce9045a5d3ac942175e47","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card4","/dev/dri/renderD132","/dev/nvidia4"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":694,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":646,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33412000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:07:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":5,"Memory":11720982528,"hash":"a2041949015a04aafdb94b1d64238123","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card5","/dev/dri/renderD133","/dev/nvidia5"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":675,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":589,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":32678000,"splittingAlgorithm":1}}}],"RAM":{"device":{"total":33693065216,"available":33693065216,"used":1532973056},"benchmarks":{"3":{"ID":3,"code":"ram-size","type":3,"result":33693065216,"splittingAlgorithm":1}}},"network":{"in":93373123,"out":97657021,"netFlags":{"flags":3},"benchmarksIn":{"5":{"ID":5,"code":"net-download","type":5,"image":"sonm/net-bandwidth@sha256:2b1569dd1ecb2dd588e448df8762e8310d1f9c6ea539a3404fc2a934bb195fe3","result":93373123,"splittingAlgorithm":1}},"benchmarksOut":{"6":{"ID":6,"code":"net-upload","type":6,"image":"sonm/net-bandwidth@sha256:2b1569dd1ecb2dd588e448df8762e8310d1f9c6ea539a3404fc2a934bb195fe3","result":97657021,"splittingAlgorithm":1}}},"storage":{"device":{"bytesAvailable":39983321088},"benchmarks":{"4":{"ID":4,"code":"storage-size","type":4,"result":39983321088,"splittingAlgorithm":1}}}}`
	devices := &sonm.DevicesReply{}
	err = json.Unmarshal([]byte(devicesJSON), devices)
	require.NoError(t, err)

	fmt.Printf("`%v`-\n", devicesJSON)

	freeDevicesJSON := `{"CPU":{"device":{"modelName":"Intel(R) Celeron(R) CPU G3930 @ 2.90GHz","cores":2,"sockets":2},"benchmarks":{"0":{"code":"cpu-sysbench-multi","type":1,"image":"sonm/cpu-sysbench@sha256:8eeb78e04954c07b2f72f9311ac2f7eb194456a4af77b2c883f99f8949701924","result":2239,"splittingAlgorithm":1},"1":{"ID":1,"code":"cpu-sysbench-single","type":1,"image":"sonm/cpu-sysbench@sha256:8eeb78e04954c07b2f72f9311ac2f7eb194456a4af77b2c883f99f8949701924","result":1140},"12":{"ID":12,"code":"cpu-cryptonight","type":1,"image":"sonm/cpu-cryptonight@sha256:c0cbff06d21fcb7ab156945c9b347216bf4c7336efcb892dbb9be0e41146897c","result":52,"splittingAlgorithm":1},"2":{"ID":2,"code":"cpu-cores","type":1,"result":2}}},"GPUs":[{"device":{"ID":"0000:01:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"Memory":11720982528,"hash":"cdd6d38430c1a1fbb86e68ed097b50d7","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card0","/dev/dri/renderD128","/dev/nvidia0"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":689,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":731,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33394000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:03:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":1,"Memory":11720982528,"hash":"dbadaf5cd086181570f9b3d4d67679ee","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card1","/dev/dri/renderD129","/dev/nvidia1"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":683,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":587,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":32681000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:04:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":2,"Memory":11720982528,"hash":"0be1e7a2eb8ccafe4df5340e2140ae9e","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card2","/dev/dri/renderD130","/dev/nvidia2"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":688,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":636,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33398000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:05:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":3,"Memory":11720982528,"hash":"669aefa304dc052283796637c14e38c1","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card3","/dev/dri/renderD131","/dev/nvidia3"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":645,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33387000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:06:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":4,"Memory":11720982528,"hash":"d6fe583923fce9045a5d3ac942175e47","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card4","/dev/dri/renderD132","/dev/nvidia4"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":694,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":646,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":33412000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:07:00.0","vendorID":4318,"vendorName":"NVidia","deviceID":6918,"deviceName":"GeForce GTX 1080 Ti","majorNumber":226,"minorNumber":5,"Memory":11720982528,"hash":"a2041949015a04aafdb94b1d64238123","deviceFiles":["/dev/nvidiactl","/dev/nvidia-uvm","/dev/nvidia-uvm-tools","/dev/dri/card5","/dev/dri/renderD133","/dev/nvidia5"],"driverVolumes":{"nvidia-docker":"nvidia-docker_396.54:/usr/local/nvidia"}},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:7a610940ab68c43ca2c22d9cff6c9bd1aee8bc1f6c6e86f483c52b5de4652711","result":675,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:bfc1cf25e4d4c39ad126709b6df28ac9aac2ec8caa95289ff5b80905f372d562","splittingAlgorithm":1},"13":{"ID":13,"code":"gpu-nvidia","type":2,"result":1,"splittingAlgorithm":2},"14":{"ID":14,"code":"gpu-radeon","type":2,"splittingAlgorithm":2},"15":{"ID":15,"code":"gpu-cuckaroo29","type":2,"image":"sonm/gpu-cuckaroo29-bench@sha256:790dea8595c3062ed454eace8cd5fa70eff78a59653d96953d6d7cf413c66048","result":589,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":11720982528,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":32678000,"splittingAlgorithm":1}}}],"RAM":{"device":{"total":33693065216,"available":33693065216,"used":1532973056},"benchmarks":{"3":{"ID":3,"code":"ram-size","type":3,"result":33693065216,"splittingAlgorithm":1}}},"network":{"in":93373123,"out":97657021,"netFlags":{"flags":3},"benchmarksIn":{"5":{"ID":5,"code":"net-download","type":5,"image":"sonm/net-bandwidth@sha256:2b1569dd1ecb2dd588e448df8762e8310d1f9c6ea539a3404fc2a934bb195fe3","result":93373123,"splittingAlgorithm":1}},"benchmarksOut":{"6":{"ID":6,"code":"net-upload","type":6,"image":"sonm/net-bandwidth@sha256:2b1569dd1ecb2dd588e448df8762e8310d1f9c6ea539a3404fc2a934bb195fe3","result":97657021,"splittingAlgorithm":1}}},"storage":{"device":{"bytesAvailable":39983321088},"benchmarks":{"4":{"ID":4,"code":"storage-size","type":4,"result":39983321088,"splittingAlgorithm":1}}}}`
	freeDevices := &sonm.DevicesReply{}
	err = json.Unmarshal([]byte(freeDevicesJSON), freeDevices)
	require.NoError(t, err)

	fmt.Printf("`%v`-\n", freeDevicesJSON)

	deviceManager, err := newDeviceManager(devices, freeDevices, newMappingMock(mockController))
	require.NoError(t, err)
	require.NotNil(t, deviceManager)

	model := &BranchBoundModel{
		Log: zap.NewExample().Sugar(),
	}

	virtualFreeOrdersJSON := `[{"order":{"id":"2314297","dealID":"288555","orderType":2,"orderStatus":1,"authorID":"0x0Fc6E1838a3FaC5C415fF087E4f922beeA8b0c2E","counterpartyID":"0x0000000000000000000000000000000000000000","price":"173611111111110","netflags":{"flags":2},"identityLevel":1,"blacklist":"0x0000000000000000000000000000000000000000","tag":"b3B0aW11cy92MC40LjI4LTIxMjVmNjFkAAAAAAAAAAA=","benchmarks":{"values":[1501,1140,2,30000000000,12000000000,10000000,9000000,6,11720982528,198950000,3429,0,35,1,0,3834]},"frozenSum":"0"},"CreatedTS":"2019-07-17T07:28:33.594943981Z"}]`
	virtualFreeOrders := make([]*MarketOrder, 0)
	err = json.Unmarshal([]byte(virtualFreeOrdersJSON), &virtualFreeOrders)
	require.NoError(t, err)

	fmt.Printf("[%d] `%v`-\n", len(virtualFreeOrders), virtualFreeOrdersJSON)
	fmt.Printf("O: %v\n", virtualFreeOrders[0].Order.Benchmarks.Values)
	fmt.Printf("D: %v\n", deviceManager.freeBenchmarks)

	knapsack := NewKnapsack(deviceManager)
	err = model.Optimize(context.Background(), knapsack, virtualFreeOrders)
	require.NoError(t, err)

	fmt.Printf("PPS=`%v`\n", knapsack.PPSf64())
}
