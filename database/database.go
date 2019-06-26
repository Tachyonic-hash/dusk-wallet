package database

import (
	"bytes"
	"dusk-wallet/transactions/v3"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/bwesterb/go-ristretto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type DB struct {
	storage *leveldb.DB
}

var (
	inputPrefix = []byte("input")
)

func New(path string) (*DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("wallet cannot be used without database %s", err.Error())
	}
	return &DB{storage: db}, nil
}

func (db *DB) Put(key, value []byte) error {
	return db.storage.Put(key, value, nil)
}

func (db *DB) PutInput(encryptionKey []byte, pubkey ristretto.Point, amount, mask, privkey ristretto.Scalar) error {

	// XXX: Encrypt data using priv view key

	fmt.Println("Received Input amount:", amount.BigInt().Int64())

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, amount.Bytes())
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, mask.Bytes())
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, privkey.Bytes())
	if err != nil {
		return err
	}

	encryptedBytes, err := encrypt(buf.Bytes(), encryptionKey)
	if err != nil {
		return err
	}

	key := append(inputPrefix, pubkey.Bytes()...)

	return db.Put(key, encryptedBytes)
}

func (db *DB) RemoveInput(pubkey []byte) error {
	key := append(inputPrefix, pubkey...)
	return db.Delete(key)
}

func (db DB) FetchInputs(decryptionKey []byte, amount int64) ([]*transactions.Input, int64, error) {

	var inputs []*inputDB

	var totalAmount = amount

	iter := db.storage.NewIterator(util.BytesPrefix(inputPrefix), nil)
	for iter.Next() {
		val := iter.Value()

		encryptedBytes := make([]byte, len(val))
		copy(encryptedBytes[:], val)

		decryptedBytes, err := decrypt(encryptedBytes, decryptionKey)
		if err != nil {
			return nil, 0, err
		}
		idb := &inputDB{}

		buf := bytes.NewBuffer(decryptedBytes)
		err = idb.Decode(buf)
		if err != nil {
			return nil, 0, err
		}

		inputs = append(inputs, idb)

		// Check if we need more inputs
		totalAmount = totalAmount - idb.amount.BigInt().Int64()
		if totalAmount <= 0 {
			break
		}
	}

	if totalAmount > 0 {
		return nil, 0, errors.New("accumulated value of all of your inputs do not account for the total amount inputted")
	}

	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, 0, err
	}

	var changeAmount int64
	if totalAmount < 0 {
		changeAmount = -totalAmount
	}

	// convert inputDb to transaction input
	var tInputs []*transactions.Input
	for _, input := range inputs {
		tInputs = append(tInputs, transactions.NewInput(input.amount, input.mask, input.privKey))
	}

	return tInputs, changeAmount, nil
}

func (db DB) Get(key []byte) ([]byte, error) {
	return db.storage.Get(key, nil)
}

func (db *DB) Delete(key []byte) error {
	return db.storage.Delete(key, nil)
}

func (db *DB) Close() error {
	return db.storage.Close()
}
