// Copyright (c) 2018 The bcext developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txsizes_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/bcext/gcash/wire"
	. "github.com/bcext/cashwallet/wallet/internal/txsizes"
)

const (
	p2pkhScriptSize = P2PKHPkScriptSize
	p2shScriptSize  = 23
)

func makeInts(value int, n int) []int {
	v := make([]int, n)
	for i := range v {
		v[i] = value
	}
	return v
}

func TestEstimateSerializeSize(t *testing.T) {
	tests := []struct {
		InputCount           int
		OutputScriptLengths  []int
		AddChangeOutput      bool
		ExpectedSizeEstimate int
	}{
		0: {1, []int{}, false, 159},
		1: {1, []int{p2pkhScriptSize}, false, 193},
		2: {1, []int{}, true, 193},
		3: {1, []int{p2pkhScriptSize}, true, 227},
		4: {1, []int{p2shScriptSize}, false, 191},
		5: {1, []int{p2shScriptSize}, true, 225},

		6:  {2, []int{}, false, 308},
		7:  {2, []int{p2pkhScriptSize}, false, 342},
		8:  {2, []int{}, true, 342},
		9:  {2, []int{p2pkhScriptSize}, true, 376},
		10: {2, []int{p2shScriptSize}, false, 340},
		11: {2, []int{p2shScriptSize}, true, 374},

		// 0xfd is discriminant for 16-bit compact ints, compact int
		// total size increases from 1 byte to 3.
		12: {1, makeInts(p2pkhScriptSize, 0xfc), false, 8727},
		13: {1, makeInts(p2pkhScriptSize, 0xfd), false, 8727 + P2PKHOutputSize + 2},
		14: {1, makeInts(p2pkhScriptSize, 0xfc), true, 8727 + P2PKHOutputSize + 2},
		15: {0xfc, []int{}, false, 37558},
		16: {0xfd, []int{}, false, 37558 + RedeemP2PKHInputSize + 2},
	}
	for i, test := range tests {
		outputs := make([]*wire.TxOut, 0, len(test.OutputScriptLengths))
		for _, l := range test.OutputScriptLengths {
			outputs = append(outputs, &wire.TxOut{PkScript: make([]byte, l)})
		}
		actualEstimate := EstimateSerializeSize(test.InputCount, outputs, test.AddChangeOutput)
		if actualEstimate != test.ExpectedSizeEstimate {
			t.Errorf("Test %d: Got %v: Expected %v", i, actualEstimate, test.ExpectedSizeEstimate)
		}
	}
}

func TestEstimateVirtualSize(t *testing.T) {

	type estimateVSizeTest struct {
		tx       func() (*wire.MsgTx, error)
		p2pkhIns int
		change   bool
		result   int
	}

	// TODO(halseth): add tests for more combination out inputs/outputs.
	tests := []estimateVSizeTest{
		// Spending one P2PKH to one P2SH and one P2PKH outputs.
		{
			tx: func() (*wire.MsgTx, error) {
				txHex := "02000000017df1002e89430d1c08606e18c618633a3526af63855deb160538af947b869d45010000006a47304402207d4c38661b4a2142a504d1074b35d013a44d15b222776502fa1f0a795558b13402201fe366c033aee7a3ccf5ad0d20735354008d07ce48071be9a11d4e4f17d2ad7e4121037a8e4aca34a7ca97f6670fa2024811ad14bda6ef9043c45934a006862271b9c2ffffffff0260c8d93e0000000017a9140508b85a8d8dd0215d0b34285cee9b95656d04cf870ddfb727030000001976a914cebf375400a0322c5369e05c50ce71612d885c2f88ac00000000"
				b, err := hex.DecodeString(txHex)
				if err != nil {
					return nil, err
				}
				tx := &wire.MsgTx{}
				err = tx.Deserialize(bytes.NewReader(b))
				if err != nil {
					return nil, err
				}

				return tx, nil
			},
			result: 76,
		},
		{
			// Spending one P2SH to one P2SH and one P2PKH outputs.
			tx: func() (*wire.MsgTx, error) {
				txHex := "020000000171355c62b8db41206ad49b80d718bc99416f02d2a2246736cd42d41d600bddb500000000fdfd0000483045022100dc5805177d2e41522ae3a9dd5f1458f7814590bb672942475867b35d038a4a0e022045e2b63fe86183a1ed9816e2bfa565e2ff2e95b5c09a7f1ce112d655a7685f1d4147304402202775ef51607b729723b42a68d1be3b0a4fe940299958846a180fcf413743f47902202893ba63014bc7f3dce5b57de3b364bb7ec24d9d9ec00d2bedbe46520f619369414c69522102d955260daed9d68adabeaf0e0396fb7bd58d3bc22944f9dcafa0fb967c6a7f0c2103853e4bdbbeb811f2c29de7928216574d37602a9e821e15d51d3359db761a2e6f21028d2a7a90fdf8062bef18e0c34a14f2a3715168c15d6aa40532dad5b613087da953aeffffffff02884357540100000017a9143f0a5f886488451171bc00d6b98f1473660c111c87d4859600000000001976a914808e2a91b3f94a9824ad8ee46c6f983ab44383d688ac6d210800"
				b, err := hex.DecodeString(txHex)
				if err != nil {
					return nil, err
				}
				tx := &wire.MsgTx{}
				err = tx.Deserialize(bytes.NewReader(b))
				if err != nil {
					return nil, err
				}

				return tx, nil
			},
			result: 76,
		},
		{
			// Spending one P2PKH to two P2PKH outputs.
			tx: func() (*wire.MsgTx, error) {
				txHex := "0100000001a4c91c9720157a5ee582a7966471d9c70d0a860fa7757b4c42a535a12054a4c9000000006c493046022100d49c452a00e5b1213ac84d92269510a05a584a4d0949bd7d0ad4e3408ac8e80a022100bf98707ffaf1eb9dff146f7da54e68651c0a27e3653ec3882b7a95202328579c01210332d98672a4246fe917b9c724c339e757d46b1ffde3fb27fdc680b4bb29b6ad59ffffffff02a0860100000000001976a9144fb55ee0524076acd4c14e7773561e4c298c8e2788ac20688a0b000000001976a914cb7f6bb8e95a2cd06423932cfbbce73d16a18df088ac00000000"
				b, err := hex.DecodeString(txHex)
				if err != nil {
					return nil, err
				}
				tx := &wire.MsgTx{}
				err = tx.Deserialize(bytes.NewReader(b))
				if err != nil {
					return nil, err
				}

				return tx, nil
			},
			p2pkhIns: 1,
			result:   227,
		},
	}

	for _, test := range tests {
		tx, err := test.tx()
		if err != nil {
			t.Fatalf("unable to get test tx: %v", err)
		}

		est := EstimateVirtualSize(test.p2pkhIns, tx.TxOut, test.change)

		if est != test.result {
			t.Fatalf("expected estimated vsize to be %d, "+
				"instead got %d", test.result, est)
		}
	}
}
