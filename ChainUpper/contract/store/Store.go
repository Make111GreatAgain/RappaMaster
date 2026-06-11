//go:build ignore

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package Store

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

// StoreABI is the input ABI used to generate the binding from.
const StoreABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"value\",\"type\":\"bytes32\"}],\"name\":\"ItemSet\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"}],\"name\":\"getItem\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"items\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"value\",\"type\":\"bytes32\"}],\"name\":\"setItem\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"keys\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"values\",\"type\":\"bytes32[]\"}],\"name\":\"setItems\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// StoreBin is the compiled bytecode used for deploying new contracts.
var StoreBin = "0x608060405234801561001057600080fd5b50610605806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806348f343f314610051578063728e99f914610081578063aa0372e71461009d578063f56256c7146100cd575b600080fd5b61006b600480360381019061006691906102d5565b6100e9565b6040516100789190610311565b60405180910390f35b61009b60048036038101906100969190610391565b610101565b005b6100b760048036038101906100b291906102d5565b610225565b6040516100c49190610311565b60405180910390f35b6100e760048036038101906100e29190610412565b610241565b005b60006020528060005260406000206000915090505481565b818190508484905014610149576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610140906104d5565b60405180910390fd5b60005b8484905081101561021e5782828281811061016a576101696104f5565b5b90506020020135600080878785818110610187576101866104f5565b5b905060200201358152602001908152602001600020819055507fe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d48585838181106101d4576101d36104f5565b5b905060200201358484848181106101ee576101ed6104f5565b5b90506020020135604051610203929190610524565b60405180910390a1808061021690610586565b91505061014c565b5050505050565b6000806000838152602001908152602001600020549050919050565b80600080848152602001908152602001600020819055507fe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d48282604051610289929190610524565b60405180910390a15050565b600080fd5b600080fd5b6000819050919050565b6102b28161029f565b81146102bd57600080fd5b50565b6000813590506102cf816102a9565b92915050565b6000602082840312156102eb576102ea610295565b5b60006102f9848285016102c0565b91505092915050565b61030b8161029f565b82525050565b60006020820190506103266000830184610302565b92915050565b600080fd5b600080fd5b600080fd5b60008083601f8401126103515761035061032c565b5b8235905067ffffffffffffffff81111561036e5761036d610331565b5b60208301915083602082028301111561038a57610389610336565b5b9250929050565b600080600080604085870312156103ab576103aa610295565b5b600085013567ffffffffffffffff8111156103c9576103c861029a565b5b6103d58782880161033b565b9450945050602085013567ffffffffffffffff8111156103f8576103f761029a565b5b6104048782880161033b565b925092505092959194509250565b6000806040838503121561042957610428610295565b5b6000610437858286016102c0565b9250506020610448858286016102c0565b9150509250929050565b600082825260208201905092915050565b7f4b65797320616e642076616c756573206172726179206d757374206265206f6660008201527f2073616d65206c656e6774680000000000000000000000000000000000000000602082015250565b60006104bf602c83610452565b91506104ca82610463565b604082019050919050565b600060208201905081810360008301526104ee816104b2565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60006040820190506105396000830185610302565b6105466020830184610302565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000819050919050565b60006105918261057c565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8214156105c4576105c361054d565b5b60018201905091905056fea264697066735822122020797a2c013153c432906ece8813d3793982568d2f8ad4c4a3f544c5ac70b12d64736f6c634300080b0033"
var StoreSMBin = "0x"

// DeployStore deploys a new contract, binding an instance of Store to it.
func DeployStore(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Receipt, *Store, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	var bytecode []byte
	if backend.SMCrypto() {
		bytecode = common.FromHex(StoreSMBin)
	} else {
		bytecode = common.FromHex(StoreBin)
	}
	if len(bytecode) == 0 {
		return common.Address{}, nil, nil, fmt.Errorf("cannot deploy empty bytecode")
	}
	address, receipt, contract, err := bind.DeployContract(auth, parsed, bytecode, StoreABI, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, receipt, &Store{StoreCaller: StoreCaller{contract: contract}, StoreTransactor: StoreTransactor{contract: contract}, StoreFilterer: StoreFilterer{contract: contract}}, nil
}

func AsyncDeployStore(auth *bind.TransactOpts, handler func(*types.Receipt, error), backend bind.ContractBackend) (*types.Transaction, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		return nil, err
	}

	var bytecode []byte
	if backend.SMCrypto() {
		bytecode = common.FromHex(StoreSMBin)
	} else {
		bytecode = common.FromHex(StoreBin)
	}
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("cannot deploy empty bytecode")
	}
	tx, err := bind.AsyncDeployContract(auth, handler, parsed, bytecode, StoreABI, backend)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Store is an auto generated Go binding around a Solidity contract.
type Store struct {
	StoreCaller     // Read-only binding to the contract
	StoreTransactor // Write-only binding to the contract
	StoreFilterer   // Log filterer for contract events
}

