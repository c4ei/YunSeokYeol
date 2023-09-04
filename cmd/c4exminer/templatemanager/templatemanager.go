package templatemanager

import (
	"fmt"
	"sync"

	"github.com/c4ei/c4exd/app/appmessage"
	"github.com/c4ei/c4exd/domain/consensus/model/externalapi"
	"github.com/c4ei/c4exd/domain/consensus/utils/pow"
)

var currentTemplate *externalapi.DomainBlock
var currentState *pow.State
var isSynced bool
var lock = &sync.Mutex{}

// Get returns the template to work on // Get은 작업할 템플릿을 반환합니다.
func Get() (*externalapi.DomainBlock, *pow.State, bool) {
	lock.Lock()
	defer lock.Unlock()
	// Shallow copy the block so when the user replaces the header it won't affect the template here.
	// 사용자가 헤더를 교체할 때 여기의 템플릿에 영향을 미치지 않도록 블록을 얕은 복사합니다.
	if currentTemplate == nil {
		fmt.Printf("#######################\ntemplatemanager.go 25 line Get() currentTemplate:%+v\n#######################\n", currentTemplate)
		return nil, nil, false
	}
	block := *currentTemplate
	fmt.Printf("templatemanager.go 29 line Get() block:%+v\n", block) // block:{Header:0xc0005fec00 Transactions:[0xc0006166c0]}
	state := *currentState
	return &block, &state, isSynced
}

// Set sets the current template to work on // Set은 작업할 현재 템플릿을 설정합니다.
func Set(template *appmessage.GetBlockTemplateResponseMessage) error {
	block, err := appmessage.RPCBlockToDomainBlock(template.Block)
	if err != nil {
		return err
	}
	lock.Lock()
	defer lock.Unlock()
	// fmt.Printf("40 templatemanager.go Set() block:%+v\n", block)
	currentTemplate = block
	currentState = pow.NewState(block.Header.ToMutable())
	isSynced = template.IsSynced
	return nil
}
