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

package options

const (
	DECODE_ERR_THORW_OUT = 0
	DECODE_ERR_IGNORE_ROW = 1
	DECODE_ERR_IGNORE_FIELD = 2
)

type FindOptions struct {
	QueryHook interface{}
	// How to deal with the data if decode error, 0-default throw out error, 1-ignore row, 2-ignore field
	DecodeErrIgnore uint32
}