// StoreCaller is an auto generated read-only Go binding around a Solidity contract.
type StoreCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreTransactor is an auto generated write-only Go binding around a Solidity contract.
type StoreTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreFilterer is an auto generated log filtering Go binding around a Solidity contract events.
type StoreFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreSession is an auto generated Go binding around a Solidity contract,
// with pre-set call and transact options.
type StoreSession struct {
	Contract     *Store            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StoreCallerSession is an auto generated read-only Go binding around a Solidity contract,
// with pre-set call options.
type StoreCallerSession struct {
	Contract *StoreCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// StoreTransactorSession is an auto generated write-only Go binding around a Solidity contract,
// with pre-set transact options.
type StoreTransactorSession struct {
	Contract     *StoreTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StoreRaw is an auto generated low-level Go binding around a Solidity contract.
type StoreRaw struct {
	Contract *Store // Generic contract binding to access the raw methods on
}

// StoreCallerRaw is an auto generated low-level read-only Go binding around a Solidity contract.
type StoreCallerRaw struct {
	Contract *StoreCaller // Generic read-only contract binding to access the raw methods on
}

// StoreTransactorRaw is an auto generated low-level write-only Go binding around a Solidity contract.
type StoreTransactorRaw struct {
	Contract *StoreTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStore creates a new instance of Store, bound to a specific deployed contract.
func NewStore(address common.Address, backend bind.ContractBackend) (*Store, error) {
	contract, err := bindStore(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Store{StoreCaller: StoreCaller{contract: contract}, StoreTransactor: StoreTransactor{contract: contract}, StoreFilterer: StoreFilterer{contract: contract}}, nil
}

// NewStoreCaller creates a new read-only instance of Store, bound to a specific deployed contract.
func NewStoreCaller(address common.Address, caller bind.ContractCaller) (*StoreCaller, error) {
	contract, err := bindStore(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StoreCaller{contract: contract}, nil
}

// NewStoreTransactor creates a new write-only instance of Store, bound to a specific deployed contract.
func NewStoreTransactor(address common.Address, transactor bind.ContractTransactor) (*StoreTransactor, error) {
	contract, err := bindStore(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StoreTransactor{contract: contract}, nil
}

// NewStoreFilterer creates a new log filterer instance of Store, bound to a specific deployed contract.
func NewStoreFilterer(address common.Address, filterer bind.ContractFilterer) (*StoreFilterer, error) {
	contract, err := bindStore(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StoreFilterer{contract: contract}, nil
}

// bindStore binds a generic wrapper to an already deployed contract.
func bindStore(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Store *StoreRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Store.Contract.StoreCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Store *StoreRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.StoreTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Store *StoreRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.StoreTransactor.contract.TransactWithResult(opts, result, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Store *StoreCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Store.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Store *StoreTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Store *StoreTransactorRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.contract.TransactWithResult(opts, result, method, params...)
}

// GetItem is a free data retrieval call binding the contract method 0xaa0372e7.
//
// Solidity: function getItem(bytes32 key) constant returns(bytes32)
func (_Store *StoreCaller) GetItem(opts *bind.CallOpts, key [32]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Store.contract.Call(opts, out, "getItem", key)
	return *ret0, err
}

// GetItem is a free data retrieval call binding the contract method 0xaa0372e7.
//
// Solidity: function getItem(bytes32 key) constant returns(bytes32)
func (_Store *StoreSession) GetItem(key [32]byte) ([32]byte, error) {
	return _Store.Contract.GetItem(&_Store.CallOpts, key)
}

// GetItem is a free data retrieval call binding the contract method 0xaa0372e7.
//
// Solidity: function getItem(bytes32 key) constant returns(bytes32)
func (_Store *StoreCallerSession) GetItem(key [32]byte) ([32]byte, error) {
	return _Store.Contract.GetItem(&_Store.CallOpts, key)
}

// Items is a free data retrieval call binding the contract method 0x48f343f3.
//
// Solidity: function items(bytes32 ) constant returns(bytes32)
func (_Store *StoreCaller) Items(opts *bind.CallOpts, arg0 [32]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Store.contract.Call(opts, out, "items", arg0)
	return *ret0, err
}

// Items is a free data retrieval call binding the contract method 0x48f343f3.
//
// Solidity: function items(bytes32 ) constant returns(bytes32)
func (_Store *StoreSession) Items(arg0 [32]byte) ([32]byte, error) {
	return _Store.Contract.Items(&_Store.CallOpts, arg0)
}

// Items is a free data retrieval call binding the contract method 0x48f343f3.
//
// Solidity: function items(bytes32 ) constant returns(bytes32)
func (_Store *StoreCallerSession) Items(arg0 [32]byte) ([32]byte, error) {
	return _Store.Contract.Items(&_Store.CallOpts, arg0)
}

// SetItem is a paid mutator transaction binding the contract method 0xf56256c7.
//
// Solidity: function setItem(bytes32 key, bytes32 value) returns()
func (_Store *StoreTransactor) SetItem(opts *bind.TransactOpts, key [32]byte, value [32]byte) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _Store.contract.TransactWithResult(opts, out, "setItem", key, value)
	return transaction, receipt, err
}

func (_Store *StoreTransactor) AsyncSetItem(handler func(*types.Receipt, error), opts *bind.TransactOpts, key [32]byte, value [32]byte) (*types.Transaction, error) {
	return _Store.contract.AsyncTransact(opts, handler, "setItem", key, value)
}

// SetItem is a paid mutator transaction binding the contract method 0xf56256c7.
//
// Solidity: function setItem(bytes32 key, bytes32 value) returns()
func (_Store *StoreSession) SetItem(key [32]byte, value [32]byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItem(&_Store.TransactOpts, key, value)
}

func (_Store *StoreSession) AsyncSetItem(handler func(*types.Receipt, error), key [32]byte, value [32]byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItem(handler, &_Store.TransactOpts, key, value)
}

// SetItem is a paid mutator transaction binding the contract method 0xf56256c7.
//
// Solidity: function setItem(bytes32 key, bytes32 value) returns()
func (_Store *StoreTransactorSession) SetItem(key [32]byte, value [32]byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItem(&_Store.TransactOpts, key, value)
}

func (_Store *StoreTransactorSession) AsyncSetItem(handler func(*types.Receipt, error), key [32]byte, value [32]byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItem(handler, &_Store.TransactOpts, key, value)
}

// SetItems is a paid mutator transaction binding the contract method 0x728e99f9.
//
// Solidity: function setItems(bytes32[] keys, bytes32[] values) returns()
func (_Store *StoreTransactor) SetItems(opts *bind.TransactOpts, keys [][32]byte, values [][32]byte) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _Store.contract.TransactWithResult(opts, out, "setItems", keys, values)
	return transaction, receipt, err
}

func (_Store *StoreTransactor) AsyncSetItems(handler func(*types.Receipt, error), opts *bind.TransactOpts, keys [][32]byte, values [][32]byte) (*types.Transaction, error) {
	return _Store.contract.AsyncTransact(opts, handler, "setItems", keys, values)
}

// SetItems is a paid mutator transaction binding the contract method 0x728e99f9.
//
// Solidity: function setItems(bytes32[] keys, bytes32[] values) returns()
func (_Store *StoreSession) SetItems(keys [][32]byte, values [][32]byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItems(&_Store.TransactOpts, keys, values)
}

func (_Store *StoreSession) AsyncSetItems(handler func(*types.Receipt, error), keys [][32]byte, values [][32]byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItems(handler, &_Store.TransactOpts, keys, values)
}

// SetItems is a paid mutator transaction binding the contract method 0x728e99f9.
//
// Solidity: function setItems(bytes32[] keys, bytes32[] values) returns()
func (_Store *StoreTransactorSession) SetItems(keys [][32]byte, values [][32]byte) (*types.Transaction, *types.Receipt, error) {
	return _Store.Contract.SetItems(&_Store.TransactOpts, keys, values)
}

func (_Store *StoreTransactorSession) AsyncSetItems(handler func(*types.Receipt, error), keys [][32]byte, values [][32]byte) (*types.Transaction, error) {
	return _Store.Contract.AsyncSetItems(handler, &_Store.TransactOpts, keys, values)
}

// StoreItemSet represents a ItemSet event raised by the Store contract.
type StoreItemSet struct {
	Key   [32]byte
	Value [32]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// WatchItemSet is a free log subscription operation binding the contract event 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4.
//
// Solidity: event ItemSet(bytes32 key, bytes32 value)
func (_Store *StoreFilterer) WatchItemSet(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _Store.contract.WatchLogs(fromBlock, handler, "ItemSet")
}

func (_Store *StoreFilterer) WatchAllItemSet(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _Store.contract.WatchLogs(fromBlock, handler, "ItemSet")
}

// ParseItemSet is a log parse operation binding the contract event 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4.
//
// Solidity: event ItemSet(bytes32 key, bytes32 value)
func (_Store *StoreFilterer) ParseItemSet(log types.Log) (*StoreItemSet, error) {
	event := new(StoreItemSet)
	if err := _Store.contract.UnpackLog(event, "ItemSet", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WatchItemSet is a free log subscription operation binding the contract event 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4.
//
// Solidity: event ItemSet(bytes32 key, bytes32 value)
func (_Store *StoreSession) WatchItemSet(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _Store.Contract.WatchItemSet(fromBlock, handler)
}

func (_Store *StoreSession) WatchAllItemSet(fromBlock *int64, handler func(int, []types.Log)) (string, error) {
	return _Store.Contract.WatchAllItemSet(fromBlock, handler)
}

// ParseItemSet is a log parse operation binding the contract event 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4.
//
// Solidity: event ItemSet(bytes32 key, bytes32 value)
func (_Store *StoreSession) ParseItemSet(log types.Log) (*StoreItemSet, error) {
	return _Store.Contract.ParseItemSet(log)
}
