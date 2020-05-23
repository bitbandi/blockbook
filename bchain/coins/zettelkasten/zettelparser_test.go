// +build unittest

package zettelkasten

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
)

var (
	testTx1, testTx2 bchain.Tx

	testTxPacked1 = "0001e2408bb2d6f66801000000010000000000000000000000000000000000000000000000000000000000000000ffffffff020102ffffffff017ad8c33900000000161449d9e77c83f75478027ee98e8b8409e3e0a02466ac00000000"
	testTxPacked2 = "0003e8598bbaa4b0740100000001bf8b0442e040862ccd1b519700887760bc2333fa280a158de736acdcff0b8ddc0000000043421c11e25c209fcaf5dd690619939d3ca81a681dc3d4e08cbe5c5e3c6ab0f8c023a20bedadfc298bd44ca3c9f12566b38cac9c5b9da5622fc8ae8b945c63ce2a7bc901ffffffff020084d717000000001614ac39d30c739f8176fba2c6f4c00ba56f5d34bf5bac4c679ea616000000161405dcd5cf03a945d52f79faaa093759a4685eb690ac00000000"
)

func init() {
	testTx1 = bchain.Tx{
		Hex:       "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff020102ffffffff017ad8c33900000000161449d9e77c83f75478027ee98e8b8409e3e0a02466ac00000000",
		Blocktime: 1529535924,
		Txid:      "b8f10b3e14704c528753ef65622b17ad4925eb464b0f8f3e3dad3e0857f4521e",
		Version:   1,
		LockTime:  0,
		Vin: []bchain.Vin{
			{
				Coinbase: "0102",
				Sequence: 4294967295,
			},
		},
		Vout: []bchain.Vout{
			{
				ValueSat: *big.NewInt(969136250),
				N:        0,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "1449d9e77c83f75478027ee98e8b8409e3e0a02466ac",
					Addresses: []string{
						"Zi7M83zBd9Z9busnBxr6zpZu3BokGjY5od",
					},
				},
			},
		},
	}

	testTx2 = bchain.Tx{
		Hex:       "0100000001bf8b0442e040862ccd1b519700887760bc2333fa280a158de736acdcff0b8ddc0000000043421c11e25c209fcaf5dd690619939d3ca81a681dc3d4e08cbe5c5e3c6ab0f8c023a20bedadfc298bd44ca3c9f12566b38cac9c5b9da5622fc8ae8b945c63ce2a7bc901ffffffff020084d717000000001614ac39d30c739f8176fba2c6f4c00ba56f5d34bf5bac4c679ea616000000161405dcd5cf03a945d52f79faaa093759a4685eb690ac00000000",
		Blocktime: 1537510458,
		Txid:      "870e38a8a0b7708fe9269ab191fdb0fc3f2bc2c7b21230d9bb18bba867c6937b",
		Version:   1,
		LockTime:  0,
		Vin: []bchain.Vin{
			{
				ScriptSig: bchain.ScriptSig{
					Hex: "421c11e25c209fcaf5dd690619939d3ca81a681dc3d4e08cbe5c5e3c6ab0f8c023a20bedadfc298bd44ca3c9f12566b38cac9c5b9da5622fc8ae8b945c63ce2a7bc901",
				},
				Txid:     "dc8d0bffdcac36e78d150a28fa3323bc6077880097511bcd2c8640e042048bbf",
				Vout:     0,
				Sequence: 4294967295,
			},
		},
		Vout: []bchain.Vout{
			{
				ValueSat: *big.NewInt(400000000),
				N:        0,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "14ac39d30c739f8176fba2c6f4c00ba56f5d34bf5bac",
					Addresses: []string{
						"Zs5WFW3aysAT1adwftRiWNQcvzEnDrcTSX",
					},
				},
			},
			{
				ValueSat: *big.NewInt(97284679500),
				N:        1,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "1405dcd5cf03a945d52f79faaa093759a4685eb690ac",
					Addresses: []string{
						"ZburgeroFZKfaDbuhbTjT3VgCnHjZskYTD",
					},
				},
			},
		},
	}
}

