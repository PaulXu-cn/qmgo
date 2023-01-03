/*
 Copyright 2020 The Qmgo Authors.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package qmgo

import (
	"context"
	"log"
	"reflect"
	"strings"

	"github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/mongo"
)

// Cursor struct define
type Cursor struct {
	ctx       context.Context
	cursor    *mongo.Cursor
	err       error
	ignoreErr uint32
}

// Next gets the next document for this cursor. It returns true if there were no errors and the cursor has not been
// exhausted.
func (c *Cursor) Next(result interface{}) bool {
	if c.err != nil {
		return false
	}
	var err error
	if c.cursor.Next(c.ctx) {
		err = c.cursor.Decode(result)
		if err == nil {
			return true
		} else if options.DECODE_ERR_IGNORE_FIELD != c.ignoreErr {
			return false
		} else {
			resultv := reflect.ValueOf(result)
			if resultv.Kind() != reflect.Ptr ||
				!(resultv.Elem().Kind() == reflect.Struct || resultv.Elem().Kind() == reflect.Ptr) {
				log.Fatalf("result argument must be a slice、map or address, result-k %d, ele-k %d",
					resultv.Kind(), resultv.Elem().Kind())
			}

			// 取指针指向的结构体变量
			v := resultv.Elem()
			if resultv.Elem().Kind() == reflect.Ptr {
				v = resultv.Elem().Elem()
			}

			// 解析字段, NumField() 4 个字段。
			for i := 0; i < v.NumField(); i++ {
				// 获取结构体字段信息
				structField := v.Type().Field(i)
				// 取tag
				tag := structField.Tag
				// 解析label tag，获取tag值
				label := tag.Get("bson")
				if label == "" {
					continue
				}
				fields := strings.Split(label, ",")
				if rawVal := c.cursor.Current.Lookup(fields[0]); rawVal.Validate() == nil {
					if unmErr := rawVal.Unmarshal(v.Field(i).Addr().Interface()); unmErr != nil {
						// TODO Log something ?
						continue
					}
				}
			}
			return true
		}
	}
	return false
}

// All iterates the cursor and decodes each document into results. The results parameter must be a pointer to a slice.
// recommend to use All() in struct Query or Aggregate
func (c *Cursor) All(results interface{}) error {
	if c.err != nil {
		return c.err
	}
	if options.DECODE_ERR_THORW_OUT == c.ignoreErr {
		return c.cursor.All(c.ctx, results)
	} else {
		resultv := reflect.ValueOf(results)
		if resultv.Kind() != reflect.Ptr ||
			!(resultv.Elem().Kind() == reflect.Slice || resultv.Elem().Kind() == reflect.Ptr) {
			log.Fatalf("result argument must be a slice、map or address, result-k %d, ele-k %d",
				resultv.Kind(), resultv.Elem().Kind())
		}
		slicev := resultv.Elem()
		slicev = slicev.Slice(0, slicev.Cap())
		elemt := slicev.Type().Elem()
		i := 0
		for {
			if slicev.Len() == i {
				elemp := reflect.New(elemt)
				if !c.Next(elemp.Interface()) {
					break
				}
				slicev = reflect.Append(slicev, elemp.Elem())
				slicev = slicev.Slice(0, slicev.Cap())
			} else {
				if !c.Next(slicev.Index(i).Addr().Interface()) {
					break
				}
			}
			i++
		}
		resultv.Elem().Set(slicev.Slice(0, i))
		return c.Close()
	}
}

// ID returns the ID of this cursor, or 0 if the cursor has been closed or exhausted.
//func (c *Cursor) ID() int64 {
//	if c.err != nil {
//		return 0
//	}
//	return c.cursor.ID()
//}

// Close closes this cursor. Next and TryNext must not be called after Close has been called.
// When the cursor object is no longer in use, it should be actively closed
func (c *Cursor) Close() error {
	if c.err != nil {
		return c.err
	}
	return c.cursor.Close(c.ctx)
}

// Err return the last error of Cursor, if no error occurs, return nil
func (c *Cursor) Err() error {
	if c.err != nil {
		return c.err
	}
	return c.cursor.Err()
}
