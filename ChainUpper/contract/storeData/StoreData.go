//go:build ignore

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package StoreData

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/FISCO-BCOS/go-sdk/v3/abi"
	"github.com/FISCO-BCOS/go-sdk/v3/abi/bind"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
)

// StoreDataInvalidEntry is an auto generated low-level Go binding around an user-defined struct.
type StoreDataInvalidEntry struct {
	Hash   [32]byte
	Reason uint8
}

// StoreDataABI is the input ABI used to generate the binding from.
const StoreDataABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32[]\",\"name\":\"justified\",\"type\":\"bytes32[]\"},{\"indexed\":false,\"internalType\":\"bytes32[]\",\"name\":\"commits\",\"type\":\"bytes32[]\"},{\"indexed\":false,\"internalType\":\"bytes32[]\",\"name\":\"invalidHashes\",\"type\":\"bytes32[]\"},{\"indexed\":false,\"internalType\":\"uint8[]\",\"name\":\"invalidReasons\",\"type\":\"uint8[]\"}],\"name\":\"EpochRecordStored\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sign\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"size\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"model\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isReliable\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"}],\"name\":\"InitTaskStored\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sign\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"slot\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"process\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"id\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"epoch\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"TaskProcessStored\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"getEpochRecordFull\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"justified\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"commits\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"reason\",\"type\":\"uint8\"}],\"internalType\":\"structStoreData.InvalidEntry[]\",\"name\":\"invalids\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sign\",\"type\":\"bytes32\"}],\"name\":\"getInitTask\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint32\",\"name\":\"size\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"model\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"isReliable\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"getTaskProcess\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"sign\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"slot\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"process\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"id\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"epoch\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"justified\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"commits\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"invalidHashes\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint8[]\",\"name\":\"invalidReasons\",\"type\":\"uint8[]\"}],\"name\":\"storeEpochRecord\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sign\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint32\",\"name\":\"size\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"model\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"isReliable\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"}],\"name\":\"storeInitTask\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sign\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"slot\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"process\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"id\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"epoch\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"storeTaskProcess\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// StoreDataBin is the compiled bytecode used for deploying new contracts.
var StoreDataBin = "0x608060405234801561001057600080fd5b50612143806100206000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c80633f566321146100675780636ba3c2c714610083578063cb68c5d7146100ba578063cc307441146100ec578063e391cfaf14610108578063ec2722d01461013c575b600080fd5b610081600480360381019061007c9190610f8f565b610158565b005b61009d6004803603810190610098919061109b565b610361565b6040516100b198979695949392919061128b565b60405180910390f35b6100d460048036038101906100cf919061134d565b610565565b6040516100e393929190611532565b60405180910390f35b6101066004803603810190610101919061162a565b6106bf565b005b610122600480360381019061011d919061109b565b61087c565b604051610133959493929190611796565b60405180910390f35b61015660048036038101906101519190611879565b6109fb565b005b6040518061012001604052808c81526020018a63ffffffff1681526020018963ffffffff1681526020018863ffffffff1681526020018763ffffffff1681526020018b815260200186815260200185858080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505081526020018383906101fd9190611b2b565b815250600160008c81526020019081526020016000206000820151816000015560208201518160010160006101000a81548163ffffffff021916908363ffffffff16021790555060408201518160010160046101000a81548163ffffffff021916908363ffffffff16021790555060608201518160010160086101000a81548163ffffffff021916908363ffffffff160217905550608082015181600101600c6101000a81548163ffffffff021916908363ffffffff16021790555060a0820151816002015560c0820151816003015560e08201518160040190805190602001906102e9929190610ba0565b50610100820151816005019080519060200190610307929190610c26565b50905050898b7f7c99774bedb5b2f1536984fdee6a8adeaca958ef5866ba04a522c829bdb7b12f8b8b8b8b8b8b8b8b8b60405161034c99989796959493929190611caf565b60405180910390a35050505050505050505050565b6000806000806000806060806000600160008b8152602001908152602001600020905080600001548160010160009054906101000a900463ffffffff168260010160049054906101000a900463ffffffff168360010160089054906101000a900463ffffffff1684600101600c9054906101000a900463ffffffff16856003015486600401876005018180546103f690611d5f565b80601f016020809104026020016040519081016040528092919081815260200182805461042290611d5f565b801561046f5780601f106104445761010080835404028352916020019161046f565b820191906000526020600020905b81548152906001019060200180831161045257829003601f168201915b5050505050915080805480602002602001604051908101604052809291908181526020016000905b828210156105435783829060005260206000200180546104b690611d5f565b80601f01602080910402602001604051908101604052809291908181526020018280546104e290611d5f565b801561052f5780601f106105045761010080835404028352916020019161052f565b820191906000526020600020905b81548152906001019060200180831161051257829003601f168201915b505050505081526020019060010190610497565b5050505090509850985098509850985098509850985050919395975091939597565b60608060606000600260008681526020019081526020016000209050806000018160010182600201828054806020026020016040519081016040528092919081815260200182805480156105d857602002820191906000526020600020905b8154815260200190600101908083116105c4575b505050505092508180548060200260200160405190810160405280929190818152602001828054801561062a57602002820191906000526020600020905b815481526020019060010190808311610616575b5050505050915080805480602002602001604051908101604052809291908181526020016000905b828210156106ab5783829060005260206000209060020201604051806040016040529081600082015481526020016001820160009054906101000a900460ff1660ff1660ff168152505081526020019060010190610652565b505050509050935093509350509193909250565b818190508484905014610707576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106fe90611ddd565b60405180910390fd5b6000600260008b815260200190815260200160002090508888826000019190610731929190610c86565b508686826001019190610745929190610c86565b508060020160006107569190610cd3565b60005b858590508110156108295781600201604051806040016040528088888581811061078657610785611dfd565b5b9050602002013581526020018686858181106107a5576107a4611dfd565b5b90506020020160208101906107ba9190611e58565b60ff1681525090806001815401808255809150506001900390600052602060002090600202016000909190919091506000820151816000015560208201518160010160006101000a81548160ff021916908360ff1602179055505050808061082190611eb4565b915050610759565b50897f9d60ff80180e842ed1a0a7833a8f4778d41c07a2a7672ff540b33c1861b7be0b8a8a8a8a8a8a8a8a604051610868989796959493929190612012565b60405180910390a250505050505050505050565b60606000806000606060008060008881526020019081526020016000209050806000018160010160009054906101000a900463ffffffff1682600201548360030160009054906101000a900460ff16846004018480546108db90611d5f565b80601f016020809104026020016040519081016040528092919081815260200182805461090790611d5f565b80156109545780601f1061092957610100808354040283529160200191610954565b820191906000526020600020905b81548152906001019060200180831161093757829003601f168201915b5050505050945080805461096790611d5f565b80601f016020809104026020016040519081016040528092919081815260200182805461099390611d5f565b80156109e05780601f106109b5576101008083540402835291602001916109e0565b820191906000526020600020905b8154815290600101906020018083116109c357829003601f168201915b50505050509050955095509550955095505091939590929450565b6040518060a0016040528088888080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505081526020018663ffffffff168152602001858152602001841515815260200183838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050508152506000808a81526020019081526020016000206000820151816000019080519060200190610ae0929190610cf7565b5060208201518160010160006101000a81548163ffffffff021916908363ffffffff1602179055506040820151816002015560608201518160030160006101000a81548160ff0219169083151502179055506080820151816004019080519060200190610b4e929190610ba0565b50905050877f89be94e685cc8ee875843439a7d50f21713114c97d7cb5091ad2fe4542b2343f88888888888888604051610b8e97969594939291906120a8565b60405180910390a25050505050505050565b828054610bac90611d5f565b90600052602060002090601f016020900481019282610bce5760008555610c15565b82601f10610be757805160ff1916838001178555610c15565b82800160010185558215610c15579182015b82811115610c14578251825591602001919060010190610bf9565b5b509050610c229190610d7d565b5090565b828054828255906000526020600020908101928215610c75579160200282015b82811115610c74578251829080519060200190610c64929190610ba0565b5091602001919060010190610c46565b5b509050610c829190610d9a565b5090565b828054828255906000526020600020908101928215610cc2579160200282015b82811115610cc1578235825591602001919060010190610ca6565b5b509050610ccf9190610dbe565b5090565b5080546000825560020290600052602060002090810190610cf49190610ddb565b50565b828054610d0390611d5f565b90600052602060002090601f016020900481019282610d255760008555610d6c565b82601f10610d3e57805160ff1916838001178555610d6c565b82800160010185558215610d6c579182015b82811115610d6b578251825591602001919060010190610d50565b5b509050610d799190610d7d565b5090565b5b80821115610d96576000816000905550600101610d7e565b5090565b5b80821115610dba5760008181610db19190610e0e565b50600101610d9b565b5090565b5b80821115610dd7576000816000905550600101610dbf565b5090565b5b80821115610e0a576000808201600090556001820160006101000a81549060ff021916905550600201610ddc565b5090565b508054610e1a90611d5f565b6000825580601f10610e2c5750610e4b565b601f016020900490600052602060002090810190610e4a9190610d7d565b5b50565b6000604051905090565b600080fd5b600080fd5b6000819050919050565b610e7581610e62565b8114610e8057600080fd5b50565b600081359050610e9281610e6c565b92915050565b600063ffffffff82169050919050565b610eb181610e98565b8114610ebc57600080fd5b50565b600081359050610ece81610ea8565b92915050565b600080fd5b600080fd5b600080fd5b60008083601f840112610ef957610ef8610ed4565b5b8235905067ffffffffffffffff811115610f1657610f15610ed9565b5b602083019150836001820283011115610f3257610f31610ede565b5b9250929050565b60008083601f840112610f4f57610f4e610ed4565b5b8235905067ffffffffffffffff811115610f6c57610f6b610ed9565b5b602083019150836020820283011115610f8857610f87610ede565b5b9250929050565b60008060008060008060008060008060006101208c8e031215610fb557610fb4610e58565b5b6000610fc38e828f01610e83565b9b50506020610fd48e828f01610e83565b9a50506040610fe58e828f01610ebf565b9950506060610ff68e828f01610ebf565b98505060806110078e828f01610ebf565b97505060a06110188e828f01610ebf565b96505060c06110298e828f01610e83565b95505060e08c013567ffffffffffffffff81111561104a57611049610e5d565b5b6110568e828f01610ee3565b94509450506101008c013567ffffffffffffffff81111561107a57611079610e5d565b5b6110868e828f01610f39565b92509250509295989b509295989b9093969950565b6000602082840312156110b1576110b0610e58565b5b60006110bf84828501610e83565b91505092915050565b6110d181610e62565b82525050565b6110e081610e98565b82525050565b600081519050919050565b600082825260208201905092915050565b60005b83811015611120578082015181840152602081019050611105565b8381111561112f576000848401525b50505050565b6000601f19601f8301169050919050565b6000611151826110e6565b61115b81856110f1565b935061116b818560208601611102565b61117481611135565b840191505092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b600082825260208201905092915050565b60006111c7826110e6565b6111d181856111ab565b93506111e1818560208601611102565b6111ea81611135565b840191505092915050565b600061120183836111bc565b905092915050565b6000602082019050919050565b60006112218261117f565b61122b818561118a565b93508360208202850161123d8561119b565b8060005b85811015611279578484038952815161125a85826111f5565b945061126583611209565b925060208a01995050600181019050611241565b50829750879550505050505092915050565b6000610100820190506112a1600083018b6110c8565b6112ae602083018a6110d7565b6112bb60408301896110d7565b6112c860608301886110d7565b6112d560808301876110d7565b6112e260a08301866110c8565b81810360c08301526112f48185611146565b905081810360e08301526113088184611216565b90509998505050505050505050565b6000819050919050565b61132a81611317565b811461133557600080fd5b50565b60008135905061134781611321565b92915050565b60006020828403121561136357611362610e58565b5b600061137184828501611338565b91505092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b6113af81610e62565b82525050565b60006113c183836113a6565b60208301905092915050565b6000602082019050919050565b60006113e58261137a565b6113ef8185611385565b93506113fa83611396565b8060005b8381101561142b57815161141288826113b5565b975061141d836113cd565b9250506001810190506113fe565b5085935050505092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b600060ff82169050919050565b61147a81611464565b82525050565b60408201600082015161149660008501826113a6565b5060208201516114a96020850182611471565b50505050565b60006114bb8383611480565b60408301905092915050565b6000602082019050919050565b60006114df82611438565b6114e98185611443565b93506114f483611454565b8060005b8381101561152557815161150c88826114af565b9750611517836114c7565b9250506001810190506114f8565b5085935050505092915050565b6000606082019050818103600083015261154c81866113da565b9050818103602083015261156081856113da565b9050818103604083015261157481846114d4565b9050949350505050565b60008083601f84011261159457611593610ed4565b5b8235905067ffffffffffffffff8111156115b1576115b0610ed9565b5b6020830191508360208202830111156115cd576115cc610ede565b5b9250929050565b60008083601f8401126115ea576115e9610ed4565b5b8235905067ffffffffffffffff81111561160757611606610ed9565b5b60208301915083602082028301111561162357611622610ede565b5b9250929050565b600080600080600080600080600060a08a8c03121561164c5761164b610e58565b5b600061165a8c828d01611338565b99505060208a013567ffffffffffffffff81111561167b5761167a610e5d565b5b6116878c828d0161157e565b985098505060408a013567ffffffffffffffff8111156116aa576116a9610e5d565b5b6116b68c828d0161157e565b965096505060608a013567ffffffffffffffff8111156116d9576116d8610e5d565b5b6116e58c828d0161157e565b945094505060808a013567ffffffffffffffff81111561170857611707610e5d565b5b6117148c828d016115d4565b92509250509295985092959850929598565b600081519050919050565b600082825260208201905092915050565b600061174d82611726565b6117578185611731565b9350611767818560208601611102565b61177081611135565b840191505092915050565b60008115159050919050565b6117908161177b565b82525050565b600060a08201905081810360008301526117b08188611742565b90506117bf60208301876110d7565b6117cc60408301866110c8565b6117d96060830185611787565b81810360808301526117eb8184611146565b90509695505050505050565b60008083601f84011261180d5761180c610ed4565b5b8235905067ffffffffffffffff81111561182a57611829610ed9565b5b60208301915083600182028301111561184657611845610ede565b5b9250929050565b6118568161177b565b811461186157600080fd5b50565b6000813590506118738161184d565b92915050565b60008060008060008060008060c0898b03121561189957611898610e58565b5b60006118a78b828c01610e83565b985050602089013567ffffffffffffffff8111156118c8576118c7610e5d565b5b6118d48b828c016117f7565b975097505060406118e78b828c01610ebf565b95505060606118f88b828c01610e83565b94505060806119098b828c01611864565b93505060a089013567ffffffffffffffff81111561192a57611929610e5d565b5b6119368b828c01610ee3565b92509250509295985092959890939650565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b61198082611135565b810181811067ffffffffffffffff8211171561199f5761199e611948565b5b80604052505050565b60006119b2610e4e565b90506119be8282611977565b919050565b600067ffffffffffffffff8211156119de576119dd611948565b5b602082029050602081019050919050565b600080fd5b600067ffffffffffffffff821115611a0f57611a0e611948565b5b611a1882611135565b9050602081019050919050565b82818337600083830152505050565b6000611a47611a42846119f4565b6119a8565b905082815260208101848484011115611a6357611a626119ef565b5b611a6e848285611a25565b509392505050565b600082601f830112611a8b57611a8a610ed4565b5b8135611a9b848260208601611a34565b91505092915050565b6000611ab7611ab2846119c3565b6119a8565b90508083825260208201905060208402830185811115611ada57611ad9610ede565b5b835b81811015611b2157803567ffffffffffffffff811115611aff57611afe610ed4565b5b808601611b0c8982611a76565b85526020850194505050602081019050611adc565b5050509392505050565b6000611b38368484611aa4565b905092915050565b6000611b4c83856110f1565b9350611b59838584611a25565b611b6283611135565b840190509392505050565b6000819050919050565b6000611b8383856111ab565b9350611b90838584611a25565b611b9983611135565b840190509392505050565b6000611bb1848484611b77565b90509392505050565b600080fd5b600080fd5b600080fd5b60008083356001602003843603038112611be657611be5611bc4565b5b83810192508235915060208301925067ffffffffffffffff821115611c0e57611c0d611bba565b5b600182023603841315611c2457611c23611bbf565b5b509250929050565b6000602082019050919050565b6000611c45838561118a565b935083602084028501611c5784611b6d565b8060005b87811015611c9d578484038952611c728284611bc9565b611c7d868284611ba4565b9550611c8884611c2c565b935060208b019a505050600181019050611c5b565b50829750879450505050509392505050565b600060e082019050611cc4600083018c6110d7565b611cd1602083018b6110d7565b611cde604083018a6110d7565b611ceb60608301896110d7565b611cf860808301886110c8565b81810360a0830152611d0b818688611b40565b905081810360c0830152611d20818486611c39565b90509a9950505050505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b60006002820490506001821680611d7757607f821691505b60208210811415611d8b57611d8a611d30565b5b50919050565b7f4d69736d61746368656420696e76616c696420656e7472696573000000000000600082015250565b6000611dc7601a83611731565b9150611dd282611d91565b602082019050919050565b60006020820190508181036000830152611df681611dba565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b611e3581611464565b8114611e4057600080fd5b50565b600081359050611e5281611e2c565b92915050565b600060208284031215611e6e57611e6d610e58565b5b6000611e7c84828501611e43565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000611ebf82611317565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff821415611ef257611ef1611e85565b5b600182019050919050565b600080fd5b6000611f0e8385611385565b93507f07ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff831115611f4157611f40611efd565b5b602083029250611f52838584611a25565b82840190509392505050565b600082825260208201905092915050565b6000819050919050565b6000611f858383611471565b60208301905092915050565b6000611fa06020840184611e43565b905092915050565b6000602082019050919050565b6000611fc18385611f5e565b9350611fcc82611f6f565b8060005b8581101561200557611fe28284611f91565b611fec8882611f79565b9750611ff783611fa8565b925050600181019050611fd0565b5085925050509392505050565b6000608082019050818103600083015261202d818a8c611f02565b9050818103602083015261204281888a611f02565b90508181036040830152612057818688611f02565b9050818103606083015261206c818486611fb5565b90509998505050505050505050565b60006120878385611731565b9350612094838584611a25565b61209d83611135565b840190509392505050565b600060a08201905081810360008301526120c381898b61207b565b90506120d260208301886110d7565b6120df60408301876110c8565b6120ec6060830186611787565b81810360808301526120ff818486611b40565b90509897505050505050505056fea2646970667358221220a3a832473b9b9cbe34042f153601e5f6082a9a31eb482cf1723490f64197748864736f6c634300080b0033"
var StoreDataSMBin = "0x"

