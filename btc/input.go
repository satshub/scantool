package btc

import (
	"fmt"
	"encoding/hex"
	"strconv"
	"html/template"

	"btctx/app"
)

type Input struct {
	previousOutputTxId [32] byte
	previousOutputIndex uint32
	coinbase bool
	spendType string
	inputScript Script
	redeemScript Script
	segwit Segwit
	sequence uint32
}

func NewInput (coinbase bool,
				previousOutputTxId [32] byte, 
				previousOutputIndex uint32, 
spendType string, 
inputScript Script, 
redeemScript Script, 
segwit Segwit, 
				sequence uint32) Input {

	i := Input {
		coinbase: coinbase,
		previousOutputTxId: previousOutputTxId,
		previousOutputIndex: previousOutputIndex,
		spendType: spendType,
		inputScript: inputScript,
		redeemScript: redeemScript,
		segwit: segwit,
		sequence: sequence }

	i.setFieldTypes ()

	return i
}

func (i *Input) setFieldTypes () {

	// P2SH-wrapped types
	if i.spendType == "P2SH-P2WPKH" || i.spendType == "P2SH-P2WSH" {

		// input script
		if i.redeemScript.IsNil () { fmt.Println (i.spendType + " input with no input script.") }
		parsedFieldCount := i.inputScript.GetParsedFieldCount ()
		if parsedFieldCount != 1 { fmt.Println (i.spendType + " input script has wrong field count. Found ", parsedFieldCount, ", expected 1.") }

		inputScriptFieldTypes := [...] string { "Serialized Redeem Script" }
		i.inputScript.SetFieldTypes (inputScriptFieldTypes [:])

		// redeem script should always exist for these types
		if i.redeemScript.IsNil () { fmt.Println (i.spendType + " input with no redeem script.") }
		parsedFieldCount = i.redeemScript.GetParsedFieldCount ()
		if parsedFieldCount != 2 { fmt.Println (i.spendType + " redeem script has wrong field count. Found ", parsedFieldCount, ", expected 2.") }

		redeemScriptFieldTypes := make ([] string, 2)
		redeemScriptFieldTypes [0] = "OP_0"
		if i.spendType == "P2SH-P2WPKH" { redeemScriptFieldTypes [1] = "20-Byte Witness Program" } else 
		if i.spendType == "P2SH-P2WSH" { redeemScriptFieldTypes [1] = "32-Byte Witness Program" }

		i.redeemScript.SetFieldTypes (redeemScriptFieldTypes [:])

	// witness types
	} else if i.spendType == "P2WPKH" || i.spendType == "P2WSH" || i.spendType == "Taproot Key Path" || i.spendType == "Taproot Script Path" {
		if !i.inputScript.IsEmpty () { fmt.Println (i.spendType + " input has non-empty input script.") }

		switch i.spendType {
			case "P2WPKH": break
			case "P2WSH": break
			case "Taproot Key Path": break
			case "Taproot Script Path": break
		}

	// coinbase, legacy and non-standard types
	} else {
		if !i.segwit.IsEmpty () { fmt.Println (i.spendType + " has non-empty segwit.") }

		// input script
		inputScriptFields := i.inputScript.GetFields ()
		inputScriptStackItemStr := i.inputScript.GetRawFieldTypes ()
		inputScriptFieldCount := len (inputScriptFields)
		inputScriptFieldTypes := make ([] string, inputScriptFieldCount)
		for f, field := range inputScriptFields {
			if inputScriptStackItemStr [f] == 'o' {
				inputScriptFieldTypes [f] = field
				continue
			}

			inputScriptFieldTypes [f] = GetStackItemType (field, false, false)
		}

		if !i.redeemScript.IsNil () {
			// it would have identified the redeem script as a data field, so we modify that here
			inputScriptFieldTypes [inputScriptFieldCount - 1] = "Serialized Redeem Script"
		}
		i.inputScript.SetFieldTypes (inputScriptFieldTypes)

		// redeem script
		redeemScriptFields := i.redeemScript.GetFields ()
		redeemScriptFieldCount := len (redeemScriptFields)
		redeemScriptStackItemStr := i.redeemScript.GetRawFieldTypes ()
		redeemScriptFieldTypes := make ([] string, redeemScriptFieldCount)
		for f, field := range redeemScriptFields {
			if redeemScriptStackItemStr [f] == 'o' {
				redeemScriptFieldTypes [f] = field
				continue
			}

			redeemScriptFieldTypes [f] = GetStackItemType (field, false, false)
		}
		i.redeemScript.SetFieldTypes (redeemScriptFieldTypes)
	}
}

func (i *Input) GetInputScript () Script {
	return i.inputScript
}

func (i *Input) HasRedeemScript () bool {
	return !i.redeemScript.IsNil ()
}

func (i *Input) GetRedeemScript () Script {
	return i.redeemScript
}

func (i *Input) HasSegwitFields () bool {
	return !i.segwit.IsNil () && !i.segwit.IsEmpty ()
}

func (i *Input) GetSegwit () Segwit {
	return i.segwit
}

