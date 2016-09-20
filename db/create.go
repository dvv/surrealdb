// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"github.com/abcum/surreal/kvs"
	"github.com/abcum/surreal/sql"
	"github.com/abcum/surreal/util/item"
	"github.com/abcum/surreal/util/keys"
	"github.com/abcum/surreal/util/uuid"
)

func executeCreateStatement(txn kvs.TX, ast *sql.CreateStatement) (out []interface{}, err error) {

	var local bool

	if txn == nil {
		local = true
		txn, err = db.Txn(true)
		if err != nil {
			return
		}
		defer txn.Rollback()
	}

	for _, w := range ast.What {

		if what, ok := w.(*sql.Thing); ok {
			key := &keys.Thing{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: what.TB, ID: what.ID}
			kv, _ := txn.Get(key.Encode())
			doc := item.New(kv, txn, key)
			if ret, err := create(doc, ast); err != nil {
				return nil, err
			} else if ret != nil {
				out = append(out, ret)
			}
		}

		if what, ok := w.(*sql.Table); ok {
			key := &keys.Thing{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: what.TB, ID: uuid.NewV5(uuid.NewV4().UUID, ast.KV).String()}
			kv, _ := txn.Get(key.Encode())
			doc := item.New(kv, txn, key)
			if ret, err := create(doc, ast); err != nil {
				return nil, err
			} else if ret != nil {
				out = append(out, ret)
			}
		}

	}

	if local {
		txn.Commit()
	}

	return

}

func create(doc *item.Doc, ast *sql.CreateStatement) (out interface{}, err error) {

	if err = doc.Merge(ast.Data); err != nil {
		return
	}

	if !doc.Allow("CREATE") {
		return nil, nil
	}

	if err = doc.StoreIndex(); err != nil {
		return
	}

	if err = doc.StartThing(); err != nil {
		return
	}

	if err = doc.StorePatch(); err != nil {
		return
	}

	out = doc.Yield(ast.Echo, sql.AFTER)

	return

}