// DeployStoreData deploys a new contract, binding an instance of StoreData to it.
func DeployStoreData(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Receipt, *StoreData, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreDataABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	var bytecode []byte
	if backend.SMCrypto() {
		bytecode = common.FromHex(StoreDataSMBin)
	} else {
		bytecode = common.FromHex(StoreDataBin)
	}
	if len(bytecode) == 0 {
		return common.Address{}, nil, nil, fmt.Errorf("cannot deploy empty bytecode")
	}
	address, receipt, contract, err := bind.DeployContract(auth, parsed, bytecode, StoreDataABI, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, receipt, &StoreData{StoreDataCaller: StoreDataCaller{contract: contract}, StoreDataTransactor: StoreDataTransactor{contract: contract}, StoreDataFilterer: StoreDataFilterer{contract: contract}}, nil
}

func AsyncDeployStoreData(auth *bind.TransactOpts, handler func(*types.Receipt, error), backend bind.ContractBackend) (*types.Transaction, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreDataABI))
	if err != nil {
		return nil, err
	}

	var bytecode []byte
	if backend.SMCrypto() {
		bytecode = common.FromHex(StoreDataSMBin)
	} else {
		bytecode = common.FromHex(StoreDataBin)
	}
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("cannot deploy empty bytecode")
	}
	tx, err := bind.AsyncDeployContract(auth, handler, parsed, bytecode, StoreDataABI, backend)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// StoreData is an auto generated Go binding around a Solidity contract.