func (i *Input) IsCoinbase () bool {
	return i.coinbase
}

func (i *Input) GetPreviousOutputTxId () [32] byte {
	return i.previousOutputTxId
}

func (i *Input) GetPreviousOutputIndex () uint32 {
	return i.previousOutputIndex
}

func (i *Input) GetSpendType () string {
	return i.spendType
}

func (i *Input) GetSequence () uint32 {
	return i.sequence
}

type InputHtmlData struct {
	WidthCh uint16
	InputIndex uint32
	IsCoinbase bool
	SpendType string
	ValueIn template.HTML
	TxBaseUrl string
	PreviousOutputTxId string
	PreviousOutputIndex uint32
	Sequence uint32
	InputScript ScriptHtmlData
	RedeemScript ScriptHtmlData
	Bip141 bool
	Segwit SegwitHtmlData
}

func (i *Input) GetHtmlData (inputIndex uint32, satoshis uint64, bip141 bool, widthCh uint16) InputHtmlData {

	htmlData := InputHtmlData { WidthCh: widthCh, InputIndex: inputIndex, SpendType: i.spendType, Sequence: i.sequence, Bip141: bip141 }
	htmlId := "input-script-" + strconv.FormatUint (uint64 (inputIndex), 10)

	if i.IsCoinbase () {
		htmlData.IsCoinbase = true
		htmlData.ValueIn = template.HTML (GetValueHtml (satoshis))
		htmlData.InputScript = i.inputScript.GetHtmlData ("Coinbase Script", htmlId, widthCh - 6, "text")
	} else {
		settings := app.GetSettings ()
		htmlData.TxBaseUrl = "http://" + settings.Website.GetFullUrl () + "/tx"
		htmlData.PreviousOutputTxId = hex.EncodeToString (i.previousOutputTxId [:])
		htmlData.PreviousOutputIndex = i.previousOutputIndex
		htmlData.InputScript = i.inputScript.GetHtmlData ("Input Script", htmlId, widthCh - 6, "hex")
	}

	// redeem script and segwit
	htmlData.RedeemScript = i.redeemScript.GetHtmlData ("Redeem Script", "redeem-script-" + strconv.FormatUint (uint64 (inputIndex), 10), widthCh - 6, "hex")
	htmlData.Segwit = i.segwit.GetHtmlData (inputIndex, widthCh - 6)

	return htmlData
}

/*
// This function can be used to read a raw transaction as a byte array.
// This method has been abandoned because it does not include bitcoin addresses.
// However, it is still included here, commented out, in case it becomes more
// convenient to read transactions this way if/when other bitcoin node types are supported.
func NewInput (raw_bytes [] byte) (Input, int) {

	value_reader := ValueReader {}

	pos := 0

	prev_out_hash := value_reader.ReverseBytes (raw_bytes [pos : pos + 32])
	pos += 32

	prev_out_index := value_reader.ReadNumeric (raw_bytes [pos : pos + 4])
	pos += 4

	coinbase := true
	if hex.EncodeToString (prev_out_hash) != "0000000000000000000000000000000000000000000000000000000000000000" {
		coinbase = false
	}

	script_len, byte_count := value_reader.ReadVarInt (raw_bytes [pos:])
	pos += byte_count

	script, byte_count := NewScript (raw_bytes [pos : pos + int (script_len)])
	pos += byte_count

	sequence := value_reader.ReadNumeric (raw_bytes [pos : pos + 4])
	pos += 4

	return Input {
		coinbase: coinbase,
		prev_out_hash: [32] byte (prev_out_hash),
		prev_out_index: uint32 (prev_out_index),
		tx_type: "",
		script: script,
		has_redeem_script: false,
		has_segwit_fields: false,
		sequence: uint32 (sequence) }, pos
}

// attempt to parse the serialized script(s) without knowing the output type
// SetSegwit must be called first
func (i *Input) ParseSerializedScripts () {

	// inputs with a previous p2sh output type have redeem scripts
	// non-segwit p2sh inputs have no segwit
	if i.script.IsP2shP2wshInput () || i.script.IsP2shP2wpkhInput () || !i.has_segwit_fields {
		redeem_script, _ := NewScript (i.script.GetSerializedScript ())
		if !redeem_script.parse_error {
			i.tx_type = "P2SH"
			if i.script.IsP2shP2wshInput () {
				i.tx_type += "-P2WSH"
			} else if i.script.IsP2shP2wpkhInput () {
				i.tx_type += "-P2WSH"
			}
			i.has_redeem_script = true
			i.redeem_script = redeem_script
		}
	}

	// p2sh-p2wsh and p2wsh inputs have a witness script
	// Taproot Script Path inputs have a tap script
	// Taproot and p2wsh inputs have an empty input script
	if i.script.IsP2shP2wshInput () || i.script.IsEmpty () {
		if i.segwit.ParseSerializedScript () {
			if i.script.IsP2shP2wshInput () {
				i.tx_type += "P2SH-P2WSH"
			} else {
				i.tx_type += "Taproot Script Path"
			}
		}
	}
}

*/
