/*
 * Copyright (C) 2017 Sylvain Afchain
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package kv

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/spf13/viper"
)

type KVStore struct {
	db *bolt.DB
}

// GetString returns the string value if found or an error for the given bucket, key.
func (k *KVStore) GetString(bucket string, key string) (value string, found bool, err error) {
	err = k.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		if v := b.Get([]byte(key)); v != nil {
			value = string(v)
			found = true
		}
		return nil
	})

	return
}

// GetInt64 returns the int64 value if found or an error for the given bucket, key.
func (k *KVStore) GetInt64(bucket string, key string) (value int64, found bool, err error) {
	err = k.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		if v := b.Get([]byte(key)); v != nil {
			value = int64(binary.BigEndian.Uint64(v))
			found = true
		}
		return nil
	})

	return
}

// SetString stores the string value for the given bucket/key.
func (k *KVStore) SetString(bucket string, key string, value string) error {
	return k.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte(value))
	})
}

// SetInt64 stores the int64 value for the given bucket/key.
func (k *KVStore) SetInt64(bucket string, key string, value int64) error {
	return k.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(value))
		return b.Put([]byte(key), buf[:])
	})
}

// Inc increments the int64 value for the given bucket/key. It return the new value
// if found or an error.
func (k *KVStore) Inc(bucket string, key string) (value int64, found bool, err error) {
	err = k.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		v := b.Get([]byte(key))
		if v == nil {
			return err
		}
		value = int64(binary.BigEndian.Uint64(v))
		found = true

		value++
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(value))
		return b.Put([]byte(key), buf[:])
	})

	return
}

func NewKVStore(cfg *viper.Viper) *KVStore {
	dbname := fmt.Sprintf("%s.db", filepath.Base(os.Args[0]))
	path := filepath.Join(cfg.GetString("data"), url.QueryEscape(dbname))

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &KVStore{db: db}
}