type StoreData struct {
	StoreDataCaller     // Read-only binding to the contract
	StoreDataTransactor // Write-only binding to the contract
	StoreDataFilterer   // Log filterer for contract events
}

// StoreDataCaller is an auto generated read-only Go binding around a Solidity contract.
type StoreDataCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreDataTransactor is an auto generated write-only Go binding around a Solidity contract.
type StoreDataTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreDataFilterer is an auto generated log filtering Go binding around a Solidity contract events.
type StoreDataFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreDataSession is an auto generated Go binding around a Solidity contract,
// with pre-set call and transact options.
type StoreDataSession struct {
	Contract     *StoreData        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StoreDataCallerSession is an auto generated read-only Go binding around a Solidity contract,
// with pre-set call options.
type StoreDataCallerSession struct {
	Contract *StoreDataCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// StoreDataTransactorSession is an auto generated write-only Go binding around a Solidity contract,
// with pre-set transact options.
type StoreDataTransactorSession struct {
	Contract     *StoreDataTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// StoreDataRaw is an auto generated low-level Go binding around a Solidity contract.
type StoreDataRaw struct {
	Contract *StoreData // Generic contract binding to access the raw methods on
}

// StoreDataCallerRaw is an auto generated low-level read-only Go binding around a Solidity contract.
type StoreDataCallerRaw struct {
	Contract *StoreDataCaller // Generic read-only contract binding to access the raw methods on
}

// StoreDataTransactorRaw is an auto generated low-level write-only Go binding around a Solidity contract.
type StoreDataTransactorRaw struct {
	Contract *StoreDataTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStoreData creates a new instance of StoreData, bound to a specific deployed contract.
func NewStoreData(address common.Address, backend bind.ContractBackend) (*StoreData, error) {
	contract, err := bindStoreData(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StoreData{StoreDataCaller: StoreDataCaller{contract: contract}, StoreDataTransactor: StoreDataTransactor{contract: contract}, StoreDataFilterer: StoreDataFilterer{contract: contract}}, nil
}

// NewStoreDataCaller creates a new read-only instance of StoreData, bound to a specific deployed contract.
func NewStoreDataCaller(address common.Address, caller bind.ContractCaller) (*StoreDataCaller, error) {
	contract, err := bindStoreData(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StoreDataCaller{contract: contract}, nil
}

// NewStoreDataTransactor creates a new write-only instance of StoreData, bound to a specific deployed contract.
func NewStoreDataTransactor(address common.Address, transactor bind.ContractTransactor) (*StoreDataTransactor, error) {
	contract, err := bindStoreData(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StoreDataTransactor{contract: contract}, nil
}

// NewStoreDataFilterer creates a new log filterer instance of StoreData, bound to a specific deployed contract.
func NewStoreDataFilterer(address common.Address, filterer bind.ContractFilterer) (*StoreDataFilterer, error) {
	contract, err := bindStoreData(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StoreDataFilterer{contract: contract}, nil
}

// bindStoreData binds a generic wrapper to an already deployed contract.
func bindStoreData(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreDataABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StoreData *StoreDataRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _StoreData.Contract.StoreDataCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StoreData *StoreDataRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreDataTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StoreData *StoreDataRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreDataTransactor.contract.TransactWithResult(opts, result, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StoreData *StoreDataCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _StoreData.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StoreData *StoreDataTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StoreData *StoreDataTransactorRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.contract.TransactWithResult(opts, result, method, params...)
}

// GetEpochRecordFull is a free data retrieval call binding the contract method 0xcb68c5d7.
//
// Solidity: function getEpochRecordFull(uint256 id) constant returns(bytes32[] justified, bytes32[] commits, []StoreDataInvalidEntry invalids)
func (_StoreData *StoreDataCaller) GetEpochRecordFull(opts *bind.CallOpts, id *big.Int) (struct {
	Justified [][32]byte
	Commits   [][32]byte
	Invalids  []StoreDataInvalidEntry
}, error) {
	ret := new(struct {
		Justified [][32]byte
		Commits   [][32]byte
		Invalids  []StoreDataInvalidEntry
	})
	out := ret
	err := _StoreData.contract.Call(opts, out, "getEpochRecordFull", id)
	return *ret, err
}

// GetEpochRecordFull is a free data retrieval call binding the contract method 0xcb68c5d7.
//
// Solidity: function getEpochRecordFull(uint256 id) constant returns(bytes32[] justified, bytes32[] commits, []StoreDataInvalidEntry invalids)
func (_StoreData *StoreDataSession) GetEpochRecordFull(id *big.Int) (struct {
	Justified [][32]byte
	Commits   [][32]byte
	Invalids  []StoreDataInvalidEntry
}, error) {
	return _StoreData.Contract.GetEpochRecordFull(&_StoreData.CallOpts, id)
}

// GetEpochRecordFull is a free data retrieval call binding the contract method 0xcb68c5d7.
//
// Solidity: function getEpochRecordFull(uint256 id) constant returns(bytes32[] justified, bytes32[] commits, []StoreDataInvalidEntry invalids)
func (_StoreData *StoreDataCallerSession) GetEpochRecordFull(id *big.Int) (struct {
	Justified [][32]byte
	Commits   [][32]byte
	Invalids  []StoreDataInvalidEntry
}, error) {
	return _StoreData.Contract.GetEpochRecordFull(&_StoreData.CallOpts, id)
}

// GetInitTask is a free data retrieval call binding the contract method 0xe391cfaf.
//
// Solidity: function getInitTask(bytes32 sign) constant returns(string name, uint32 size, bytes32 model, bool isReliable, bytes params)
func (_StoreData *StoreDataCaller) GetInitTask(opts *bind.CallOpts, sign [32]byte) (struct {
	Name       string
	Size       uint32
	Model      [32]byte
	IsReliable bool
	Params     []byte
}, error) {
	ret := new(struct {
		Name       string
		Size       uint32
		Model      [32]byte
		IsReliable bool
		Params     []byte
	})
	out := ret
	err := _StoreData.contract.Call(opts, out, "getInitTask", sign)
	return *ret, err
}

// GetInitTask is a free data retrieval call binding the contract method 0xe391cfaf.
//
// Solidity: function getInitTask(bytes32 sign) constant returns(string name, uint32 size, bytes32 model, bool isReliable, bytes params)
func (_StoreData *StoreDataSession) GetInitTask(sign [32]byte) (struct {
	Name       string
	Size       uint32
	Model      [32]byte
	IsReliable bool
	Params     []byte
}, error) {
	return _StoreData.Contract.GetInitTask(&_StoreData.CallOpts, sign)
}

// GetInitTask is a free data retrieval call binding the contract method 0xe391cfaf.
//
// Solidity: function getInitTask(bytes32 sign) constant returns(string name, uint32 size, bytes32 model, bool isReliable, bytes params)
func (_StoreData *StoreDataCallerSession) GetInitTask(sign [32]byte) (struct {
	Name       string
	Size       uint32
	Model      [32]byte
	IsReliable bool
	Params     []byte
}, error) {
	return _StoreData.Contract.GetInitTask(&_StoreData.CallOpts, sign)
}

// GetTaskProcess is a free data retrieval call binding the contract method 0x6ba3c2c7.
//
// Solidity: function getTaskProcess(bytes32 hash) constant returns(bytes32 sign, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures)
func (_StoreData *StoreDataCaller) GetTaskProcess(opts *bind.CallOpts, hash [32]byte) (struct {
	Sign       [32]byte
	Slot       uint32
	Process    uint32
	Id         uint32
	Epoch      uint32
	Commitment [32]byte
	Proof      []byte
	Signatures [][]byte
}, error) {
	ret := new(struct {
		Sign       [32]byte
		Slot       uint32
		Process    uint32
		Id         uint32
		Epoch      uint32
		Commitment [32]byte
		Proof      []byte
		Signatures [][]byte
	})
	out := ret
	err := _StoreData.contract.Call(opts, out, "getTaskProcess", hash)
	return *ret, err
}

// GetTaskProcess is a free data retrieval call binding the contract method 0x6ba3c2c7.
//
// Solidity: function getTaskProcess(bytes32 hash) constant returns(bytes32 sign, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures)
func (_StoreData *StoreDataSession) GetTaskProcess(hash [32]byte) (struct {
	Sign       [32]byte
	Slot       uint32
	Process    uint32
	Id         uint32
	Epoch      uint32
	Commitment [32]byte
	Proof      []byte
	Signatures [][]byte
}, error) {
	return _StoreData.Contract.GetTaskProcess(&_StoreData.CallOpts, hash)
}

// GetTaskProcess is a free data retrieval call binding the contract method 0x6ba3c2c7.
//
// Solidity: function getTaskProcess(bytes32 hash) constant returns(bytes32 sign, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures)
func (_StoreData *StoreDataCallerSession) GetTaskProcess(hash [32]byte) (struct {
	Sign       [32]byte
	Slot       uint32
	Process    uint32
	Id         uint32
	Epoch      uint32
	Commitment [32]byte
	Proof      []byte
	Signatures [][]byte
}, error) {
	return _StoreData.Contract.GetTaskProcess(&_StoreData.CallOpts, hash)
}

// StoreEpochRecord is a paid mutator transaction binding the contract method 0xcc307441.
//
// Solidity: function storeEpochRecord(uint256 id, bytes32[] justified, bytes32[] commits, bytes32[] invalidHashes, uint8[] invalidReasons) returns()
func (_StoreData *StoreDataTransactor) StoreEpochRecord(opts *bind.TransactOpts, id *big.Int, justified [][32]byte, commits [][32]byte, invalidHashes [][32]byte, invalidReasons []uint8) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _StoreData.contract.TransactWithResult(opts, out, "storeEpochRecord", id, justified, commits, invalidHashes, invalidReasons)
	return transaction, receipt, err
}

func (_StoreData *StoreDataTransactor) AsyncStoreEpochRecord(handler func(*types.Receipt, error), opts *bind.TransactOpts, id *big.Int, justified [][32]byte, commits [][32]byte, invalidHashes [][32]byte, invalidReasons []uint8) (*types.Transaction, error) {
	return _StoreData.contract.AsyncTransact(opts, handler, "storeEpochRecord", id, justified, commits, invalidHashes, invalidReasons)
}

// StoreEpochRecord is a paid mutator transaction binding the contract method 0xcc307441.
//
// Solidity: function storeEpochRecord(uint256 id, bytes32[] justified, bytes32[] commits, bytes32[] invalidHashes, uint8[] invalidReasons) returns()
func (_StoreData *StoreDataSession) StoreEpochRecord(id *big.Int, justified [][32]byte, commits [][32]byte, invalidHashes [][32]byte, invalidReasons []uint8) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreEpochRecord(&_StoreData.TransactOpts, id, justified, commits, invalidHashes, invalidReasons)
}

func (_StoreData *StoreDataSession) AsyncStoreEpochRecord(handler func(*types.Receipt, error), id *big.Int, justified [][32]byte, commits [][32]byte, invalidHashes [][32]byte, invalidReasons []uint8) (*types.Transaction, error) {
	return _StoreData.Contract.AsyncStoreEpochRecord(handler, &_StoreData.TransactOpts, id, justified, commits, invalidHashes, invalidReasons)
}

// StoreEpochRecord is a paid mutator transaction binding the contract method 0xcc307441.
//
// Solidity: function storeEpochRecord(uint256 id, bytes32[] justified, bytes32[] commits, bytes32[] invalidHashes, uint8[] invalidReasons) returns()
func (_StoreData *StoreDataTransactorSession) StoreEpochRecord(id *big.Int, justified [][32]byte, commits [][32]byte, invalidHashes [][32]byte, invalidReasons []uint8) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreEpochRecord(&_StoreData.TransactOpts, id, justified, commits, invalidHashes, invalidReasons)
}

func (_StoreData *StoreDataTransactorSession) AsyncStoreEpochRecord(handler func(*types.Receipt, error), id *big.Int, justified [][32]byte, commits [][32]byte, invalidHashes [][32]byte, invalidReasons []uint8) (*types.Transaction, error) {
	return _StoreData.Contract.AsyncStoreEpochRecord(handler, &_StoreData.TransactOpts, id, justified, commits, invalidHashes, invalidReasons)
}

// StoreInitTask is a paid mutator transaction binding the contract method 0xec2722d0.
//
// Solidity: function storeInitTask(bytes32 sign, string name, uint32 size, bytes32 model, bool isReliable, bytes params) returns()
func (_StoreData *StoreDataTransactor) StoreInitTask(opts *bind.TransactOpts, sign [32]byte, name string, size uint32, model [32]byte, isReliable bool, params []byte) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _StoreData.contract.TransactWithResult(opts, out, "storeInitTask", sign, name, size, model, isReliable, params)
	return transaction, receipt, err
}

func (_StoreData *StoreDataTransactor) AsyncStoreInitTask(handler func(*types.Receipt, error), opts *bind.TransactOpts, sign [32]byte, name string, size uint32, model [32]byte, isReliable bool, params []byte) (*types.Transaction, error) {
	return _StoreData.contract.AsyncTransact(opts, handler, "storeInitTask", sign, name, size, model, isReliable, params)
}

// StoreInitTask is a paid mutator transaction binding the contract method 0xec2722d0.
//
// Solidity: function storeInitTask(bytes32 sign, string name, uint32 size, bytes32 model, bool isReliable, bytes params) returns()
func (_StoreData *StoreDataSession) StoreInitTask(sign [32]byte, name string, size uint32, model [32]byte, isReliable bool, params []byte) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreInitTask(&_StoreData.TransactOpts, sign, name, size, model, isReliable, params)
}

func (_StoreData *StoreDataSession) AsyncStoreInitTask(handler func(*types.Receipt, error), sign [32]byte, name string, size uint32, model [32]byte, isReliable bool, params []byte) (*types.Transaction, error) {
	return _StoreData.Contract.AsyncStoreInitTask(handler, &_StoreData.TransactOpts, sign, name, size, model, isReliable, params)
}

// StoreInitTask is a paid mutator transaction binding the contract method 0xec2722d0.
//
// Solidity: function storeInitTask(bytes32 sign, string name, uint32 size, bytes32 model, bool isReliable, bytes params) returns()
func (_StoreData *StoreDataTransactorSession) StoreInitTask(sign [32]byte, name string, size uint32, model [32]byte, isReliable bool, params []byte) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreInitTask(&_StoreData.TransactOpts, sign, name, size, model, isReliable, params)
}

func (_StoreData *StoreDataTransactorSession) AsyncStoreInitTask(handler func(*types.Receipt, error), sign [32]byte, name string, size uint32, model [32]byte, isReliable bool, params []byte) (*types.Transaction, error) {
	return _StoreData.Contract.AsyncStoreInitTask(handler, &_StoreData.TransactOpts, sign, name, size, model, isReliable, params)
}

// StoreTaskProcess is a paid mutator transaction binding the contract method 0x3f566321.
//
// Solidity: function storeTaskProcess(bytes32 sign, bytes32 hash, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures) returns()
func (_StoreData *StoreDataTransactor) StoreTaskProcess(opts *bind.TransactOpts, sign [32]byte, hash [32]byte, slot uint32, process uint32, id uint32, epoch uint32, commitment [32]byte, proof []byte, signatures [][]byte) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _StoreData.contract.TransactWithResult(opts, out, "storeTaskProcess", sign, hash, slot, process, id, epoch, commitment, proof, signatures)
	return transaction, receipt, err
}

func (_StoreData *StoreDataTransactor) AsyncStoreTaskProcess(handler func(*types.Receipt, error), opts *bind.TransactOpts, sign [32]byte, hash [32]byte, slot uint32, process uint32, id uint32, epoch uint32, commitment [32]byte, proof []byte, signatures [][]byte) (*types.Transaction, error) {
	return _StoreData.contract.AsyncTransact(opts, handler, "storeTaskProcess", sign, hash, slot, process, id, epoch, commitment, proof, signatures)
}

// StoreTaskProcess is a paid mutator transaction binding the contract method 0x3f566321.
//
// Solidity: function storeTaskProcess(bytes32 sign, bytes32 hash, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures) returns()
func (_StoreData *StoreDataSession) StoreTaskProcess(sign [32]byte, hash [32]byte, slot uint32, process uint32, id uint32, epoch uint32, commitment [32]byte, proof []byte, signatures [][]byte) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreTaskProcess(&_StoreData.TransactOpts, sign, hash, slot, process, id, epoch, commitment, proof, signatures)
}

func (_StoreData *StoreDataSession) AsyncStoreTaskProcess(handler func(*types.Receipt, error), sign [32]byte, hash [32]byte, slot uint32, process uint32, id uint32, epoch uint32, commitment [32]byte, proof []byte, signatures [][]byte) (*types.Transaction, error) {
	return _StoreData.Contract.AsyncStoreTaskProcess(handler, &_StoreData.TransactOpts, sign, hash, slot, process, id, epoch, commitment, proof, signatures)
}

// StoreTaskProcess is a paid mutator transaction binding the contract method 0x3f566321.
//
// Solidity: function storeTaskProcess(bytes32 sign, bytes32 hash, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures) returns()
func (_StoreData *StoreDataTransactorSession) StoreTaskProcess(sign [32]byte, hash [32]byte, slot uint32, process uint32, id uint32, epoch uint32, commitment [32]byte, proof []byte, signatures [][]byte) (*types.Transaction, *types.Receipt, error) {
	return _StoreData.Contract.StoreTaskProcess(&_StoreData.TransactOpts, sign, hash, slot, process, id, epoch, commitment, proof, signatures)
}

func (_StoreData *StoreDataTransactorSession) AsyncStoreTaskProcess(handler func(*types.Receipt, error), sign [32]byte, hash [32]byte, slot uint32, process uint32, id uint32, epoch uint32, commitment [32]byte, proof []byte, signatures [][]byte) (*types.Transaction, error) {
	return _StoreData.Contract.AsyncStoreTaskProcess(handler, &_StoreData.TransactOpts, sign, hash, slot, process, id, epoch, commitment, proof, signatures)
}

// StoreDataEpochRecordStored represents a EpochRecordStored event raised by the StoreData contract.
type StoreDataEpochRecordStored struct {
	Id             *big.Int
	Justified      [][32]byte
	Commits        [][32]byte
	InvalidHashes  [][32]byte
	InvalidReasons []uint8
	Raw            types.Log // Blockchain specific contextual infos
}

// WatchEpochRecordStored is a free log subscription operation binding the contract event 0x9d60ff80180e842ed1a0a7833a8f4778d41c07a2a7672ff540b33c1861b7be0b.
//
// Solidity: event EpochRecordStored(uint256 indexed id, bytes32[] justified, bytes32[] commits, bytes32[] invalidHashes, uint8[] invalidReasons)
func (_StoreData *StoreDataFilterer) WatchEpochRecordStored(fromBlock *int64, handler func(int, []types.Log), id *big.Int) (string, error) {
	return _StoreData.contract.WatchLogs(fromBlock, handler, "EpochRecordStored", id)
}

func (_StoreData *StoreDataFilterer) WatchAllEpochRecordStored(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _StoreData.contract.WatchLogs(fromBlock, handler, "EpochRecordStored")
}

// ParseEpochRecordStored is a log parse operation binding the contract event 0x9d60ff80180e842ed1a0a7833a8f4778d41c07a2a7672ff540b33c1861b7be0b.
//
// Solidity: event EpochRecordStored(uint256 indexed id, bytes32[] justified, bytes32[] commits, bytes32[] invalidHashes, uint8[] invalidReasons)
func (_StoreData *StoreDataFilterer) ParseEpochRecordStored(log types.Log) (*StoreDataEpochRecordStored, error) {
	event := new(StoreDataEpochRecordStored)
	if err := _StoreData.contract.UnpackLog(event, "EpochRecordStored", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WatchEpochRecordStored is a free log subscription operation binding the contract event 0x9d60ff80180e842ed1a0a7833a8f4778d41c07a2a7672ff540b33c1861b7be0b.
//
// Solidity: event EpochRecordStored(uint256 indexed id, bytes32[] justified, bytes32[] commits, bytes32[] invalidHashes, uint8[] invalidReasons)
func (_StoreData *StoreDataSession) WatchEpochRecordStored(fromBlock *int64, handler func(int, []types.Log), id *big.Int) (string, error) {
	return _StoreData.Contract.WatchEpochRecordStored(fromBlock, handler, id)
}

func (_StoreData *StoreDataSession) WatchAllEpochRecordStored(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _StoreData.Contract.WatchAllEpochRecordStored(fromBlock, handler)
}

// ParseEpochRecordStored is a log parse operation binding the contract event 0x9d60ff80180e842ed1a0a7833a8f4778d41c07a2a7672ff540b33c1861b7be0b.
//
// Solidity: event EpochRecordStored(uint256 indexed id, bytes32[] justified, bytes32[] commits, bytes32[] invalidHashes, uint8[] invalidReasons)
func (_StoreData *StoreDataSession) ParseEpochRecordStored(log types.Log) (*StoreDataEpochRecordStored, error) {
	return _StoreData.Contract.ParseEpochRecordStored(log)
}

// StoreDataInitTaskStored represents a InitTaskStored event raised by the StoreData contract.
type StoreDataInitTaskStored struct {
	Sign       [32]byte
	Name       string
	Size       uint32
	Model      [32]byte
	IsReliable bool
	Params     []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// WatchInitTaskStored is a free log subscription operation binding the contract event 0x89be94e685cc8ee875843439a7d50f21713114c97d7cb5091ad2fe4542b2343f.
//
// Solidity: event InitTaskStored(bytes32 indexed sign, string name, uint32 size, bytes32 model, bool isReliable, bytes params)
func (_StoreData *StoreDataFilterer) WatchInitTaskStored(fromBlock *int64, handler func(int, []types.Log), sign [32]byte) (string, error) {
	return _StoreData.contract.WatchLogs(fromBlock, handler, "InitTaskStored", sign)
}

func (_StoreData *StoreDataFilterer) WatchAllInitTaskStored(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _StoreData.contract.WatchLogs(fromBlock, handler, "InitTaskStored")
}

// ParseInitTaskStored is a log parse operation binding the contract event 0x89be94e685cc8ee875843439a7d50f21713114c97d7cb5091ad2fe4542b2343f.
//
// Solidity: event InitTaskStored(bytes32 indexed sign, string name, uint32 size, bytes32 model, bool isReliable, bytes params)
func (_StoreData *StoreDataFilterer) ParseInitTaskStored(log types.Log) (*StoreDataInitTaskStored, error) {
	event := new(StoreDataInitTaskStored)
	if err := _StoreData.contract.UnpackLog(event, "InitTaskStored", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WatchInitTaskStored is a free log subscription operation binding the contract event 0x89be94e685cc8ee875843439a7d50f21713114c97d7cb5091ad2fe4542b2343f.
//
// Solidity: event InitTaskStored(bytes32 indexed sign, string name, uint32 size, bytes32 model, bool isReliable, bytes params)
func (_StoreData *StoreDataSession) WatchInitTaskStored(fromBlock *int64, handler func(int, []types.Log), sign [32]byte) (string, error) {
	return _StoreData.Contract.WatchInitTaskStored(fromBlock, handler, sign)
}

func (_StoreData *StoreDataSession) WatchAllInitTaskStored(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _StoreData.Contract.WatchAllInitTaskStored(fromBlock, handler)
}

// ParseInitTaskStored is a log parse operation binding the contract event 0x89be94e685cc8ee875843439a7d50f21713114c97d7cb5091ad2fe4542b2343f.
//
// Solidity: event InitTaskStored(bytes32 indexed sign, string name, uint32 size, bytes32 model, bool isReliable, bytes params)
func (_StoreData *StoreDataSession) ParseInitTaskStored(log types.Log) (*StoreDataInitTaskStored, error) {
	return _StoreData.Contract.ParseInitTaskStored(log)
}

// StoreDataTaskProcessStored represents a TaskProcessStored event raised by the StoreData contract.
type StoreDataTaskProcessStored struct {
	Sign       [32]byte
	Hash       [32]byte
	Slot       uint32
	Process    uint32
	Id         uint32
	Epoch      uint32
	Commitment [32]byte
	Proof      []byte
	Signatures [][]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// WatchTaskProcessStored is a free log subscription operation binding the contract event 0x7c99774bedb5b2f1536984fdee6a8adeaca958ef5866ba04a522c829bdb7b12f.
//
// Solidity: event TaskProcessStored(bytes32 indexed sign, bytes32 indexed hash, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures)
func (_StoreData *StoreDataFilterer) WatchTaskProcessStored(fromBlock *int64, handler func(int, []types.Log), sign [32]byte, hash [32]byte) (string, error) {
	return _StoreData.contract.WatchLogs(fromBlock, handler, "TaskProcessStored", sign, hash)
}

func (_StoreData *StoreDataFilterer) WatchAllTaskProcessStored(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _StoreData.contract.WatchLogs(fromBlock, handler, "TaskProcessStored")
}

// ParseTaskProcessStored is a log parse operation binding the contract event 0x7c99774bedb5b2f1536984fdee6a8adeaca958ef5866ba04a522c829bdb7b12f.
//
// Solidity: event TaskProcessStored(bytes32 indexed sign, bytes32 indexed hash, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures)
func (_StoreData *StoreDataFilterer) ParseTaskProcessStored(log types.Log) (*StoreDataTaskProcessStored, error) {
	event := new(StoreDataTaskProcessStored)
	if err := _StoreData.contract.UnpackLog(event, "TaskProcessStored", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WatchTaskProcessStored is a free log subscription operation binding the contract event 0x7c99774bedb5b2f1536984fdee6a8adeaca958ef5866ba04a522c829bdb7b12f.
//
// Solidity: event TaskProcessStored(bytes32 indexed sign, bytes32 indexed hash, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures)
func (_StoreData *StoreDataSession) WatchTaskProcessStored(fromBlock *int64, handler func(int, []types.Log), sign [32]byte, hash [32]byte) (string, error) {
	return _StoreData.Contract.WatchTaskProcessStored(fromBlock, handler, sign, hash)
}

func (_StoreData *StoreDataSession) WatchAllTaskProcessStored(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _StoreData.Contract.WatchAllTaskProcessStored(fromBlock, handler)
}

// ParseTaskProcessStored is a log parse operation binding the contract event 0x7c99774bedb5b2f1536984fdee6a8adeaca958ef5866ba04a522c829bdb7b12f.
//
// Solidity: event TaskProcessStored(bytes32 indexed sign, bytes32 indexed hash, uint32 slot, uint32 process, uint32 id, uint32 epoch, bytes32 commitment, bytes proof, bytes[] signatures)
func (_StoreData *StoreDataSession) ParseTaskProcessStored(log types.Log) (*StoreDataTaskProcessStored, error) {
	return _StoreData.Contract.ParseTaskProcessStored(log)
}