func TestMain(m *testing.M) {
	c := m.Run()
	chaincfg.ResetParams()
	os.Exit(c)
}

func TestGetAddressesFromAddrDesc(t *testing.T) {
	type args struct {
		script string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want2   bool
		wantErr bool
	}{
		{
			name:    "Genesis",
			args:    args{script: "149f7db318cb93848108d3d6f91ce4517db50d5dd2ac"},
			want:    []string{"ZqvAn4MEHxqHgk8WHHTm6jVxhRAX5Ag6wY"},
			want2:   true,
			wantErr: false,
		},
		{
			name:    "First reward",
			args:    args{script: "148ba6fcede976e803edc7b3625e258710ff023098ac"},
			want:    []string{"Zp7GizRRtx14WAA6HxypuqUxhihdDAewGa"},
			want2:   true,
			wantErr: false,
		},
	}

	parser := NewZettelkastenParser(GetChainParams("main"), &btc.Configuration{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := hex.DecodeString(tt.args.script)
			got, got2, err := parser.GetAddressesFromAddrDesc(b)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressesFromAddrDesc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetAddrDesc(t *testing.T) {
	type args struct {
		tx     bchain.Tx
		parser *ZettelkastenParser
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "zettelkasten-1",
			args: args{
				tx:     testTx1,
				parser: NewZettelkastenParser(GetChainParams("main"), &btc.Configuration{}),
			},
		},
		{
			name: "zettelkasten-2",
			args: args{
				tx:     testTx2,
				parser: NewZettelkastenParser(GetChainParams("test"), &btc.Configuration{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for n, vout := range tt.args.tx.Vout {
				got1, err := tt.args.parser.GetAddrDescFromVout(&vout)
				if err != nil {
					t.Errorf("getAddrDescFromVout() error = %v, vout = %d", err, n)
					return
				}
				got2, err := tt.args.parser.GetAddrDescFromAddress(vout.ScriptPubKey.Addresses[0])
				if err != nil {
					t.Errorf("getAddrDescFromAddress() error = %v, vout = %d", err, n)
					return
				}
				if !bytes.Equal(got1, got2) {
					t.Errorf("Address descriptors mismatch: got1 = %v, got2 = %v", got1, got2)
				}
			}
		})
	}
}

func TestPackTx(t *testing.T) {
	type args struct {
		tx	bchain.Tx
		height    uint32
		blockTime int64
		parser    *ZettelkastenParser
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "zettelkasten-1",
			args: args{
				tx:        testTx1,
				height:    123456,
				blockTime: 1529535924,
				parser:    NewZettelkastenParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    testTxPacked1,
			wantErr: false,
		},
		{
			name: "zettelkasten-2",
			args: args{
				tx:        testTx2,
				height:    256089,
				blockTime: 1537510458,
				parser:    NewZettelkastenParser(GetChainParams("test"), &btc.Configuration{}),
			},
			want:    testTxPacked2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.parser.PackTx(&tt.args.tx, tt.args.height, tt.args.blockTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("packTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			h := hex.EncodeToString(got)
			if !reflect.DeepEqual(h, tt.want) {
				t.Errorf("packTx() = %v, want %v", h, tt.want)
			}
		})
	}
}

func TestUnpackTx(t *testing.T) {
	type args struct {
		packedTx string
		parser   *ZettelkastenParser
	}
	tests := []struct {
		name    string
		args    args
		want    *bchain.Tx
		want1   uint32
		wantErr bool
	}{
		{
			name: "zettelkasten-1",
			args: args{
				packedTx: testTxPacked1,
				parser:   NewZettelkastenParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    &testTx1,
			want1:   123456,
			wantErr: false,
		},
		{
			name: "zettelkasten-2",
			args: args{
				packedTx: testTxPacked2,
				parser:   NewZettelkastenParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    &testTx2,
			want1:   256089,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := hex.DecodeString(tt.args.packedTx)
			got, got1, err := tt.args.parser.UnpackTx(b)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpackTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unpackTx() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("unpackTx() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
