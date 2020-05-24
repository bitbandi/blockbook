package zettelkasten

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"

	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil"
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/martinboehm/btcutil/txscript"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
	"github.com/trezor/blockbook/bchain/coins/utils"
)

// magic numbers
const (
	MainnetMagic wire.BitcoinNet = 0xc2e3cbfa
	TestnetMagic wire.BitcoinNet = 0x0709110b
)

// chain parameters
var (
	MainNetParams chaincfg.Params
	TestNetParams chaincfg.Params
)

func init() {
	MainNetParams = chaincfg.MainNetParams
	MainNetParams.Net = MainnetMagic

	// Address encoding magics
	MainNetParams.PubKeyHashAddrID = []byte{81}
	MainNetParams.ScriptHashAddrID = []byte{5}

	TestNetParams = chaincfg.TestNet3Params
	TestNetParams.Net = TestnetMagic

	// Address encoding magics
	TestNetParams.PubKeyHashAddrID = []byte{111}
	TestNetParams.ScriptHashAddrID = []byte{196}
}

// ZettelkastenParser handle
type ZettelkastenParser struct {
	*btc.BitcoinParser
	baseparser *bchain.BaseParser
	BitcoinOutputScriptToAddressesFunc btc.OutputScriptToAddressesFunc
}

// NewZettelkastenParser returns new ZettelkastenParser instance
func NewZettelkastenParser(params *chaincfg.Params, c *btc.Configuration) *ZettelkastenParser {
	p := &ZettelkastenParser{
		BitcoinParser: btc.NewBitcoinParser(params, c),
		baseparser:    &bchain.BaseParser{},
	}
	p.BitcoinOutputScriptToAddressesFunc = p.OutputScriptToAddressesFunc
	p.OutputScriptToAddressesFunc = p.outputScriptToAddresses
	return p
}

// GetChainParams contains network parameters for the main Zettelkasten network,
// the regression test Zettelkasten network, the test Zettelkasten network and
// the simulation test Zettelkasten network, in this order
func GetChainParams(chain string) *chaincfg.Params {
	if !chaincfg.IsRegistered(&MainNetParams) {
		err := chaincfg.Register(&MainNetParams)
		if err == nil {
			err = chaincfg.Register(&TestNetParams)
		}
		if err != nil {
			panic(err)
		}
	}
	switch chain {
	case "test":
		return &TestNetParams
	default:
		return &MainNetParams
	}
}
/*
// PackTx packs transaction to byte array using protobuf
func (p *ZettelkastenParser) PackTx(tx *bchain.Tx, height uint32, blockTime int64) ([]byte, error) {
	return p.baseparser.PackTx(tx, height, blockTime)
}

// UnpackTx unpacks transaction from protobuf byte array
func (p *ZettelkastenParser) UnpackTx(buf []byte) (*bchain.Tx, uint32, error) {
	return p.baseparser.UnpackTx(buf)
}
*/

// ParseBlock parses raw block to our Block struct
// it has special handling for minersignature blocks that cannot be parsed by standard btc wire parser
func (p *ZettelkastenParser) ParseBlock(b []byte) (*bchain.Block, error) {
	r := bytes.NewReader(b)
	w := wire.MsgBlock{}
	h := wire.BlockHeader{}

	err := binary.Read(r, binary.LittleEndian, &h.Version)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, h.PrevBlock[:])
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, h.MerkleRoot[:])
	if err != nil {
		return nil, err
	}
	var t uint64
	err = binary.Read(r, binary.LittleEndian, &t)
	if err != nil {
		return nil, err
	}
	h.Timestamp = time.Unix(int64(t), 0)
	err = binary.Read(r, binary.LittleEndian, &h.Bits)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &h.Nonce)
	if err != nil {
		return nil, err
	}
	// hashWholeBlock
	_, err = r.Seek(32, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	// MinerSignature
	_, err = r.Seek(65, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	err = utils.DecodeTransactions(r, 0, wire.WitnessEncoding, &w)
	if err != nil {
		return nil, err
	}

	txs := make([]bchain.Tx, len(w.Transactions))
	for ti, t := range w.Transactions {
		txs[ti] = p.TxFromMsgTx(t, false)
	}

	return &bchain.Block{
		BlockHeader: bchain.BlockHeader{
			Size: len(b),
			Time: h.Timestamp.Unix(),
		},
		Txs: txs,
	}, nil
}

// GetAddrDescFromAddress returns internal address representation of given address
func (p *ZettelkastenParser) GetAddrDescFromAddress(address string) (bchain.AddressDescriptor, error) {
	return p.addressToOutputScript(address)
}

// outputScriptToAddresses converts ScriptPubKey to bitcoin addresses
func (p *ZettelkastenParser) outputScriptToAddresses(script []byte) ([]string, bool, error) {
	if len(script) == 22 {
		rv := make([]string, 1)
		address, err := btcutil.NewAddressPubKeyHash(script[1:len(script)-1], p.Params)
		if err != nil {
			return nil, false, err
		}
		rv[0] = address.EncodeAddress()
		return rv, true, nil
	}
	// TODO: multisign address
	return p.BitcoinOutputScriptToAddressesFunc(script)
}

// addressToOutputScript converts bitcoin address to ScriptPubKey
func (p *ZettelkastenParser) addressToOutputScript(address string) ([]byte, error) {
	da, err := btcutil.DecodeAddress(address, p.Params)
	if err != nil {
		return nil, err
	}
	script, err := txscript.NewScriptBuilder().AddData(da.ScriptAddress()).AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		return nil, err
	}
	return script, nil
}
