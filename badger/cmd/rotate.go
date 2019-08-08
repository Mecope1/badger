/*
 * Copyright 2019 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"io/ioutil"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/y"

	"github.com/spf13/cobra"
)

var oldKeyPath string
var newKeyPath string
var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate encryption key.",
	Long:  "Rotate will rotate the old key with new encryption key.",
	RunE:  doRotate,
}

func init() {
	RootCmd.AddCommand(rotateCmd)
	rotateCmd.Flags().StringVarP(&oldKeyPath, "old-key-path", "o",
		"", "Path of the old key. It'll be considered as plain text, If no path"+
			"provided.")
	rotateCmd.Flags().StringVarP(&newKeyPath, "new-key-path", "n",
		"", "Path of the new key. It'll be considered as plain text, If no path"+
			"provided.")
}

func doRotate(cmd *cobra.Command, args []string) error {
	oldKey, err := getKey(oldKeyPath)
	if err != nil {
		return y.Wrapf(err, "Error while opening old key.")
	}
	opt := badger.Options{
		Dir:           sstDir,
		ReadOnly:      false,
		EncryptionKey: oldKey,
	}
	kr, err := badger.OpenKeyRegistry(opt)
	if err != nil {
		return y.Wrapf(err, "Unable to open key registry.")
	}
	kr.Close()
	newKey, err := getKey(newKeyPath)
	if err != nil {
		return err
	}
	opt.EncryptionKey = newKey
	err = badger.WriteKeyRegistry(kr, opt)
	if err != nil {
		return y.Wrapf(err, "Error while writing key registry.")
	}
	return nil
}

func getKey(path string) ([]byte, error) {
	if path == "" {
		// Empty bytes for plain text to encryption(vice versa).
		return []byte{}, nil
	}
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return ioutil.ReadAll(fp)
}
