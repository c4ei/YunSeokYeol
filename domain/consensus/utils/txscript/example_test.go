// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript_test

import (
	"encoding/hex"
	"fmt"

	"github.com/c4ei/c4exd/domain/consensus/model/externalapi"

	"github.com/c4ei/c4exd/domain/consensus/utils/txscript"
	"github.com/c4ei/c4exd/domain/dagconfig"
	"github.com/c4ei/c4exd/util"
)

// This example demonstrates creating a script which pays to a c4ex address.
// It also prints the created script hex and uses the DisasmString function to
// display the disassembled script.
// 이 예는 c4ex 주소로 지불하는 스크립트를 생성하는 방법을 보여줍니다.
// 또한 생성된 스크립트 16진수를 인쇄하고 DisasmString 함수를 사용하여
// 디스어셈블된 스크립트를 표시합니다.
func ExamplePayToAddrScript() {
	// Parse the address to send the coins to into a util.Address
	// which is useful to ensure the accuracy of the address and determine
	// the address type. It is also required for the upcoming call to
	// PayToAddrScript.
	// 코인을 보낼 주소를 util.Address로 구문 분석합니다.
	// 이는 주소의 정확성을 보장하고 결정하는 데 유용합니다.
	// 주소 유형. 이는 향후 통화에도 필요합니다.
	// PayToAddrScript.
	addressStr := "c4ex:qqj9fg59mptxkr9j0y53j5mwurcmda5mtza9n6v9pm9uj8h0wgk6uma5pvumr"
	address, err := util.DecodeAddress(addressStr, util.Bech32PrefixC4ex)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a public key script that pays to the address.
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Script Hex: %x\n", script.Script)

	disasm, err := txscript.DisasmString(script.Version, script.Script)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Script Disassembly:", disasm)

	// Output:
	// Script Hex: 202454a285d8566b0cb2792919536ee0f1b6f69b58ba59e9850ecbc91eef722daeac
	// Script Disassembly: 2454a285d8566b0cb2792919536ee0f1b6f69b58ba59e9850ecbc91eef722dae OP_CHECKSIG
}

// This example demonstrates extracting information from a standard public key
// script.
func ExampleExtractScriptPubKeyAddress() {
	// Start with a standard pay-to-pubkey script.
	scriptHex := "2089ac24ea10bb751af4939623ccc5e550d96842b64e8fca0f63e94b4373fd555eac"
	script, err := hex.DecodeString(scriptHex)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Extract and print details from the script.
	scriptClass, address, err := txscript.ExtractScriptPubKeyAddress(
		&externalapi.ScriptPublicKey{
			Script:  script,
			Version: 0,
		}, &dagconfig.MainnetParams)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Script Class:", scriptClass)
	fmt.Println("Address:", address)

	// Output:
	// Script Class: pubkey
	// Address: c4ex:qzy6cf82zzah2xh5jwtz8nx9u4gdj6zzke8gljs0v055ksmnl424u6fv7ajrs
}
